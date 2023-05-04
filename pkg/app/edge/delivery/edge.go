package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-empty/pkg/app/ctxcache"
	dlvModel "go-empty/pkg/app/deliverymodel"
	devUCase "go-empty/pkg/app/device/usecase"
	edgeUCase "go-empty/pkg/app/edge/usecase"
	errDef "go-empty/pkg/app/errordef"
	dlv "go-empty/pkg/delivery"
)

type NewReserveResp struct {
	GameServerIP string `json:"game_server_ip"`
}

func NewReserve(ctx *gin.Context) { //TODO:
	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}

	id, err := strconv.Atoi(ctx.Param("id")) //TODO: uint
	if err != nil {
		dlv.RespError(ctx, errDef.ErrUrlParamError, nil)
		return
	}
	nCtx := ctxcache.NewContext(ctx)
	ip, err := dev.NewReserve(nCtx, uint(id))
	if err != nil {
		code, _ := dlv.GetStatusCode(err)
		if code == dlv.RES_ERROR_UNKNOWN {
			dlv.RespError(ctx, errDef.ErrNoResource, err)
			return
		}
		dlv.RespError(ctx, err, nil)
		return
	}

	if ip == nil {
		dlv.RespError(ctx, errDef.ErrNoResource, nil)
		return
	}

	if *ip == "127.0.0.1" && dlv.IP != "" { // TODO
		*ip = dlv.IP
	}

	data := NewReserveResp{
		GameServerIP: *ip, //TODO: !!!
	}

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK
	response.Data = data

	ctx.JSON(http.StatusOK, response)
}

func ReleaseReserve(ctx *gin.Context) { //TODO:
	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}

	err := dev.ReleaseReserve(ctxcache.NewContext(ctx), false)
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK

	ctx.JSON(http.StatusOK, response)
}

func DeviceResume(ctx *gin.Context) { //TODO:
	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}

	err := dev.Resume(ctxcache.NewContext(ctx))
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK

	ctx.JSON(http.StatusOK, response)

}

type StartAppResp struct {
}

func StartApp(ctx *gin.Context) { //TODO:
	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}

	err := dev.StartApp(ctxcache.NewContext(ctx))
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK

	ctx.JSON(http.StatusOK, response)
}

type StopAppResp struct {
}

func StopApp(ctx *gin.Context) {
	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}

	err := dev.StopApp(ctxcache.NewContext(ctx))
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK

	ctx.JSON(http.StatusOK, response)
}

// =========================================== //

type EdgeStatusReq struct {
	DevStatus int    `json:"device_status"`
	StatusDes string `json:"status_des"`
}

type EdgeStatusResp struct {
	Edge *dlvModel.EdgeInfo `json:"edge"`
}

func edgeStatusParam(ctx *gin.Context) (*EdgeStatusReq, error) {
	// .. //
	param := EdgeStatusReq{}
	req, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		// Handle error
		e := errors.New("read request body failed")
		return nil, e
	}
	err = json.Unmarshal(req, &param)
	if err != nil {
		e := errors.New("unmarshal body failed")
		return nil, e
	}
	return &param, nil
}

func EdgeStatus(ctx *gin.Context) {
	dev := devUCase.GetCacheDevice(ctx)
	if dev == nil {
		e := errors.New("GetCacheDevice Nil")
		dlv.RespError(ctx, e, nil)
		return
	}
	// .. //
	param, err := edgeStatusParam(ctx)
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}
	dev.UpdateStatus(ctxcache.NewContext(ctx), devUCase.DevStatus(param.DevStatus))

	// .. //
	edge := dev.GetEdgeInfo()
	devInfo := dev.GetDeviceInfo()

	data := EdgeStatusResp{}
	if edge != nil {
		tmp := dlvModel.WarpEdgeInfo(edge, &devInfo)
		// tmp := dlvModel.EdgeInfo{
		// 	ID:     edge.ID,
		// 	IP:     edge.IP,
		// 	Port:   edge.Port,
		// 	Status: edge.Status,
		// 	Online: edge.Online,
		// 	ActRet: edge.ActRet,
		// 	AppID:  dev.GetAppID(),
		// 	Device: dlvModel.WarpDeviceInfo(&devInfo),
		// }

		data.Edge = tmp

	}

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK
	response.Data = data

	ctx.JSON(http.StatusOK, response)
}

// =========================================== //

type EdgeListResp struct {
	Total int                 `json:"total_num"`
	Edges []dlvModel.EdgeInfo `json:"edges"`
}

func EdgeList(ctx *gin.Context) {
	edgeM := edgeUCase.GetEdgeManager()
	devM := devUCase.GetDeviceManager()
	ret := edgeM.GetEdgeList()

	data := EdgeListResp{}

	for _, v := range ret {
		// tmp := dlvModel.EdgeInfo{
		// 	ID:     v.ID,
		// 	IP:     v.IP,
		// 	Port:   v.Port,
		// 	Status: v.Status,
		// 	Online: v.Online,
		// 	ActRet: v.ActRet,
		// }

		dev := devM.GetDevInfoWithEdge(v.ID)
		// if dev != nil {
		// 	tmp.AppID = dev.GetAppID()
		// 	tmp.Device = dlvModel.WarpDeviceInfo(dev)
		// }
		tmp := dlvModel.WarpEdgeInfo(&v, dev)
		data.Edges = append(data.Edges, *tmp)
	}

	data.Total = len(data.Edges)

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK
	response.Data = data

	ctx.JSON(http.StatusOK, response)
}

// =========================================== //

type EdgeRegReq struct {
	AppsID []uint `json:"apps_id"`
	Port   int    `json:"port"`
}

func edgeRegParam(ctx *gin.Context) (*EdgeRegReq, error) {
	// .. //
	param := EdgeRegReq{}
	req, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		// Handle error
		e := errors.New("read request body failed")
		return nil, e
	}
	err = json.Unmarshal(req, &param)
	if err != nil {
		e := errors.New("unmarshal body failed")
		return nil, e
	}

	if param.Port <= 0 {
		e := fmt.Errorf("error port: %d", param.Port)
		return nil, e
	}
	return &param, nil
}

func EdgeReg(ctx *gin.Context) { //TODO: 處理掉刪除的部分 多一個欄位叫valid
	// .. //
	param, err := edgeRegParam(ctx)
	fmt.Printf("EdgeReg: apps %#v", param)
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}
	//ip := ctx.Request.RemoteAddr
	ip := ctx.ClientIP() //TODO:127.0.0.1 ::1換成真的ip 否則device會打不到
	manager := edgeUCase.GetEdgeManager()
	edge, err := manager.RegEdge(ip, param.Port)
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}
	edge.RegApps(param.AppsID)

	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK
	//response.Data = data

	ctx.JSON(http.StatusOK, response)

}

type AppInfo struct {
	ID uint `json:"id"`
}

type EdgeAppListResp struct {
	Apps []AppInfo `json:"apps"`
}

func EdgeAppList(ctx *gin.Context) { //TODO: 處理掉刪除的部分 多一個欄位叫valid

	id, err := strconv.Atoi(ctx.Param("id")) //TODO: uint
	if err != nil {
		dlv.RespError(ctx, errDef.ErrUrlParamError, nil)
		return
	}
	manager := edgeUCase.GetEdgeManager()
	apps, err := manager.GetAppsOfEdge(uint(id))
	if err != nil {
		dlv.RespError(ctx, err, nil)
		return
	}

	data := EdgeAppListResp{}
	for _, v := range apps {
		tmp := AppInfo{
			ID: v.ID,
		}
		data.Apps = append(data.Apps, tmp)
	}
	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK
	response.Data = data

	ctx.JSON(http.StatusOK, response)

}
