//package repo
package mysql

import (
	"go-empty/pkg/db"
	"go-empty/pkg/models"

	"gorm.io/gorm"
)

type User struct {
}

func GetDB() *gorm.DB {
	return db.MainDB
}
func (t *User) CreateUser(user *models.User) (*models.User, error) {
	ddb := GetDB()
	out := &models.User{}
	*out = *user

	dbc := ddb.Where("account = ?",
		out.Account).
		FirstOrCreate(&out) //TODO:  return error if user is exist

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil

}

func (t *User) GetUser(account string) (*models.User, error) {
	ddb := GetDB()
	out := &models.User{}

	dbc := ddb.Where("account = ?",
		account).
		First(&out)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return out, nil

}
