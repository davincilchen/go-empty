package usecase

import (
	"fmt"
	"go-empty/pkg/app/ctxcache"
	repo "go-empty/pkg/app/device/repo/mysql"
	"go-empty/pkg/models"
	"sync"
	"time"

	edgeUCase "go-empty/pkg/app/edge/usecase"
	errDef "go-empty/pkg/app/errordef"
	userUCase "go-empty/pkg/app/user/usecase"
)

type DevStatus int

const (
	STATUS_FREE DevStatus = 0
	//STATUS_RESERVE_INIT           DevStatus = 110
	STATUS_RESERVE_XR_NOT_CONNECT DevStatus = 120
	STATUS_RESERVE_XR_CONNECT     DevStatus = 130
	//STATUS_RX_START_APP           DevStatus = 140
	STATUS_PLAYING DevStatus = 150
	// STATUS_RX_STOP_APP            DevStatus = 160
	// STATUS_RX_RELEASE             DevStatus = 170
)

var deviceRepo repo.Device

type DeviceLoginProc struct {
	IP     string
	Device models.Device
}

func NewDeviceLoginProc(Type int, UUID, IP string) *DeviceLoginProc {
	d := &DeviceLoginProc{
		IP: IP,
		Device: models.Device{
			Type: Type,
			UUID: UUID,
		},
	}
	return d
}

func (t *DeviceLoginProc) DevLoginSucess(ctx ctxcache.Context, user userUCase.LoginUser) error {

	//TODO: save ip and login/logout
	device, err := deviceRepo.RegDevice(&t.Device)
	if err != nil {
		ctx.CacheDBError(err)
		return err
	}

	loginDev := LoginDevice{
		ip:     t.IP,
		device: *device,
		user:   user,
	}

	manager := GetDeviceManager()
	return manager.Add(ctx, &loginDev)

}

// ============================================= //
type QLoginDeviceRet struct {
	Device models.Device
	User   userUCase.LoginUser
	appID  uint
	IP     string
}

func (t *QLoginDeviceRet) GetAppID() *uint {
	if t.appID == 0 {
		return nil
	}
	return &t.appID
}

type QLoginDeviceRetDetail struct {
	QLoginDeviceRet
	Edge *edgeUCase.EdgeInfoStatus
}

type LoginDevice struct {
	edgeMux sync.Mutex
	//每次呼叫edge的ctx會不同,不能在new的時候跟著綁
	edge   *edgeUCase.Edge //not nil when post reserve success
	device models.Device
	user   userUCase.LoginUser

	processMux sync.Mutex
	inProcess  bool

	statusMux sync.RWMutex
	status    DevStatus

	appMux sync.RWMutex
	appID  uint
	ip     string
}

func (t *LoginDevice) GetDeviceManager() *DeviceManager {
	return GetDeviceManager()
}

func (t *LoginDevice) GetEdgeManager() *edgeUCase.EdgeManager {
	return edgeUCase.GetEdgeManager()
}

func (t *LoginDevice) Logout(ctx ctxcache.Context) error {

	err := t.ReleaseReserve(ctx, true)
	if err == errDef.ErrInOldProcess {
		return err
	}
	manager := t.GetDeviceManager()
	manager.Delete(t)
	return nil
}

func (t *LoginDevice) ToProcess(do bool) bool {
	t.processMux.Lock()
	defer t.processMux.Unlock()

	if !do {
		t.inProcess = false
		return true
	}

	//do == true
	if t.inProcess {
		return false
	}

	t.inProcess = true
	return true
}

func (t *LoginDevice) IsInProcess() bool {
	t.processMux.Lock()
	defer t.processMux.Unlock()
	return t.inProcess
}

func (t *LoginDevice) NewReserve(ctx ctxcache.Context, appID uint) (*string, error) {

	if !t.ToProcess(true) {
		fmt.Println("#(LoginDevice) NewReserve [still in old process]: can not process ")
		return nil, errDef.ErrInOldProcess
	}

	fmt.Println("#(LoginDevice) NewReserve [start]: can process")

	//can process
	defer t.ToProcess(false)
	//time.Sleep(15 * time.Second)
	if t.IsReserve() {
		return nil, errDef.ErrRepeatedReserve
	}

	edgeManager := t.GetEdgeManager()
	edge, err := edgeManager.Reserve(ctx, appID)
	if err != nil {
		return nil, err
	}
	ctx.ResetHttpError()

	devM := t.GetDeviceManager()
	devM.reserveFor(edge.GetInfo().ID, t.device.UUID)
	t.AttachEdge(edge)
	e := edge.GetInfo()
	t.SetAppID(appID)
	fmt.Println("#(LoginDevice) NewReserve [success]")
	return &e.IP, nil
}

func (t *LoginDevice) ReleaseReserve(ctx ctxcache.Context, isLogout bool) error {

	try := 1000 //50*1000 = 50 sec
	if isLogout {
		try = 10000 //50*10000 = 500 sec
	}
	inOldProcess := true
	for i := 0; i < try; i++ {
		if t.ToProcess(true) {
			inOldProcess = false
			break
		}
		time.Sleep(50 * time.Millisecond)
		if i%50 == 0 {
			fmt.Println(time.Now(), i+1,
				"#(LoginDevice) ReleaseReserve wait for processing",
				"[ isLogout:", isLogout, "]")
		}

	}

	if inOldProcess {
		fmt.Println("#(LoginDevice) ReleaseReserve [still in old process]: can not process ",
			"[ isLogout:", isLogout, "]")
		return errDef.ErrInOldProcess
	}

	var edge *edgeUCase.Edge
	//can process
	fmt.Println("#(LoginDevice) ReleaseReserve [start]: can process ",
		"[ isLogout:", isLogout, "]")
	defer func() {
		t.ToProcess(false)
		edgeID := uint(0)
		if edge != nil {
			edgeID = edge.GetInfo().ID
		}
		devM := t.GetDeviceManager()
		devM.releseReserve(edgeID, t.device.UUID)
		//delete cache 會觸發OnEvicted(reserveTimeout)
		//會兩邊重複lock inProcess
		//所以先release edge和DetachEdge
		//OnEvicted 有加timeout flag,所以不會發生了
		fmt.Println("#(LoginDevice) ReleaseReserve [after function]",
			"[ isLogout:", isLogout, "]")
	}()

	t.statusMux.Lock()
	t.status = STATUS_FREE
	t.statusMux.Unlock()

	edge = t.getEdge()
	if edge == nil {
		return errDef.ErrDevNoReserve
	}

	edge.ReleaseReserve(ctx)
	t.DetachEdge()

	t.SetAppID(0)

	fmt.Println("#(LoginDevice) ReleaseReserve [success]",
		"[ isLogout:", isLogout, "]")
	return nil
}

func (t *LoginDevice) StartApp(ctx ctxcache.Context) error {
	if !t.ToProcess(true) {
		fmt.Println("#(LoginDevice) StartApp [still in old process]: can not process ")
		return errDef.ErrInOldProcess
	}

	//can process
	fmt.Println("#(LoginDevice) StartApp [start]: can process ")
	defer t.ToProcess(false)
	//time.Sleep(15 * time.Second)
	edge := t.getEdge()
	if edge == nil {
		return errDef.ErrDevNoReserve
	}
	appID := t.GetAppID()
	if appID == nil {
		return errDef.ErrNoResource
	}

	err := edge.StartAPP(ctx, *appID)
	if err != nil {
		return err
	}
	fmt.Println("#(LoginDevice) StartApp [success]: can process ")
	return nil
}

func (t *LoginDevice) StopApp(ctx ctxcache.Context) error {

	if !t.ToProcess(true) {
		fmt.Println("#(LoginDevice) StopApp [still in old process]: can not process ")
		return errDef.ErrInOldProcess
	}

	//can process
	fmt.Println("#(LoginDevice) StopApp [start]: can process ")
	defer t.ToProcess(false)
	//time.Sleep(15 * time.Second)
	edge := t.getEdge()
	if edge == nil {
		return errDef.ErrDevNoReserve
	}
	err := edge.StopAPP(ctx)

	if err != nil {
		return err
	}
	fmt.Println("#(LoginDevice) StopApp [success]")
	return nil
}

func (t *LoginDevice) Resume(ctx ctxcache.Context) error {
	edge := t.getEdge()
	if edge == nil {
		return errDef.ErrDevNoReserve
	}
	return edge.Resume(ctx)
}

func (t *LoginDevice) UpdateStatus(ctx ctxcache.Context, status DevStatus) error {
	edge := t.getEdge()
	if edge == nil {
		return errDef.ErrDevNoReserve
	}
	if status == STATUS_RESERVE_XR_CONNECT {
		edge.OnCloudXRConnect(ctx) //TODO: double check
	}
	if status != STATUS_FREE {
		devM := t.GetDeviceManager()
		devM.Alive(t.device.UUID)
	}
	//TODO: 當以連線狀態下 丟未連線 判斷是否真的未連線

	t.statusMux.Lock()
	defer t.statusMux.Unlock()
	t.status = status

	return nil
}

func (t *LoginDevice) IsReserve() bool {
	t.edgeMux.Lock()
	defer t.edgeMux.Unlock()

	return t.edge != nil
}

func (t *LoginDevice) AttachEdge(edge *edgeUCase.Edge) {
	t.edgeMux.Lock()
	defer t.edgeMux.Unlock()

	t.edge = edge
}

func (t *LoginDevice) DetachEdge() {
	t.edgeMux.Lock()
	defer t.edgeMux.Unlock()

	t.edge = nil
}

func (t *LoginDevice) getEdge() *edgeUCase.Edge {
	t.edgeMux.Lock()
	defer t.edgeMux.Unlock()
	if t.edge == nil {
		return nil
	}
	return t.edge
}

func (t *LoginDevice) GetEdgeInfo() *edgeUCase.EdgeInfoStatus {
	t.edgeMux.Lock()
	defer t.edgeMux.Unlock()

	if t.edge == nil {
		return nil
	}

	e := t.edge.GetInfo()
	return &e
}

func (t *LoginDevice) GetDeviceInfo() QLoginDeviceRet {
	return QLoginDeviceRet{
		User:   t.user,
		Device: t.device,
		appID:  t.appID,
		IP:     t.ip,
	}
}

func (t *LoginDevice) GetAppID() *uint {
	t.appMux.RLock()
	defer t.appMux.RUnlock()
	if t.appID == 0 {
		return nil
	}

	return &t.appID
}

func (t *LoginDevice) SetAppID(appID uint) {
	t.appMux.Lock()
	defer t.appMux.Unlock()
	t.appID = appID
}
