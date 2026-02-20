package service

import (
	"perfect-pic-server/internal/model"
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
