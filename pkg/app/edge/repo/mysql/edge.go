//package repo
package mysql

import (
	"fmt"
	"go-empty/pkg/db"
	"go-empty/pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Edge struct {
}

func GetDB() *gorm.DB {
	return db.MainDB
}

func (t *Edge) GetEdges() ([]models.Edge, error) {
	ddb := GetDB()
	out := []models.Edge{}

	dbc := ddb.Find(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil
}

func (t *Edge) RegEdge(ip string, port int) (*models.Edge, error) {
	ddb := GetDB()
	out := models.Edge{
		IP:   ip,
		Port: port,
	}

	dbc := ddb.FirstOrCreate(&out,
		"ip = ? AND port = ?", ip, port)

	if dbc.Error != nil {
		return nil, dbc.Error
	}

	out.Online = true
	ddb.Model(&out).Updates(models.Edge{Online: out.Online})
	return &out, nil
}

// func (t *Edge) RegApps(edgeID uint, appsID []uint) ([]models.EdgeApp, error) {
// 	ddb := GetDB()

// 	var ret []models.EdgeApp
// 	for _, v := range appsID {
// 		eApp := models.EdgeApp{
// 			EdgeID: edgeID,
// 			AppID:  v,
// 			Valid:  true,
// 		}
// 		dbc := ddb.Where(models.EdgeApp{EdgeID: edgeID,
// 			AppID: v}).
// 			FirstOrCreate(&eApp)

// 		if dbc.Error != nil {
// 			continue
// 		}

// 		fmt.Printf("[REG APP] ID = %d eApp.EdgeID = %d  eApp.AppID =%d \n", eApp.ID, eApp.EdgeID, eApp.AppID)
// 		ret = append(ret, eApp)
// 	}

// 	return ret, nil
// }

func (t *Edge) RegApps(edgeID uint, appsID []uint) ([]models.EdgeApp, error) {
	ddb := GetDB()

	t.SetAppsValid(edgeID, false)

	var ret []models.EdgeApp
	for _, v := range appsID {
		eApp := models.EdgeApp{
			EdgeID: edgeID,
			AppID:  v,
			Valid:  true,
		}
		// dbc := ddb.Where(models.EdgeApp{EdgeID: edgeID,
		// 	AppID: v}).
		// 	FirstOrCreate(&eApp)

		//Update or Create
		// Update columns to default value on `edge_id` and  `app_id` conflict
		dbc := ddb.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "edge_id"}, {Name: "app_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"valid": true}),
		}).Create(&eApp)

		if dbc.Error != nil {
			continue
		}

		fmt.Printf("[REG APP] ID = %d eApp.EdgeID = %d  eApp.AppID =%d eApp.Valid = %t \n",
			eApp.ID, eApp.EdgeID, eApp.AppID, eApp.Valid)
		//fmt.Printf("[REG APP] %#v  \n", eApp)
		ret = append(ret, eApp)
	}

	return ret, nil
}

func (t *Edge) SetAppsValid(edgeID uint, valid bool) error {
	ddb := GetDB()
	// Update with conditions
	ddb.Model(&models.EdgeApp{}).Where("edge_id = ?", edgeID).
		Update("valid", valid)

	return ddb.Error
}

func (t *Edge) FindEdgesWithEdgeID(edgeID uint) ([]models.EdgeApp, error) {
	ddb := GetDB()
	out := []models.EdgeApp{}

	dbc := ddb.Where("edge_id = ?", edgeID).Find(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil

}

func (t *Edge) FindEdgesWithAppID(appID uint) ([]models.EdgeApp, error) {
	ddb := GetDB()
	out := []models.EdgeApp{}

	dbc := ddb.Where("app_id = ? and valid = 1", appID).Find(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil

}
