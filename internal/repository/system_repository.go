package repository

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type SystemRepository struct{}

var System = &SystemRepository{}

func (r *SystemRepository) InitializeSystem(settingValues map[string]string, admin *model.User) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
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
