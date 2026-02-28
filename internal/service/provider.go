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
	dbConfig    config.DBConfig
}

type PasskeyService struct {
	dbConfig     config.DBConfig
	passkeyStore repo.PasskeyStore
}

type SettingsService struct {
	settingStore repo.SettingStore
	dbConfig     config.DBConfig
}

func NewAuthService(dbConfig *config.DBConfig) *AuthService {
	return &AuthService{dbConfig: *dbConfig}
}

func NewUserService(userStore repo.UserStore, dbConfig *config.DBConfig) *UserService {
	return &UserService{userStore: userStore, dbConfig: *dbConfig}
}

func NewImageService(imageStore repo.ImageStore, dbConfig *config.DBConfig) *ImageService {
	return &ImageService{imageStore: imageStore, dbConfig: *dbConfig}
}

func NewEmailService(dbConfig *config.DBConfig) *EmailService {
	return &EmailService{dbConfig: *dbConfig}
}

func NewCaptchaService(dbConfig *config.DBConfig) *CaptchaService {
	return &CaptchaService{dbConfig: *dbConfig}
}

func NewInitService(systemStore repo.SystemStore, dbConfig *config.DBConfig) *InitService {
	return &InitService{systemStore: systemStore, dbConfig: *dbConfig}
}

func NewPasskeyService(passkeyStore repo.PasskeyStore, dbConfig *config.DBConfig) *PasskeyService {
	return &PasskeyService{passkeyStore: passkeyStore, dbConfig: *dbConfig}
}

func NewSettingsService(settingStore repo.SettingStore, dbConfig *config.DBConfig) *SettingsService {
	return &SettingsService{settingStore: settingStore, dbConfig: *dbConfig}
}
