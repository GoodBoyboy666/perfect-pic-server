package dto

import "perfect-pic-server/internal/model"

type PaginationRequest struct {
	Page     int
	PageSize int
}

type ImageResponse struct {
	model.Image
	Username string `json:"username"`
}

type BatchDeleteImagesRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

type ListImagesRequest struct {
	PaginationRequest
	UserID      *uint
	Username    string
	Filename    string
	ID          *uint
	PreloadUser bool
}
