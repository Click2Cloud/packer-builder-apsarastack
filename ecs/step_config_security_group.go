package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigApsaraStackSecurityGroup struct {
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	VpcId             string
	RegionId          string
	isCreate          bool
}

var createSecurityGroupRetryErrors = []string{
	"IdempotentProcessing",
}

var deleteSecurityGroupRetryErrors = []string{
	"DependencyViolation",
}

func (s *stepConfigApsaraStackSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	networkType := state.Get("networktype").(InstanceNetWork)

	if len(s.SecurityGroupId) != 0 {
		describeSecurityGroupsRequest := ecs.CreateDescribeSecurityGroupsRequest()
		describeSecurityGroupsRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		describeSecurityGroupsRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		describeSecurityGroupsRequest.RegionId = s.RegionId
		describeSecurityGroupsRequest.SecurityGroupId = s.SecurityGroupId
		if networkType == InstanceNetworkVpc {
			vpcId := state.Get("vpcid").(string)
			describeSecurityGroupsRequest.VpcId = vpcId
		}

		securityGroupsResponse, err := client.DescribeSecurityGroups(describeSecurityGroupsRequest)
		if err != nil {
			return halt(state, err, "Failed querying security group")
		}

		securityGroupItems := securityGroupsResponse.SecurityGroups.SecurityGroup
		for _, securityGroupItem := range securityGroupItems {
			if securityGroupItem.SecurityGroupId == s.SecurityGroupId {
				state.Put("securitygroupid", s.SecurityGroupId)
				s.isCreate = false
				return multistep.ActionContinue
			}
		}

		s.isCreate = false
		err = fmt.Errorf("The specified security group {%s} doesn't exist.", s.SecurityGroupId)
		return halt(state, err, "")
	}

	ui.Say("Creating security group...")

	createSecurityGroupRequest := s.buildCreateSecurityGroupRequest(state)
	securityGroupResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.CreateSecurityGroup(createSecurityGroupRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createSecurityGroupRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Failed creating security group")
	}

	securityGroupId := securityGroupResponse.(*ecs.CreateSecurityGroupResponse).SecurityGroupId

	ui.Message(fmt.Sprintf("Created security group: %s", securityGroupId))
	state.Put("securitygroupid", securityGroupId)
	s.isCreate = true
	s.SecurityGroupId = securityGroupId

	authorizeSecurityGroupEgressRequest := ecs.CreateAuthorizeSecurityGroupEgressRequest()
	authorizeSecurityGroupEgressRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	authorizeSecurityGroupEgressRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	authorizeSecurityGroupEgressRequest.SecurityGroupId = securityGroupId
	authorizeSecurityGroupEgressRequest.RegionId = s.RegionId
	authorizeSecurityGroupEgressRequest.IpProtocol = IpProtocolAll
	authorizeSecurityGroupEgressRequest.PortRange = DefaultPortRange
	authorizeSecurityGroupEgressRequest.NicType = NicTypeInternet
	authorizeSecurityGroupEgressRequest.DestCidrIp = DefaultCidrIp

	if _, err := client.AuthorizeSecurityGroupEgress(authorizeSecurityGroupEgressRequest); err != nil {
		return halt(state, err, "Failed authorizing security group")
	}

	authorizeSecurityGroupRequest := ecs.CreateAuthorizeSecurityGroupRequest()
	authorizeSecurityGroupRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	authorizeSecurityGroupRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	authorizeSecurityGroupRequest.SecurityGroupId = securityGroupId
	authorizeSecurityGroupRequest.RegionId = s.RegionId
	authorizeSecurityGroupRequest.IpProtocol = IpProtocolAll
	authorizeSecurityGroupRequest.PortRange = DefaultPortRange
	authorizeSecurityGroupRequest.NicType = NicTypeInternet
	authorizeSecurityGroupRequest.SourceCidrIp = DefaultCidrIp

	if _, err := client.AuthorizeSecurityGroup(authorizeSecurityGroupRequest); err != nil {
		return halt(state, err, "Failed authorizing security group")
	}

	return multistep.ActionContinue
}

func (s *stepConfigApsaraStackSecurityGroup) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	cleanUpMessage(state, "security group")

	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDeleteSecurityGroupRequest()
			request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
			request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

			request.RegionId = s.RegionId
			request.SecurityGroupId = s.SecurityGroupId
			return client.DeleteSecurityGroup(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteSecurityGroupRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete security group, it may still be around: %s", err))
	}
}

func (s *stepConfigApsaraStackSecurityGroup) buildCreateSecurityGroupRequest(state multistep.StateBag) *ecs.CreateSecurityGroupRequest {
	networkType := state.Get("networktype").(InstanceNetWork)

	config := state.Get("config").(*Config)
	request := ecs.CreateCreateSecurityGroupRequest()
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = s.RegionId
	request.SecurityGroupName = s.SecurityGroupName

	if networkType == InstanceNetworkVpc {
		vpcId := state.Get("vpcid").(string)
		request.VpcId = vpcId
	}

	return request
}
