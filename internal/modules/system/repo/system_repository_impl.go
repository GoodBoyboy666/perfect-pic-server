package repo

import (
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

func (r *SystemRepository) InitializeSystem(settingValues map[string]string, admin *model.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range settingValues {
			if err := tx.Model(&model.Setting{}).Where("key = ?", key).Update("value", value).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(admin).Error; err != nil {
			return err
		}
		return nil
	})
}
