package handler

import "perfect-pic-server/internal/service"

type AuthHandler struct {
	authService *service.Service
}

type UserHandler struct {
	userService  *service.Service
	imageService *service.Service
	authService  *service.Service
}

type ImageHandler struct {
	imageService *service.Service
}

type SystemHandler struct {
	systemService *service.Service
}

type SettingsHandler struct {
	settingsService *service.Service
}

func NewAuthHandler(appService *service.Service) *AuthHandler {
	return &AuthHandler{authService: appService}
}

func NewUserHandler(appService *service.Service) *UserHandler {
	return &UserHandler{
		userService:  appService,
		imageService: appService,
		authService:  appService,
	}
}

func NewImageHandler(appService *service.Service) *ImageHandler {
	return &ImageHandler{imageService: appService}
}

func NewSystemHandler(appService *service.Service) *SystemHandler {
	return &SystemHandler{systemService: appService}
}

func NewSettingsHandler(appService *service.Service) *SettingsHandler {
	return &SettingsHandler{settingsService: appService}
}
