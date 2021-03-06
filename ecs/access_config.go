package ecs

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/endpoints"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/packer-builder-apsarastack/ascm"
	"log"
	"net/http"
	"strconv"
	"strings"

	//"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	//"github.com/aliyun/alibaba-cloud-sdk-go/sdk/endpoints"

	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/hashicorp/packer/version"
	"github.com/mitchellh/go-homedir"
)

// Config of ApsaraStack
type ApsaraStackAccessConfig struct {
	// ApsaraStack access key must be provided unless `profile` is set, but it can
	// also be sourced from the `APSARASTACK_ACCESS_KEY` environment variable.
	ApsaraStackAccessKey string `mapstructure:"access_key" required:"true"`
	// ApsaraStack secret key must be provided unless `profile` is set, but it can
	// also be sourced from the `APSARASTACK_SECRET_KEY` environment variable.
	ApsaraStackSecretKey string `mapstructure:"secret_key" required:"true"`
	// ApsaraStack region must be provided unless `profile` is set, but it can
	// also be sourced from the `APSARASTACK_REGION` environment variable.
	ApsaraStackRegion string `mapstructure:"region" required:"true"`
	ResourceSetName   string `mapstructure:"resource_group_set_name" required:"true"`
	// The region validation can be skipped if this value is true, the default
	// value is false.
	ApsaraStackSkipValidation bool `mapstructure:"skip_region_validation" required:"false"`
	// The image validation can be skipped if this value is true, the default
	// value is false.
	ApsaraStackSkipImageValidation bool `mapstructure:"skip_image_validation" required:"true"`
	// ApsaraStack profile must be set unless `access_key` is set; it can also be
	// sourced from the `APSARASTACK_PROFILE` environment variable.
	ApsaraStackProfile string `mapstructure:"profile" required:"false"`
	// ApsaraStack shared credentials file path. If this file exists, access and
	// secret keys will be read from this file.
	ApsaraStackSharedCredentialsFile string `mapstructure:"shared_credentials_file" required:"false"`
	Protocol                         string `mapstructure:"protocol" required:"false"`

	// STS access token, can be set through template or by exporting as
	// environment variable such as `export SECURITY_TOKEN=value`.
	SecurityToken string `mapstructure:"security_token" required:"false"`
	AS_Insecure   bool   `mapstructure:"insecure" required:"true"`
	Proxy         string `mapstructure:"proxy" required:"true"`
	Endpoint      string `mapstructure:"endpoint" required:"true"`
	Product       string `mapstructure:"product" required:"true"`
	Department    string `mapstructure:"department" required:"true"`
	ResourceGroup string `mapstructure:"resource_group" required:"true"`

	BootCommand []string `mapstructure:"boot_command" required:"false"`

	client *ClientWrapper
}

const Packer = "HashiCorp-Packer"
const DefaultRequestReadTimeout = 10 * time.Minute

// Client for ApsaraStackClient
func (c *ApsaraStackAccessConfig) Client() (*ClientWrapper, error) {
	if c.client != nil {
		return c.client, nil
	}
	if c.SecurityToken == "" {
		c.SecurityToken = os.Getenv("SECURITY_TOKEN")
	}

	var getProviderConfig = func(str string, key string) string {
		value, err := getConfigFromProfile(c, key)
		if err == nil && value != nil {
			str = value.(string)
		}
		return str
	}

	if c.ApsaraStackAccessKey == "" || c.ApsaraStackSecretKey == "" {
		c.ApsaraStackAccessKey = getProviderConfig(c.ApsaraStackAccessKey, "access_key_id")
		c.ApsaraStackSecretKey = getProviderConfig(c.ApsaraStackSecretKey, "access_key_secret")
		c.ApsaraStackRegion = getProviderConfig(c.ApsaraStackRegion, "region_id")
		c.SecurityToken = getProviderConfig(c.SecurityToken, "sts_token")
	}

	client, err := ecs.NewClientWithStsToken(c.ApsaraStackRegion, c.ApsaraStackAccessKey, c.ApsaraStackSecretKey, c.SecurityToken)
	if err != nil {
		return nil, err
	}
	/*
		client, err := ecs.NewClientWithAccessKey(c.ApsaraStackRegion,c.ApsaraStackAccessKey,c.ApsaraStackSecretKey)
			if err != nil {
				return nil, err
			}
		    client.Domain = "ecs.inter.env66.shuguang.com"
		   client.EndpointMap = map[string]string{c.ApsaraStackRegion:"http://ecs.inter.env66.shuguang.com"}
		client.SetHTTPSInsecure(c.AS_Insecure)
		if c.Proxy != "" {
			client.SetHttpsProxy(c.Proxy)
		}
		if c.client != nil {
			return c.client, nil
		}
		client.AppendUserAgent(Packer, version.FormattedVersion())
		client.SetReadTimeout(DefaultequestReadTimeout)
		c.client = &ClientWrapper{client}
		return c.client, nil*/
	//req.Headers = map[string]string{"RegionId": c.ApsaraStackRegion}
	//req.QueryParams = map[string]string{"AccessKeySecret": c.ApsaraStackSecretKey, "Product": "ecs", "Department": c.Department, "ResourceGroup": c.ResourceGroup}

	if c.client != nil {
		return c.client, nil
	}
	endpoints.AddEndpointMapping(c.ApsaraStackRegion, "ECS", c.Endpoint)
	//endpoints.AddEndpointMapping(c.ApsaraStackRegion,"OSS","oss-cn-qingdao-env66-d01-a.intra.env66.shuguang.com")
	//	client, err := ecs.NewClientWithAccessKey(c.ApsaraStackRegion,c.ApsaraStackAccessKey,c.ApsaraStackSecretKey)
	//	client, err := ecs.NewClientWithOptions(c.ApsaraStackRegion, c.getSdkConfig().WithTimeout(time.Duration(60)*time.Second), credentials.NewAccessKeyCredential(c.ApsaraStackRegion, c.ApsaraStackAccessKey))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize the ECS client: %#v", err)
	}
	client.Domain = c.Endpoint
	//client.Domain = "oss-cn-qingdao-env66-d01-a.intra.env66.shuguang.com"
	//c.OSS_Endpoint= "oss-cn-qingdao-env66-d01-a.intra.env66.shuguang.com"
	//c.Product = "ecs"

	client.SetHTTPSInsecure(true)
	if c.Proxy != "" {
		client.SetHttpProxy(c.Proxy)
	}
	/*if c.ResourceSetName == "" {
		return fmt.Errorf("errror while fetching resource group details, resource group set name can not be empty")
	}
	request := requests.NewCommonRequest()
	request.RegionId = c.ApsaraStackRegion
	request.Method = "GET"         // Set request method
	request.Product = "ascm"       // Specify product
	request.Domain = c.Endpoint      // Location Service will not be enabled if the host is specified. For example, service with a Certification type-Bearer Token should be specified
	request.Version = "2019-05-10" // Specify product version
	request.Scheme = "http"        // Set request scheme. Default: http
	request.ApiName = "ListResourceGroup"
	request.QueryParams = map[string]string{
		"AccessKeySecret": c.ApsaraStackSecretKey,
		"Product":         "ascm",
		//"Department":        config.Department,
		//"ResourceGroup":     config.ResourceGroup,
		"RegionId":          c.ApsaraStackRegion,
		"Action":            "ListResourceGroup",
		"Version":           "2019-05-10",
		"SignatureVersion":  "1.0",
		"resourceGroupName": c.ResourceSetName,
	}*/
	client.AppendUserAgent(Packer, version.FormattedVersion())
	client.SetReadTimeout(DefaultRequestReadTimeout)
	c.client = &ClientWrapper{client}
	return c.client, nil
}

func (c *ApsaraStackAccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.Config(); err != nil {
		errs = append(errs, err)
	}

	if c.ApsaraStackRegion == "" {
		c.ApsaraStackRegion = os.Getenv("APSARASTACK_REGION")
	}

	if c.ApsaraStackRegion == "" {
		errs = append(errs, fmt.Errorf("region option or APSARASTACK_REGION must be provided in template file or environment variables."))
	}
	/*if c.ResourceSetName == "" {
		errs = append(errs,fmt.Errorf("error while fetching resource group details, resource group set name can not be empty"))
	}*/

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *ApsaraStackAccessConfig) Config() error {
	if c.ApsaraStackAccessKey == "" {
		c.ApsaraStackAccessKey = os.Getenv("APSARASTACK_ACCESS_KEY")
	}
	if c.ApsaraStackSecretKey == "" {
		c.ApsaraStackSecretKey = os.Getenv("APSARASTACK_SECRET_KEY")
	}
	if c.ApsaraStackProfile == "" {
		c.ApsaraStackProfile = os.Getenv("APSARASTACK_PROFILE")
	}
	if c.ResourceSetName == "" {
		c.ResourceSetName = os.Getenv("RESOURCE_GROUP_SET_NAME")
	}
	if c.Department == "" || c.ResourceGroup == "" {
		dept, rg, err := getResourceCredentials(c)
		if err != nil {
			return err
		}
		c.Department = dept
		c.ResourceGroup = rg
	}
	if c.ApsaraStackSharedCredentialsFile == "" {
		c.ApsaraStackSharedCredentialsFile = os.Getenv("APSARASTACK_SHARED_CREDENTIALS_FILE")
	}
	if (c.ApsaraStackAccessKey == "" || c.ApsaraStackSecretKey == "") && c.ApsaraStackProfile == "" {
		return fmt.Errorf("APSARASTACK_ACCESS_KEY and APSARASTACK_SECRET_KEY must be set in template file or environment variables.")
	}
	return nil

}

func (c *ApsaraStackAccessConfig) ValidateRegion(region string) error {

	supportedRegions, err := c.getSupportedRegions()
	if err != nil {
		return err
	}

	for _, supportedRegion := range supportedRegions {
		if region == supportedRegion {
			return nil
		}
	}

	return fmt.Errorf("Not a valid ApsaraStack region: %s", region)
}

func (c *ApsaraStackAccessConfig) getSupportedRegions() ([]string, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	regionsRequest := ecs.CreateDescribeRegionsRequest()
	if strings.ToLower(c.Protocol) == "https" {
		regionsRequest.Scheme = "https"
	} else {
		regionsRequest.Scheme = "http"
	}
	regionsRequest.Headers = map[string]string{"RegionId": c.ApsaraStackRegion}
	regionsRequest.QueryParams = map[string]string{"AccessKeySecret": c.ApsaraStackSecretKey, "Product": "ecs", "Department": c.Department, "ResourceGroup": c.ResourceGroup}

	regionsResponse, err := client.DescribeRegions(regionsRequest)
	if err != nil {
		return nil, err
	}

	validRegions := make([]string, len(regionsResponse.Regions.Region))
	for _, valid := range regionsResponse.Regions.Region {
		validRegions = append(validRegions, valid.RegionId)
	}

	return validRegions, nil
}

func getConfigFromProfile(c *ApsaraStackAccessConfig, ProfileKey string) (interface{}, error) {
	providerConfig := make(map[string]interface{})
	current := c.ApsaraStackProfile
	if current != "" {
		profilePath, err := homedir.Expand(c.ApsaraStackSharedCredentialsFile)
		if err != nil {
			return nil, err
		}
		if profilePath == "" {
			profilePath = fmt.Sprintf("%s/.aliyun/config.json", os.Getenv("HOME"))
			if runtime.GOOS == "windows" {
				profilePath = fmt.Sprintf("%s/.aliyun/config.json", os.Getenv("USERPROFILE"))
			}
		}
		_, err = os.Stat(profilePath)
		if !os.IsNotExist(err) {
			data, err := ioutil.ReadFile(profilePath)
			if err != nil {
				return nil, err
			}
			config := map[string]interface{}{}
			err = json.Unmarshal(data, &config)
			if err != nil {
				return nil, err
			}
			for _, v := range config["profiles"].([]interface{}) {
				if current == v.(map[string]interface{})["name"] {
					providerConfig = v.(map[string]interface{})
				}
			}
		}
	}
	mode := ""
	if v, ok := providerConfig["mode"]; ok {
		mode = v.(string)
	} else {
		return v, nil
	}
	switch ProfileKey {
	case "access_key_id", "access_key_secret":
		if mode == "EcsRamRole" {
			return "", nil
		}
	case "ram_role_name":
		if mode != "EcsRamRole" {
			return "", nil
		}
	case "sts_token":
		if mode != "StsToken" {
			return "", nil
		}
	case "ram_role_arn", "ram_session_name":
		if mode != "RamRoleArn" {
			return "", nil
		}
	case "expired_seconds":
		if mode != "RamRoleArn" {
			return float64(0), nil
		}
	}
	return providerConfig[ProfileKey], nil
}
func (c *ApsaraStackAccessConfig) getSdkConfig() *sdk.Config {
	return sdk.NewConfig().
		WithMaxRetryTime(5).
		WithTimeout(time.Duration(30) * time.Minute).
		WithEnableAsync(true).
		WithGoRoutinePoolSize(100).
		WithMaxTaskQueueSize(10000).
		WithDebug(false).
		WithHttpTransport(c.getTransport()).
		WithScheme("http")

}

func (c *ApsaraStackAccessConfig) getTransport() *http.Transport {
	handshakeTimeout, err := strconv.Atoi(os.Getenv("TLSHandshakeTimeout"))
	if err != nil {
		handshakeTimeout = 120
	}
	transport := &http.Transport{}
	transport.TLSHandshakeTimeout = time.Duration(handshakeTimeout) * time.Second
	return transport
}

func getResourceCredentials(config *ApsaraStackAccessConfig) (string, string, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		return "", "", fmt.Errorf("unable to initialize the ascm client: endpoint or domain is not provided for ascm service")
	}
	if endpoint != "" {
		//endpoints.AddEndpointMapping(config.ApsaraStackRegion, string(connectivity.ASCMCode), endpoint)
		endpoints.AddEndpointMapping(config.ApsaraStackRegion, "ECS", config.Endpoint)
	}
	ascmClient, err := sdk.NewClientWithAccessKey(config.ApsaraStackRegion, config.ApsaraStackAccessKey, config.ApsaraStackSecretKey)
	if err != nil {
		return "", "", fmt.Errorf("unable to initialize the ascm client: %#v", err)
	}
	ascmClient.AppendUserAgent(Packer, version.FormattedVersion())
	ascmClient.SetReadTimeout(DefaultRequestReadTimeout)
	/*
		ascmClient.AppendUserAgent(connectivity.Terraform, connectivity.TerraformVersion)
		ascmClient.AppendUserAgent(connectivity.Provider, connectivity.ProviderVersion)
		ascmClient.AppendUserAgent(connectivity.Module, config.ConfigurationSource)*/
	ascmClient.SetHTTPSInsecure(config.AS_Insecure)
	ascmClient.Domain = endpoint
	if config.Proxy != "" {
		ascmClient.SetHttpProxy(config.Proxy)
	}
	if config.ResourceSetName == "" {
		return "", "", fmt.Errorf("errror while fetching resource group details, resource group set name can not be empty")
	}
	request := requests.NewCommonRequest()
	request.RegionId = config.ApsaraStackRegion
	request.Method = "GET"         // Set request method
	request.Product = "ascm"       // Specify product
	request.Domain = endpoint      // Location Service will not be enabled if the host is specified. For example, service with a Certification type-Bearer Token should be specified
	request.Version = "2019-05-10" // Specify product version

	//request.Scheme = "http"        // Set request scheme. Default: http
	if strings.ToLower(config.Protocol) == "https" {
		request.Scheme = "https"
	} else {
		request.Scheme = "http"
	}
	request.SetHTTPSInsecure(true)
	request.ApiName = "ListResourceGroup"
	request.Headers = map[string]string{"RegionId": config.ApsaraStackRegion}
	request.QueryParams = map[string]string{
		"AccessKeySecret": config.ApsaraStackSecretKey,
		"Product":         "ascm",
		//"Department":        config.Department,
		//"ResourceGroup":     config.ResourceGroup,
		"RegionId":          config.ApsaraStackRegion,
		"Action":            "ListResourceGroup",
		"Version":           "2019-05-10",
		"SignatureVersion":  "1.0",
		"resourceGroupName": config.ResourceSetName,
	}
	resp := responses.BaseResponse{}
	request.TransToAcsRequest()
	err = ascmClient.DoAction(request, &resp)
	if err != nil {
		return "", "", err
	}
	response := &ascm.ResourceGroup{}
	err = json.Unmarshal(resp.GetHttpContentBytes(), response)

	if len(response.Data) != 1 || response.Code != "200" {
		if len(response.Data) == 0 {
			return "", "", fmt.Errorf("resource group ID and organization not found for resource set %s", config.ResourceSetName)
		}
		return "", "", fmt.Errorf("unable to initialize the ascm client: department or resource_group is not provided")
	}

	log.Printf("[INFO] Get Resource Group Details Succssfull for Resource set: %s : Department: %s, ResourceGroupId: %s", config.ResourceSetName, fmt.Sprint(response.Data[0].OrganizationID), fmt.Sprint(response.Data[0].ResourceGroupID))
	return fmt.Sprint(response.Data[0].OrganizationID), fmt.Sprint(response.Data[0].ResourceGroupID), err
}
