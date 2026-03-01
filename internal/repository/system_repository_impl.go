package repository

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

func (r *SystemRepository) InitializeSystem(settingValues map[string]string, admin *model.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		claim := tx.Model(&model.Setting{}).
			Where("key = ? AND value = ?", consts.ConfigAllowInit, "true").
			Update("value", "false")
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected == 0 {
			return ErrSystemAlreadyInitialized
		}

		for key, value := range settingValues {
			// Skip ConfigAllowInit as it was already updated in the optimistic lock claim above (lines 16-18).
			if key == consts.ConfigAllowInit {
				continue
			}
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
