package handler

import authservice "perfect-pic-server/internal/modules/auth/service"

type Handler struct {
	authService *authservice.Service
}

func New(authService *authservice.Service) *Handler {
	return &Handler{authService: authService}
}
