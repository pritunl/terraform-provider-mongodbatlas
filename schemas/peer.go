package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Peer struct {
	Id           string
	ContainerId  string
	GroupId      string
	AwsAccountId string
	VpcId        string
	VpcCidr      string
}

func LoadPeer(d *schema.ResourceData) (sch *Peer) {
	sch = &Peer{
		Id:           d.Id(),
		ContainerId:  d.Get("container_id").(string),
		GroupId:      d.Get("group_id").(string),
		AwsAccountId: d.Get("aws_account_id").(string),
		VpcId:        d.Get("vpc_id").(string),
		VpcCidr:      d.Get("vpc_cidr").(string),
	}

	return
}
