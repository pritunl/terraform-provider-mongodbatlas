package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pritunl/terraform-provider-mongodbatlas/resources"
	"github.com/pritunl/terraform-provider-mongodbatlas/schemas"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ConfigureFunc: providerConfigure,
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"api_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"mongodbatlas_group":     resources.Group(),
			"mongodbatlas_cluster":   resources.Cluster(),
			"mongodbatlas_user":      resources.User(),
			"mongodbatlas_peer":      resources.Peer(),
			"mongodbatlas_whitelist": resources.Whitelist(),
		},
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return schemas.LoadProvider(d), nil
}
