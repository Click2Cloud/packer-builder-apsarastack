package ecs

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
)

// FlatApsaraStackDiskDevice is an auto-generated flat version of ApsaraStackDiskDevice.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatApsaraStackDiskDevice struct {
	DiskName           *string `mapstructure:"disk_name" required:"false" cty:"disk_name" hcl:"disk_name"`
	DiskCategory       *string `mapstructure:"disk_category" required:"false" cty:"disk_category" hcl:"disk_category"`
	DiskSize           *int    `mapstructure:"disk_size" required:"false" cty:"disk_size" hcl:"disk_size"`
	SnapshotId         *string `mapstructure:"disk_snapshot_id" required:"false" cty:"disk_snapshot_id" hcl:"disk_snapshot_id"`
	Description        *string `mapstructure:"disk_description" required:"false" cty:"disk_description" hcl:"disk_description"`
	DeleteWithInstance *bool   `mapstructure:"disk_delete_with_instance" required:"false" cty:"disk_delete_with_instance" hcl:"disk_delete_with_instance"`
	Device             *string `mapstructure:"disk_device" required:"false" cty:"disk_device" hcl:"disk_device"`
	Encrypted          *bool   `mapstructure:"disk_encrypted" required:"false" cty:"disk_encrypted" hcl:"disk_encrypted"`
}

// FlatMapstructure returns a new FlatApsaraStackDiskDevice.
// FlatApsaraStackDiskDevice is an auto-generated flat version of ApsaraStackDiskDevice.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*ApsaraStackDiskDevice) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatApsaraStackDiskDevice)
}

// HCL2Spec returns the hcl spec of a ApsaraStackDiskDevice.
// This spec is used by HCL to read the fields of ApsaraStackDiskDevice.
// The decoded values from this spec will then be applied to a FlatApsaraStackDiskDevice.
func (*FlatApsaraStackDiskDevice) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"disk_name":                 &hcldec.AttrSpec{Name: "disk_name", Type: cty.String, Required: false},
		"disk_category":             &hcldec.AttrSpec{Name: "disk_category", Type: cty.String, Required: false},
		"disk_size":                 &hcldec.AttrSpec{Name: "disk_size", Type: cty.Number, Required: false},
		"disk_snapshot_id":          &hcldec.AttrSpec{Name: "disk_snapshot_id", Type: cty.String, Required: false},
		"disk_description":          &hcldec.AttrSpec{Name: "disk_description", Type: cty.String, Required: false},
		"disk_delete_with_instance": &hcldec.AttrSpec{Name: "disk_delete_with_instance", Type: cty.Bool, Required: false},
		"disk_device":               &hcldec.AttrSpec{Name: "disk_device", Type: cty.String, Required: false},
		"disk_encrypted":            &hcldec.AttrSpec{Name: "disk_encrypted", Type: cty.Bool, Required: false},
	}
	return s
}

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	PackerBuildName                      *string                     `mapstructure:"packer_build_name" cty:"packer_build_name" hcl:"packer_build_name"`
	PackerBuilderType                    *string                     `mapstructure:"packer_builder_type" cty:"packer_builder_type" hcl:"packer_builder_type"`
	PackerDebug                          *bool                       `mapstructure:"packer_debug" cty:"packer_debug" hcl:"packer_debug"`
	PackerForce                          *bool                       `mapstructure:"packer_force" cty:"packer_force" hcl:"packer_force"`
	PackerOnError                        *string                     `mapstructure:"packer_on_error" cty:"packer_on_error" hcl:"packer_on_error"`
	PackerUserVars                       map[string]string           `mapstructure:"packer_user_variables" cty:"packer_user_variables" hcl:"packer_user_variables"`
	PackerSensitiveVars                  []string                    `mapstructure:"packer_sensitive_variables" cty:"packer_sensitive_variables" hcl:"packer_sensitive_variables"`
	ApsaraStackAccessKey                 *string                     `mapstructure:"access_key" required:"true" cty:"access_key" hcl:"access_key"`
	ApsaraStackSecretKey                 *string                     `mapstructure:"secret_key" required:"true" cty:"secret_key" hcl:"secret_key"`
	ApsaraStackRegion                    *string                     `mapstructure:"region" required:"true" cty:"region" hcl:"region"`
	ApsaraStackSkipValidation            *bool                       `mapstructure:"skip_region_validation" required:"false" cty:"skip_region_validation" hcl:"skip_region_validation"`
	ApsaraStackSkipImageValidation       *bool                       `mapstructure:"skip_image_validation" required:"false" cty:"skip_image_validation" hcl:"skip_image_validation"`
	ApsaraStackProfile                   *string                     `mapstructure:"profile" required:"false" cty:"profile" hcl:"profile"`
	ApsaraStackSharedCredentialsFile     *string                     `mapstructure:"shared_credentials_file" required:"false" cty:"shared_credentials_file" hcl:"shared_credentials_file"`
	SecurityToken                        *string                     `mapstructure:"security_token" required:"false" cty:"security_token" hcl:"security_token"`
	ApsaraStackImageName                 *string                     `mapstructure:"image_name" required:"true" cty:"image_name" hcl:"image_name"`
	ApsaraStackImageVersion              *string                     `mapstructure:"image_version" required:"false" cty:"image_version" hcl:"image_version"`
	ApsaraStackImageDescription          *string                     `mapstructure:"image_description" required:"false" cty:"image_description" hcl:"image_description"`
	ApsaraStackImageShareAccounts        []string                    `mapstructure:"image_share_account" required:"false" cty:"image_share_account" hcl:"image_share_account"`
	ApsaraStackImageUNShareAccounts      []string                    `mapstructure:"image_unshare_account" cty:"image_unshare_account" hcl:"image_unshare_account"`
	ApsaraStackImageDestinationRegions   []string                    `mapstructure:"image_copy_regions" required:"false" cty:"image_copy_regions" hcl:"image_copy_regions"`
	ApsaraStackImageDestinationNames     []string                    `mapstructure:"image_copy_names" required:"false" cty:"image_copy_names" hcl:"image_copy_names"`
	ImageEncrypted                       *bool                       `mapstructure:"image_encrypted" required:"false" cty:"image_encrypted" hcl:"image_encrypted"`
	ApsaraStackImageForceDelete          *bool                       `mapstructure:"image_force_delete" required:"false" cty:"image_force_delete" hcl:"image_force_delete"`
	ApsaraStackImageForceDeleteSnapshots *bool                       `mapstructure:"image_force_delete_snapshots" required:"false" cty:"image_force_delete_snapshots" hcl:"image_force_delete_snapshots"`
	ApsaraStackImageForceDeleteInstances *bool                       `mapstructure:"image_force_delete_instances" cty:"image_force_delete_instances" hcl:"image_force_delete_instances"`
	ApsaraStackImageIgnoreDataDisks      *bool                       `mapstructure:"image_ignore_data_disks" required:"false" cty:"image_ignore_data_disks" hcl:"image_ignore_data_disks"`
	ApsaraStackImageTags                 map[string]string           `mapstructure:"tags" required:"false" cty:"tags" hcl:"tags"`
	ApsaraStackImageTag                  []hcl2template.FlatKeyValue `mapstructure:"tag" required:"false" cty:"tag" hcl:"tag"`
	ECSSystemDiskMapping                 *FlatApsaraStackDiskDevice  `mapstructure:"system_disk_mapping" required:"false" cty:"system_disk_mapping" hcl:"system_disk_mapping"`
	ECSImagesDiskMappings                []FlatApsaraStackDiskDevice `mapstructure:"image_disk_mappings" required:"false" cty:"image_disk_mappings" hcl:"image_disk_mappings"`
	AssociatePublicIpAddress             *bool                       `mapstructure:"associate_public_ip_address" cty:"associate_public_ip_address" hcl:"associate_public_ip_address"`
	ZoneId                               *string                     `mapstructure:"zone_id" required:"false" cty:"zone_id" hcl:"zone_id"`
	IOOptimized                          *bool                       `mapstructure:"io_optimized" required:"false" cty:"io_optimized" hcl:"io_optimized"`
	InstanceType                         *string                     `mapstructure:"instance_type" required:"true" cty:"instance_type" hcl:"instance_type"`
	Description                          *string                     `mapstructure:"description" cty:"description" hcl:"description"`
	ApsaraStackSourceImage               *string                     `mapstructure:"source_image" required:"true" cty:"source_image" hcl:"source_image"`
	ForceStopInstance                    *bool                       `mapstructure:"force_stop_instance" required:"false" cty:"force_stop_instance" hcl:"force_stop_instance"`
	DisableStopInstance                  *bool                       `mapstructure:"disable_stop_instance" required:"false" cty:"disable_stop_instance" hcl:"disable_stop_instance"`
	SecurityGroupId                      *string                     `mapstructure:"security_group_id" required:"false" cty:"security_group_id" hcl:"security_group_id"`
	SecurityGroupName                    *string                     `mapstructure:"security_group_name" required:"false" cty:"security_group_name" hcl:"security_group_name"`
	UserData                             *string                     `mapstructure:"user_data" required:"false" cty:"user_data" hcl:"user_data"`
	UserDataFile                         *string                     `mapstructure:"user_data_file" required:"false" cty:"user_data_file" hcl:"user_data_file"`
	VpcId                                *string                     `mapstructure:"vpc_id" required:"false" cty:"vpc_id" hcl:"vpc_id"`
	VpcName                              *string                     `mapstructure:"vpc_name" required:"false" cty:"vpc_name" hcl:"vpc_name"`
	CidrBlock                            *string                     `mapstructure:"vpc_cidr_block" required:"false" cty:"vpc_cidr_block" hcl:"vpc_cidr_block"`
	VSwitchId                            *string                     `mapstructure:"vswitch_id" required:"false" cty:"vswitch_id" hcl:"vswitch_id"`
	VSwitchName                          *string                     `mapstructure:"vswitch_name" required:"false" cty:"vswitch_name" hcl:"vswitch_name"`
	InstanceName                         *string                     `mapstructure:"instance_name" required:"false" cty:"instance_name" hcl:"instance_name"`
	InternetMaxBandwidthOut              *int                        `mapstructure:"internet_max_bandwidth_out" required:"false" cty:"internet_max_bandwidth_out" hcl:"internet_max_bandwidth_out"`
	WaitSnapshotReadyTimeout             *int                        `mapstructure:"wait_snapshot_ready_timeout" required:"false" cty:"wait_snapshot_ready_timeout" hcl:"wait_snapshot_ready_timeout"`
	Type                                 *string                     `mapstructure:"communicator" cty:"communicator" hcl:"communicator"`
	PauseBeforeConnect                   *string                     `mapstructure:"pause_before_connecting" cty:"pause_before_connecting" hcl:"pause_before_connecting"`

	SSHPrivateIp                         *bool                       `mapstructure:"ssh_private_ip" required:"false" cty:"ssh_private_ip" hcl:"ssh_private_ip"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"packer_build_name":            &hcldec.AttrSpec{Name: "packer_build_name", Type: cty.String, Required: false},
		"packer_builder_type":          &hcldec.AttrSpec{Name: "packer_builder_type", Type: cty.String, Required: false},
		"packer_debug":                 &hcldec.AttrSpec{Name: "packer_debug", Type: cty.Bool, Required: false},
		"packer_force":                 &hcldec.AttrSpec{Name: "packer_force", Type: cty.Bool, Required: false},
		"packer_on_error":              &hcldec.AttrSpec{Name: "packer_on_error", Type: cty.String, Required: false},
		"packer_user_variables":        &hcldec.AttrSpec{Name: "packer_user_variables", Type: cty.Map(cty.String), Required: false},
		"packer_sensitive_variables":   &hcldec.AttrSpec{Name: "packer_sensitive_variables", Type: cty.List(cty.String), Required: false},
		"access_key":                   &hcldec.AttrSpec{Name: "access_key", Type: cty.String, Required: false},
		"secret_key":                   &hcldec.AttrSpec{Name: "secret_key", Type: cty.String, Required: false},
		"region":                       &hcldec.AttrSpec{Name: "region", Type: cty.String, Required: false},
		"skip_region_validation":       &hcldec.AttrSpec{Name: "skip_region_validation", Type: cty.Bool, Required: false},
		"skip_image_validation":        &hcldec.AttrSpec{Name: "skip_image_validation", Type: cty.Bool, Required: false},
		"profile":                      &hcldec.AttrSpec{Name: "profile", Type: cty.String, Required: false},
		"shared_credentials_file":      &hcldec.AttrSpec{Name: "shared_credentials_file", Type: cty.String, Required: false},
		"security_token":               &hcldec.AttrSpec{Name: "security_token", Type: cty.String, Required: false},
		"image_name":                   &hcldec.AttrSpec{Name: "image_name", Type: cty.String, Required: false},
		"image_version":                &hcldec.AttrSpec{Name: "image_version", Type: cty.String, Required: false},
		"image_description":            &hcldec.AttrSpec{Name: "image_description", Type: cty.String, Required: false},
		"image_share_account":          &hcldec.AttrSpec{Name: "image_share_account", Type: cty.List(cty.String), Required: false},
		"image_unshare_account":        &hcldec.AttrSpec{Name: "image_unshare_account", Type: cty.List(cty.String), Required: false},
		"image_copy_regions":           &hcldec.AttrSpec{Name: "image_copy_regions", Type: cty.List(cty.String), Required: false},
		"image_copy_names":             &hcldec.AttrSpec{Name: "image_copy_names", Type: cty.List(cty.String), Required: false},
		"image_encrypted":              &hcldec.AttrSpec{Name: "image_encrypted", Type: cty.Bool, Required: false},
		"image_force_delete":           &hcldec.AttrSpec{Name: "image_force_delete", Type: cty.Bool, Required: false},
		"image_force_delete_snapshots": &hcldec.AttrSpec{Name: "image_force_delete_snapshots", Type: cty.Bool, Required: false},
		"image_force_delete_instances": &hcldec.AttrSpec{Name: "image_force_delete_instances", Type: cty.Bool, Required: false},
		"image_ignore_data_disks":      &hcldec.AttrSpec{Name: "image_ignore_data_disks", Type: cty.Bool, Required: false},
		"tags":                         &hcldec.AttrSpec{Name: "tags", Type: cty.Map(cty.String), Required: false},
		"tag":                          &hcldec.BlockListSpec{TypeName: "tag", Nested: hcldec.ObjectSpec((*hcl2template.FlatKeyValue)(nil).HCL2Spec())},
		"system_disk_mapping":          &hcldec.BlockSpec{TypeName: "system_disk_mapping", Nested: hcldec.ObjectSpec((*FlatApsaraStackDiskDevice)(nil).HCL2Spec())},
		"image_disk_mappings":          &hcldec.BlockListSpec{TypeName: "image_disk_mappings", Nested: hcldec.ObjectSpec((*FlatApsaraStackDiskDevice)(nil).HCL2Spec())},
		"associate_public_ip_address":  &hcldec.AttrSpec{Name: "associate_public_ip_address", Type: cty.Bool, Required: false},
		"zone_id":                      &hcldec.AttrSpec{Name: "zone_id", Type: cty.String, Required: false},
		"io_optimized":                 &hcldec.AttrSpec{Name: "io_optimized", Type: cty.Bool, Required: false},
		"instance_type":                &hcldec.AttrSpec{Name: "instance_type", Type: cty.String, Required: false},
		"description":                  &hcldec.AttrSpec{Name: "description", Type: cty.String, Required: false},
		"source_image":                 &hcldec.AttrSpec{Name: "source_image", Type: cty.String, Required: false},
		"force_stop_instance":          &hcldec.AttrSpec{Name: "force_stop_instance", Type: cty.Bool, Required: false},
		"disable_stop_instance":        &hcldec.AttrSpec{Name: "disable_stop_instance", Type: cty.Bool, Required: false},
		"security_group_id":            &hcldec.AttrSpec{Name: "security_group_id", Type: cty.String, Required: false},
		"security_group_name":          &hcldec.AttrSpec{Name: "security_group_name", Type: cty.String, Required: false},
		"user_data":                    &hcldec.AttrSpec{Name: "user_data", Type: cty.String, Required: false},
		"user_data_file":               &hcldec.AttrSpec{Name: "user_data_file", Type: cty.String, Required: false},
		"vpc_id":                       &hcldec.AttrSpec{Name: "vpc_id", Type: cty.String, Required: false},
		"vpc_name":                     &hcldec.AttrSpec{Name: "vpc_name", Type: cty.String, Required: false},
		"vpc_cidr_block":               &hcldec.AttrSpec{Name: "vpc_cidr_block", Type: cty.String, Required: false},
		"vswitch_id":                   &hcldec.AttrSpec{Name: "vswitch_id", Type: cty.String, Required: false},
		"vswitch_name":                 &hcldec.AttrSpec{Name: "vswitch_name", Type: cty.String, Required: false},
		"instance_name":                &hcldec.AttrSpec{Name: "instance_name", Type: cty.String, Required: false},
		"internet_max_bandwidth_out":   &hcldec.AttrSpec{Name: "internet_max_bandwidth_out", Type: cty.Number, Required: false},
		"wait_snapshot_ready_timeout":  &hcldec.AttrSpec{Name: "wait_snapshot_ready_timeout", Type: cty.Number, Required: false},
		"communicator":                 &hcldec.AttrSpec{Name: "communicator", Type: cty.String, Required: false},
		"pause_before_connecting":      &hcldec.AttrSpec{Name: "pause_before_connecting", Type: cty.String, Required: false},
		"ssh_private_ip":               &hcldec.AttrSpec{Name: "ssh_private_ip", Type: cty.Bool, Required: false},
	}
	return s
}
