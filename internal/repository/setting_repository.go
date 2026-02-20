package repository

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type SettingRepository struct{}

type UpdateSettingItem struct {
	Key   string
	Value string
}

var Setting = &SettingRepository{}

func (r *SettingRepository) InitializeDefaults(defaults []model.Setting) {
	for _, def := range defaults {
		var count int64
		db.DB.Model(&model.Setting{}).Where("key = ?", def.Key).Count(&count)
		if count == 0 {
			db.DB.Create(&def)
		} else {
			db.DB.Model(&model.Setting{}).Where("key = ?", def.Key).Updates(map[string]interface{}{
				"category":  def.Category,
				"desc":      def.Desc,
				"sensitive": def.Sensitive,
			})
		}
	}
}

func (r *SettingRepository) FindByKey(key string) (*model.Setting, error) {
	var setting model.Setting
	if err := db.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *SettingRepository) Create(setting *model.Setting) error {
	return db.DB.Create(setting).Error
}

func (r *SettingRepository) FindAll() ([]model.Setting, error) {
	var settings []model.Setting
	if err := db.DB.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingRepository) UpdateSettings(items []UpdateSettingItem, maskedValue string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
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
