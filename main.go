package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

func main() {
	server, err := bastion.NewProviderServer(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// plugin.Serve keeps the log setup of the SDKv2 provider and serves protocol 5.
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: server,
	})
}
