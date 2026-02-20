package service

import (
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
)

type UpdateSettingPayload struct {
	Key   string
	Value string
}

// AdminListSettings 获取全部系统设置。
func AdminListSettings() ([]model.Setting, error) {
	settings, err := repository.Setting.FindAll()
	if err != nil {
		return nil, err
	}

	maskSensitiveSettings(settings)
	return settings, nil
}

// AdminUpdateSettings 批量更新系统设置，并在成功后清理配置缓存。
func AdminUpdateSettings(items []UpdateSettingPayload) error {
	repoItems := make([]repository.UpdateSettingItem, 0, len(items))
	for _, item := range items {
		repoItems = append(repoItems, repository.UpdateSettingItem{
			Key:   item.Key,
			Value: item.Value,
		})
	}

	if err := repository.Setting.UpdateSettings(repoItems, maskedSettingValue); err != nil {
		return err
	}

	ClearCache()
	return nil
}
