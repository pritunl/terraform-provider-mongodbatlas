package digest

import (
	"net/http"

	"crypto/md5"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/terraform-provider-mongodbatlas/errortypes"
	"strings"
)

func Do(client *http.Client, req *http.Request,
	username, password string) (resp *http.Response, err error) {

	digestReq, err := http.NewRequest(
		req.Method,
		req.URL.String(),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "digest: Digest request failed"),
		}
		return
	}

	digestReq.Header = req.Header

	digestResp, err := client.Do(digestReq)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "digest: Digest request failed"),
		}
		return
	}
	defer digestResp.Body.Close()

	auth := map[string]string{}

	authSpl := strings.Split(digestResp.Header.Get("WWW-Authenticate"), ",")

	for _, item := range authSpl {
		itemSpl := strings.Split(item, "=")
		if len(itemSpl) != 2 {
			continue
		}

		key := strings.TrimSpace(itemSpl[0])
		val := strings.Trim(strings.TrimSpace(itemSpl[1]), "\"")
		auth[key] = val
	}

	nc := "00000001"
	realm := auth["Digest realm"]
	nonce := auth["nonce"]
	qop := auth["qop"]
	cnonce, err := randStr(32)
	if err != nil {
		return
	}

	a1Hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf(
		"%s:%s:%s",
		username,
		realm,
		password,
	))))
	a2Hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf(
		"%s:%s",
		req.Method,
		req.URL.Path,
	))))
	respHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s",
		a1Hash,
		nonce,
		nc,
		cnonce,
		qop,
		a2Hash,
	))))

	authHeader := fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc=%s, qop=%s, response="%s", algorithm="MD5"`,
		username,
		realm,
		nonce,
		req.URL.Path,
		cnonce,
		nc,
		qop,
		respHash,
	)

	req.Header.Set("Authorization", authHeader)

	resp, err = client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "digest: Request failed"),
		}
		return
	}

	return
}
