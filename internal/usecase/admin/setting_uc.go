package admin

import (
	platformservice "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/utils"
)

// AdminSendTestEmail 发送管理员测试邮件。
func (c *SettingsUseCase) AdminSendTestEmail(toEmail string) error {
	if ok, msg := utils.ValidateEmail(toEmail); !ok {
		return platformservice.NewValidationError(msg)
	}

	if err := c.emailService.SendTestEmail(toEmail); err != nil {
		return platformservice.NewInternalError("发送失败")
	}

	return nil
}
