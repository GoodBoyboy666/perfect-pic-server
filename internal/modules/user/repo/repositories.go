package repo

import (
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type ImageStore interface {
	FindUnscopedByUserID(userID uint) ([]model.Image, error)
}

func NewUserRepository(db *gorm.DB) UserStore {
	return &UserRepository{db: db}
}
