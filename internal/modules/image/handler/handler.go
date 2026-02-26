package handler

import imageservice "perfect-pic-server/internal/modules/image/service"

type Handler struct {
	imageService *imageservice.Service
}

func New(imageService *imageservice.Service) *Handler {
	return &Handler{imageService: imageService}
}
