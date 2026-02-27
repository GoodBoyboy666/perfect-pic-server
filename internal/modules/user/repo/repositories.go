package repo

import (
	"gorm.io/gorm"
)

func NewUserRepository(db *gorm.DB) UserStore {
	return &UserRepository{db: db}
}
