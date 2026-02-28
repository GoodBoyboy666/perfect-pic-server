package service

import (
	"perfect-pic-server/internal/config"
	repo "perfect-pic-server/internal/repository"
)

type AuthService struct {
	dbConfig config.DBConfig
}

type UserService struct {
	userStore repo.UserStore
	dbConfig  config.DBConfig
}

type ImageService struct {
	imageStore repo.ImageStore
	dbConfig   config.DBConfig
}

type EmailService struct {
	dbConfig config.DBConfig
}

type CaptchaService struct {
	dbConfig config.DBConfig
}

type InitService struct {
	systemStore repo.SystemStore
	dbConfig config.DBConfig
}

type PasskeyService struct {
	dbConfig config.DBConfig
	passkeyStore repo.PasskeyStore
}

type SettingsService struct {
	settingStore repo.SettingStore
	dbConfig     config.DBConfig
}
