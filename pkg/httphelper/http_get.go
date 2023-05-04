package httphelper

import (
	"crypto/tls"
	"fmt"
	"go-empty/version"
	"net"
	"net/http"
	"time"
)

const (
	APICommentNormal       = "Normal"
	APICommentRoutineRetry = "Routine retry"
)

var APICommentNormalV = fmt.Sprintf("[%s], [%s]", version.Version, APICommentNormal)
var APICommentRoutineRetryV = fmt.Sprintf("[%s], [%s]", version.Version, APICommentRoutineRetry)

var client = &http.Client{
	Timeout: 20 * time.Second,
	//Timeout: 8 * time.Second,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			// Timeout:   5 * time.Second,
			// KeepAlive: 5 * time.Second,
			Timeout:   10 * time.Second, //TODO:
			KeepAlive: 5 * time.Second,  //TODO:
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //取消驗證憑證
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func Get(url string) (res *http.Response, err error) {
	now := time.Now()
	res, err = client.Get(url)

	if res == nil {
		NilResponseLog(APICommentNormalV, http.MethodGet, url, "", time.Since(now), err)
		if err == nil {
			err = fmt.Errorf("url:%s,  method:%s,  response is nil", url, http.MethodGet)
		}
	} else { //no error if http 405
		ResponseLog(APICommentNormalV, "", res, time.Since(now), err)
	}

	return
}
