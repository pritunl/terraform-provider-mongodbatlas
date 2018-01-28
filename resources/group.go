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
)

func Group() *schema.Resource {
	return &schema.Resource{
		Create: groupCreate,
		Read:   groupRead,
		Delete: groupDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

type groupPostData struct {
	Name  string `json:"name"`
	OrgId string `json:"orgId"`
}

type groupData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func groupGet(prvdr *schemas.Provider, name string) (
	data *groupData, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/byName/%s",
			name,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Group request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Group request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode == 401 {
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
				"resources: Group request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &groupData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Group decode failed"),
		}
		return
	}

	return
}

func groupPost(prvdr *schemas.Provider, grp *schemas.Group) (
	data *groupData, err error) {

	postData := groupPostData{
		Name:  grp.Name,
		OrgId: prvdr.OrgId,
	}

	body, err := json.Marshal(postData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Group marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"POST",
		constants.BaseUrl+"/api/atlas/v1.0/groups",
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Groups request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Groups request failed"),
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
				"resources: Groups request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &groupData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Group decode failed"),
		}
		return
	}

	return
}

func groupCreate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	grp := schemas.LoadGroup(d)

	grpData, err := groupGet(prvdr, grp.Name)
	if err != nil {
		return
	}

	if grpData == nil {
		grpData, err = groupPost(prvdr, grp)
		if err != nil {
			return
		}
	}

	d.SetId(grpData.Id)

	return
}

func groupRead(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	clst := schemas.LoadGroup(d)

	grpData, err := groupGet(prvdr, clst.Name)
	if err != nil {
		return
	}

	if grpData == nil {
		d.SetId("")
		return
	}

	d.SetId(grpData.Id)

	return
}

func groupDelete(d *schema.ResourceData, m interface{}) (err error) {
	d.SetId("")
	return
}
