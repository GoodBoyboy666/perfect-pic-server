package service

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

// AdminListImages 分页查询后台图片列表，支持按用户名与图片 ID 过滤。
func AdminListImages(params AdminImageListParams) ([]model.Image, int64, int, int, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)

	var total int64
	var images []model.Image

	query := db.DB.Model(&model.Image{})
	if params.Username != "" {
		query = query.Joins("JOIN users ON users.id = images.user_id").Where("users.username LIKE ?", "%"+params.Username+"%")
	}
	if params.ID != "" {
		query = query.Where("images.id = ?", params.ID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	if err := query.Preload("User").Order("images.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&images).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	return images, total, page, pageSize, nil
}

// AdminGetImageByID 获取后台指定图片。
func AdminGetImageByID(id string) (*model.Image, error) {
	var image model.Image
	if err := db.DB.First(&image, id).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// AdminGetImagesByIDs 按 ID 列表获取后台图片。
func AdminGetImagesByIDs(ids []uint) ([]model.Image, error) {
	var images []model.Image
	if err := db.DB.Where("id IN ?", ids).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}
