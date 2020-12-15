package ecs

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/packer/common/uuid"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	confighelper "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateApsaraStackInstance struct {
	IOOptimized             confighelper.Trilean
	InstanceType            string
	UserData                string
	UserDataFile            string
	instanceId              string
	RegionId                string
	InternetChargeType      string
	InternetMaxBandwidthOut int
	InstanceName            string
	ZoneId                  string
	instance                *ecs.Instance
}

var createInstanceRetryErrors = []string{
	"IdempotentProcessing",
}

var deleteInstanceRetryErrors = []string{
	"IncorrectInstanceStatus.Initializing",
}

func (s *stepCreateApsaraStackInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating instance...")
	createInstanceRequest, err := s.buildCreateInstanceRequest(state)
	if err != nil {
		return halt(state, err, "")
	}

	createInstanceResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.RunInstances(createInstanceRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createInstanceRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating instance")
	}

	instanceId := createInstanceResponse.(*ecs.RunInstancesResponse).InstanceIdSets.InstanceIdSet[0]

	_, err = client.WaitForInstanceStatus(s.RegionId, instanceId, InstanceStatusRunning, state)
	if err != nil {
		return halt(state, err, "Error waiting create instance")
	}
	ui.Say("[EXTRA] Creation Complete, describing instance...")
	describeInstancesRequest := ecs.CreateDescribeInstancesRequest()
	if strings.ToLower(config.Protocol) == "https" {
		describeInstancesRequest.Scheme = "https"
	} else {
		describeInstancesRequest.Scheme = "http"
	}
	describeInstancesRequest.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	describeInstancesRequest.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	describeInstancesRequest.InstanceIds = fmt.Sprintf("[\"%s\"]", instanceId)
	instances, err := client.DescribeInstances(describeInstancesRequest)
	if err != nil {
		return halt(state, err, "")
	}

	ui.Message(fmt.Sprintf("Created instance: %s", instanceId))
	s.instance = &instances.Instances.Instance[0]
	state.Put("instance", s.instance)
	//state.Put("ipaddress",s.instance.VpcAttributes.PrivateIpAddress)
	state.Put("ipaddress", s.instance.VpcAttributes.PrivateIpAddress.IpAddress[0])
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", instanceId)

	return multistep.ActionContinue
}

func (s *stepCreateApsaraStackInstance) Cleanup(state multistep.StateBag) {
	if s.instance == nil {
		return
	}
	cleanUpMessage(state, "instance")

	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDeleteInstanceRequest()
			if strings.ToLower(config.Protocol) == "https" {
				request.Scheme = "https"
			} else {
				request.Scheme = "http"
			}

			request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
			request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

			request.InstanceId = s.instance.InstanceId
			request.Force = requests.NewBoolean(true)
			return client.DeleteInstance(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteInstanceRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Failed to clean up instance %s: %s", s.instance.InstanceId, err))
	}
}

func (s *stepCreateApsaraStackInstance) buildCreateInstanceRequest(state multistep.StateBag) (*ecs.RunInstancesRequest, error) {
	request := ecs.CreateRunInstancesRequest()

	config := state.Get("config").(*Config)
	if strings.ToLower(config.Protocol) == "https" {
		request.Scheme = "https"
	} else {
		request.Scheme = "http"
	}
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{"AccessKeySecret": config.ApsaraStackSecretKey, "Product": "ecs", "Department": config.Department, "ResourceGroup": config.ResourceGroup}

	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = s.RegionId
	request.InstanceType = s.InstanceType
	request.InstanceName = s.InstanceName
	request.ZoneId = s.ZoneId

	sourceImage := state.Get("source_image").(*ecs.Image)
	request.ImageId = sourceImage.ImageId

	securityGroupId := state.Get("securitygroupid").(string)
	request.SecurityGroupId = securityGroupId

	networkType := state.Get("networktype").(InstanceNetWork)
	if networkType == InstanceNetworkVpc {
		vswitchId := state.Get("vswitchid").(string)
		request.VSwitchId = vswitchId

		userData, err := s.getUserData(state)
		if err != nil {
			return nil, err
		}

		request.UserData = userData
	} else {

		if s.InternetMaxBandwidthOut == 0 {
			s.InternetMaxBandwidthOut = 5
		}
	}
	request.InternetChargeType = s.InternetChargeType
	request.InternetMaxBandwidthOut = requests.Integer(convertNumber(s.InternetMaxBandwidthOut))

	if s.IOOptimized.True() {
		request.IoOptimized = IOOptimizedOptimized
	} else if s.IOOptimized.False() {
		request.IoOptimized = IOOptimizedNone
	}

	password := config.Comm.SSHPassword
	if password == "" && config.Comm.WinRMPassword != "" {
		password = config.Comm.WinRMPassword
	}
	request.Password = password

	systemDisk := config.ApsaraStackImageConfig.ECSSystemDiskMapping
	request.SystemDiskDiskName = systemDisk.DiskName
	request.SystemDiskCategory = systemDisk.DiskCategory
	request.SystemDiskSize = (convertNumber(systemDisk.DiskSize))
	request.SystemDiskDescription = systemDisk.Description

	imageDisks := config.ApsaraStackImageConfig.ECSImagesDiskMappings
	var dataDisks []ecs.RunInstancesDataDisk
	for _, imageDisk := range imageDisks {
		var dataDisk ecs.RunInstancesDataDisk
		dataDisk.DiskName = imageDisk.DiskName
		dataDisk.Category = imageDisk.DiskCategory
		dataDisk.Size = string(convertNumber(imageDisk.DiskSize))
		dataDisk.SnapshotId = imageDisk.SnapshotId
		dataDisk.Description = imageDisk.Description
		dataDisk.DeleteWithInstance = strconv.FormatBool(imageDisk.DeleteWithInstance)
		dataDisk.Device = imageDisk.Device
		if imageDisk.Encrypted != confighelper.TriUnset {
			dataDisk.Encrypted = strconv.FormatBool(imageDisk.Encrypted.True())
		}

		dataDisks = append(dataDisks, dataDisk)
	}
	request.DataDisk = &dataDisks

	return request, nil
}

func (s *stepCreateApsaraStackInstance) getUserData(state multistep.StateBag) (string, error) {
	userData := s.UserData

	if s.UserDataFile != "" {
		data, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", err
		}

		userData = string(data)
	}
	/*	if s.UserDataFile != "" {
		data, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", err
		}

		userData = string(data)
	}*/
	if userData != "" {
		userData = base64.StdEncoding.EncodeToString([]byte(userData))
	}
	/*if userData != "" {
		_, base64DecodeError := base64.StdEncoding.DecodeString(userData)
		if base64DecodeError == nil {
			s.UserData = userData
		} else {
			s.UserData = base64.StdEncoding.EncodeToString([]byte(userData))
		}
	}*/
	/*if a := userData; a != "" {
		_, base64DecodeError := base64.StdEncoding.DecodeString(a)
		if base64DecodeError == nil {
			s.UserData = a
		} else {
			s.UserData = base64.StdEncoding.EncodeToString([]byte(a))
		}
	}*/

	return userData, nil

}
