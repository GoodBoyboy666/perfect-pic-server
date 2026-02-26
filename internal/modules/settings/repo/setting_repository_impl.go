package repo

import (
	"fmt"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func (r *SettingRepository) InitializeDefaults(defaults []model.Setting) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, def := range defaults {
			var count int64
			if err := tx.Model(&model.Setting{}).Where("key = ?", def.Key).Count(&count).Error; err != nil {
				return fmt.Errorf("count default setting %q failed: %w", def.Key, err)
			}
			if count == 0 {
				if err := tx.Create(&def).Error; err != nil {
					return fmt.Errorf("create default setting %q failed: %w", def.Key, err)
				}
			} else {
				if err := tx.Model(&model.Setting{}).Where("key = ?", def.Key).Updates(map[string]interface{}{
					"category":  def.Category,
					"desc":      def.Desc,
					"sensitive": def.Sensitive,
				}).Error; err != nil {
					return fmt.Errorf("update default setting metadata %q failed: %w", def.Key, err)
				}
			}
		}
		return nil
	})
}

func (r *SettingRepository) DeleteNotInKeys(allowedKeys []string) error {
	query := r.db.Model(&model.Setting{})
	if len(allowedKeys) == 0 {
		return query.Where("1 = 1").Delete(&model.Setting{}).Error
	}
	return query.Where("key NOT IN ?", allowedKeys).Delete(&model.Setting{}).Error
}

func (r *SettingRepository) FindByKey(key string) (*model.Setting, error) {
	var setting model.Setting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *SettingRepository) Create(setting *model.Setting) error {
	return r.db.Create(setting).Error
}

func (r *SettingRepository) FindAll() ([]model.Setting, error) {
	var settings []model.Setting
	if err := r.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingRepository) UpdateSettings(items []UpdateSettingItem, maskedValue string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			skip, err := shouldSkipMaskedSensitiveSettingUpdate(tx, item, maskedValue)
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
}

func shouldSkipMaskedSensitiveSettingUpdate(tx *gorm.DB, item UpdateSettingItem, maskedValue string) (bool, error) {
	if item.Value != maskedValue {
		return false, nil
	}

	var currentSetting model.Setting
	if err := tx.Where("key = ?", item.Key).First(&currentSetting).Error; err != nil {
		return false, nil
	}

	return currentSetting.Sensitive, nil
}

func upsertSettingValue(tx *gorm.DB, item UpdateSettingItem) error {
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
