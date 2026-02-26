package handler

import (
	authservice "perfect-pic-server/internal/modules/auth/service"
	imageservice "perfect-pic-server/internal/modules/image/service"
	userservice "perfect-pic-server/internal/modules/user/service"
)

type Handler struct {
	userService  *userservice.Service
	authService  *authservice.Service
	imageService *imageservice.Service
}

func New(
	userService *userservice.Service,
	authService *authservice.Service,
	imageService *imageservice.Service,
) *Handler {
	return &Handler{
		userService:  userService,
		authService:  authService,
		imageService: imageService,
	}
}
