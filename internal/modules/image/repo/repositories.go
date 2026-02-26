package repo

import (
	"gorm.io/gorm"
)

func NewImageRepository(db *gorm.DB) ImageStore {
	return &ImageRepository{db: db}
}
