package repo

import (
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type UserStore interface {
	FindByID(id uint) (*model.User, error)
	UpdateAvatar(user *model.User, filename string) error
	ClearAvatar(user *model.User) error
}

func NewImageRepository(db *gorm.DB) ImageStore {
	return &ImageRepository{db: db}
}
