//package repo
package mysql

import (
	"xr-central/pkg/db"
	"xr-central/pkg/models"

	"gorm.io/gorm"
)

type AppGenre struct {
}

func GetDB() *gorm.DB {
	return db.MainDB
}
func (t *AppGenre) RegType(dev *models.AppGenre) (*models.AppGenre, error) {
	ddb := GetDB()
	out := &models.AppGenre{}
	*out = *dev

	dbc := ddb.FirstOrCreate(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}

	return out, nil

}

func (t *AppGenre) Get(id uint) (*models.AppGenre, error) {

	ddb := GetDB()
	out := &models.AppGenre{}
	out.ID = id
	dbc := ddb.
		First(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil

}
