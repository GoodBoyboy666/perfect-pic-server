package service

import (
	"perfect-pic-server/internal/model"
	moduledto "perfect-pic-server/internal/modules/settings/dto"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	platformservice "perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/utils"
)

// AdminListSettings 获取全部系统设置。
func (s *Service) AdminListSettings() ([]model.Setting, error) {
	settings, err := s.settingStore.FindAll()
	if err != nil {
		return nil, platformservice.NewInternalError("获取配置失败")
	}

	sortSettingsForAdmin(settings)
	maskSensitiveSettings(settings)
	return settings, nil
}

// AdminUpdateSettings 批量更新系统设置，并在成功后清理配置缓存。
func (s *Service) AdminUpdateSettings(items []moduledto.UpdateSettingRequest) error {
	for _, item := range items {
		if item.Key == "" {
			return platformservice.NewValidationError("配置键不能为空")
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
		return platformservice.NewInternalError("更新失败")
	}

	s.ClearCache()
	return nil
}

// AdminSendTestEmail 发送管理员测试邮件。
func (s *Service) AdminSendTestEmail(toEmail string) error {
	if ok, msg := utils.ValidateEmail(toEmail); !ok {
		return platformservice.NewValidationError(msg)
	}

	if err := s.SendTestEmail(toEmail); err != nil {
		return platformservice.NewInternalError("发送失败")
	}

	return nil
}
