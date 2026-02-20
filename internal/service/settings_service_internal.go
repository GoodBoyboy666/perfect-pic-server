package service

import (
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

const maskedSettingValue = "**********"

// maskSensitiveSettings 对敏感配置值进行统一脱敏。
func maskSensitiveSettings(settings []model.Setting) {
	for i := range settings {
		if settings[i].Sensitive {
			settings[i].Value = maskedSettingValue
		}
	}
}

// shouldSkipMaskedSensitiveSettingUpdate 判断敏感配置在提交掩码值时是否应跳过更新。
func shouldSkipMaskedSensitiveSettingUpdate(tx *gorm.DB, item UpdateSettingPayload) (bool, error) {
	if item.Value != maskedSettingValue {
		return false, nil
	}

	var currentSetting model.Setting
	if err := tx.Where("key = ?", item.Key).First(&currentSetting).Error; err != nil {
		// 历史行为是查询失败时继续按普通更新逻辑处理，这里保持一致。
		return false, nil
	}

	return currentSetting.Sensitive, nil
}

// upsertSettingValue 按 key 更新配置值，不存在时创建新配置项。
func upsertSettingValue(tx *gorm.DB, item UpdateSettingPayload) error {
	setting := model.Setting{Key: item.Key, Value: item.Value}
	result := tx.Model(&setting).Select("Value").Updates(setting)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return tx.Create(&setting).Error
	}
	return nil
}
