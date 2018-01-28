package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pritunl/terraform-provider-mongodbatlas/constants"
	"github.com/pritunl/terraform-provider-mongodbatlas/digest"
	"github.com/pritunl/terraform-provider-mongodbatlas/errortypes"
	"github.com/pritunl/terraform-provider-mongodbatlas/schemas"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Cluster() *schema.Resource {
	return &schema.Resource{
		Create: clusterCreate,
		Read:   clusterRead,
		Update: clusterUpdate,
		Delete: clusterDelete,
		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"service_provider": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "AWS",
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "us-east-2",
			},
			"size": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "M10",
			},
			"disk_size_gb": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"replication_factor": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"mongodb_version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3.6",
			},
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"atlas_vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"atlas_cidr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type clusterAutoScalingData struct {
	DiskGbEnabled bool `json:"diskGBEnabled"`
}

type clusterProviderData struct {
	ProviderName     string `json:"providerName"`
	RegionName       string `json:"regionName"`
	InstanceSizeName string `json:"instanceSizeName"`
}

type clusterPostData struct {
	AutoScaling         clusterAutoScalingData `json:"autoScaling"`
	Name                string                 `json:"name"`
	MongoDbMajorVersion string                 `json:"mongoDBMajorVersion"`
	ReplicationFactor   int                    `json:"replicationFactor"`
	BackupEnabled       bool                   `json:"backupEnabled"`
	DiskSizeGb          int                    `json:"diskSizeGB"`
	ProviderSettings    clusterProviderData    `json:"providerSettings"`
}

type clusterPutData struct {
	AutoScaling         clusterAutoScalingData `json:"autoScaling"`
	MongoDbMajorVersion string                 `json:"mongoDBMajorVersion"`
	ReplicationFactor   int                    `json:"replicationFactor"`
	BackupEnabled       bool                   `json:"backupEnabled"`
	DiskSizeGb          int                    `json:"diskSizeGB"`
	ProviderSettings    clusterProviderData    `json:"providerSettings"`
}

type clusterData struct {
	Id                  string              `json:"id"`
	Name                string              `json:"name"`
	GroupId             string              `json:"groupId"`
	StateName           string              `json:"stateName"`
	MongoUri            string              `json:"mongoURI"`
	MongoUriWithOptions string              `json:"mongoURIWithOptions"`
	MongoDbMajorVersion string              `json:"mongoDBMajorVersion"`
	Paused              bool                `json:"paused"`
	ProviderSettings    clusterProviderData `json:"providerSettings"`
}

type containerData struct {
	Id             string `json:"id"`
	ProviderName   string `json:"providerName"`
	RegionName     string `json:"regionName"`
	VpcId          string `json:"vpcId"`
	AtlasCidrBlock string `json:"atlasCidrBlock"`
	Provisioned    bool   `json:"provisioned"`
}

type containerResp struct {
	Results []*containerData `json:"results"`
}

func (c *clusterData) Available() bool {
	switch c.StateName {
	case "IDLE", "REPAIRING":
		return true
	default:
		return false
	}
}

func (c *clusterData) Updating() bool {
	switch c.StateName {
	case "UPDATING", "REPAIRING":
		return true
	default:
		return false
	}
}

func containerGet(prvdr *schemas.Provider, clst *schemas.Cluster) (
	container *containerData, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/containers",
			clst.GroupId,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Containers request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Containers request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBodyStr := ""
		respBody, _ := ioutil.ReadAll(resp.Body)
		if respBody != nil {
			respBodyStr = string(respBody)
		}

		err = &errortypes.RequestError{
			errors.Wrapf(
				err,
				"resources: Containers request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data := &containerResp{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Containers decode failed"),
		}
		return
	}

	if data.Results == nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Containers response empty"),
		}
		return
	}

	region := strings.Replace(strings.ToUpper(clst.Region), "-", "_", -1)

	for _, cntr := range data.Results {
		if cntr.Provisioned && region == cntr.RegionName {
			container = cntr
			return
		}
	}

	return
}

func clusterGet(prvdr *schemas.Provider, groupId, name string) (
	data *clusterData, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/clusters/%s",
			groupId,
			name,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Clusters request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Clusters request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return
	} else if resp.StatusCode != 200 {
		respBodyStr := ""
		respBody, _ := ioutil.ReadAll(resp.Body)
		if respBody != nil {
			respBodyStr = string(respBody)
		}

		err = &errortypes.RequestError{
			errors.Wrapf(
				err,
				"resources: Clusters request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &clusterData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Clusters decode failed"),
		}
		return
	}

	return
}

func clusterPost(prvdr *schemas.Provider, clst *schemas.Cluster) (err error) {
	region := strings.Replace(strings.ToUpper(clst.Region), "-", "_", -1)

	data := clusterPostData{
		AutoScaling: clusterAutoScalingData{
			DiskGbEnabled: true,
		},
		Name:                clst.Name,
		MongoDbMajorVersion: clst.MongoDbVersion,
		ReplicationFactor:   clst.ReplicationFactor,
		BackupEnabled:       true,
		DiskSizeGb:          clst.DiskSizeGb,
		ProviderSettings: clusterProviderData{
			ProviderName:     strings.ToUpper(clst.ServiceProvider),
			RegionName:       region,
			InstanceSizeName: strings.ToUpper(clst.Size),
		},
	}

	body, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Cluster marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"POST",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/clusters",
			clst.GroupId,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Clusters request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Clusters request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		respBodyStr := ""
		respBody, _ := ioutil.ReadAll(resp.Body)
		if respBody != nil {
			respBodyStr = string(respBody)
		}

		err = &errortypes.RequestError{
			errors.Wrapf(
				err,
				"resources: Clusters request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	return
}

func clusterPut(prvdr *schemas.Provider, clst *schemas.Cluster) (
	data *clusterData, err error) {

	region := strings.Replace(strings.ToUpper(clst.Region), "-", "_", -1)

	putData := clusterPutData{
		AutoScaling: clusterAutoScalingData{
			DiskGbEnabled: true,
		},
		MongoDbMajorVersion: clst.MongoDbVersion,
		ReplicationFactor:   clst.ReplicationFactor,
		BackupEnabled:       true,
		DiskSizeGb:          clst.DiskSizeGb,
		ProviderSettings: clusterProviderData{
			ProviderName:     strings.ToUpper(clst.ServiceProvider),
			RegionName:       region,
			InstanceSizeName: strings.ToUpper(clst.Size),
		},
	}

	body, err := json.Marshal(putData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Cluster marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"PATCH",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/clusters/%s",
			clst.GroupId,
			clst.Name,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Clusters request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Clusters request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return
	} else if resp.StatusCode != 200 {
		respBodyStr := ""
		respBody, _ := ioutil.ReadAll(resp.Body)
		if respBody != nil {
			respBodyStr = string(respBody)
		}

		err = &errortypes.RequestError{
			errors.Wrapf(
				err,
				"resources: Clusters request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &clusterData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Clusters decode failed"),
		}
		return
	}

	return
}

func clusterCreate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	clst := schemas.LoadCluster(d)

	var clstData *clusterData
	for {
		clstData, err = clusterGet(prvdr, clst.GroupId, clst.Name)
		if err != nil {
			return
		}

		if clstData == nil {
			err = clusterPost(prvdr, clst)
			if err != nil {
				return
			}
		} else if clstData.Available() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	cntr, err := containerGet(prvdr, clst)
	if err != nil {
		return
	}

	if cntr == nil {
		err = errortypes.NotFoundError{
			errors.New("resources: Container not found"),
		}
		return
	}

	d.Set("container_id", cntr.Id)
	d.Set("atlas_vpc_id", cntr.VpcId)
	d.Set("atlas_cidr", cntr.AtlasCidrBlock)
	d.SetId(clst.Name)

	return
}

func clusterRead(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	clst := schemas.LoadCluster(d)

	clstData, err := clusterGet(prvdr, clst.GroupId, clst.Name)
	if err != nil {
		return
	}

	if clstData == nil {
		d.SetId("")
		return
	}

	cntr, err := containerGet(prvdr, clst)
	if err != nil {
		return
	}

	if cntr == nil {
		err = errortypes.NotFoundError{
			errors.New("resources: Container not found"),
		}
		return
	}

	d.Set("container_id", cntr.Id)
	d.Set("atlas_vpc_id", cntr.VpcId)
	d.Set("atlas_cidr", cntr.AtlasCidrBlock)
	d.SetId(clst.Name)

	return
}

func clusterUpdate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	clst := schemas.LoadCluster(d)

	clstData, err := clusterPut(prvdr, clst)
	if err != nil {
		return
	}

	if clstData == nil {
		d.SetId("")
		return
	}

	for {
		clstData, err = clusterGet(prvdr, clst.GroupId, clst.Name)
		if err != nil {
			return
		}

		if clstData == nil {
			d.SetId("")
			return
		} else if clstData.Available() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	cntr, err := containerGet(prvdr, clst)
	if err != nil {
		return
	}

	if cntr == nil {
		err = errortypes.NotFoundError{
			errors.New("resources: Container not found"),
		}
		return
	}

	d.Set("container_id", cntr.Id)
	d.Set("atlas_vpc_id", cntr.VpcId)
	d.Set("atlas_cidr", cntr.AtlasCidrBlock)
	d.SetId(clst.Name)

	return
}

func clusterDelete(d *schema.ResourceData, m interface{}) (err error) {
	d.SetId("")
	return
}
