package usecase

import (
	"fmt"
	"go-empty/pkg/app/ctxcache"
	repo "go-empty/pkg/app/edge/repo/mysql"
	"go-empty/pkg/app/errordef"
	"go-empty/pkg/models"
	"sync"
)

type EdgeManager struct {
	//TODO: lock
	edges   []*Edge
	edgeMap map[uint]*Edge //edge.ID
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

func (t *EdgeManager) findEdgeAppsWithAppID(appID uint) ([]models.EdgeApp, error) {
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

func (t *EdgeManager) GetAppsOfEdge(edgeID uint) ([]models.EdgeApp, error) {
	return t.findEdgeAppsWithEdgeID(edgeID)
}

func (t *EdgeManager) FindUnusedEdgesWithAppID(appID uint) ([]*Edge, error) {

	eapp, err := t.findEdgeAppsWithAppID(appID)

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
		fmt.Printf("[Reserve List for APP = %d] [status = %d] [valid = %t] IP:%s : %d \n", appID, info.Status, info.Valid, e.info.IP, e.info.Port)
		// if !info.Online { //try it
		// 	continue
		// }
		if info.Status != models.STATUS_FREE {
			continue
		}

		if !info.Valid {
			continue
		}
		edges = append(edges, e)
		fmt.Printf("[Reserve List for APP = %d] [Wait to try] IP:%s : %d \n", appID, e.info.IP, e.info.Port)
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
