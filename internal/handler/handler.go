package handler

import "perfect-pic-server/internal/service"

type Handler struct {
	service *service.AppService
}

func NewHandler(appService *service.AppService) *Handler {
	return &Handler{service: appService}
}
