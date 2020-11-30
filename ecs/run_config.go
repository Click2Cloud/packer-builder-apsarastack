package ecs

import (
	"errors"
	"fmt"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/template/interpolate"
	"os"
	"strings"
)

type RunConfig struct {
	AssociatePublicIpAddress bool `mapstructure:"associate_public_ip_address"`
	ZoneId string `mapstructure:"zone_id" required:"false"`
	IOOptimized config.Trilean `mapstructure:"io_optimized" required:"false"`
	InstanceType string `mapstructure:"instance_type" required:"true"`
	Description  string `mapstructure:"description"`
	ApsaraStackSourceImage string `mapstructure:"source_image" required:"true"`
	ForceStopInstance bool `mapstructure:"force_stop_instance" required:"false"`
	DisableStopInstance bool `mapstructure:"disable_stop_instance" required:"false"`
	SecurityGroupId string `mapstructure:"security_group_id" required:"false"`
	SecurityGroupName string `mapstructure:"security_group_name" required:"false"`
	UserData string `mapstructure:"user_data" required:"false"`
	UserDataFile string `mapstructure:"user_data_file" required:"false"`
	VpcId string `mapstructure:"vpc_id" required:"false"`
	VpcName string `mapstructure:"vpc_name" required:"false"`
	CidrBlock string `mapstructure:"vpc_cidr_block" required:"false"`
	VSwitchId string `mapstructure:"vswitch_id" required:"false"`
	VSwitchName string `mapstructure:"vswitch_name" required:"false"`
	InstanceName string `mapstructure:"instance_name" required:"false"`
	InternetMaxBandwidthOut int `mapstructure:"internet_max_bandwidth_out" required:"false"`
	WaitSnapshotReadyTimeout int `mapstructure:"wait_snapshot_ready_timeout" required:"false"`
	Comm communicator.Config `mapstructure:",squash"`
	SSHPrivateIp bool `mapstructure:"ssh_private_ip" required:"false"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	/*if c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" && c.Comm.WinRMPassword == "" {

		c.Comm.SSHTimeout = 10 * time.Minute
		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
		
	}*/
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" && c.Comm.WinRMPassword == "" {

		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}
	// Validation
	errs := c.Comm.Prepare(ctx)
	if c.ApsaraStackSourceImage == "" {
		errs = append(errs, errors.New("A source_image must be specified"))
	}

	if strings.TrimSpace(c.ApsaraStackSourceImage) != c.ApsaraStackSourceImage {
		errs = append(errs, errors.New("The source_image can't include spaces"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An ApsaraStack_instance_type must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	return errs
}
