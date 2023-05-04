package usecase

import (
	"fmt"
	"go-empty/pkg/app/ctxcache"
	"sync"
	"time"

	errDef "go-empty/pkg/app/errordef"

	cache "github.com/patrickmn/go-cache"
)

type DeviceManager struct {
	userDeviceMap      map[uint]*LoginDevice   //KEY: userID
	uuidDeviceMap      map[string]*LoginDevice //KEY: UUID
	tokenDeviceMap     map[string]*LoginDevice //KEY: Token
	edgeIDtoDevUUIDMap map[uint]string         //KEY: edgeID
	mux                sync.RWMutex

	uuidCache *cache.Cache //uuid , beTimeout
}

var deviceManager *DeviceManager

var dafaultKeepAliveInterval time.Duration = 5 * time.Minute
var dafaultCleanAliveInterval time.Duration = 1 * time.Minute

// var dafaultKeepAliveInterval time.Duration = 1 * time.Minute
// var dafaultCleanAliveInterval time.Duration = 1 * time.Second

// var dafaultKeepAliveInterval time.Duration = 10 * time.Second
// var dafaultCleanAliveInterval time.Duration = 11 * time.Second

func newDeviceManager() *DeviceManager {
	d := &DeviceManager{}
	d.userDeviceMap = make(map[uint]*LoginDevice)
	d.uuidDeviceMap = make(map[string]*LoginDevice)
	d.tokenDeviceMap = make(map[string]*LoginDevice)
	d.edgeIDtoDevUUIDMap = make(map[uint]string)
	d.uuidCache = cache.New(dafaultKeepAliveInterval,
		dafaultCleanAliveInterval)
	d.uuidCache.OnEvicted(d.reserveEvicted)
	return d
}

func GetDeviceManager() *DeviceManager {
	if deviceManager == nil {
		deviceManager = newDeviceManager()

	}
	return deviceManager
}

func (t *DeviceManager) Add(ctx ctxcache.Context, dev *LoginDevice) error {

	var oldDev *LoginDevice

	t.mux.Lock()
	defer func() {
		t.mux.Unlock()
		if oldDev != nil {
			oldDev.Logout(ctx)
		}

		//release完再加,避免被清掉
		t.mux.Lock()
		t.userDeviceMap[dev.user.ID] = dev
		t.tokenDeviceMap[dev.user.Token] = dev
		t.uuidDeviceMap[dev.device.UUID] = dev
		t.mux.Unlock()

		//TODO: close
		for k, v := range t.userDeviceMap {
			fmt.Printf("userDeviceMap user %d, uuid %s\n", k, v.device.UUID)
		}
		fmt.Println()
		for k, v := range t.tokenDeviceMap {
			fmt.Printf("tokenDeviceMap token %s, uuid %s, player %d\n", k, v.device.UUID, dev.user.ID)
		}
		fmt.Println()
		for k, v := range t.uuidDeviceMap {
			fmt.Printf("uuidDeviceMap uuid %s, uuid %s, player %d\n", k, v.device.UUID, dev.user.ID)
		}
		fmt.Println()
	}()

	_, ok := t.tokenDeviceMap[dev.user.Token]
	if ok {
		return errDef.ErrRepeatedLogin //請先登出
	}

	tmpDev, ok := t.userDeviceMap[dev.user.ID]
	if ok { //同樣的帳號,可能同裝置或不同裝置
		oldDev = tmpDev
		//return errDef.ErrRepeatedLogin //請先登出
	}

	tmpDev, ok = t.uuidDeviceMap[dev.device.UUID]
	if ok { //同裝置,可能同人或不同人
		oldDev = tmpDev
		//return errDef.ErrRepeatedLogin //請先登出
	}

	return nil
}

func (t *DeviceManager) Alive(uuid string) {
	_, ok := t.uuidCache.Get(uuid)
	if !ok {
		return
	}
	needTimeout := true
	t.uuidCache.Set(uuid, needTimeout, cache.DefaultExpiration)
}

func (t *DeviceManager) reserveEvicted(uuid string, value interface{}) {

	timeout := false
	if tmp, ok := value.(bool); ok {
		timeout = tmp
	}
	fmt.Println(time.Now(), " [reserveEvicted] [ timeout:", timeout, "] for uuid:", uuid)
	if !timeout {
		return
	}

	edgeID := uint(0)
	edgeIP := ""
	devID := uint(0)
	t.mux.Lock()
	dev, ok := t.uuidDeviceMap[uuid] //!ok if logout
	t.mux.Unlock()

	if ok && dev != nil {
		devID = dev.device.ID
		edge := dev.GetEdgeInfo()
		if edge != nil {
			edgeID = edge.ID
			edgeIP = edge.IP
		}
		ctx := ctxcache.NewContextLogger("reserveEvicted")
		dev.ReleaseReserve(ctx, false)
	}
	fmt.Println(time.Now(), " [reserveEvicted] edge_id:", edgeID,
		",IP:", edgeIP,
		",dev_id:", devID,
		", ", uuid)
}

func (t *DeviceManager) reserveFor(edgeID uint, devUUID string) error {
	needTimeout := true
	t.uuidCache.Set(devUUID, needTimeout, cache.DefaultExpiration)

	t.mux.Lock()
	defer t.mux.Unlock()
	t.edgeIDtoDevUUIDMap[edgeID] = devUUID

	return nil
}

func (t *DeviceManager) releseReserve(edgeID uint, uuid string) error {

	needTimeout := false
	_, ok := t.uuidCache.Get(uuid)
	if ok { //set for OnEvicted check
		t.uuidCache.Set(uuid, needTimeout, cache.DefaultExpiration)
	}

	t.uuidCache.Delete(uuid)

	t.mux.Lock()
	defer t.mux.Unlock()

	delete(t.edgeIDtoDevUUIDMap, edgeID)
	return nil
}

func (t *DeviceManager) GetByUUID(uuid string) *LoginDevice {

	t.mux.RLock()
	defer t.mux.RUnlock()

	dev, ok := t.uuidDeviceMap[uuid]
	if ok {
		return dev
	}
	return nil
}

func (t *DeviceManager) GetByToken(token string) *LoginDevice {

	t.mux.RLock()
	defer t.mux.RUnlock()

	dev, ok := t.tokenDeviceMap[token]
	if ok {
		return dev
	}
	return nil
}

func (t *DeviceManager) Delete(dev *LoginDevice) {

	t.mux.Lock()
	defer t.mux.Unlock()

	delete(t.userDeviceMap, dev.user.ID)
	delete(t.tokenDeviceMap, dev.user.Token)
	delete(t.uuidDeviceMap, dev.device.UUID)

	if dev.edge != nil {
		delete(t.edgeIDtoDevUUIDMap, dev.edge.GetInfo().ID)
	}
}

func (t *DeviceManager) GetDevInfoWithEdge(edgeID uint) *QLoginDeviceRet {
	t.mux.Lock()
	defer t.mux.Unlock()

	uuid, ok := t.edgeIDtoDevUUIDMap[edgeID]
	if !ok {
		return nil
	}

	dev, ok := t.uuidDeviceMap[uuid]
	if !ok {
		return nil
	}

	ret := dev.GetDeviceInfo()
	return &ret

}

func (t *DeviceManager) GetDevices() []QLoginDeviceRetDetail {

	devs := make([]*LoginDevice, 0)
	t.mux.Lock()
	for _, v := range t.uuidDeviceMap {
		devs = append(devs, v)
	}
	t.mux.Unlock()

	ret := make([]QLoginDeviceRetDetail, 0)
	for _, v := range devs {
		tmp := QLoginDeviceRetDetail{
			QLoginDeviceRet: v.GetDeviceInfo(),
			Edge:            v.GetEdgeInfo(),
		}
		ret = append(ret, tmp)
	}

	return ret

}
