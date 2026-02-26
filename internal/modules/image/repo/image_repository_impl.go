package repo

import (
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func (r *ImageRepository) CreateAndIncreaseUserStorage(image *model.Image, userID uint, size int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(image).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("storage_used", gorm.Expr("storage_used + ?", size)).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ImageRepository) DeleteAndDecreaseUserStorage(image *model.Image) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(image).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.User{}).Where("id = ?", image.UserID).
			UpdateColumn("storage_used", gorm.Expr("storage_used - ?", image.Size)).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ImageRepository) BatchDeleteAndDecreaseUserStorage(imageIDs []uint, userSizeMap map[uint]int64) error {
	if len(imageIDs) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ?", imageIDs).Delete(&model.Image{}).Error; err != nil {
			return err
		}
		for uid, size := range userSizeMap {
			if err := tx.Model(&model.User{}).Where("id = ?", uid).
				UpdateColumn("storage_used", gorm.Expr("storage_used - ?", size)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ImageRepository) ListImages(params ListImagesParams) ([]model.Image, int64, error) {
	var images []model.Image
	var total int64

	query := r.db.Model(&model.Image{})
	if params.UserID != nil {
		query = query.Where("images.user_id = ?", *params.UserID)
	}
	if params.Username != "" {
		query = query.Joins("JOIN users ON users.id = images.user_id").
			Where("users.username LIKE ?", "%"+params.Username+"%")
	}
	if params.Filename != "" {
		query = query.Where("images.filename LIKE ?", "%"+params.Filename+"%")
	}
	if params.ID != nil {
		query = query.Where("images.id = ?", *params.ID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.PreloadUser {
		query = query.Preload("User")
	}

	if err := query.Order("images.id desc").Offset(params.Offset).Limit(params.Limit).Find(&images).Error; err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

func (r *ImageRepository) ListUserImages(
	userID uint,
	filename string,
	id *uint,
	offset int,
	limit int,
) ([]model.Image, int64, error) {
	return r.ListImages(ListImagesParams{
		UserID:   &userID,
		Filename: filename,
		ID:       id,
		Offset:   offset,
		Limit:    limit,
	})
}

func (r *ImageRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Image{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ImageRepository) FindByIDAndUserID(imageID uint, userID uint) (*model.Image, error) {
	var image model.Image
	if err := r.db.Where("id = ? AND user_id = ?", imageID, userID).First(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) FindByIDsAndUserID(ids []uint, userID uint) ([]model.Image, error) {
	var images []model.Image
	if err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *ImageRepository) AdminListImages(
	username string,
	filename string,
	userID *uint,
	id *uint,
	offset int,
	limit int,
) ([]model.Image, int64, error) {
	return r.ListImages(ListImagesParams{
		UserID:      userID,
		Username:    username,
		Filename:    filename,
		ID:          id,
		Offset:      offset,
		Limit:       limit,
		PreloadUser: true,
	})
}

func (r *ImageRepository) FindByID(id uint) (*model.Image, error) {
	var image model.Image
	if err := r.db.First(&image, id).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) FindByIDs(ids []uint) ([]model.Image, error) {
	var images []model.Image
	if err := r.db.Where("id IN ?", ids).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *ImageRepository) FindUnscopedByUserID(userID uint) ([]model.Image, error) {
	var images []model.Image
	if err := r.db.Unscoped().Where("user_id = ?", userID).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *ImageRepository) CountAll() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Image{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ImageRepository) SumAllSize() (int64, error) {
	var total int64
	if err := r.db.Model(&model.Image{}).Select("COALESCE(SUM(size), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
