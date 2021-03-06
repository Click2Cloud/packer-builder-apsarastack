package ecs

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"strings"
)

type stepPreValidate struct {
	ApsaraStackDestImageName string
	ForceDelete              bool
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if err := s.validateRegions(state); err != nil {
		return halt(state, err, "")
	}

	if err := s.validateDestImageName(state); err != nil {
		return halt(state, err, "")
	}

	if err := s.validateinsecure(state); err != nil {
		return halt(state, err, "")
	}

	return multistep.ActionContinue
}

func (s *stepPreValidate) validateRegions(state multistep.StateBag) error {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	if config.ApsaraStackSkipValidation {
		ui.Say("Skip region validation flag found, skipping prevalidating source region and copied regions.")
		return nil
	}

	ui.Say("Prevalidating source region and copied regions...")

	var errs *packer.MultiError
	if err := config.ValidateRegion(config.ApsaraStackRegion); err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}
	for _, region := range config.ApsaraStackImageDestinationRegions {
		if err := config.ValidateRegion(region); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *stepPreValidate) validateDestImageName(state multistep.StateBag) error {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)

	if s.ForceDelete {
		ui.Say("Force delete flag found, skipping prevalidating image name.")
		return nil
	}

	ui.Say("Prevalidating image name...")

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	if strings.ToLower(config.Protocol) == "https" {
		describeImagesRequest.Scheme = "https"
	} else {
		describeImagesRequest.Scheme = "http"
	}
	describeImagesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeImagesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	describeImagesRequest.RegionId = config.ApsaraStackRegion
	describeImagesRequest.ImageName = s.ApsaraStackDestImageName
	describeImagesRequest.Status = ImageStatusQueried

	imagesResponse, err := client.DescribeImages(describeImagesRequest)

	if err != nil {
		return fmt.Errorf("Error querying ApsaraStack image: %s", err)
	}

	images := imagesResponse.Images.Image
	if len(images) > 0 {
		return fmt.Errorf("Error: Image Name: '%s' is used by an existing ApsaraStack image: %s", images[0].ImageName, images[0].ImageId)
	}

	return nil
}
func (s *stepPreValidate) validateinsecure(state multistep.StateBag) error {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	if config.ApsaraStackSkipValidation {
		ui.Say("Skip validation of insecure value.")
		return nil
	}

	ui.Say("Prevalidating insecure value...")

	var errs *packer.MultiError
	/*
		if err := config.ValidateInsecure(config.Insecure); err != nil{
			errs = packer.MultiErrorAppend(errs)
		}*/

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil

}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {}
