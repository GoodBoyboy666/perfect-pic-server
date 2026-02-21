package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	User    UserStore
	Image   ImageStore
	Setting SettingStore
	System  SystemStore
}

func NewUserRepository(db *gorm.DB) UserStore {
	return &UserRepository{db: db}
}

func NewImageRepository(db *gorm.DB) ImageStore {
	return &ImageRepository{db: db}
}

func NewSettingRepository(db *gorm.DB) SettingStore {
	return &SettingRepository{db: db}
}

func NewSystemRepository(db *gorm.DB) SystemStore {
	return &SystemRepository{db: db}
}

func NewRepositories(user UserStore, image ImageStore, setting SettingStore, system SystemStore) *Repositories {
	return &Repositories{
		User:    user,
		Image:   image,
		Setting: setting,
		System:  system,
	}
}
