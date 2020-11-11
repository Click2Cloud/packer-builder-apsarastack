package ecs

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCheckApsaraStackSourceImage struct {
	SourceECSImageId string
}

func (s *stepCheckApsaraStackSourceImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeImagesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs"}
	//describeImagesRequest.Headers = map[string]string{"RegionId": }
	//describeImagesRequest.QueryParams = map[string]string{"AccessKeySecret": client.SecretKey, "Product": "ecs"}

	describeImagesRequest.RegionId = config.ApsaraStackRegion
	describeImagesRequest.ImageId = config.ApsaraStackSourceImage
	if config.ApsaraStackSkipImageValidation {
		describeImagesRequest.ShowExpired = "true"
	}
	client.Domain = config.Endpoint
	client.SetHTTPSInsecure(true)

	imagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "Error querying ApsaraStack image")
	}

	images := imagesResponse.Images.Image

	// Describe marketplace image
	describeImagesRequest.ImageOwnerAlias = "system"
	marketImagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "Error querying ApsaraStack system image")
	}

	marketImages := marketImagesResponse.Images.Image
	if len(marketImages) > 0 {
		images = append(images, marketImages...)
	}

	if len(images) == 0 {
		err := fmt.Errorf("No ApsaraStack image was found matching filters: %v", config.ApsaraStackSourceImage)
		return halt(state, err, "")
	}

	ui.Message(fmt.Sprintf("Found image ID: %s", images[0].ImageId))

	state.Put("source_image", &images[0])
	return multistep.ActionContinue
}

func (s *stepCheckApsaraStackSourceImage) Cleanup(multistep.StateBag) {

}
