package service

import (
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
)

// AdminListImages 分页查询后台图片列表，支持按用户名、用户ID、文件名、图片 ID 过滤。
func AdminListImages(params AdminImageListParams) ([]model.Image, int64, int, int, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)

	images, total, err := repository.Image.AdminListImages(
		params.Username,
		params.Filename,
		params.UserID,
		params.ID,
		(page-1)*pageSize,
		pageSize,
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	return images, total, page, pageSize, nil
}

// AdminGetImageByID 获取后台指定图片。
func AdminGetImageByID(id uint) (*model.Image, error) {
	return repository.Image.FindByID(id)
}

// AdminGetImagesByIDs 按 ID 列表获取后台图片。
func AdminGetImagesByIDs(ids []uint) ([]model.Image, error) {
	return repository.Image.FindByIDs(ids)
}
