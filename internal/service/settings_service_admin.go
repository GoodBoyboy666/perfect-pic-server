package service

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type UpdateSettingPayload struct {
	Key   string
	Value string
}

// AdminListSettings 获取全部系统设置。
func AdminListSettings() ([]model.Setting, error) {
	var settings []model.Setting
	if err := db.DB.Find(&settings).Error; err != nil {
		return nil, err
	}

	maskSensitiveSettings(settings)
	return settings, nil
}

// AdminUpdateSettings 批量更新系统设置，并在成功后清理配置缓存。
func AdminUpdateSettings(items []UpdateSettingPayload) error {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			skip, err := shouldSkipMaskedSensitiveSettingUpdate(tx, item)
			if err != nil {
				return err
			}
			if skip {
				continue
			}

			if err := upsertSettingValue(tx, item); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	ClearCache()
	return nil
}
