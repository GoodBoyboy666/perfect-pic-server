package runtime

import (
	"errors"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

// 测试内容：验证初始化设置会清理非默认配置键，并保留默认键已有值。
func TestInitializeSettings_RemovesLegacySettings(t *testing.T) {
	gdb := setupTestDB(t)
	svc := mustTestService(t)

	if err := gdb.Create(&model.Setting{
		Key:   "legacy_custom_key",
		Value: "legacy-value",
	}).Error; err != nil {
		t.Fatalf("create legacy setting failed: %v", err)
	}

	if err := gdb.Create(&model.Setting{
		Key:   consts.ConfigSiteName,
		Value: "My Custom Site Name",
	}).Error; err != nil {
		t.Fatalf("create default setting row failed: %v", err)
	}

	if err := svc.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}

	var legacy model.Setting
	err := gdb.Where("key = ?", "legacy_custom_key").First(&legacy).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("legacy setting should be removed, got err=%v", err)
	}

	var siteName model.Setting
	if err := gdb.Where("key = ?", consts.ConfigSiteName).First(&siteName).Error; err != nil {
		t.Fatalf("default setting should remain: %v", err)
	}
	if siteName.Value != "My Custom Site Name" {
		t.Fatalf("default setting value should be preserved, got=%q", siteName.Value)
	}
}
