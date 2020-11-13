package main

//Register apsarastack packer builder plugin
import (
	"github.com/aliyun/packer-builder-apsarastack/ecs"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(ecs.Builder))
	server.Serve()
}

