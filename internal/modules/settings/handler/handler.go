package handler

import settingsservice "perfect-pic-server/internal/modules/settings/service"

type Handler struct {
	settingsService *settingsservice.Service
}

func New(settingsService *settingsservice.Service) *Handler {
	return &Handler{settingsService: settingsService}
}
