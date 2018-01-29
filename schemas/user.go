package schemas

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type User struct {
	Id           string
	GroupId      string
	Name         string
	ClusterName  string
	DatabaseName string
	Password     string
	MongoDbUri   string
}

func LoadUser(d *schema.ResourceData) (sch *User) {
	sch = &User{
		Id:           d.Id(),
		GroupId:      d.Get("group_id").(string),
		Name:         d.Get("name").(string),
		ClusterName:  d.Get("cluster_name").(string),
		DatabaseName: d.Get("database_name").(string),
		Password:     d.Get("password").(string),
		MongoDbUri:   d.Get("mongodb_uri").(string),
	}

	return
}
