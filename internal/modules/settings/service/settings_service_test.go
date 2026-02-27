package service

import (
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	moduledto "perfect-pic-server/internal/modules/settings/dto"
	platformservice "perfect-pic-server/internal/platform/service"
)

func assertSettingsServiceErrorCode(t *testing.T, err error, code platformservice.ErrorCode) *platformservice.ServiceError {
	t.Helper()
	serviceErr, ok := platformservice.AsServiceError(err)
	if !ok {
		t.Fatalf("期望 ServiceError，实际为: %v", err)
	}
	if serviceErr.Code != code {
		t.Fatalf("期望错误码 %q，实际为 %q", code, serviceErr.Code)
	}
	return serviceErr
}

// 测试内容：验证读取字符串设置时会插入默认值并与数据库一致。
func TestGetString_DefaultSettingInserted(t *testing.T) {
	setupTestDB(t)

	testService.ClearCache()
	val := testService.GetString(consts.ConfigSiteName)
	if val == "" {
		t.Fatalf("期望 default site_name to be non-empty")
	}

	var s model.Setting
	if err := db.DB.Where("key = ?", consts.ConfigSiteName).First(&s).Error; err != nil {
		t.Fatalf("期望 default setting row to be created: %v", err)
	}
	if s.Value != val {
		t.Fatalf("db value mismatch: got=%q 期望=%q", s.Value, val)
	}
}

// 测试内容：验证未知 key 返回空值且缓存未找到结果。
func TestGetString_UnknownKeyReturnsEmpty(t *testing.T) {
	setupTestDB(t)

	testService.ClearCache()
	val := testService.GetString("unknown_key_not_exists")
	if val != "" {
		t.Fatalf("期望 empty for unknown key，实际为 %q", val)
	}
	// 第二次调用仍应返回空值（缓存了未找到标记）。
	val2 := testService.GetString("unknown_key_not_exists")
	if val2 != "" {
		t.Fatalf("期望 empty for unknown key，实际为 %q", val2)
	}
}

// 测试内容：验证整数配置解析失败时返回 0。
func TestGetInt_ParseFailureReturnsZero(t *testing.T) {
	db := setupTestDB(t)

	_ = db.Create(&model.Setting{Key: "k", Value: "not-int"}).Error
	testService.ClearCache()

	if got := testService.GetInt("k"); got != 0 {
		t.Fatalf("期望 0 for parse failure，实际为 %d", got)
	}
}

// 测试内容：验证浮点配置的正常解析与解析失败回退为 0。
func TestGetFloat64_ParseAndFailure(t *testing.T) {
	db := setupTestDB(t)

	_ = db.Create(&model.Setting{Key: "f1", Value: "0.5"}).Error
	_ = db.Create(&model.Setting{Key: "f2", Value: "bad"}).Error
	testService.ClearCache()

	if got := testService.GetFloat64("f1"); got != 0.5 {
		t.Fatalf("期望 0.5，实际为 %v", got)
	}
	if got := testService.GetFloat64("f2"); got != 0 {
		t.Fatalf("期望 0 on parse 错误，实际为 %v", got)
	}
}

// 测试内容：验证管理员读取设置时敏感字段会被掩码。
func TestListSettingsForAdmin_MasksSensitive(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Create(&model.Setting{Key: "k1", Value: "v1", Sensitive: false}).Error
	_ = db.DB.Create(&model.Setting{Key: "k2", Value: "secret", Sensitive: true}).Error

	settings, err := testService.AdminListSettings()
	if err != nil {
		t.Fatalf("AdminListSettings: %v", err)
	}

	m := map[string]string{}
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	if m["k1"] != "v1" {
		t.Fatalf("期望 k1=v1，实际为 %q", m["k1"])
	}
	if m["k2"] != "**********" {
		t.Fatalf("期望 sensitive masked，实际为 %q", m["k2"])
	}
}

// 测试内容：验证更新设置时敏感掩码值不会覆盖真实敏感值。
func TestUpdateSettingsForAdmin_MaskedSensitiveIsNotOverwritten(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Create(&model.Setting{Key: "s1", Value: "secret", Sensitive: true}).Error
	_ = db.DB.Create(&model.Setting{Key: "n1", Value: "old", Sensitive: false}).Error

	err := testService.AdminUpdateSettings([]moduledto.UpdateSettingRequest{
		{Key: "s1", Value: "**********"}, // 应被忽略
		{Key: "n1", Value: "**********"}, // 应覆盖（非敏感）
		{Key: "new", Value: "val"},       // 更新或插入
	})
	if err != nil {
		t.Fatalf("AdminUpdateSettings: %v", err)
	}

	var s1 model.Setting
	_ = db.DB.Where("key = ?", "s1").First(&s1).Error
	if s1.Value != "secret" {
		t.Fatalf("期望 sensitive value preserved，实际为 %q", s1.Value)
	}
	var n1 model.Setting
	_ = db.DB.Where("key = ?", "n1").First(&n1).Error
	if n1.Value != "**********" {
		t.Fatalf("期望 non-sensitive overwritten，实际为 %q", n1.Value)
	}
	var n model.Setting
	_ = db.DB.Where("key = ?", "new").First(&n).Error
	if n.Value != "val" {
		t.Fatalf("期望 new=val，实际为 %q", n.Value)
	}
}

// 测试内容：验证默认存储配额配置仅允许正整数，非法值会被拒绝。
func TestUpdateSettingsForAdmin_RejectInvalidDefaultStorageQuota(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigDefaultStorageQuota, Value: "1073741824"}).Error
	testService.ClearCache()

	invalidValues := []string{"-1", "0", "abc", ""}
	for _, value := range invalidValues {
		err := testService.AdminUpdateSettings([]moduledto.UpdateSettingRequest{
			{Key: consts.ConfigDefaultStorageQuota, Value: value},
		})
		assertSettingsServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

		var setting model.Setting
		_ = db.DB.Where("key = ?", consts.ConfigDefaultStorageQuota).First(&setting).Error
		if setting.Value != "1073741824" {
			t.Fatalf("非法值不应写入数据库，实际为 %q", setting.Value)
		}
	}
}

// 测试内容：验证默认存储配额配置在合法值下可成功更新。
func TestUpdateSettingsForAdmin_AcceptValidDefaultStorageQuota(t *testing.T) {
	setupTestDB(t)

	err := testService.AdminUpdateSettings([]moduledto.UpdateSettingRequest{
		{Key: consts.ConfigDefaultStorageQuota, Value: "2048"},
	})
	if err != nil {
		t.Fatalf("AdminUpdateSettings: %v", err)
	}

	var setting model.Setting
	_ = db.DB.Where("key = ?", consts.ConfigDefaultStorageQuota).First(&setting).Error
	if setting.Value != "2048" {
		t.Fatalf("期望 quota=2048，实际为 %q", setting.Value)
	}
}

// 测试内容：验证管理员读取设置时顺序稳定，优先按默认配置定义顺序返回。
func TestListSettingsForAdmin_OrderStableByDefaultSettings(t *testing.T) {
	setupTestDB(t)

	// 乱序写入默认配置项 + 自定义配置项。
	_ = db.DB.Create(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "image", Category: "验证码"}).Error
	_ = db.DB.Create(&model.Setting{Key: consts.ConfigAllowRegister, Value: "true", Category: "安全"}).Error
	_ = db.DB.Create(&model.Setting{Key: consts.ConfigSiteName, Value: "Perfect Pic", Category: "常规"}).Error
	_ = db.DB.Create(&model.Setting{Key: "z_custom", Value: "1", Category: "自定义"}).Error
	_ = db.DB.Create(&model.Setting{Key: "a_custom", Value: "2", Category: "自定义"}).Error

	// 更新其中一项，模拟“修改后重新拉取列表”场景。
	err := testService.AdminUpdateSettings([]moduledto.UpdateSettingRequest{
		{Key: consts.ConfigAllowRegister, Value: "false"},
	})
	if err != nil {
		t.Fatalf("AdminUpdateSettings: %v", err)
	}

	settings, err := testService.AdminListSettings()
	if err != nil {
		t.Fatalf("AdminListSettings: %v", err)
	}

	pos := map[string]int{}
	for i, item := range settings {
		pos[item.Key] = i
	}

	// 默认项顺序应固定，不受写入/更新顺序影响。
	if pos[consts.ConfigSiteName] >= pos[consts.ConfigAllowRegister] {
		t.Fatalf("期望 %s 在 %s 之前，实际顺序异常", consts.ConfigSiteName, consts.ConfigAllowRegister)
	}
	if pos[consts.ConfigAllowRegister] >= pos[consts.ConfigCaptchaProvider] {
		t.Fatalf("期望 %s 在 %s 之前，实际顺序异常", consts.ConfigAllowRegister, consts.ConfigCaptchaProvider)
	}

	// 自定义项应排在默认项之后，并在同分类内按 key 稳定排序。
	if pos[consts.ConfigCaptchaProvider] >= pos["a_custom"] {
		t.Fatalf("期望默认项在自定义项之前，实际顺序异常")
	}
	if pos["a_custom"] >= pos["z_custom"] {
		t.Fatalf("期望自定义项按 key 排序，实际顺序异常")
	}
}
