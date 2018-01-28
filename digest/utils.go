package digest

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/terraform-provider-mongodbatlas/errortypes"
	"math"
	"regexp"
)

var (
	randRe = regexp.MustCompile("[^a-zA-Z0-9]+")
)

func randStr(n int) (str string, err error) {
	for i := 0; i < 10; i++ {
		input, e := randBytes(int(math.Ceil(float64(n) * 1.25)))
		if e != nil {
			err = e
			return
		}

		output := base64.RawStdEncoding.EncodeToString(input)
		output = randRe.ReplaceAllString(output, "")

		if len(output) < n {
			continue
		}

		str = output[:n]
		break
	}

	if str == "" {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "utils: Random generate error"),
		}
		return
	}

	return
}

func randBytes(size int) (bytes []byte, err error) {
	bytes = make([]byte, size)
	_, err = rand.Read(bytes)
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "utils: Random read error"),
		}
		return
	}

	return
}
