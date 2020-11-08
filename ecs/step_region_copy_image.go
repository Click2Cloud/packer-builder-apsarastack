package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	confighelper "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepRegionCopyApsaraStackImage struct {
	ApsaraStackImageDestinationRegions []string
	ApsaraStackImageDestinationNames   []string
	RegionId                           string
}

func (s *stepRegionCopyApsaraStackImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)

	if config.ImageEncrypted != confighelper.TriUnset {
		s.ApsaraStackImageDestinationRegions = append(s.ApsaraStackImageDestinationRegions, s.RegionId)
		s.ApsaraStackImageDestinationNames = append(s.ApsaraStackImageDestinationNames, config.ApsaraStackImageName)
	}

	if len(s.ApsaraStackImageDestinationRegions) == 0 {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)

	srcImageId := state.Get("ApsaraStackimage").(string)
	ApsaraStackImages := state.Get("ApsaraStackimages").(map[string]string)
	numberOfName := len(s.ApsaraStackImageDestinationNames)

	ui.Say(fmt.Sprintf("Coping image %s from %s...", srcImageId, s.RegionId))
	for index, destinationRegion := range s.ApsaraStackImageDestinationRegions {
		if destinationRegion == s.RegionId && config.ImageEncrypted == confighelper.TriUnset {
			continue
		}

		ecsImageName := ""
		if numberOfName > 0 && index < numberOfName {
			ecsImageName = s.ApsaraStackImageDestinationNames[index]
		}

		copyImageRequest := ecs.CreateCopyImageRequest()
		copyImageRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		copyImageRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		copyImageRequest.RegionId = s.RegionId
		copyImageRequest.ImageId = srcImageId
		copyImageRequest.DestinationRegionId = destinationRegion
		copyImageRequest.DestinationImageName = ecsImageName
		if config.ImageEncrypted != confighelper.TriUnset {
			copyImageRequest.Encrypted = requests.NewBoolean(config.ImageEncrypted.True())
		}

		imageResponse, err := client.CopyImage(copyImageRequest)
		if err != nil {
			return halt(state, err, "Error copying images")
		}

		ApsaraStackImages[destinationRegion] = imageResponse.ImageId
		ui.Message(fmt.Sprintf("Copy image from %s(%s) to %s(%s)", s.RegionId, srcImageId, destinationRegion, imageResponse.ImageId))
	}

	if config.ImageEncrypted != confighelper.TriUnset {
		if _, err := client.WaitForImageStatus(s.RegionId, ApsaraStackImages[s.RegionId], ImageStatusAvailable, time.Duration(APSARASTACK_DEFAULT_LONG_TIMEOUT)*time.Second, state); err != nil {
			return halt(state, err, fmt.Sprintf("Timeout waiting image %s finish copying", ApsaraStackImages[s.RegionId]))
		}
	}

	return multistep.ActionContinue
}

func (s *stepRegionCopyApsaraStackImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Stopping copy image because cancellation or error..."))

	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ApsaraStackImages := state.Get("ApsaraStackimages").(map[string]string)
	srcImageId := state.Get("ApsaraStackimage").(string)

	for copiedRegionId, copiedImageId := range ApsaraStackImages {
		if copiedImageId == srcImageId {
			continue
		}

		cancelCopyImageRequest := ecs.CreateCancelCopyImageRequest()
		cancelCopyImageRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		cancelCopyImageRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		cancelCopyImageRequest.RegionId = copiedRegionId
		cancelCopyImageRequest.ImageId = copiedImageId
		if _, err := client.CancelCopyImage(cancelCopyImageRequest); err != nil {

			ui.Error(fmt.Sprintf("Error cancelling copy image: %v", err))
		}
	}
}
