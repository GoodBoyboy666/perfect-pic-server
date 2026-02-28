package service

import (
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	settingsrepo "perfect-pic-server/internal/repository"
	"strconv"
	"strings"
)

// AdminListSettings 获取全部系统设置。
func (s *SettingsService) AdminListSettings() ([]model.Setting, error) {
	settings, err := s.settingStore.FindAll()
	if err != nil {
		return nil, commonpkg.NewInternalError("获取配置失败")
	}

	sortSettingsForAdmin(settings)
	maskSensitiveSettings(settings)
	return settings, nil
}

// AdminUpdateSettings 批量更新系统设置，并在成功后清理配置缓存。
func (s *SettingsService) AdminUpdateSettings(items []moduledto.UpdateSettingRequest) error {
	for _, item := range items {
		if err := validateSettingUpdate(item); err != nil {
			return err
		}
	}

	repoItems := make([]settingsrepo.UpdateSettingItem, 0, len(items))
	for _, item := range items {
		repoItems = append(repoItems, settingsrepo.UpdateSettingItem{
			Key:   item.Key,
			Value: item.Value,
		})
	}

	if err := s.settingStore.UpdateSettings(repoItems, maskedSettingValue); err != nil {
		return commonpkg.NewInternalError("更新失败")
	}

	s.dbConfig.ClearCache()
	return nil
}

func validateSettingUpdate(item moduledto.UpdateSettingRequest) error {
	if strings.TrimSpace(item.Key) == "" {
		return commonpkg.NewValidationError("配置键不能为空")
	}

	switch item.Key {
	case consts.ConfigDefaultStorageQuota:
		quota, err := strconv.ParseInt(strings.TrimSpace(item.Value), 10, 64)
		if err != nil || quota <= 0 {
			return commonpkg.NewValidationError("默认存储配额必须为正整数（单位：Bytes）")
		}
	}

	return nil
}
