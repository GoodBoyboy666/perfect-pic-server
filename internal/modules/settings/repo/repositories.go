package repo

import "gorm.io/gorm"

func NewSettingRepository(db *gorm.DB) SettingStore {
	return &SettingRepository{db: db}
}
