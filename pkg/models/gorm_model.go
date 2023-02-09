package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type GormModel struct {
	gorm.Model
}

func (t *GormModel) GetID() uint {
	return t.ID
}

func (t *GormModel) GetCreatedAt() time.Time {
	return t.CreatedAt
}

func (t *GormModel) GetUpdatedAt() time.Time {
	return t.UpdatedAt
}

func (t *GormModel) GetDeletedAt() *time.Time {
	return t.DeletedAt
}
