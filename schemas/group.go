package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Group struct {
	Id   string
	Name string
}

func LoadGroup(d *schema.ResourceData) (sch *Group) {
	sch = &Group{
		Id:   d.Id(),
		Name: d.Get("name").(string),
	}

	return
}
