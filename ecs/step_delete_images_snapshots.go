package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepDeleteApsaraStackImageSnapshots struct {
	ApsaraStackImageForceDelete          bool
	ApsaraStackImageForceDeleteSnapshots bool
	ApsaraStackImageName                 string
	ApsaraStackImageDestinationRegions   []string
	ApsaraStackImageDestinationNames     []string
}

func (s *stepDeleteApsaraStackImageSnapshots) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)

	// Check for force delete
	if s.ApsaraStackImageForceDelete {
		err := s.deleteImageAndSnapshots(state, s.ApsaraStackImageName, config.ApsaraStackRegion)
		if err != nil {
			return halt(state, err, "")
		}

		numberOfName := len(s.ApsaraStackImageDestinationNames)
		if numberOfName == 0 {
			return multistep.ActionContinue
		}

		for index, destinationRegion := range s.ApsaraStackImageDestinationRegions {
			if destinationRegion == config.ApsaraStackRegion {
				continue
			}

			if index < numberOfName {
				err = s.deleteImageAndSnapshots(state, s.ApsaraStackImageDestinationNames[index], destinationRegion)
				if err != nil {
					return halt(state, err, "")
				}
			} else {
				break
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepDeleteApsaraStackImageSnapshots) deleteImageAndSnapshots(state multistep.StateBag, imageName string, region string) error {

	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeImagesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	describeImagesRequest.RegionId = region
	describeImagesRequest.ImageName = imageName
	describeImagesRequest.Status = ImageStatusQueried
	imageResponse, _ := client.DescribeImages(describeImagesRequest)
	images := imageResponse.Images.Image
	if len(images) < 1 {
		return nil
	}

	ui.Say(fmt.Sprintf("Deleting duplicated image and snapshot in %s: %s", region, imageName))

	for _, image := range images {
		if image.ImageOwnerAlias != ImageOwnerSelf {
			log.Printf("You can not delete non-customized images: %s ", image.ImageId)
			continue
		}

		deleteImageRequest := ecs.CreateDeleteImageRequest()
		deleteImageRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		deleteImageRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		deleteImageRequest.RegionId = region
		deleteImageRequest.ImageId = image.ImageId
		if _, err := client.DeleteImage(deleteImageRequest); err != nil {
			err := fmt.Errorf("Failed to delete image: %s", err)
			return err
		}

		if s.ApsaraStackImageForceDeleteSnapshots {
			for _, diskDevice := range image.DiskDeviceMappings.DiskDeviceMapping {
				deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
				deleteSnapshotRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
				deleteSnapshotRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

				deleteSnapshotRequest.SnapshotId = diskDevice.SnapshotId
				if _, err := client.DeleteSnapshot(deleteSnapshotRequest); err != nil {
					err := fmt.Errorf("Deleting ECS snapshot failed: %s", err)
					return err
				}
			}
		}
	}

	return nil
}

func (s *stepDeleteApsaraStackImageSnapshots) Cleanup(state multistep.StateBag) {
}
