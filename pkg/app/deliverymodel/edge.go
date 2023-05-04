package deliverymodel

import (
	"go-empty/pkg/models"

	//devUCase "go-empty/pkg/app/device/usecase"
	devUCase "go-empty/pkg/app/device/usecase"
	edgeUCase "go-empty/pkg/app/edge/usecase"
)

type EdgeInfo struct {
	ID     uint                `json:"id"`
	IP     string              `json:"ip"`
	Port   int                 `json:"port"`
	Status models.EdgeStatus   `json:"status"`
	Online bool                `json:"online"`
	ActRet edgeUCase.ActionRet `json:"last_act_ret"`

	Device *DeviceInfo `json:"device"`
	AppID  *uint       `json:"app_id"`
}

func WarpEdgeInfo(edgeIn *edgeUCase.EdgeInfoStatus,
	devIn *devUCase.QLoginDeviceRet) *EdgeInfo {

	if edgeIn == nil {
		return nil
	}

	out := &EdgeInfo{
		ID:     edgeIn.ID,
		IP:     edgeIn.IP,
		Port:   edgeIn.Port,
		Status: edgeIn.Status,
		Online: edgeIn.Online,
		ActRet: edgeIn.ActRet,
	}

	if devIn != nil {
		out.AppID = devIn.GetAppID()
		out.Device = WarpDeviceInfo(devIn)
	}

	return out

}
