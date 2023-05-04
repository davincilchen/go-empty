package delivery

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-empty/pkg/app/ctxcache"

	devUCase "go-empty/pkg/app/device/usecase"
	userUCase "go-empty/pkg/app/user/usecase"
	dlv "go-empty/pkg/delivery"
)

type loginSuccess func(ctxcache.Context, userUCase.LoginUser) error

type UserLoginParams struct {
	Account  *string
	Password *string
}

type DevInfo struct {
	Type *int `json:"device_type"`
	UUID *string
}

type DevLoginParams struct {
	UserLoginParams
	DevInfo
}

type DevLoginResponse struct {
	ID    uint
	Name  string
	Token string
}

type LoginResponse struct { //:TODO for device login
	ID    uint   `json:"user_id"`
	Name  string `json:"user_name"`
	Token string `json:"token"`
}

// ======================================== //

func DevLogin(ctx *gin.Context) {
	req := &DevLoginParams{}
	err := dlv.GetBodyFromRawData(ctx, req)
	if err != nil {
		dlv.RespBadRequest(ctx, err)
		return
	}

	if req.DevInfo.Type == nil || req.DevInfo.UUID == nil {
		e := fmt.Errorf("nil device_type or UUID")
		dlv.RespBadRequest(ctx, e)
		return
	}

	ip := ctx.ClientIP()
	d := devUCase.NewDeviceLoginProc(*req.DevInfo.Type, *req.DevInfo.UUID, ip)
	handle := NewLoginController(ctx, req.UserLoginParams, d.DevLoginSucess)
	handle.Do()

}

// ======================================== //

type LoginController struct {
	ctx         *gin.Context
	loginParams UserLoginParams
	//loginUser   *userUCase.LoginUser
	fnSuccess loginSuccess
}

func (t *LoginController) Do() {

	ctx := t.ctx
	if t.loginParams.Account == nil || t.loginParams.Password == nil {
		e := fmt.Errorf("nil Account or Password")
		dlv.RespBadRequest(ctx, e)
		return
	}

	loginUser := t.authWhenLogin()
	if loginUser == nil {
		dlv.RespInvalidPassword(ctx)
		return
	}

	err := t.fnSuccess(ctxcache.NewContext(ctx), *loginUser)

	if err != nil {
		dlv.RespError(ctx, err, nil) //TODO:
		return
	}
	LoginSucessReponse(ctx, loginUser)
}

func (t *LoginController) authWhenLogin() *userUCase.LoginUser {

	param := t.loginParams
	user := userUCase.User{}
	ret, err := user.Login(*param.Account, *param.Password)
	if err != nil {
		return nil
	}

	return ret

}

func NewLoginController(ctx *gin.Context,
	loginParams UserLoginParams,
	fn loginSuccess) *LoginController {
	return &LoginController{
		ctx:         ctx,
		loginParams: loginParams,
		fnSuccess:   fn,
	}
}

func LoginSucessReponse(ctx *gin.Context, user *userUCase.LoginUser) {
	response := dlv.ResBody{}
	data := LoginResponse{
		ID:    user.ID,
		Name:  user.Name,
		Token: user.Token,
	}
	response.ResCode = dlv.RES_OK
	response.Data = data
	ctx.JSON(http.StatusOK, response)

}

// ======================================== //

func Login(ctx *gin.Context) {
	req := &UserLoginParams{}
	err := dlv.GetBodyFromRawData(ctx, req)
	if err != nil {
		dlv.RespBadRequest(ctx, err)
		return
	}

	handle := NewLoginController(ctx, *req, UserLoginSucess)
	handle.Do()

}

func UserLoginSucess(ctx ctxcache.Context, user userUCase.LoginUser) error {
	return nil
}

// ======================================== //
func DevLogout(ctx *gin.Context) {

	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}

	nCtx := ctxcache.NewContext(ctx)
	err := dev.Logout(nCtx)
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}
	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK

	ctx.JSON(http.StatusOK, response)
}

func Logout(ctx *gin.Context) {
	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK

	ctx.JSON(http.StatusOK, response)
}
