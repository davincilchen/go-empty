package delivery

import (
	"net/http"

	// "go-empty/pkg/app/ctxcache"

	"github.com/gin-gonic/gin"

	dlvModel "go-empty/pkg/app/deliverymodel"
	devUCase "go-empty/pkg/app/device/usecase"
	dlv "go-empty/pkg/delivery"
)

type DeviceDetailListResp struct {
	Total int                         `json:"total_num"`
	List  []dlvModel.DeviceInfoDetail `json:"device"`
}

func DeviceList(ctx *gin.Context) {

	devM := devUCase.GetDeviceManager()
	list := make([]dlvModel.DeviceInfoDetail, 0)

	devices := devM.GetDevices()
	for _, v := range devices {
		warpDev := dlvModel.WarpDeviceInfo(&v.QLoginDeviceRet)
		tmp := dlvModel.DeviceInfoDetail{
			DeviceInfo: *warpDev,
			Edge:       dlvModel.WarpEdgeInfo(v.Edge, &v.QLoginDeviceRet),
		}
		list = append(list, tmp)
	}

	data := DeviceDetailListResp{
		Total: len(list),
		List:  list,
	}
	response := dlv.ResBody{}
	response.ResCode = dlv.RES_OK
	response.Data = data

	ctx.JSON(http.StatusOK, response)
}
