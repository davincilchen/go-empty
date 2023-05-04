package server

import (
	"encoding/json"
	"fmt"
	"go-empty/pkg/app/ctxcache"
	devUCase "go-empty/pkg/app/device/usecase"
	"time"

	// gcpLogging "cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type BasicLog struct {
	SessionToken *string `json:"sessionToken"`
	Version      string
	RequestURI   string
	Method       string
	Duration     time.Duration
	DurationText string
	InfoTxt      string
}

func (t *BasicLog) MakeTokenString() string {
	SessionToken := ""
	if t.SessionToken != nil {
		SessionToken = *t.SessionToken
	}
	return SessionToken
}

func (t *BasicLog) MakeBasicString() string {

	s := fmt.Sprintf("%s [version:%s] ,%s %s %s [info:%s]",
		t.MakeTokenString(), t.Version,
		t.Method, t.RequestURI, t.DurationText,
		t.InfoTxt)
	return s
}

type APILog struct {
	BasicLog
	DBErrorTxt   string
	HttpErrorTxt string
	ErrorTxt     string
	AdvErrorTxt  string

	RequestBody *string `json:"requestBody"`

	DevData interface{}
}

func (t *APILog) HaveError() bool {
	s := t.AdvErrorTxt + t.ErrorTxt +
		t.DBErrorTxt + t.HttpErrorTxt

	return s != ""
}

func (t *APILog) MakeErrorString() string {

	if !t.HaveError() {
		return ""
	}
	s := t.MakeTokenString() +
		" ,ErrorTxt:" + t.ErrorTxt +
		" ,AdvErrorTxt:" + t.AdvErrorTxt +
		" ,DBErrorTxt:" + t.DBErrorTxt +
		" ,HttpErrorTxt:" + t.HttpErrorTxt

	return s
}

func (t *APILog) MakeDevString() string {

	//不用檢查t.DevData interface null
	//null marshal 還是會成功,會寫"null" string
	//if log.DevData != nil { 無法檢查interface, 這裡可以不檢查

	s := ""
	b, err := json.Marshal(t.DevData)
	if err == nil {
		s = string(b)
	}
	ret := t.MakeTokenString() +
		" ,DevData:" + s
	return ret
}

// type Gcp struct {
// 	client    *gcpLogging.Client
// 	logger    *gcpLogging.Logger
// 	stdLogger *log.Logger
// }

//gcloud auth application-default login
//gcloud auth application-default revoke

// var gcp *Gcp

func InitLogger(key string, project string, logName string) error {
	// //client, err := gcpLogging.NewClient(context.Background(), project, option.WithCredentialsFile(key))
	// client, err := gcpLogging.NewClient(context.Background(), project)
	// if err != nil {
	// 	return err
	// }
	// gcp = &Gcp{client: client,
	// 	logger:    client.Logger(logName),
	// 	stdLogger: client.Logger(logName).StandardLogger(gcpLogging.Info)}

	return nil
}

func CloseLogger() {
	// if gcp != nil {
	// 	gcp.client.Close()
	// }
}

func Logger(ctx *gin.Context) {
	now := time.Now()

	ctx.Next()

	log := APILog{}

	theError := ctxcache.GetError(ctx)
	theAdvError := ctxcache.GetAdvError(ctx)
	theDBError := ctxcache.GetDBError(ctx)
	theHttpError := ctxcache.GetHttpError(ctx)
	log.RequestURI = ctx.Request.RequestURI
	log.Method = ctx.Request.Method
	log.Duration = time.Since(now)
	log.DurationText = fmt.Sprintf("%v", log.Duration)

	log.RequestBody, _ = ctxcache.GetRequestBodyInGin(ctx)
	log.SessionToken = ctxcache.GetSessionToken(ctx)
	log.DevData = devUCase.GetCacheDevice(ctx)
	// log.PlayerSession, _ = GetPlayerSessionInGin(ctx)

	if theError != nil {
		log.ErrorTxt = theError.Error()
	}
	if theAdvError != nil {
		log.AdvErrorTxt = theAdvError.Error()
	}
	if theDBError != nil {
		log.DBErrorTxt = theDBError.Error()
	}
	if theHttpError != nil {
		log.HttpErrorTxt = theHttpError.Error()
	}

	logger(log)

	// if theError != nil {
	// 	log.Error = theError
	// 	log.ErrorTxt = log.Error.Error()
	// 	Warning(log)
	// } else {
	// 	Info(log)
	// }

}

func logger(log APILog) {

	// if gcp == nil {
	// 	fmt.Printf("[gcp == nil] %#v", log)
	// 	fmt.Println(log)
	// 	return
	// }
	//logBasicInfo(log)

	s := log.MakeBasicString()
	devString := log.MakeDevString()

	if log.HaveError() {
		logrus.Error(s)
		sErr := log.MakeErrorString()
		logrus.Error(sErr)
		logrus.Error(devString)

	} else {
		logrus.Info(s)
		logrus.Info(devString)
	}

	fmt.Println()
	// gcp.stdLogger.Println(s)

	// //defer gcp.logger.Flush() // Ensure the entry is written.
	// gcp.logger.Log(gcpLogging.Entry{
	// 	// Log anything that can be marshaled to JSON.
	// 	Payload:  log,
	// 	Severity: gcpLogging.Debug,
	// })
}
