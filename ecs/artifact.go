package ecs

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer"
)

type Artifact struct {
	// A map of regions to ApsaraStack image IDs.
	ApsaraStackImages map[string]string

	// BuilderId is the unique ID for the builder that created this ApsaraStack image
	BuilderIdValue string

	// Alcloud connection for performing API stuff.
	Config *Config
	Client *ClientWrapper
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.ApsaraStackImages))
	for region, ecsImageId := range a.ApsaraStackImages {
		parts = append(parts, fmt.Sprintf("%s:%s", region, ecsImageId))
	}

	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	ApsaraStackImageStrings := make([]string, 0, len(a.ApsaraStackImages))
	for region, id := range a.ApsaraStackImages {
		single := fmt.Sprintf("%s: %s", region, id)
		ApsaraStackImageStrings = append(ApsaraStackImageStrings, single)
	}

	sort.Strings(ApsaraStackImageStrings)
	return fmt.Sprintf("ApsaraStack images were created:\n\n%s", strings.Join(ApsaraStackImageStrings, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {

	errors := make([]error, 0)
	//config := state.Get("config").(*Config)

	copyingImages := make(map[string]string, len(a.ApsaraStackImages))
	sourceImage := make(map[string]*ecs.Image, 1)
	for regionId, imageId := range a.ApsaraStackImages {
		describeImagesRequest := ecs.CreateDescribeImagesRequest()
		if strings.ToLower(a.Config.Protocol) == "https" {
			describeImagesRequest.Scheme = "https"
		} else {
			describeImagesRequest.Scheme = "http"
		}
		describeImagesRequest.Headers = map[string]string{"RegionId": a.Config.ApsaraStackRegion}
		describeImagesRequest.QueryParams = map[string]string{"AccessKeySecret": a.Config.ApsaraStackSecretKey, "Product": "ecs", "Department": a.Config.Department, "ResourceGroup": a.Config.ResourceGroup}

		describeImagesRequest.RegionId = regionId
		describeImagesRequest.ImageId = imageId
		describeImagesRequest.Status = ImageStatusQueried

		imagesResponse, err := a.Client.DescribeImages(describeImagesRequest)
		if err != nil {
			errors = append(errors, err)
		}

		images := imagesResponse.Images.Image
		if len(images) == 0 {
			err := fmt.Errorf("Error retrieving details for ApsaraStack image(%s), no ApsaraStack images found", imageId)
			errors = append(errors, err)
			continue
		}

		if images[0].IsCopied && images[0].Status != ImageStatusAvailable {
			copyingImages[regionId] = imageId
		} else {
			sourceImage[regionId] = &images[0]
		}
	}

	for regionId, imageId := range copyingImages {
		log.Printf("Cancel copying ApsaraStack image (%s) from region (%s)", imageId, regionId)

		errs := a.unsharedAccountsOnImages(regionId, imageId)
		if errs != nil {
			errors = append(errors, errs...)
		}

		cancelImageCopyRequest := ecs.CreateCancelCopyImageRequest()
		if strings.ToLower(a.Config.Protocol) == "https" {
			cancelImageCopyRequest.Scheme = "https"
		} else {
			cancelImageCopyRequest.Scheme = "http"
		}
		cancelImageCopyRequest.Headers = map[string]string{"RegionId": a.Config.ApsaraStackRegion}
		cancelImageCopyRequest.QueryParams = map[string]string{"AccessKeySecret": a.Config.ApsaraStackSecretKey, "Product": "ecs", "Department": a.Config.Department, "ResourceGroup": a.Config.ResourceGroup}

		cancelImageCopyRequest.RegionId = regionId
		cancelImageCopyRequest.ImageId = imageId
		if _, err := a.Client.CancelCopyImage(cancelImageCopyRequest); err != nil {
			errors = append(errors, err)
		}
	}

	for regionId, image := range sourceImage {
		imageId := image.ImageId
		log.Printf("Delete ApsaraStack image (%s) from region (%s)", imageId, regionId)

		errs := a.unsharedAccountsOnImages(regionId, imageId)
		if errs != nil {
			errors = append(errors, errs...)
		}

		deleteImageRequest := ecs.CreateDeleteImageRequest()
		if strings.ToLower(a.Config.Protocol) == "https" {
			deleteImageRequest.Scheme = "https"
		} else {
			deleteImageRequest.Scheme = "http"
		}
		deleteImageRequest.Headers = map[string]string{"RegionId": a.Config.ApsaraStackRegion}
		deleteImageRequest.QueryParams = map[string]string{"AccessKeySecret": a.Config.ApsaraStackSecretKey, "Product": "ecs", "Department": a.Config.Department, "ResourceGroup": a.Config.ResourceGroup}

		deleteImageRequest.RegionId = regionId
		deleteImageRequest.ImageId = imageId
		if _, err := a.Client.DeleteImage(deleteImageRequest); err != nil {
			errors = append(errors, err)
		}

		//Delete the snapshot of this images
		for _, diskDevices := range image.DiskDeviceMappings.DiskDeviceMapping {
			deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
			if strings.ToLower(a.Config.Protocol) == "https" {
				deleteSnapshotRequest.Scheme = "https"
			} else {
				deleteSnapshotRequest.Scheme = "http"
			}
			deleteSnapshotRequest.Headers = map[string]string{"RegionId": a.Config.ApsaraStackRegion}
			deleteSnapshotRequest.QueryParams = map[string]string{"AccessKeySecret": a.Config.ApsaraStackSecretKey, "Product": "ecs", "Department": a.Config.Department, "ResourceGroup": a.Config.ResourceGroup}

			deleteSnapshotRequest.SnapshotId = diskDevices.SnapshotId
			_, err := a.Client.DeleteSnapshot(deleteSnapshotRequest)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packer.MultiError{Errors: errors}
		}
	}

	return nil
}

func (a *Artifact) unsharedAccountsOnImages(regionId string, imageId string) []error {
	var errors []error
	//ig := state.Get("config").(*Config)
	describeImageShareRequest := ecs.CreateDescribeImageSharePermissionRequest()
	if strings.ToLower(a.Config.Protocol) == "https" {
		describeImageShareRequest.Scheme = "https"
	} else {
		describeImageShareRequest.Scheme = "http"
	}
	describeImageShareRequest.Headers = map[string]string{"RegionId": a.Config.ApsaraStackRegion}
	describeImageShareRequest.QueryParams = map[string]string{"AccessKeySecret": a.Config.ApsaraStackSecretKey, "Product": "ecs", "Department": a.Config.Department, "ResourceGroup": a.Config.ResourceGroup}

	describeImageShareRequest.RegionId = regionId
	describeImageShareRequest.ImageId = imageId
	imageShareResponse, err := a.Client.DescribeImageSharePermission(describeImageShareRequest)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	accountsNumber := len(imageShareResponse.Accounts.Account)
	if accountsNumber > 0 {
		accounts := make([]string, accountsNumber)
		for index, account := range imageShareResponse.Accounts.Account {
			accounts[index] = account.AliyunId
		}

		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
		if strings.ToLower(a.Config.Protocol) == "https" {
			modifyImageShareRequest.Scheme = "https"
		} else {
			modifyImageShareRequest.Scheme = "http"
		}
		modifyImageShareRequest.Headers = map[string]string{"RegionId": a.Config.ApsaraStackRegion}
		modifyImageShareRequest.QueryParams = map[string]string{"AccessKeySecret": a.Config.ApsaraStackSecretKey, "Product": "ecs", "Department": a.Config.Department, "ResourceGroup": a.Config.ResourceGroup}

		modifyImageShareRequest.RegionId = regionId
		modifyImageShareRequest.ImageId = imageId
		modifyImageShareRequest.RemoveAccount = &accounts
		_, err := a.Client.ModifyImageSharePermission(modifyImageShareRequest)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.ApsaraStackImages {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
