package repo

import "perfect-pic-server/internal/model"

type UpdateSettingItem struct {
	Key   string
	Value string
}

type SettingStore interface {
	InitializeDefaults(defaults []model.Setting) error
	DeleteNotInKeys(allowedKeys []string) error
	FindByKey(key string) (*model.Setting, error)
	Create(setting *model.Setting) error
	FindAll() ([]model.Setting, error)
	UpdateSettings(items []UpdateSettingItem, maskedValue string) error
}
