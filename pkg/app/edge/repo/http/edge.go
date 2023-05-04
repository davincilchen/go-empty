package http

// import "go-empty/pkg/models"
import (
	"encoding/json"
	"fmt"
	"go-empty/pkg/app/ctxcache"
	errDef "go-empty/pkg/app/errordef"
	httph "go-empty/pkg/httphelper"
	"io/ioutil"
	"net/http"
)

type ResCode int

const (
	RES_OK                ResCode = 0
	RES_START_TIME_OUT    ResCode = 1
	RES_CLOUDXR_NOT_RUN   ResCode = 2
	RES_CLOUDXR_UNCONNECT ResCode = 3
)

type ResError struct {
	Title string `json:"title"`
	Desc  string `json:"description"`
}

type ResBody struct {
	ResCode ResCode `json:"resp_code"`
	// Error   *ResError   `json:"error,omitempty"`
	// Data    interface{} `json:"data,omitempty"`
}

func ResCodeToErr(code ResCode) error {
	switch code {
	case RES_START_TIME_OUT:
		return errDef.ErrStartAppTimeout
	case RES_CLOUDXR_NOT_RUN:
		return errDef.ErrInvalidStramVR
	case RES_CLOUDXR_UNCONNECT:
		return errDef.ErrCloudXRUnconect
	}

	return nil
}
func NewEdge(URL string) *Edge {
	return &Edge{
		URL: URL,
	}

}

type Edge struct {
	URL string
}

func (t *Edge) SetURL(url string) {
	t.URL = url
}

func (t *Edge) parseRespBody(res *http.Response, out interface{}) error {

	defer res.Body.Close() //20190815
	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return fmt.Errorf("ParseRespBody body ReadAll error : %s,", bodyErr)
	}

	err := json.Unmarshal(body, &out)
	if err != nil {
		return fmt.Errorf("ParseRespBody body Unmarshal error : %s,", err)
	}
	return nil

}

func (t *Edge) Reserve(ctx ctxcache.Context, appID uint) error {
	url := fmt.Sprintf("http://%s/reserve/apps/%d", t.URL, appID)
	resp, err := httph.Post(url)
	if err != nil {
		ctx.CacheHttpError(err)
		return errDef.ErrEdgeLost
	}
	fmt.Println("[Response for no error] ", resp)
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("resp.StatusCode = %d", resp.StatusCode)
		ctx.CacheHttpError(err)
		return err
	}

	body := ResBody{}
	err = t.parseRespBody(resp, &body) //TODO: 容易有錯如果沒加&
	if err != nil {
		return err
	}
	return ResCodeToErr(body.ResCode)

}

func (t *Edge) Release(ctx ctxcache.Context) error {
	url := fmt.Sprintf("http://%s/reserve", t.URL)
	resp, err := httph.Delete(url)

	if err != nil {
		fmt.Println(err)
		ctx.CacheHttpError(err)
		return errDef.ErrEdgeLost
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("resp.StatusCode = %d", resp.StatusCode)
		ctx.CacheHttpError(err)
		return err
	}
	return err
}

func (t *Edge) Resume(ctx ctxcache.Context) error {

	return nil
}

func (t *Edge) GetStatus(ctx ctxcache.Context) error {

	return nil
}

func (t *Edge) Status(ctx ctxcache.Context) error {

	return nil
}

func (t *Edge) StartAPP(ctx ctxcache.Context, appID uint) error {
	url := fmt.Sprintf("http://%s/apps/%d/start_app", t.URL, appID)
	resp, err := httph.Post(url)

	if err != nil {
		fmt.Println(err)
		ctx.CacheHttpError(err)
		return errDef.ErrEdgeLost
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("resp.StatusCode = %d", resp.StatusCode)
		ctx.CacheHttpError(err)
		return err
	}

	body := ResBody{}
	err = t.parseRespBody(resp, &body) //TODO: 容易有錯如果沒加&
	if err != nil {
		return err
	}
	return ResCodeToErr(body.ResCode)
}

func (t *Edge) StopAPP(ctx ctxcache.Context) error {
	url := fmt.Sprintf("http://%s/stop_app", t.URL)
	resp, err := httph.Post(url)

	if err != nil {
		fmt.Println(err)
		ctx.CacheHttpError(err)
		return errDef.ErrEdgeLost
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("resp.StatusCode = %d", resp.StatusCode)
		ctx.CacheHttpError(err)
		return err
	}

	return err
	//return errDef.ErrEdgeLost
	//return nil
}
