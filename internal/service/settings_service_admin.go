package service

import (
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/utils"
)

type UpdateSettingPayload struct {
	Key   string
	Value string
}

// AdminListSettings 获取全部系统设置。
func (s *AppService) AdminListSettings() ([]model.Setting, error) {
	settings, err := s.repos.Setting.FindAll()
	if err != nil {
		return nil, NewInternalError("获取配置失败")
	}

	sortSettingsForAdmin(settings)
	maskSensitiveSettings(settings)
	return settings, nil
}

// AdminUpdateSettings 批量更新系统设置，并在成功后清理配置缓存。
func (s *AppService) AdminUpdateSettings(items []UpdateSettingPayload) error {
	for _, item := range items {
		if item.Key == "" {
			return NewValidationError("配置键不能为空")
		}
	}

	repoItems := make([]repository.UpdateSettingItem, 0, len(items))
	for _, item := range items {
		repoItems = append(repoItems, repository.UpdateSettingItem{
			Key:   item.Key,
			Value: item.Value,
		})
	}

	if err := s.repos.Setting.UpdateSettings(repoItems, maskedSettingValue); err != nil {
		return NewInternalError("更新失败")
	}

	s.ClearCache()
	return nil
}

// AdminSendTestEmail 发送管理员测试邮件。
func (s *AppService) AdminSendTestEmail(toEmail string) error {
	if ok, msg := utils.ValidateEmail(toEmail); !ok {
		return NewValidationError(msg)
	}

	if err := s.SendTestEmail(toEmail); err != nil {
		return NewInternalError("发送失败")
	}

	return nil
}
