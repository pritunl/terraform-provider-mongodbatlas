package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pritunl/terraform-provider-mongodbatlas/logger"
	"github.com/pritunl/terraform-provider-mongodbatlas/provider"
)

func main() {
	logger.Init()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return provider.Provider()
		},
	})
}
