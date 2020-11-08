package ecs

import (
	"context"
	errorsNew "errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	//"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigApsaraStackVPC struct {
	VpcId     string
	CidrBlock string //192.168.0.0/16 or 172.16.0.0/16 (default)
	VpcName   string
	isCreate  bool
}

var createVpcRetryErrors = []string{
	"TOKEN_PROCESSING",
}

var deleteVpcRetryErrors = []string{
	"DependencyViolation.Instance",
	"DependencyViolation.RouteEntry",
	"DependencyViolation.VSwitch",
	"DependencyViolation.SecurityGroup",
	"Forbbiden",
	"TaskConflict",
}

func (s *stepConfigApsaraStackVPC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	vpcclient, err := vpc.NewClientWithAccessKey(config.ApsaraStackRegion, config.ApsaraStackAccessKey, config.ApsaraStackSecretKey)
	if err != nil {
		panic(err) // TODO eRROR STATEMENT
	}
	vpcclient.Domain = client.Domain
	if client.GetHttpProxy() != "" {
		vpcclient.SetHttpProxy(client.GetHttpProxy())
	}
	if len(s.VpcId) != 0 {
		describeVpcsRequest := vpc.CreateDescribeVpcsRequest()
		describeVpcsRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		describeVpcsRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		describeVpcsRequest.VpcId = s.VpcId
		describeVpcsRequest.RegionId = config.ApsaraStackRegion

		vpcsResponse, err := vpcclient.DescribeVpcs(describeVpcsRequest)
		if err != nil {
			return halt(state, err, "Failed querying vpcs")
		}

		vpcs := vpcsResponse.Vpcs.Vpc
		if len(vpcs) > 0 {
			state.Put("vpcid", vpcs[0].VpcId)
			s.isCreate = false
			return multistep.ActionContinue
		}

		message := fmt.Sprintf("The specified vpc {%s} doesn't exist.", s.VpcId)
		return halt(state, errorsNew.New(message), "")
	}

	ui.Say("Creating vpc...")

	createVpcRequest := s.buildCreateVpcRequest(state)
	createVpcRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	createVpcRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	createVpcResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return vpcclient.CreateVpc(createVpcRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createVpcRetryErrors, EvalRetryErrorType),
	})
	if err != nil {
		return halt(state, err, "Failed creating vpc")
	}

	vpcId := createVpcResponse.(*vpc.CreateVpcResponse).VpcId
	_, err = client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := vpc.CreateDescribeVpcsRequest()
			request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
			request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

			request.RegionId = config.ApsaraStackRegion
			request.VpcId = vpcId
			return vpcclient.DescribeVpcs(request)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			vpcsResponse := response.(*vpc.DescribeVpcsResponse)
			vpcs := vpcsResponse.Vpcs.Vpc
			if len(vpcs) > 0 {
				for _, vpc := range vpcs {
					if vpc.Status == VpcStatusAvailable {
						return WaitForExpectSuccess
					}
				}
			}

			return WaitForExpectToRetry
		},
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		return halt(state, err, "Failed waiting for vpc to become available")
	}

	ui.Message(fmt.Sprintf("Created vpc: %s", vpcId))
	state.Put("vpcid", vpcId)
	s.isCreate = true
	s.VpcId = vpcId
	return multistep.ActionContinue
}

func (s *stepConfigApsaraStackVPC) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}
	config := state.Get("config").(*Config)
	cleanUpMessage(state, "VPC")

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
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
			request := vpc.CreateDeleteVpcRequest()
			request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
			request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

			request.VpcId = s.VpcId
			return vpcclient.DeleteVpc(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteVpcRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting vpc, it may still be around: %s", err))
	}
}

func (s *stepConfigApsaraStackVPC) buildCreateVpcRequest(state multistep.StateBag) *vpc.CreateVpcRequest {
	config := state.Get("config").(*Config)

	request := vpc.CreateCreateVpcRequest()
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "vpc", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = config.ApsaraStackRegion
	request.CidrBlock = s.CidrBlock
	request.VpcName = s.VpcName

	return request
}
