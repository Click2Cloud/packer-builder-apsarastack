package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigApsaraStackPublicIP struct {
	publicIPAddress string
	RegionId        string
	SSHPrivateIp    bool
}

func (s *stepConfigApsaraStackPublicIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.Instance)

	if s.SSHPrivateIp {
		ipaddress := instance.InnerIpAddress.IpAddress
		if len(ipaddress) == 0 {
			ui.Say("Failed to get private ip of instance")
			return multistep.ActionHalt
		}
		state.Put("ipaddress", ipaddress[0])
		return multistep.ActionContinue
	}

	allocatePublicIpAddressRequest := ecs.CreateAllocatePublicIpAddressRequest()
	if strings.ToLower(config.Protocol) == "https" {
		allocatePublicIpAddressRequest.Scheme = "https"
	} else {
		allocatePublicIpAddressRequest.Scheme = "http"
	}

	allocatePublicIpAddressRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	allocatePublicIpAddressRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	allocatePublicIpAddressRequest.InstanceId = instance.InstanceId
	ipaddress, err := client.AllocatePublicIpAddress(allocatePublicIpAddressRequest)
	if err != nil {
		return halt(state, err, "Error allocating public ip")
	}

	s.publicIPAddress = ipaddress.IpAddress
	ui.Say(fmt.Sprintf("Allocated public ip address %s.", ipaddress.IpAddress))
	state.Put("ipaddress", ipaddress.IpAddress)
	return multistep.ActionContinue
}

func (s *stepConfigApsaraStackPublicIP) Cleanup(state multistep.StateBag) {

}
