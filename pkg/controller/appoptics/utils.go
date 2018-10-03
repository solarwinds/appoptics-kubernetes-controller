package appoptics

import (
	"crypto/sha1"
	aoApi "github.com/appoptics/appoptics-api-go"
	"io"
	"strings"

	"encoding/json"
	"github.com/golang/glog"
)

const (
	Dashboard = "dashboard"
	Alert     = "alert"
	Service   = "service"
)

func CheckIfErrorIsAppOpticsNotFoundError(err error, kind string, id int) bool {
	if errorResponse, ok := err.(*aoApi.ErrorResponse); ok {
		errorObj := errorResponse.Errors.(map[string]interface{})
		if requestErr, ok := errorObj["request"]; ok {
			for _, errorType := range requestErr.([]interface{}) {
				// The ID does not exist in AppOptics so create a new space
				if strings.Compare(errorType.(string), "Not Found") == 0 {
					glog.Warningf("Not Found in APPOPTICS type: %s (id: %d) was not found in AppOptics", kind, id)
					return true
				}
			}
		}
	}
	return false
}

func Hash(s interface{}) ([]byte, error) {
	byteArr, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	h := sha1.New()
	io.WriteString(h, string(byteArr))
	return h.Sum(nil), nil
}
