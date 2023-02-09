package controllers

type ResCode int

const (
	RES_OK                    ResCode = 0
	RES_BALANCE_NOT_ENOUGH    ResCode = 100
	RES_ERROR_UNKNOWN         ResCode = 200
	RES_ERROR_BAD_REQUEST     ResCode = 201
	RES_INVALID_USER_PASSWORD ResCode = 202
	// RES_INVALID_USER          ResCode = 203
	RES_INVALID_USER_TOKEN ResCode = 204
)

type ResError struct {
	Title string `json:"title"`
	Desc  string `json:"description"`
}

type ResBody struct {
	ResCode ResCode     `json:"resCode"`
	Error   *ResError   `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
