package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepShareApsaraStackImage struct {
	ApsaraStackImageShareAccounts   []string
	ApsaraStackImageUNShareAccounts []string
	RegionId                        string
}

func (s *stepShareApsaraStackImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ApsaraStackImages := state.Get("ApsaraStackimages").(map[string]string)

	for regionId, imageId := range ApsaraStackImages {
		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
		if strings.ToLower(config.Protocol) == "https" {
			modifyImageShareRequest.Scheme = "https"
		} else {
			modifyImageShareRequest.Scheme = "http"
		}
		modifyImageShareRequest.Headers = map[string]string{"RegionId": s.RegionId}
		modifyImageShareRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		modifyImageShareRequest.RegionId = regionId
		modifyImageShareRequest.ImageId = imageId
		modifyImageShareRequest.AddAccount = &s.ApsaraStackImageShareAccounts
		modifyImageShareRequest.RemoveAccount = &s.ApsaraStackImageUNShareAccounts

		if _, err := client.ModifyImageSharePermission(modifyImageShareRequest); err != nil {
			return halt(state, err, "Failed modifying image share permissions")
		}
	}
	return multistep.ActionContinue
}

func (s *stepShareApsaraStackImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ApsaraStackImages := state.Get("ApsaraStackimages").(map[string]string)

	ui.Say("Restoring image share permission because cancellations or error...")

	for regionId, imageId := range ApsaraStackImages {
		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
		if strings.ToLower(config.Protocol) == "https" {
			modifyImageShareRequest.Scheme = "https"
		} else {
			modifyImageShareRequest.Scheme = "http"
		}
		modifyImageShareRequest.Headers = map[string]string{"RegionId": s.RegionId}
		modifyImageShareRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		modifyImageShareRequest.RegionId = regionId
		modifyImageShareRequest.ImageId = imageId
		modifyImageShareRequest.AddAccount = &s.ApsaraStackImageUNShareAccounts
		modifyImageShareRequest.RemoveAccount = &s.ApsaraStackImageShareAccounts
		if _, err := client.ModifyImageSharePermission(modifyImageShareRequest); err != nil {
			ui.Say(fmt.Sprintf("Restoring image share permission failed: %s", err))
		}
	}
}
