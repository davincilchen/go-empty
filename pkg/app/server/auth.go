package server

import (
	"fmt"
	"go-empty/pkg/app/ctxcache"

	"github.com/gin-gonic/gin"

	deviceUCase "go-empty/pkg/app/device/usecase"
	dlv "go-empty/pkg/delivery"
)

// ===================== //

func getSessionToken(ctx *gin.Context) (string, error) {

	key := "authorization"
	// _, ok := ctx.Request.Header[key] //authorization 取不到//"X-Session-Token"可以
	// if !ok {
	// 	return "", fmt.Errorf("getSessionToken failed")
	// }

	authorization := ctx.GetHeader(key)
	if authorization == "" {
		return "", fmt.Errorf("getSessionToken failed")
	}

	// s := strings.Split(authorization, "Bearer ")
	// if len(s) < 2 {
	// 	return "", fmt.Errorf("Split Bearer failed form [%s]", authorization)
	// }
	// token := s[1]

	//return token, nil

	return authorization, nil
}

func AuthDevSession(ctx *gin.Context) {
	sessionToken, err := getSessionToken(ctx)
	if err != nil {
		dlv.RespUnauthorized(ctx, err)
		return
	}
	ctxcache.CacheSessionToken(ctx, sessionToken)
	ok := deviceUCase.AuthDeviceToken(ctx, sessionToken)
	if !ok {
		dlv.RespUnauthorized(ctx, err)
	}
}

// func auth(ctx *gin.Context, allowOffline bool) {
// 	sessionToken, err := getSessionTokenInAuth(ctx)
// 	if err != nil {
// 		ResponseRequestError(ctx, err)
// 		return
// 	}
// 	CacheSessionTokenInGin(ctx, sessionToken)

// 	loginPlayer, errAuth := AuthLoginToken(sessionToken, allowOffline)
// 	if errAuth != nil {
// 		ResponseAuthError(ctx, errAuth)
// 		return
// 	}
// 	CachePlayerSessionInGin(ctx, loginPlayer)
// }

// func authBlockPlayer() {
// }

// func authDuplicatePlayer() {

// }
