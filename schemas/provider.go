package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Provider struct {
	Username string
	ApiKey   string
	OrgId    string
}

func LoadProvider(d *schema.ResourceData) (sch *Provider) {
	sch = &Provider{
		Username: d.Get("username").(string),
		ApiKey:   d.Get("api_key").(string),
		OrgId:    d.Get("org_id").(string),
	}

	return
}
