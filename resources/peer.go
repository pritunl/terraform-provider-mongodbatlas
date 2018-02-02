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
	"time"
)

func Peer() *schema.Resource {
	return &schema.Resource{
		Create: peerCreate,
		Read:   peerRead,
		Update: peerUpdate,
		Delete: peerDelete,
		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"aws_account_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_cidr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"connection_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type peerPostData struct {
	VpcId               string `json:"vpcId"`
	AwsAccountId        string `json:"awsAccountId"`
	RouteTableCidrBlock string `json:"routeTableCidrBlock"`
	ContainerId         string `json:"containerId"`
}

type peerPutData struct {
	VpcId               string `json:"vpcId"`
	AwsAccountId        string `json:"awsAccountId"`
	RouteTableCidrBlock string `json:"routeTableCidrBlock"`
	ContainerId         string `json:"containerId"`
}

type peerData struct {
	Id                  string `json:"id"`
	VpcId               string `json:"vpcId"`
	AwsAccountId        string `json:"awsAccountId"`
	ConnectionId        string `json:"connectionId"`
	RouteTableCidrBlock string `json:"routeTableCidrBlock"`
	ContainerId         string `json:"containerId"`
	StatusName          string `json:"statusName"`
	ErrorStateName      string `json:"errorStateName"`
}

type peerResp struct {
	Results []*peerData `json:"results"`
}

func (p *peerData) Available() bool {
	switch p.StatusName {
	case "PENDING_ACCEPTANCE", "FINALIZING", "AVAILABLE":
		return true
	default:
		return false
	}
}

func (p *peerData) Failed() bool {
	switch p.StatusName {
	case "FAILED", "TERMINATING":
		return true
	}

	if p.ErrorStateName != "" {
		return true
	}

	return false
}

func peerFind(prvdr *schemas.Provider, pr *schemas.Peer) (
	data *peerData, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/peers",
			pr.GroupId,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
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
				"resources: Peer request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	respData := &peerResp{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Peer decode failed"),
		}
		return
	}

	if respData.Results == nil {
		return
	}

	for _, p := range respData.Results {
		if p.VpcId == pr.VpcId {
			data = p
			return
		}
	}

	return
}

func peerGet(prvdr *schemas.Provider, pr *schemas.Peer) (
	data *peerData, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/peers/%s",
			pr.GroupId,
			pr.Id,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
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
				"resources: Peer request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &peerData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Peer decode failed"),
		}
		return
	}

	return
}

func peerPost(prvdr *schemas.Provider, pr *schemas.Peer) (
	data *peerData, err error) {

	postData := peerPostData{
		VpcId:               pr.VpcId,
		AwsAccountId:        pr.AwsAccountId,
		RouteTableCidrBlock: pr.VpcCidr,
		ContainerId:         pr.ContainerId,
	}

	body, err := json.Marshal(postData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Peer marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"POST",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/peers",
			pr.GroupId,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
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
				"resources: Peer request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &peerData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Peer decode failed"),
		}
		return
	}

	return
}

func peerPut(prvdr *schemas.Provider, pr *schemas.Peer) (
	data *peerData, err error) {

	putData := peerPutData{
		VpcId:               pr.VpcId,
		AwsAccountId:        pr.AwsAccountId,
		RouteTableCidrBlock: pr.VpcCidr,
	}

	body, err := json.Marshal(putData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Peer marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"PATCH",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/peers/%s",
			pr.GroupId,
			pr.Id,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
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
				"resources: Peer request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &peerData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: Peer decode failed"),
		}
		return
	}

	return
}

func peerDel(prvdr *schemas.Provider, pr *schemas.Peer) (
	err error) {

	req, err := http.NewRequest(
		"DELETE",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/peers/%s",
			pr.GroupId,
			pr.Id,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: Peer request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return
	} else if resp.StatusCode != 200 && resp.StatusCode != 202 {
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

	return
}

func peerCreate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	pr := schemas.LoadPeer(d)

	prData, err := peerFind(prvdr, pr)
	if err != nil {
		return
	}

	if prData == nil {
		prData, err = peerPost(prvdr, pr)
		if err != nil {
			return
		}
	}

	pr.Id = prData.Id

	for {
		prData, err = peerGet(prvdr, pr)
		if err != nil {
			return
		}

		if prData.Available() {
			break
		}

		if prData.Failed() {
			err = &errortypes.RequestError{
				errors.Wrap(err, "resources: Peer in failed state"),
			}
			return
		}

		time.Sleep(1 * time.Second)
	}

	d.Set("connection_id", prData.ConnectionId)
	d.SetId(prData.Id)

	return
}

func peerRead(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	pr := schemas.LoadPeer(d)

	if pr.Id != "" {
		prData, e := peerGet(prvdr, pr)
		if e != nil {
			err = e
			return
		}

		if prData != nil {
			if prData.Failed() {
				err = peerDel(prvdr, pr)
				if err != nil {
					return
				}
				d.SetId("")
			}

			d.Set("connection_id", prData.ConnectionId)
			d.SetId(prData.Id)
			return
		}
	}

	d.SetId("")

	return
}

func peerUpdate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	pr := schemas.LoadPeer(d)

	if pr.Id != "" {
		prData, e := peerPut(prvdr, pr)
		if e != nil {
			err = e
			return
		}

		if prData != nil {
			if prData.Failed() {
				err = &errortypes.RequestError{
					errors.Wrap(err, "resources: Peer in failed state"),
				}
				return
			}

			d.Set("connection_id", prData.ConnectionId)
			d.SetId(prData.Id)
			return
		}
	}

	d.SetId("")

	return
}

func peerDelete(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	pr := schemas.LoadPeer(d)

	if pr.Id != "" {
		err = peerDel(prvdr, pr)
		if err != nil {
			return
		}
	}

	d.SetId("")

	return
}
