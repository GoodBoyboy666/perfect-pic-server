package handler

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/usecase/admin"
	"perfect-pic-server/internal/usecase/app"
)

type AuthHandler struct {
	authService    *service.AuthService
	captchaService *service.CaptchaService
	authUseCase    *app.AuthUseCase
	initService    *service.InitService
	dbConfig       *config.DBConfig
	passkeyUseCase *app.PasskeyUseCase
}

type UserHandler struct {
	userService       *service.UserService
	userUseCase       *app.UserUseCase
	userManageUseCase *admin.UserManageUseCase
	imageService      *service.ImageService
	imageUseCase      *app.ImageUseCase
	authService       *service.AuthService
	passkeyService    *service.PasskeyService
	passkeyUseCase    *app.PasskeyUseCase
}

type ImageHandler struct {
	imageService *service.ImageService
	imageUseCase *app.ImageUseCase
}

type SystemHandler struct {
	initService *service.InitService
	statUseCase *admin.StatUseCase
	dbConfig    *config.DBConfig
	userService *service.UserService
}

type SettingsHandler struct {
	settingsService *service.SettingsService
	settingsUseCase *admin.SettingsUseCase
}

func NewAuthHandler(
	authService *service.AuthService,
	captchaService *service.CaptchaService,
	authUseCase *app.AuthUseCase,
	initService *service.InitService,
	dbConfig *config.DBConfig,
	passkeyUseCase *app.PasskeyUseCase,
) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		captchaService: captchaService,
		authUseCase:    authUseCase,
		initService:    initService,
		dbConfig:       dbConfig,
		passkeyUseCase: passkeyUseCase,
	}
}

func NewUserHandler(
	userService *service.UserService,
	userUseCase *app.UserUseCase,
	userManageUseCase *admin.UserManageUseCase,
	imageService *service.ImageService,
	imageUseCase *app.ImageUseCase,
	authService *service.AuthService,
	passkeyService *service.PasskeyService,
	passkeyUseCase *app.PasskeyUseCase,
) *UserHandler {
	return &UserHandler{
		userService:       userService,
		userUseCase:       userUseCase,
		userManageUseCase: userManageUseCase,
		imageService:      imageService,
		imageUseCase:      imageUseCase,
		authService:       authService,
		passkeyService:    passkeyService,
		passkeyUseCase:    passkeyUseCase,
	}
}

func NewImageHandler(imageService *service.ImageService, imageUseCase *app.ImageUseCase) *ImageHandler {
	return &ImageHandler{imageService: imageService, imageUseCase: imageUseCase}
}

func NewSystemHandler(
	initService *service.InitService,
	statUseCase *admin.StatUseCase,
	dbConfig *config.DBConfig,
	userService *service.UserService) *SystemHandler {
	return &SystemHandler{
		initService: initService,
		statUseCase: statUseCase,
		dbConfig:    dbConfig,
		userService: userService,
	}
}

func NewSettingsHandler(
	settingsService *service.SettingsService,
	settingsUseCase *admin.SettingsUseCase,
) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		settingsUseCase: settingsUseCase,
	}
}
