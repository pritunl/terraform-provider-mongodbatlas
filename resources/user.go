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
	"time"
)

func User() *schema.Resource {
	return &schema.Resource{
		Create: userCreate,
		Read:   userRead,
		Update: userUpdate,
		Delete: userDelete,
		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"database_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"mongodb_uri": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type userRoleData struct {
	DatabaseName   string `json:"databaseName"`
	CollectionName string `json:"collectionName"`
	RoleName       string `json:"roleName"`
}

type userData struct {
	DatabaseName string         `json:"databaseName"`
	Username     string         `json:"username"`
	GroupId      string         `json:"groupId"`
	Roles        []userRoleData `json:"roles"`
}

type userPostData struct {
	DatabaseName string         `json:"databaseName"`
	Username     string         `json:"username"`
	Password     string         `json:"password"`
	GroupId      string         `json:"groupId"`
	Roles        []userRoleData `json:"roles"`
}

type userPutData struct {
	Password string         `json:"password"`
	Roles    []userRoleData `json:"roles"`
}

func userGet(prvdr *schemas.Provider, usr *schemas.User) (
	data *userData, err error) {

	req, err := http.NewRequest(
		"GET",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/databaseUsers/admin/%s",
			usr.GroupId,
			usr.Name,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
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
				"resources: User request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &userData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: User decode failed"),
		}
		return
	}

	return
}

func userPost(prvdr *schemas.Provider, usr *schemas.User) (err error) {
	data := userPostData{
		DatabaseName: "admin",
		Username:     usr.Name,
		Password:     usr.Password,
		GroupId:      usr.GroupId,
		Roles: []userRoleData{
			userRoleData{
				DatabaseName: usr.DatabaseName,
				RoleName:     "readWrite",
			},
		},
	}

	body, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: User marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"POST",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/databaseUsers",
			usr.GroupId,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
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
				"resources: User request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	return
}

func userPut(prvdr *schemas.Provider, usr *schemas.User) (
	data *userData, err error) {

	putData := userPutData{
		Password: usr.Password,
		Roles: []userRoleData{
			userRoleData{
				DatabaseName: usr.DatabaseName,
				RoleName:     "readWrite",
			},
		},
	}

	body, err := json.Marshal(putData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: User marshal failed"),
		}
		return
	}

	req, err := http.NewRequest(
		"PATCH",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/databaseUsers/admin/%s",
			usr.GroupId,
			usr.Name,
		),
		bytes.NewBuffer(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
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
				"resources: User request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	data = &userData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "resources: User decode failed"),
		}
		return
	}

	return
}

func userDel(prvdr *schemas.Provider, usr *schemas.User) (err error) {
	req, err := http.NewRequest(
		"DELETE",
		constants.BaseUrl+fmt.Sprintf(
			"/api/atlas/v1.0/groups/%s/databaseUsers/admin/%s",
			usr.GroupId,
			usr.Name,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := digest.Do(client, req, prvdr.Username, prvdr.ApiKey)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "resources: User request failed"),
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
				"resources: User request bad status %d %s",
				resp.StatusCode,
				respBodyStr,
			),
		}
		return
	}

	return
}

func userUriParse(usr *schemas.User, inputUri string) (
	uriStr string, err error) {

	uri, err := url.Parse(inputUri)
	if err != nil {
		err = errortypes.ParseError{
			errors.New("resources: Invalid MongoDB uri"),
		}
		return
	}

	uri.Path = "/" + usr.Name
	uri.User = url.UserPassword(usr.Name, usr.Password)

	uriStr = uri.String()

	return
}

func userCreate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	usr := schemas.LoadUser(d)

	clstData, err := clusterGet(prvdr, usr.GroupId, usr.ClusterName)
	if err != nil {
		return
	}

	if clstData == nil {
		err = errortypes.NotFoundError{
			errors.New("resources: Cluster not found"),
		}
		return
	}

	usrData, err := userGet(prvdr, usr)
	if err != nil {
		return
	}

	if usrData != nil {
		err = userDel(prvdr, usr)
		if err != nil {
			return
		}
	}

	err = userPost(prvdr, usr)
	if err != nil {
		return
	}

	for {
		clstData, err = clusterGet(prvdr, usr.GroupId, usr.ClusterName)
		if err != nil {
			return
		}

		if clstData == nil {
			err = errortypes.NotFoundError{
				errors.New("resources: Cluster not found"),
			}
			return
		} else if clstData.Available() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	uri, err := userUriParse(usr, clstData.MongoUriWithOptions)
	if err != nil {
		return
	}

	d.Set("mongodb_uri", uri)
	d.SetId(usr.Name)

	return
}

func userRead(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	usr := schemas.LoadUser(d)

	clstData, err := clusterGet(prvdr, usr.GroupId, usr.ClusterName)
	if err != nil {
		return
	}

	if clstData == nil {
		err = errortypes.NotFoundError{
			errors.New("resources: Cluster not found"),
		}
		return
	}

	usrData, err := userGet(prvdr, usr)
	if err != nil {
		return
	}

	if usrData == nil {
		d.SetId("")
		return
	}

	uri, err := userUriParse(usr, clstData.MongoUriWithOptions)
	if err != nil {
		return
	}

	d.Set("mongodb_uri", uri)
	d.SetId(usr.Name)

	return
}

func userUpdate(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	usr := schemas.LoadUser(d)

	clstData, err := clusterGet(prvdr, usr.GroupId, usr.ClusterName)
	if err != nil {
		return
	}

	if clstData == nil {
		err = errortypes.NotFoundError{
			errors.New("resources: Cluster not found"),
		}
		return
	}

	usrData, err := userPut(prvdr, usr)
	if err != nil {
		return
	}

	if usrData == nil {
		d.SetId("")
		return
	}

	uri, err := userUriParse(usr, clstData.MongoUriWithOptions)
	if err != nil {
		return
	}

	d.Set("mongodb_uri", uri)
	d.SetId(usr.Name)

	return
}

func userDelete(d *schema.ResourceData, m interface{}) (err error) {
	prvdr := m.(*schemas.Provider)
	usr := schemas.LoadUser(d)

	err = userDel(prvdr, usr)
	if err != nil {
		return
	}

	return
}
