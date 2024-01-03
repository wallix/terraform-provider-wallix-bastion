package main

import (
	"github.com/wallix/terraform-provider-wallix-bastion/bastion"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: bastion.Provider,
	})
}
