package ecs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepStopApsaraStackInstance struct {
	ForceStop   bool
	DisableStop bool
}

func (s *stepStopApsaraStackInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	instance := state.Get("instance").(*ecs.Instance)
	ui := state.Get("ui").(packer.Ui)

	if !s.DisableStop {
		ui.Say(fmt.Sprintf("Stopping instance: %s", instance.InstanceId))

		stopInstanceRequest := ecs.CreateStopInstanceRequest()
		stopInstanceRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
		stopInstanceRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

		stopInstanceRequest.InstanceId = instance.InstanceId
		stopInstanceRequest.ForceStop = requests.Boolean(strconv.FormatBool(s.ForceStop))
		if _, err := client.StopInstance(stopInstanceRequest); err != nil {
			return halt(state, err, "Error stopping ApsaraStack instance")
		}
	}

	ui.Say(fmt.Sprintf("Waiting instance stopped: %s", instance.InstanceId))

	_, err := client.WaitForInstanceStatus(instance.RegionId, instance.InstanceId, InstanceStatusStopped, state)
	if err != nil {
		return halt(state, err, "Error waiting for ApsaraStack instance to stop")
	}

	return multistep.ActionContinue
}

func (s *stepStopApsaraStackInstance) Cleanup(multistep.StateBag) {
	// No cleanup...
}
