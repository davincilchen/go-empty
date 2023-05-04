package delivery

import (
	"errors"
	"net/http"

	"go-empty/pkg/app/ctxcache"
	"go-empty/pkg/app/errordef"

	"github.com/gin-gonic/gin"
)

var IP = ""

type ResCode int

const (
	RES_OK ResCode = 0

	RES_NO_RESOURCE ResCode = 100
	RES_EDGE_LOST   ResCode = 101

	RES_REPEATED_LOGIN   ResCode = 120
	RES_REPEATED_RESERVE ResCode = 121
	RES_NO_RESERVE       ResCode = 122
	RES_IN_PROCESS       ResCode = 123

	RES_START_TIME_OUT    ResCode = 140
	RES_INVALID_STEAM_VR  ResCode = 141
	RES_CLOUDXR_UNCONNECT ResCode = 142

	RES_ERROR_UNKNOWN         ResCode = 200
	RES_ERROR_BAD_REQUEST     ResCode = 201
	RES_INVALID_USER_TOKEN    ResCode = 202
	RES_INVALID_USER_PASSWORD ResCode = 203
)

type ResError struct {
	Title string `json:"title"`
	Desc  string `json:"description"`
}

type ResBody struct {
	ResCode ResCode     `json:"resp_code"`
	Error   *ResError   `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func RespBadRequest(ctx *gin.Context, err error) {

	response := FillErrorBody(ctx, err)
	response.ResCode = RES_ERROR_BAD_REQUEST
	ctxcache.CacheError(ctx, err)
	ctx.JSON(http.StatusBadRequest, response)

}

func RespUnauthorized(ctx *gin.Context, err error) {

	response := FillErrorBody(ctx, err)
	response.ResCode = RES_INVALID_USER_TOKEN
	if err == nil {
		ctxcache.CacheError(ctx, errors.New("INVALID_USER_TOKEN"))
	} else {
		ctxcache.CacheError(ctx, err)
	}

	ctx.JSON(http.StatusUnauthorized, response)
	ctx.Abort()
}

func RespInvalidPassword(ctx *gin.Context) {

	response := FillErrorBody(ctx, nil)
	response.ResCode = RES_INVALID_USER_PASSWORD
	ctx.JSON(http.StatusBadRequest, response)

}

func IsClientStatusCode(err error) bool {
	switch err {
	case errordef.ErrNotPlaying:
		return true
	case errordef.ErrAlreadyPlaying:
		return true
	case errordef.ErrAlreadyFree:
		return true
	case errordef.ErrProcessing:
		return true

	default:
		return false
	}
}

func GetStatusCode(err error) (ResCode, int) {
	//logrus.Error(err)
	switch err {
	case errordef.ErrNoResource:
		return RES_NO_RESOURCE, http.StatusOK //TODO: http code
	case errordef.ErrEdgeLost:
		return RES_EDGE_LOST, http.StatusOK //TODO: http code

	case errordef.ErrRepeatedLogin:
		return RES_REPEATED_LOGIN, http.StatusOK //TODO: http code
	case errordef.ErrRepeatedReserve:
		return RES_REPEATED_RESERVE, http.StatusOK //TODO: http code
	case errordef.ErrDevNoReserve:
		return RES_NO_RESERVE, http.StatusOK //TODO: http code
	case errordef.ErrInOldProcess:
		return RES_IN_PROCESS, http.StatusTooManyRequests

	case errordef.ErrStartAppTimeout: //TODO: http code
		return RES_START_TIME_OUT, http.StatusOK
	case errordef.ErrInvalidStramVR:
		return RES_INVALID_STEAM_VR, http.StatusOK //TODO: http code
	case errordef.ErrCloudXRUnconect:
		return RES_CLOUDXR_UNCONNECT, http.StatusOK //TODO: http code

		//
	case errordef.ErrUrlParamError:
		return RES_ERROR_BAD_REQUEST, http.StatusNotFound
		//

	default:
		if IsClientStatusCode(err) {
			return RES_ERROR_UNKNOWN, http.StatusBadRequest
		}
		return RES_ERROR_UNKNOWN, http.StatusInternalServerError
	}
}

func RespError(ctx *gin.Context, err, advErr error) {

	response := FillErrorBody(ctx, err)
	resCode, httpCode := GetStatusCode(err)
	response.ResCode = resCode
	ctxcache.CacheError(ctx, err)
	ctxcache.CacheAdvError(ctx, advErr)
	ctx.JSON(httpCode, response)

}

func FillErrorBody(ctx *gin.Context, err error) *ResBody {
	resp := &ResBody{}
	if err != nil {
		resp.Error = &ResError{
			Desc: err.Error(),
		}
	}
	return resp
}
