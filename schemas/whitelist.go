package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Whitelist struct {
	Id      string
	GroupId string
	Address string
}

func LoadWhitelist(d *schema.ResourceData) (sch *Whitelist) {
	sch = &Whitelist{
		Id:      d.Id(),
		GroupId: d.Get("group_id").(string),
		Address: d.Get("address").(string),
	}

	return
}
