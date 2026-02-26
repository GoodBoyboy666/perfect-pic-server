package handler

import systemservice "perfect-pic-server/internal/modules/system/service"

type Handler struct {
	systemService *systemservice.Service
}

func New(systemService *systemservice.Service) *Handler {
	return &Handler{systemService: systemService}
}
