package server

import (
	"fmt"
	"time"

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

func CloseLogger() {

}

func Logger(ctx *gin.Context) {
	now := time.Now()

	ctx.Next()

	log := APILog{}

	log.Duration = time.Since(now)
	log.DurationText = fmt.Sprintf("%v", log.Duration)

	logger(log)

}

func logger(log APILog) {

	s := log.MakeBasicString()

	if log.HaveError() {
		logrus.Error(s)
		sErr := log.MakeErrorString()
		logrus.Error(sErr)

	} else {
		logrus.Info(s)

	}

	fmt.Println()

}
