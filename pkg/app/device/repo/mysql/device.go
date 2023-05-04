//package repo
package mysql

import (
	"go-empty/pkg/db"
	"go-empty/pkg/models"

	"gorm.io/gorm"
)

type Device struct {
}

func GetDB() *gorm.DB {
	return db.MainDB
}
func (t *Device) RegDevice(dev *models.Device) (*models.Device, error) {
	ddb := GetDB()
	out := &models.Device{}
	*out = *dev

	dbc := ddb.Where("UUID = ?",
		out.UUID).
		FirstOrCreate(&out) //TODO:  return error if user is exist

	if dbc.Error != nil {
		return nil, dbc.Error
	}

	return out, nil

}

func (t *Device) GetDev(UUID string) (*models.Device, error) {
	ddb := GetDB()
	out := &models.Device{}

	dbc := ddb.Where("UUID = ?",
		UUID).
		First(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil

}
