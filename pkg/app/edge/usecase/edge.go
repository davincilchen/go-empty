package usecase

import (
	"fmt"
	"go-empty/pkg/models"
	"sync"

	"go-empty/pkg/app/ctxcache"
	edgeHttp "go-empty/pkg/app/edge/repo/http"
	repo "go-empty/pkg/app/edge/repo/mysql"
	errDef "go-empty/pkg/app/errordef"
)

type ActionRet int

const (
	ACTION_RET_NORMAL         ActionRet = 0
	ACTION_RET_RESERVE_FAILD  ActionRet = 1
	ACTION_RET_STARTAPP_FAILD ActionRet = 2
	ACTION_RET_STOPAPP_FAILD  ActionRet = 3
	ACTION_RET_RELEASE_FAILD  ActionRet = 4
)

type HttpEdge interface {
	SetURL(url string)
	Reserve(ctx ctxcache.Context, appID uint) error
	Release(ctx ctxcache.Context) error
	Resume(ctx ctxcache.Context) error
	StartAPP(ctx ctxcache.Context, appID uint) error
	StopAPP(ctx ctxcache.Context) error
	GetStatus(ctx ctxcache.Context) error
}

type Edge struct {
	mux    sync.Mutex
	info   models.Edge
	actRet ActionRet
	eHttp  HttpEdge
}
type EdgeInfoStatus struct {
	models.Edge
	ActRet ActionRet
}

func NewEdge(edge models.Edge) *Edge {
	e := Edge{
		info: edge,
	}
	e.eHttp = edgeHttp.NewEdge(e.GetURL())
	return &e
}

func (t *Edge) GetURL() string {
	e := t.info
	if e.Port > 0 {
		return fmt.Sprintf("%s:%d", e.IP, e.Port)
	}

	return e.IP
}

func (t *Edge) Reserve(ctx ctxcache.Context, appID uint) error {

	//online由每次reg時確認,減少api時間
	ok := t.updateStatusWhen(models.STATUS_FREE, models.STATUS_RESERVE_INIT)
	if !ok {
		return errDef.ErrNoResource
	}

	err := t.eHttp.Reserve(ctx, appID)

	status := models.STATUS_RESERVE_XR_NOT_CONNECT
	actRet := ACTION_RET_NORMAL
	var online *bool
	if err != nil {
		status = models.STATUS_FREE //reset free for next try
		if err == errDef.ErrEdgeLost {
			tmp := false
			online = &tmp
		} else {
			actRet = ACTION_RET_RESERVE_FAILD
			//status = models.STATUS_FREE
			//status = models.STATUS_FAIL
		}
	}

	t.updateStatus(status, online, &actRet)

	return err

}

func (t *Edge) ReleaseReserve(ctx ctxcache.Context) error {

	status, _ := t.GetCacheStatus()
	if status == models.STATUS_FREE {
		return errDef.ErrAlreadyFree
	}

	status = models.STATUS_RX_RELEASE
	t.updateStatus(status, nil, nil)
	err := t.eHttp.Release(ctx)

	status = models.STATUS_FREE
	actRet := ACTION_RET_NORMAL
	online := true
	if err != nil {
		if err == errDef.ErrEdgeLost {
			online = false
		} else {
			actRet = ACTION_RET_RELEASE_FAILD
			//status = models.STATUS_FREE
			//status = models.STATUS_FAIL
		}
	}

	t.updateStatus(status, &online, &actRet)
	return err

}

func (t *Edge) Resume(ctx ctxcache.Context) error {

	err := t.eHttp.Resume(ctx)

	if err == errDef.ErrEdgeLost {
		t.setOnline(false)
	}

	return err
}

func (t *Edge) GetStatus() error {

	return nil
}

func (t *Edge) GetCacheStatus() (models.EdgeStatus, bool) {
	t.mux.Lock()
	defer t.mux.Unlock()

	e := t.info
	return e.Status, e.Online
}

func (t *Edge) CanStartAPP() (bool, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	// if !t.info.Online { // try it
	// 	return false, errDef.ErrCloudXRUnconect
	// }

	if t.info.Status == models.STATUS_RX_START_APP {
		return false, errDef.ErrProcessing
	}

	if t.info.Status == models.STATUS_PLAYING {
		return false, errDef.ErrAlreadyPlaying
	}

	if t.info.Status != models.STATUS_RESERVE_XR_CONNECT {
		return false, errDef.ErrCloudXRUnconect
	}

	return true, nil
}

func (t *Edge) StartAPP(ctx ctxcache.Context, appID uint) error {
	ok, err := t.CanStartAPP()
	if !ok {
		return err
	}
	err = t.eHttp.StartAPP(ctx, appID)

	status := models.STATUS_PLAYING
	actRet := ACTION_RET_NORMAL
	var online = true
	if err != nil {
		status = models.STATUS_RESERVE_XR_CONNECT //退回狀態
		if err == errDef.ErrEdgeLost {
			online = false
		} else {
			actRet = ACTION_RET_STARTAPP_FAILD
			//status = models.STATUS_FAIL
		}
	}

	t.updateStatus(status, &online, &actRet)
	return err
}

func (t *Edge) StopAPP(ctx ctxcache.Context) error {

	ok := t.updateStatusWhen(models.STATUS_PLAYING,
		models.STATUS_RX_STOP_APP)
	if !ok {
		return errDef.ErrNotPlaying
	}

	err := t.eHttp.StopAPP(ctx)

	status := models.STATUS_RESERVE_XR_CONNECT
	actRet := ACTION_RET_NORMAL
	var online = true
	if err != nil {
		if err == errDef.ErrEdgeLost {
			online = false
		} else {
			actRet = ACTION_RET_STOPAPP_FAILD
			status = models.STATUS_PLAYING //TODO:
			//status = models.STATUS_FAIL
		}
	}

	t.updateStatus(status, &online, &actRet)

	return err
}

func (t *Edge) OnCloudXRConnect(ctx ctxcache.Context) error {
	//online := true
	//t.updateStatus(models.STATUS_RESERVE_XR_CONNECT, &online, nil)
	t.updateStatusWhen(models.STATUS_RESERVE_XR_NOT_CONNECT, models.STATUS_RESERVE_XR_CONNECT)
	return nil
}

func (t *Edge) updateStatus(status models.EdgeStatus, online *bool, actRet *ActionRet) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.info.Status = status

	if online != nil {
		t.info.Online = *online
	}

	if actRet != nil {
		t.actRet = *actRet
	}

}

func (t *Edge) updateStatusWhen(oriStatus, newStatus models.EdgeStatus) bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	// if !t.info.Online {// try it
	// 	return false
	// }

	if t.info.Status == newStatus {
		return true
	}

	if t.info.Status == oriStatus {
		t.info.Status = newStatus
		return true
	}

	return false
}

func (t *Edge) setOnline(online bool) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.info.Online = online
}

func (t *Edge) GetInfo() EdgeInfoStatus { //副本
	ret := EdgeInfoStatus{
		Edge:   t.info,
		ActRet: t.actRet,
	}
	return ret
}

// .. //
func (t *Edge) RegApps(appsID []uint) error {
	eRepo := repo.Edge{}
	_, err := eRepo.RegApps(t.info.ID, appsID)
	return err
}
