package usecase

import (
	"fmt"
	"sync"
	"xr-central/pkg/app/ctxcache"
	repo "xr-central/pkg/app/edge/repo/mysql"
	"xr-central/pkg/app/errordef"
	"xr-central/pkg/models"
)

type EdgeManager struct {
	//TODO: lock
	edges   []*Edge
	edgeMap map[uint]*Edge
	mux     sync.Mutex
}

var manager *EdgeManager

func newEdgeManager() *EdgeManager {
	d := &EdgeManager{}

	return d
}

func GetEdgeManager() *EdgeManager {
	if manager == nil {
		manager = newEdgeManager()
		manager.edges = make([]*Edge, 0)
		manager.edgeMap = make(map[uint]*Edge)
		eRepo := repo.Edge{}
		es, err := eRepo.GetEdges()
		if err != nil {
			fmt.Printf("LoadEdges error %s\n", err.Error())
		} else {
			fmt.Println("========= LoadEdges Start =========")
			fmt.Printf("LoadEdges count %d\n", len(es))
			for i, v := range es {
				fmt.Printf("%d %#v\n", i, v)
				manager.addEdge(v)
			}
			fmt.Println("========= LoadEdges Done =========")
		}

	}
	return manager
}

func (t *EdgeManager) addEdge(edge models.Edge) *Edge {
	t.mux.Lock()
	defer t.mux.Unlock()

	edgeOld, ok := t.edgeMap[edge.ID]
	if ok {
		edgeOld.info.Online = true
		return edgeOld
	}

	tmpEdge := NewEdge(edge)
	t.edgeMap[edge.ID] = tmpEdge
	t.edges = append(manager.edges, tmpEdge)

	// edgeOld, ok := manager.edgeMap[edge.ID]
	// if ok {
	// 	edgeOld.info.Online = true
	// 	return edgeOld
	// }

	// tmpEdge := NewEdge(edge)
	// manager.edgeMap[edge.ID] = tmpEdge
	// manager.edges = append(manager.edges, tmpEdge)

	return tmpEdge
}

func (t *EdgeManager) getEdge(id uint) *Edge {

	t.mux.Lock()
	defer t.mux.Unlock()

	e, ok := t.edgeMap[id]
	if !ok {
		return nil
	}
	return e
}

func (t *EdgeManager) findEdgeAppsWithEdgeID(edgeID uint) ([]models.EdgeApp, error) {

	eRepo := repo.Edge{}
	edge_app, err := eRepo.FindEdgesWithEdgeID(edgeID)

	if err != nil {
		return nil, err
	}
	return edge_app, nil
}

func (t *EdgeManager) findEdgeApp(appID uint) ([]models.EdgeApp, error) {
	eRepo := repo.Edge{}
	edge_app, err := eRepo.FindEdgesWithAppID(appID)

	if err != nil {
		return nil, err
	}
	return edge_app, nil
}

func (t *EdgeManager) Reserve(ctx ctxcache.Context, appID uint) (*Edge, error) {
	elist, err := t.FindUnusedEdgesWithAppID(appID)
	if err != nil {
		return nil, err
	}

	var edge *Edge

	for _, v := range elist {
		err := v.Reserve(ctx, appID)
		if err != nil {
			continue
		}
		edge = v
		break
	}

	if edge == nil {
		return nil, errordef.ErrNoResource
	}

	return edge, nil
}

func (t *EdgeManager) FindUnusedEdgesWithAppID(appID uint) ([]*Edge, error) {

	eapp, err := t.findEdgeApp(appID)

	if err != nil {
		return nil, err
	}

	edges := make([]*Edge, 0)
	for _, v := range eapp {

		e := t.getEdge(v.EdgeID)
		if e == nil {
			continue
		}

		info := e.GetInfo()
		if !info.Online {
			continue
		}
		if info.Status != models.STATUS_FREE {
			continue
		}
		edges = append(edges, e)
		fmt.Printf("[Reserve List for APP = %d] IP:%s : %d \n", appID, e.info.IP, e.info.Port)
	}

	return edges, nil
}

func (t *EdgeManager) GetEdgeList() []EdgeInfoStatus {

	ret := make([]EdgeInfoStatus, 0)
	for _, v := range t.edges {
		ret = append(ret, v.GetInfo())
	}
	return ret
}

func (t *EdgeManager) RegEdge(ip string, port int) (*Edge, error) {
	eRepo := repo.Edge{}
	edge, err := eRepo.RegEdge(ip, port)
	if err != nil {
		return nil, err
	}
	edgeUse := t.addEdge(*edge)

	return edgeUse, nil
}
