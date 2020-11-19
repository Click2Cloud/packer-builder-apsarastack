package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepRunApsaraStackInstance struct {
}

func (s *stepRunApsaraStackInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.Instance)

	startInstanceRequest := ecs.CreateStartInstanceRequest()
	startInstanceRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	startInstanceRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	startInstanceRequest.InstanceId = instance.InstanceId
	if _, err := client.StartInstance(startInstanceRequest); err != nil {
		return halt(state, err, "Error starting instance")
	}

	ui.Say(fmt.Sprintf("Starting instance: %s", instance.InstanceId))

	_, err := client.WaitForInstanceStatus(instance.RegionId, instance.InstanceId, InstanceStatusRunning, state)
	if err != nil {
		return halt(state, err, "Timeout waiting for instance to start")
	}

	return multistep.ActionContinue
}

func (s *stepRunApsaraStackInstance) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	instance := state.Get("instance").(*ecs.Instance)

	describeInstancesRequest := ecs.CreateDescribeInstancesRequest()
	describeInstancesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeInstancesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs","Department": config.Department, "ResourceGroup": config.ResourceGroup}

	describeInstancesRequest.InstanceIds = fmt.Sprintf("[\"%s\"]", instance.InstanceId)
	instancesResponse, _ := client.DescribeInstances(describeInstancesRequest)

	if len(instancesResponse.Instances.Instance) == 0 {
		return
	}

	instanceAttribute := instancesResponse.Instances.Instance[0]
	if instanceAttribute.Status == InstanceStatusStarting || instanceAttribute.Status == InstanceStatusRunning {
		stopInstanceRequest := ecs.CreateStopInstanceRequest()
		stopInstanceRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		stopInstanceRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		stopInstanceRequest.InstanceId = instance.InstanceId
		stopInstanceRequest.ForceStop = requests.NewBoolean(true)
		if _, err := client.StopInstance(stopInstanceRequest); err != nil {
			ui.Say(fmt.Sprintf("Error stopping instance %s, it may still be around %s", instance.InstanceId, err))
			return
		}

		_, err := client.WaitForInstanceStatus(instance.RegionId, instance.InstanceId, InstanceStatusStopped, state)
		if err != nil {
			ui.Say(fmt.Sprintf("Error stopping instance %s, it may still be around %s", instance.InstanceId, err))
		}
	}
}
