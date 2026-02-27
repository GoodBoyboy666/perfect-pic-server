package service

import (
	"errors"
	platformservice "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

// AdminListImages 分页查询后台图片列表，支持按用户名、用户ID、文件名、图片 ID 过滤。
func (s *Service) AdminListImages(params moduledto.AdminImageListRequest) ([]model.Image, int64, int, int, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)

	images, total, err := s.imageStore.AdminListImages(
		params.Username,
		params.Filename,
		params.UserID,
		params.ID,
		(page-1)*pageSize,
		pageSize,
	)
	if err != nil {
		return nil, 0, page, pageSize, platformservice.NewInternalError("获取图片列表失败")
	}

	return images, total, page, pageSize, nil
}

// AdminGetImageByID 获取后台指定图片。
func (s *Service) AdminGetImageByID(id uint) (*model.Image, error) {
	image, err := s.imageStore.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, platformservice.NewNotFoundError("图片不存在")
		}
		return nil, platformservice.NewInternalError("查找图片失败")
	}
	return image, nil
}

// AdminGetImagesByIDs 按 ID 列表获取后台图片。
func (s *Service) AdminGetImagesByIDs(ids []uint) ([]model.Image, error) {
	images, err := s.imageStore.FindByIDs(ids)
	if err != nil {
		return nil, platformservice.NewInternalError("查找图片失败")
	}
	return images, nil
}
