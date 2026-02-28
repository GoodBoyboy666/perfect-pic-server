package repository

import "gorm.io/gorm"

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

func NewPasskeyRepository(db *gorm.DB) PasskeyStore {
	return &PasskeyRepository{db: db}
}
