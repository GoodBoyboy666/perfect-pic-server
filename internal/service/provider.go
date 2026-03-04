package service

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/pkg/email"
	repo "perfect-pic-server/internal/repository"
	"sync"

	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	dbConfig *config.DBConfig
}

type UserService struct {
	userStore repo.UserStore
	dbConfig  *config.DBConfig
	redisDB   *redis.Client

	passwordResetStore      sync.Map
	passwordResetTokenStore sync.Map
	emailChangeStore        sync.Map
	emailChangeTokenStore   sync.Map
}

type ImageService struct {
	imageStore repo.ImageStore
	dbConfig   *config.DBConfig
}

type EmailService struct {
	dbConfig *config.DBConfig
	mailer   *email.Mailer
}

type InitService struct {
	systemStore repo.SystemStore
	dbConfig    *config.DBConfig
}

type PasskeyService struct {
	dbConfig     *config.DBConfig
	passkeyStore repo.PasskeyStore
	redisDB      *redis.Client
}

type SettingsService struct {
	settingStore repo.SettingStore
	dbConfig     *config.DBConfig
}

type CaptchaService struct {
	dbConfig *config.DBConfig
}

func NewAuthService(dbConfig *config.DBConfig) *AuthService {
	return &AuthService{dbConfig: dbConfig}
}

func NewUserService(userStore repo.UserStore, dbConfig *config.DBConfig, redisDB *redis.Client) *UserService {
	return &UserService{userStore: userStore, dbConfig: dbConfig, redisDB: redisDB}
}

func NewImageService(imageStore repo.ImageStore, dbConfig *config.DBConfig) *ImageService {
	return &ImageService{imageStore: imageStore, dbConfig: dbConfig}
}

func NewEmailService(dbConfig *config.DBConfig, mailer *email.Mailer) *EmailService {
	return &EmailService{dbConfig: dbConfig, mailer: mailer}
}

func NewInitService(systemStore repo.SystemStore, dbConfig *config.DBConfig) *InitService {
	return &InitService{systemStore: systemStore, dbConfig: dbConfig}
}

func NewPasskeyService(passkeyStore repo.PasskeyStore, dbConfig *config.DBConfig, redisDB *redis.Client) *PasskeyService {
	return &PasskeyService{passkeyStore: passkeyStore, dbConfig: dbConfig, redisDB: redisDB}
}

func NewSettingsService(settingStore repo.SettingStore, dbConfig *config.DBConfig) *SettingsService {
	return &SettingsService{settingStore: settingStore, dbConfig: dbConfig}
}

func NewCaptchaService(dbConfig *config.DBConfig) *CaptchaService {
	return &CaptchaService{dbConfig: dbConfig}
}
