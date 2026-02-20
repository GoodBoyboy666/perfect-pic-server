package service

import (
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

// 测试内容：验证读取字符串设置时会插入默认值并与数据库一致。
func TestGetString_DefaultSettingInserted(t *testing.T) {
	setupTestDB(t)

	ClearCache()
	val := GetString(consts.ConfigSiteName)
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

	ClearCache()
	val := GetString("unknown_key_not_exists")
	if val != "" {
		t.Fatalf("期望 empty for unknown key，实际为 %q", val)
	}
	// 第二次调用仍应返回空值（缓存了未找到标记）。
	val2 := GetString("unknown_key_not_exists")
	if val2 != "" {
		t.Fatalf("期望 empty for unknown key，实际为 %q", val2)
	}
}

// 测试内容：验证整数配置解析失败时返回 0。
func TestGetInt_ParseFailureReturnsZero(t *testing.T) {
	db := setupTestDB(t)

	_ = db.Create(&model.Setting{Key: "k", Value: "not-int"}).Error
	ClearCache()

	if got := GetInt("k"); got != 0 {
		t.Fatalf("期望 0 for parse failure，实际为 %d", got)
	}
}

// 测试内容：验证浮点配置的正常解析与解析失败回退为 0。
func TestGetFloat64_ParseAndFailure(t *testing.T) {
	db := setupTestDB(t)

	_ = db.Create(&model.Setting{Key: "f1", Value: "0.5"}).Error
	_ = db.Create(&model.Setting{Key: "f2", Value: "bad"}).Error
	ClearCache()

	if got := GetFloat64("f1"); got != 0.5 {
		t.Fatalf("期望 0.5，实际为 %v", got)
	}
	if got := GetFloat64("f2"); got != 0 {
		t.Fatalf("期望 0 on parse 错误，实际为 %v", got)
	}
}

// 测试内容：验证管理员读取设置时敏感字段会被掩码。
func TestListSettingsForAdmin_MasksSensitive(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Create(&model.Setting{Key: "k1", Value: "v1", Sensitive: false}).Error
	_ = db.DB.Create(&model.Setting{Key: "k2", Value: "secret", Sensitive: true}).Error

	settings, err := AdminListSettings()
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

	err := AdminUpdateSettings([]UpdateSettingPayload{
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
