package repo

import (
	"gorm.io/gorm"
)

func NewSystemRepository(db *gorm.DB) SystemStore {
	return &SystemRepository{db: db}
}
