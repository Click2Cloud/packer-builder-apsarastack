package ecs

import (
	"context"
	"github.com/hashicorp/hcl/v2/hcldec"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "apsarastack.apsarastack"

type Config struct {
	common.PackerConfig     `mapstructure:",squash"`
	ApsaraStackAccessConfig `mapstructure:",squash"`
	ApsaraStackImageConfig  `mapstructure:",squash"`
	RunConfig               `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type InstanceNetWork string

const (
	APSARASTACK_DEFAULT_SHORT_TIMEOUT = 180
	APSARASTACK_DEFAULT_TIMEOUT       = 1800
	APSARASTACK_DEFAULT_LONG_TIMEOUT  = 3600
)

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	b.config.ctx.EnableEnv = true
	if err != nil {
		return nil, nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.ApsaraStackImageForceDelete = true
		b.config.ApsaraStackImageForceDeleteSnapshots = true
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.ApsaraStackAccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ApsaraStackImageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(b.config.ApsaraStackAccessKey, b.config.ApsaraStackSecretKey)
	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	client, err := b.config.Client()

	if err != nil {
		return nil, err
	}
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("networktype", b.chooseNetworkType())
	var steps []multistep.Step

	// Build the steps
	steps = []multistep.Step{
		&stepPreValidate{
			ApsaraStackDestImageName: b.config.ApsaraStackImageName,
			ForceDelete:              b.config.ApsaraStackImageForceDelete,
		},
		&stepCheckApsaraStackSourceImage{
			SourceECSImageId: b.config.ApsaraStackSourceImage,
		},
	}
	if b.chooseNetworkType() == InstanceNetworkVpc {
		steps = append(steps,
			&stepConfigApsaraStackVPC{
				VpcId:     b.config.VpcId,
				CidrBlock: b.config.CidrBlock,
				VpcName:   b.config.VpcName,
			},
			&stepConfigApsaraStackVSwitch{
				VSwitchId:   b.config.VSwitchId,
				ZoneId:      b.config.ZoneId,
				CidrBlock:   b.config.CidrBlock,
				VSwitchName: b.config.VSwitchName,
			})
	}
	steps = append(steps,
		&stepConfigApsaraStackSecurityGroup{
			SecurityGroupId:   b.config.SecurityGroupId,
			SecurityGroupName: b.config.SecurityGroupId,
			RegionId:          b.config.ApsaraStackRegion,
		},
		&stepCreateApsaraStackInstance{
			IOOptimized:             b.config.IOOptimized,
			InstanceType:            b.config.InstanceType,
			UserData:                b.config.UserData,
			UserDataFile:            b.config.UserDataFile,
			RegionId:                b.config.ApsaraStackRegion,
			InternetChargeType:      b.config.InternetChargeType,
			InternetMaxBandwidthOut: b.config.InternetMaxBandwidthOut,
			InstanceName:            b.config.InstanceName,
			ZoneId:                  b.config.ZoneId,
		})
	if b.chooseNetworkType() == InstanceNetworkVpc {
		steps = append(steps, &stepConfigApsaraStackEIP{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			RegionId:                 b.config.ApsaraStackRegion,
			InternetChargeType:       b.config.InternetChargeType,
			InternetMaxBandwidthOut:  b.config.InternetMaxBandwidthOut,
			SSHPrivateIp:             b.config.SSHPrivateIp,
		})
	} else {
		steps = append(steps, &stepConfigApsaraStackPublicIP{
			RegionId:     b.config.ApsaraStackRegion,
			SSHPrivateIp: b.config.SSHPrivateIp,
		})
	}
	steps = append(steps,
		&stepRunApsaraStackInstance{},
		&stepStopApsaraStackInstance{
			ForceStop:   b.config.ForceStopInstance,
			DisableStop: b.config.DisableStopInstance,
		},
		&stepDeleteApsaraStackImageSnapshots{
			ApsaraStackImageForceDeleteSnapshots: b.config.ApsaraStackImageForceDeleteSnapshots,
			ApsaraStackImageForceDelete:          b.config.ApsaraStackImageForceDelete,
			ApsaraStackImageName:                 b.config.ApsaraStackImageName,
			ApsaraStackImageDestinationRegions:   b.config.ApsaraStackImageConfig.ApsaraStackImageDestinationRegions,
			ApsaraStackImageDestinationNames:     b.config.ApsaraStackImageConfig.ApsaraStackImageDestinationNames,
		})

	if b.config.ApsaraStackImageIgnoreDataDisks {
		steps = append(steps, &stepCreateApsaraStackSnapshot{
			WaitSnapshotReadyTimeout: b.getSnapshotReadyTimeout(),
		})
	}

	steps = append(steps,
		&stepCreateApsaraStackImage{
			ApsaraStackImageIgnoreDataDisks: b.config.ApsaraStackImageIgnoreDataDisks,
			WaitSnapshotReadyTimeout:        b.getSnapshotReadyTimeout(),
		},
		&stepCreateTags{
			Tags: b.config.ApsaraStackImageTags,
		},
		&stepRegionCopyApsaraStackImage{
			ApsaraStackImageDestinationRegions: b.config.ApsaraStackImageDestinationRegions,
			ApsaraStackImageDestinationNames:   b.config.ApsaraStackImageDestinationNames,
			RegionId:                           b.config.ApsaraStackRegion,
		},
		&stepShareApsaraStackImage{
			ApsaraStackImageShareAccounts:   b.config.ApsaraStackImageShareAccounts,
			ApsaraStackImageUNShareAccounts: b.config.ApsaraStackImageUNShareAccounts,
			RegionId:                        b.config.ApsaraStackRegion,
		})

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no ECS images, then just return
	if _, ok := state.GetOk("ApsaraStackimages"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		ApsaraStackImages: state.Get("ApsaraStackimages").(map[string]string),
		BuilderIdValue:    BuilderId,
		Client:            client,
		Config:            &b.config,
	}

	return artifact, nil
}

func (b *Builder) chooseNetworkType() InstanceNetWork {
	if b.isVpcNetRequired() {
		return InstanceNetworkVpc
	} else {
		return InstanceNetworkClassic
	}
}

func (b *Builder) isVpcNetRequired() bool {
	// UserData and KeyPair only works in VPC
	return b.isVpcSpecified() || b.isUserDataNeeded() || b.isKeyPairNeeded()
}

func (b *Builder) isVpcSpecified() bool {
	return b.config.VpcId != "" || b.config.VSwitchId != ""
}

func (b *Builder) isUserDataNeeded() bool {
	// Public key setup requires userdata
	if b.config.RunConfig.Comm.SSHPrivateKeyFile != "" {
		return true
	}

	return b.config.UserData != "" || b.config.UserDataFile != ""
}

func (b *Builder) isKeyPairNeeded() bool {
	return b.config.Comm.SSHKeyPairName != "" || b.config.Comm.SSHTemporaryKeyPairName != ""
}

func (b *Builder) getSnapshotReadyTimeout() int {
	if b.config.WaitSnapshotReadyTimeout > 0 {
		return b.config.WaitSnapshotReadyTimeout
	}

	return APSARASTACK_DEFAULT_LONG_TIMEOUT
}
