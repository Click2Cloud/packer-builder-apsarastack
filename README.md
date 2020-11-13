# Apsarastack Cloud packer provider 

This is the official repository for the ApsaraStack packer builder.  

If you are not planning to contribute to this repo, you can download the [compiled binaries](https://github.com/aliyun/packer-builder-apsarastack/releases) according to you platform, unzip and move 
them into the folder under the packer **PATH** such as **/usr/local/packer**.

## Install
- Download the correct packer from you platform from https://www.packer.io/downloads.html
- Install packer according to the guide from https://www.packer.io/docs/installation.html
- Install Go according to the guide from [https://golang.org/doc/install](https://golang.org/doc/install)
- Setup your access key and secret key in the environment variables according to platform, for example In Linux platform with default bash, open your .bashrc in your home directory and add following two lines<p>
    ```aidl
        export APSARASTACK_ACCESS_KEY="access key value"
        
        export APSARASTACK_SECRET_KEY="secret key value"
     ```
- Open a terminator and clone ApsaraStack packer builder and build,install and test<p>
  ```
  cd <$GOPATH>
  
  mkdir -p src/github.com/aliyun/
  
  cd <$GOPATH>/src/github.com/aliyun/
  
  git clone https://github.com/aliyun/packer-builder-apsarastack
  
  
  cd <$GOPTH>/src/github.com/aliyun/packer-builder-apsarastack
    
  make all
  
  sorce ~/.bashrc
  
  packer build examples/apsarastack.json
  ```
 If output similar as following, configurations, you can now start the journey of apsarastack with packer support
 ```
    apsarastack output will be in this color.
    
    ==> apsarastack: Force delete flag found, skipping prevalidating ApsaraStack ECS Image Name
        apsarastack: Found Image ID: centos_7_03_64_20G_alibase_20170818.vhd
    ==> apsarastack: allocated eip address 121.196.193.14
    ==> apsarastack: Instance starting
 
```
## Example
### Create a simple image with redis installed
```
{
  "variables": {
    "access_key": "{{env `APSARASTACK_ACCESS_KEY`}}",
    "secret_key": "{{env `APSARASTACK_SECRET_KEY`}}"
  },
  "builders": [
    {
      "type": "apsarastack",
      "access_key": "{{user `access_key`}}",
      "secret_key": "{{user `secret_key`}}",
      "region": "cn-wulan-env82-d01",
      "insecure": true,
      "proxy": "http://100.67.27.224:58201",
      "endpoint": "server.asapi.cn-wulan-env82-d01.intra.env17e.shuguang.com/asapi/v3",
      "department": "11",
      "resource_group": "27",
      "image_name": "packer_basicudvx",
      "instance_type": "ecs.se1.large",
      "source_image": "centos_7_03_64_20G_alibase_20170818.vhd",
      "io_optimized": "true",
      "communicator": "none",
      "user_data" :"yum install redis.x86_64 -y"

    }
  ]
}
 
```
### Create a simple image for windows
```aidl
{
  "variables": {
    "access_key": "{{env `APSARASTACK_ACCESS_KEY`}}",
    "secret_key": "{{env `APSARASTACK_SECRET_KEY`}}"
  },
  "builders": [{
    "type": "apsarastack",
    "access_key": "{{user `access_key`}}",
    "secret_key": "{{user `secret_key`}}",
    "region": "cn-wulan-env82-d01",
    "insecure": true,
    "proxy": "http://100.67.27.224:58201",
    "endpoint": "server.asapi.cn-wulan-env82-d01.intra.env17e.shuguang.com/asapi/v3",
    "department": "11",
    "resource_group": "27",
    "image_name":"packer_test",
    "source_image":"win2012r2_9600_x64_dtc_en-us_40G_alibase_20200314.vhd",
    "instance_type":"ecs.xn4.small",
    "io_optimized":"true",
    "communicator": "none",
    "user_data_file": "examples/apsarastack/basic/winrm_enable_userdata.ps1"
  }]
}
```
### Create a simple image with mounted disk
```
{
  "variables": {
    "access_key": "{{env `APSARASTACK_ACCESS_KEY`}}",
    "secret_key": "{{env `APSARASTACK_SECRET_KEY`}}"
  },
  "builders": [{
    "type":"apsarastack-ecs",
    "access_key":"{{user `access_key`}}",
    "secret_key":"{{user `secret_key`}}",
    "region": "cn-wulan-env82-d01",
    "insecure": true,
    "proxy": "http://100.67.27.224:58201",
    "endpoint": "server.asapi.cn-wulan-env82-d01.intra.env17e.shuguang.com/asapi/v3",
    "department": "11",
    "resource_group": "27",
    "image_name":"packer_with_data_disk",
    "source_image":"centos_6_08_32_40G_alibase_20170710.raw",
    "communicator": "none",
    "instance_type":"ecs.e4.small",
    "io_optimized":"true",
    "image_disk_mappings":[{"disk_name":"data1","disk_size":20},{"disk_name":"data1","disk_size":20,"disk_device":"/dev/xvdz"}]
  }]
}
```

### Here are [more examples](https://github.com/aliyun/packer-builder-apsarastack/tree/master/examples/apsarastack) include chef, jenkins image template etc.

## 
### How to contribute code
* If you are not sure or have any doubts, feel free to ask and/or submit an issue or PR. We appreciate all contributions and don't want to create artificial obstacles that get in the way.
* Contributions are welcome and will be merged via PRs.

### Contributors

### Refrence
* Pakcer document: https://www.packer.io/intro/

