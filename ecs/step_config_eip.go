package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/hashicorp/packer/common/uuid"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigApsaraStackEIP struct {
	AssociatePublicIpAddress bool
	RegionId                 string
	InternetChargeType       string
	InternetMaxBandwidthOut  int
	allocatedId              string
	SSHPrivateIp             bool
}

var allocateEipAddressRetryErrors = []string{
	"LastTokenProcessing",
}

func (s *stepConfigApsaraStackEIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.Instance)

	if s.SSHPrivateIp {
		ipaddress := instance.VpcAttributes.PrivateIpAddress.IpAddress
		if len(ipaddress) == 0 {
			ui.Say("Failed to get private ip of instance")
			return multistep.ActionHalt
		}
		state.Put("ipaddress", ipaddress[0])
		return multistep.ActionContinue
	}

	ui.Say("Allocating eip...")

	allocateEipAddressRequest := s.buildAllocateEipAddressRequest(state)
	allocateEipAddressResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.AllocateEipAddress(allocateEipAddressRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(allocateEipAddressRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error allocating eip")
	}

	ipaddress := allocateEipAddressResponse.(*ecs.AllocateEipAddressResponse).EipAddress
	ui.Message(fmt.Sprintf("Allocated eip: %s", ipaddress))
	/*	ipaddress := allocateEipAddressResponse.(*ecs.AllocateEipAddressResponse).EipAddress
		ui.Message(fmt.Sprintf("Allocated eip: %s", ipaddress))*/

	allocateId := allocateEipAddressResponse.(*ecs.AllocateEipAddressResponse).AllocationId
	s.allocatedId = allocateId

	err = s.waitForEipStatus(client, state, instance.RegionId, s.allocatedId, EipStatusAvailable)
	if err != nil {
		return halt(state, err, "Error wait eip available timeout")
	}

	associateEipAddressRequest := ecs.CreateAssociateEipAddressRequest()
	if strings.ToLower(config.Protocol) == "https" {
		associateEipAddressRequest.Scheme = "https"
	} else {
		associateEipAddressRequest.Scheme = "http"
	}
	associateEipAddressRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	associateEipAddressRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	associateEipAddressRequest.AllocationId = allocateId
	associateEipAddressRequest.InstanceId = instance.InstanceId
	if _, err := client.AssociateEipAddress(associateEipAddressRequest); err != nil {
		e, ok := err.(errors.Error)
		if !ok || e.ErrorCode() != "TaskConflict" {
			return halt(state, err, "Error associating eip")
		}

		ui.Error(fmt.Sprintf("Error associate eip: %s", err))
	}

	err = s.waitForEipStatus(client, state, instance.RegionId, s.allocatedId, EipStatusInUse)
	if err != nil {
		return halt(state, err, "Error wait eip associated timeout")
	}

	state.Put("ipaddress", ipaddress)
	return multistep.ActionContinue
}

func (s *stepConfigApsaraStackEIP) Cleanup(state multistep.StateBag) {
	if len(s.allocatedId) == 0 {
		return
	}

	cleanUpMessage(state, "EIP")

	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	instance := state.Get("instance").(*ecs.Instance)
	ui := state.Get("ui").(packer.Ui)

	unassociateEipAddressRequest := ecs.CreateUnassociateEipAddressRequest()
	if strings.ToLower(config.Protocol) == "https" {
		unassociateEipAddressRequest.Scheme = "https"
	} else {
		unassociateEipAddressRequest.Scheme = "http"
	}
	unassociateEipAddressRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	unassociateEipAddressRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	unassociateEipAddressRequest.AllocationId = s.allocatedId
	unassociateEipAddressRequest.InstanceId = instance.InstanceId
	if _, err := client.UnassociateEipAddress(unassociateEipAddressRequest); err != nil {
		ui.Say(fmt.Sprintf("Failed to unassociate eip: %s", err))
	}

	if err := s.waitForEipStatus(client, state, instance.RegionId, s.allocatedId, EipStatusAvailable); err != nil {
		ui.Say(fmt.Sprintf("Timeout while unassociating eip: %s", err))
	}

	releaseEipAddressRequest := ecs.CreateReleaseEipAddressRequest()
	if strings.ToLower(config.Protocol) == "https" {
		releaseEipAddressRequest.Scheme = "https"
	} else {
		releaseEipAddressRequest.Scheme = "http"
	}
	releaseEipAddressRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	releaseEipAddressRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}
	releaseEipAddressRequest.AllocationId = s.allocatedId
	if _, err := client.ReleaseEipAddress(releaseEipAddressRequest); err != nil {
		ui.Say(fmt.Sprintf("Failed to release eip: %s", err))
	}
}

func (s *stepConfigApsaraStackEIP) waitForEipStatus(client *ClientWrapper, state multistep.StateBag, regionId string, allocationId string, expectedStatus string) error {
	describeEipAddressesRequest := ecs.CreateDescribeEipAddressesRequest()
	config := state.Get("config").(*Config)
	if strings.ToLower(config.Protocol) == "https" {
		describeEipAddressesRequest.Scheme = "https"
	} else {
		describeEipAddressesRequest.Scheme = "http"
	}
	describeEipAddressesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeEipAddressesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	describeEipAddressesRequest.RegionId = regionId
	describeEipAddressesRequest.AllocationId = s.allocatedId

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			response, err := client.DescribeEipAddresses(describeEipAddressesRequest)
			if err == nil && len(response.EipAddresses.EipAddress) == 0 {
				err = fmt.Errorf("eip allocated is not find")
			}

			return response, err
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			eipAddressesResponse := response.(*ecs.DescribeEipAddressesResponse)
			eipAddresses := eipAddressesResponse.EipAddresses.EipAddress

			for _, eipAddress := range eipAddresses {
				if eipAddress.Status == expectedStatus {
					return WaitForExpectSuccess
				}
			}

			return WaitForExpectToRetry
		},
		RetryTimes: shortRetryTimes,
	})

	return err
}

func (s *stepConfigApsaraStackEIP) buildAllocateEipAddressRequest(state multistep.StateBag) *ecs.AllocateEipAddressRequest {
	instance := state.Get("instance").(*ecs.Instance)
	config := state.Get("config").(*Config)
	request := ecs.CreateAllocateEipAddressRequest()
	if strings.ToLower(config.Protocol) == "https" {
		request.Scheme = "https"
	} else {
		request.Scheme = "http"
	}
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = instance.RegionId
	request.InternetChargeType = s.InternetChargeType
	request.Bandwidth = string(convertNumber(s.InternetMaxBandwidthOut))

	return request
}
