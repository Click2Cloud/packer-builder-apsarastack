package ecs

import (
	"context"
	"fmt"
	//"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigApsaraStackVSwitch struct {
	VSwitchId   string
	ZoneId      string
	isCreate    bool
	CidrBlock   string
	VSwitchName string
}

var createVSwitchRetryErrors = []string{
	"TOKEN_PROCESSING",
}

var deleteVSwitchRetryErrors = []string{
	"IncorrectVSwitchStatus",
	"DependencyViolation",
	"DependencyViolation.HaVip",
	"IncorrectRouteEntryStatus",
	"TaskConflict",
}

func (s *stepConfigApsaraStackVSwitch) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	vpcId := state.Get("vpcid").(string)
	config := state.Get("config").(*Config)
	vpcclient, err := vpc.NewClientWithAccessKey(config.ApsaraStackRegion, config.ApsaraStackAccessKey, config.ApsaraStackSecretKey)
	if err != nil {
		panic(err) // TODO eRROR STATEMENT
	}
	vpcclient.Domain = client.Domain
	if client.GetHttpProxy() != "" {
		vpcclient.SetHttpProxy(client.GetHttpProxy())
	}
	if len(s.VSwitchId) != 0 {
		describeVSwitchesRequest := vpc.CreateDescribeVSwitchesRequest()
		describeVSwitchesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		describeVSwitchesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		describeVSwitchesRequest.VpcId = vpcId
		describeVSwitchesRequest.VSwitchId = s.VSwitchId
		describeVSwitchesRequest.ZoneId = s.ZoneId

		vswitchesResponse, err := vpcclient.DescribeVSwitches(describeVSwitchesRequest)
		if err != nil {
			return halt(state, err, "Failed querying vswitch")
		}

		vswitch := vswitchesResponse.VSwitches.VSwitch
		if len(vswitch) > 0 {
			state.Put("vswitchid", vswitch[0].VSwitchId)
			s.isCreate = false
			return multistep.ActionContinue
		}

		s.isCreate = false
		return halt(state, fmt.Errorf("The specified vswitch {%s} doesn't exist.", s.VSwitchId), "")
	}

	if s.ZoneId == "" {
		describeZonesRequest := vpc.CreateDescribeZonesRequest()
		describeZonesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		describeZonesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		describeZonesRequest.RegionId = config.ApsaraStackRegion

		zonesResponse, err := vpcclient.DescribeZones(describeZonesRequest)
		if err != nil {
			return halt(state, err, "Query for available zones failed")
		}

		var instanceTypes []string
		zones := zonesResponse.Zones.Zone
		for _, zone := range zones {

			s.ZoneId = zone.ZoneId
			//break
			//isVSwitchSupported := false
			//for _, resourceType := range zone.AvailableResourceCreation.ResourceTypes {
			//	if resourceType == "VSwitch" {
			//		isVSwitchSupported = true
			//	}
			//}
			//
			//if isVSwitchSupported {
			//	for _, instanceType := range zone.AvailableInstanceTypes.InstanceTypes {
			//		if instanceType == config.InstanceType {
			//			s.ZoneId = zone.ZoneId
			//			break
			//		}
			//		instanceTypes = append(instanceTypes, instanceType)
			//	}
			//}
		}

		if s.ZoneId == "" {
			if len(instanceTypes) > 0 {
				ui.Say(fmt.Sprintf("The instance type %s isn't available in this region."+
					"\n You can either change the instance to one of following: %v \n"+
					"or choose another region.", config.InstanceType, instanceTypes))

				state.Put("error", fmt.Errorf("The instance type %s isn't available in this region."+
					"\n You can either change the instance to one of following: %v \n"+
					"or choose another region.", config.InstanceType, instanceTypes))
				return multistep.ActionHalt
			} else {
				ui.Say(fmt.Sprintf("The instance type %s isn't available in this region."+
					"\n You can change to other regions.", config.InstanceType))

				state.Put("error", fmt.Errorf("The instance type %s isn't available in this region."+
					"\n You can change to other regions.", config.InstanceType))
				return multistep.ActionHalt
			}
		}
	}

	if config.CidrBlock == "" {
		s.CidrBlock = DefaultCidrBlock //use the default CirdBlock
	}

	ui.Say("Creating vswitch...")

	createVSwitchRequest := s.buildCreateVSwitchRequest(state)
	createVSwitchResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return vpcclient.CreateVSwitch(createVSwitchRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createVSwitchRetryErrors, EvalRetryErrorType),
	})
	if err != nil {
		return halt(state, err, "Error Creating vswitch")
	}

	vSwitchId := createVSwitchResponse.(*vpc.CreateVSwitchResponse).VSwitchId

	describeVSwitchesRequest := vpc.CreateDescribeVSwitchesRequest()
	describeVSwitchesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeVSwitchesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	describeVSwitchesRequest.VpcId = vpcId
	describeVSwitchesRequest.VSwitchId = vSwitchId

	_, err = client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return vpcclient.DescribeVSwitches(describeVSwitchesRequest)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			vSwitchesResponse := response.(*vpc.DescribeVSwitchesResponse)
			vSwitches := vSwitchesResponse.VSwitches.VSwitch
			if len(vSwitches) > 0 {
				for _, vSwitch := range vSwitches {
					if vSwitch.Status == VSwitchStatusAvailable {
						return WaitForExpectSuccess
					}
				}
			}

			return WaitForExpectToRetry
		},
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		return halt(state, err, "Timeout waiting for vswitch to become available")
	}

	ui.Message(fmt.Sprintf("Created vswitch: %s", vSwitchId))
	state.Put("vswitchid", vSwitchId)
	s.isCreate = true
	s.VSwitchId = vSwitchId
	return multistep.ActionContinue
}

func (s *stepConfigApsaraStackVSwitch) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	cleanUpMessage(state, "vSwitch")

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	vpcclient, err := vpc.NewClientWithAccessKey(config.ApsaraStackRegion, config.ApsaraStackAccessKey, config.ApsaraStackSecretKey)
	if err != nil {
		panic(err) // TODO eRROR STATEMENT
	}
	vpcclient.Domain = client.Domain
	if client.GetHttpProxy() != "" {
		vpcclient.SetHttpProxy(client.GetHttpProxy())
	}
	_, err = client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := vpc.CreateDeleteVSwitchRequest()
			request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
			request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

			request.VSwitchId = s.VSwitchId
			return vpcclient.DeleteVSwitch(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteVSwitchRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting vswitch, it may still be around: %s", err))
	}
}

func (s *stepConfigApsaraStackVSwitch) buildCreateVSwitchRequest(state multistep.StateBag) *vpc.CreateVSwitchRequest {
	vpcId := state.Get("vpcid").(string)
	config := state.Get("config").(*Config)
	request := vpc.CreateCreateVSwitchRequest()
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	request.ClientToken = uuid.TimeOrderedUUID()
	request.CidrBlock = s.CidrBlock
	request.ZoneId = s.ZoneId
	request.VpcId = vpcId
	request.VSwitchName = s.VSwitchName

	return request
}
