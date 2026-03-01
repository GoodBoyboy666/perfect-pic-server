package dto

import "perfect-pic-server/internal/model"

type PaginationRequest struct {
	Page     int
	PageSize int
}

type UserImageListRequest struct {
	PaginationRequest
	UserID   uint
	Filename string
	ID       *uint
}

type AdminImageListRequest struct {
	PaginationRequest
	Username string
	Filename string
	UserID   *uint
	ID       *uint
}

type ImageResponse struct {
	model.Image
	Username string `json:"username"`
}

type BatchDeleteImagesRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}
