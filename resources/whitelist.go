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
	"net/url"
)

func Whitelist() *schema.Resource {
	return &schema.Resource{
		Create: whitelistCreate,
		Read:   whitelistRead,
		Delete: whitelistDelete,
		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

type whitelistPostData struct {
	CidrBlock string `json:"cidrBlock"`
	Comment   string `json:"comment"`
}

func whitelistGet(prvdr *schemas.Provider, wl *schemas.Whitelist) (
	exists bool, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/whitelist/%s",
			wl.GroupId,
			url.QueryEscape(wl.Address),
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Whitelist request failed"),
		}
		return
	}

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Whitelist request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		exists = false
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
				"resources: Peer request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	exists = true

	return
}

func whitelistPost(prvdr *schemas.Provider, wl *schemas.Whitelist) (
	err error) {

	postData := []*whitelistPostData{}
	postData = append(postData, &whitelistPostData{
		CidrBlock: wl.Address,
	})

	body, err := json.Marshal(postData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Whitelist marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"POST",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/whitelist",
			wl.GroupId,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Whitelist request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Whitelist request failed"),
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
				"resources: Whitelist request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	return
}

func whitelistDel(prvdr *schemas.Provider, wl *schemas.Whitelist) (
	err error) {

	req, err := http.NewRequest(
		"DELETE",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/whitelist/%s",
			wl.GroupId,
			url.QueryEscape(wl.Address),
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Whitelist request failed"),
		}
		return
	}

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Whitelist request failed"),
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
				"resources: Whitelist request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	return
}

func whitelistCreate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	wl := schemas.LoadWhitelist(d)

	err = whitelistPost(prvdr, wl)
	if err != nil {
		return
	}

	d.SetId(wl.Address)

	return
}

func whitelistRead(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	wl := schemas.LoadWhitelist(d)

	exists, err := whitelistGet(prvdr, wl)
	if err != nil {
		return
	}

	if exists {
		d.SetId(wl.Address)
	} else {
		d.SetId("")
	}

	return
}

func whitelistDelete(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	wl := schemas.LoadWhitelist(d)

	err = whitelistDel(prvdr, wl)
	if err != nil {
		return
	}

	d.SetId("")

	return
}
