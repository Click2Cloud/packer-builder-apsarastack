package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/common/random"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateApsaraStackImage struct {
	ApsaraStackImageIgnoreDataDisks bool
	WaitSnapshotReadyTimeout        int
	image                           *ecs.Image
}

var createImageRetryErrors = []string{
	"IdempotentProcessing",
}

func (s *stepCreateApsaraStackImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)

	tempImageName := config.ApsaraStackImageName
	if config.ImageEncrypted.True() {
		tempImageName = fmt.Sprintf("packer_%s", random.AlphaNum(7))
		ui.Say(fmt.Sprintf("Creating temporary image for encryption: %s", tempImageName))
	} else {
		ui.Say(fmt.Sprintf("Creating image: %s", tempImageName))
	}

	createImageRequest := s.buildCreateImageRequest(state, tempImageName)
	createImageResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.CreateImage(createImageRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createImageRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating image")
	}

	imageId := createImageResponse.(*ecs.CreateImageResponse).ImageId

	imagesResponse, err := client.WaitForImageStatus(config.ApsaraStackRegion, imageId, ImageStatusAvailable, time.Duration(s.WaitSnapshotReadyTimeout)*time.Second, state)

	// save image first for cleaning up if timeout
	images := imagesResponse.(*ecs.DescribeImagesResponse).Images.Image
	if len(images) == 0 {
		return halt(state, err, "Unable to find created image")
	}
	s.image = &images[0]

	if err != nil {
		return halt(state, err, "Timeout waiting for image to be created")
	}

	var snapshotIds []string
	for _, device := range images[0].DiskDeviceMappings.DiskDeviceMapping {
		snapshotIds = append(snapshotIds, device.SnapshotId)
	}

	state.Put("ApsaraStackimage", imageId)
	state.Put("ApsaraStacksnapshots", snapshotIds)

	ApsaraStackImages := make(map[string]string)
	ApsaraStackImages[config.ApsaraStackRegion] = images[0].ImageId
	state.Put("ApsaraStackimages", ApsaraStackImages)

	return multistep.ActionContinue
}

func (s *stepCreateApsaraStackImage) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	config := state.Get("config").(*Config)
	encryptedSet := config.ImageEncrypted.True()

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted && !encryptedSet {
		return
	}

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)

	if !cancelled && !halted && encryptedSet {
		ui.Say(fmt.Sprintf("Deleting temporary image %s(%s) and related snapshots after finishing encryption...", s.image.ImageId, s.image.ImageName))
	} else {
		ui.Say("Deleting the image and related snapshots because of cancellation or error...")
	}

	deleteImageRequest := ecs.CreateDeleteImageRequest()
	deleteImageRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	deleteImageRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	deleteImageRequest.RegionId = config.ApsaraStackRegion
	deleteImageRequest.ImageId = s.image.ImageId
	if _, err := client.DeleteImage(deleteImageRequest); err != nil {
		ui.Error(fmt.Sprintf("Error deleting image, it may still be around: %s", err))
		return
	}

	//Delete the snapshot of this image
	for _, diskDevices := range s.image.DiskDeviceMappings.DiskDeviceMapping {
		deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
		deleteSnapshotRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		deleteSnapshotRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		deleteSnapshotRequest.SnapshotId = diskDevices.SnapshotId
		if _, err := client.DeleteSnapshot(deleteSnapshotRequest); err != nil {
			ui.Error(fmt.Sprintf("Error deleting snapshot, it may still be around: %s", err))
			return
		}
	}
}

func (s *stepCreateApsaraStackImage) buildCreateImageRequest(state multistep.StateBag, imageName string) *ecs.CreateImageRequest {
	config := state.Get("config").(*Config)

	request := ecs.CreateCreateImageRequest()
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = config.ApsaraStackRegion
	request.ImageName = imageName
	request.ImageVersion = config.ApsaraStackImageVersion
	request.Description = config.ApsaraStackImageDescription
	if s.ApsaraStackImageIgnoreDataDisks {
		snapshotId := state.Get("ApsaraStacksnapshot").(string)
		request.SnapshotId = snapshotId
	} else {
		instance := state.Get("instance").(*ecs.Instance)
		request.InstanceId = instance.InstanceId
	}

	return request
}
