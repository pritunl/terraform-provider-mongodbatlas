package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Group struct {
	Name string
}

func LoadGroup(d *schema.ResourceData) (sch *Group) {
	sch = &Group{
		Name: d.Get("name").(string),
	}

	return
}
