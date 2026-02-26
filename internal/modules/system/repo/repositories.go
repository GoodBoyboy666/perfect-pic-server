package repo

import (
	"gorm.io/gorm"
)

type UserStore interface {
	CountAll() (int64, error)
}

type ImageStore interface {
	CountAll() (int64, error)
	SumAllSize() (int64, error)
}

func NewSystemRepository(db *gorm.DB) SystemStore {
	return &SystemRepository{db: db}
}
