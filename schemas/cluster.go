package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Cluster struct {
	GroupId           string
	Name              string
	ServiceProvider   string
	Region            string
	Size              string
	DiskSizeGb        int
	ReplicationFactor int
	MongoDbVersion    string
}

func LoadCluster(d *schema.ResourceData) (sch *Cluster) {
	sch = &Cluster{
		GroupId:           d.Get("group_id").(string),
		Name:              d.Get("name").(string),
		ServiceProvider:   d.Get("service_provider").(string),
		Region:            d.Get("region").(string),
		Size:              d.Get("size").(string),
		DiskSizeGb:        d.Get("disk_size_gb").(int),
		ReplicationFactor: d.Get("replication_factor").(int),
		MongoDbVersion:    d.Get("mongodb_version").(string),
	}

	return
}
