package service

import (
	"perfect-pic-server/internal/model"
	settingsruntime "perfect-pic-server/internal/modules/settings/runtime"
	"sort"
)

const maskedSettingValue = "**********"

var (
	defaultSettingOrderByKey = buildDefaultSettingOrderByKey()
	defaultCategoryOrder     = buildDefaultCategoryOrder()
)

// maskSensitiveSettings 对敏感配置值进行统一脱敏。
func maskSensitiveSettings(settings []model.Setting) {
	for i := range settings {
		if settings[i].Sensitive {
			settings[i].Value = maskedSettingValue
		}
	}
}

// sortSettingsForAdmin 按默认配置定义顺序排序，避免数据库返回顺序不稳定导致前端位置跳动。
func sortSettingsForAdmin(settings []model.Setting) {
	sort.SliceStable(settings, func(i, j int) bool {
		left := settings[i]
		right := settings[j]

		leftIdx, leftInDefault := defaultSettingOrderByKey[left.Key]
		rightIdx, rightInDefault := defaultSettingOrderByKey[right.Key]
		if leftInDefault && rightInDefault {
			return leftIdx < rightIdx
		}
		if leftInDefault != rightInDefault {
			return leftInDefault
		}

		leftCatIdx, leftCatKnown := defaultCategoryOrder[left.Category]
		rightCatIdx, rightCatKnown := defaultCategoryOrder[right.Category]
		if leftCatKnown && rightCatKnown && leftCatIdx != rightCatIdx {
			return leftCatIdx < rightCatIdx
		}
		if leftCatKnown != rightCatKnown {
			return leftCatKnown
		}

		if left.Category != right.Category {
			return left.Category < right.Category
		}
		return left.Key < right.Key
	})
}

func buildDefaultSettingOrderByKey() map[string]int {
	order := make(map[string]int, len(settingsruntime.DefaultSettings))
	for i, setting := range settingsruntime.DefaultSettings {
		order[setting.Key] = i
	}
	return order
}

func buildDefaultCategoryOrder() map[string]int {
	order := make(map[string]int, len(settingsruntime.DefaultSettings))
	for _, setting := range settingsruntime.DefaultSettings {
		if _, exists := order[setting.Category]; exists {
			continue
		}
		order[setting.Category] = len(order)
	}
	return order
}
