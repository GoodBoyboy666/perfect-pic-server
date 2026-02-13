package service

import (
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

func TestGetString_DefaultSettingInserted(t *testing.T) {
	setupTestDB(t)

	ClearCache()
	val := GetString(consts.ConfigSiteName)
	if val == "" {
		t.Fatalf("expected default site_name to be non-empty")
	}

	var s model.Setting
	if err := db.DB.Where("key = ?", consts.ConfigSiteName).First(&s).Error; err != nil {
		t.Fatalf("expected default setting row to be created: %v", err)
	}
	if s.Value != val {
		t.Fatalf("db value mismatch: got=%q want=%q", s.Value, val)
	}
}

func TestGetString_UnknownKeyReturnsEmpty(t *testing.T) {
	setupTestDB(t)

	ClearCache()
	val := GetString("unknown_key_not_exists")
	if val != "" {
		t.Fatalf("expected empty for unknown key, got %q", val)
	}
	// Second call should still return empty (cached not-found marker).
	val2 := GetString("unknown_key_not_exists")
	if val2 != "" {
		t.Fatalf("expected empty for unknown key, got %q", val2)
	}
}

func TestGetInt_ParseFailureReturnsZero(t *testing.T) {
	db := setupTestDB(t)

	_ = db.Create(&model.Setting{Key: "k", Value: "not-int"}).Error
	ClearCache()

	if got := GetInt("k"); got != 0 {
		t.Fatalf("expected 0 for parse failure, got %d", got)
	}
}

func TestGetFloat64_ParseAndFailure(t *testing.T) {
	db := setupTestDB(t)

	_ = db.Create(&model.Setting{Key: "f1", Value: "0.5"}).Error
	_ = db.Create(&model.Setting{Key: "f2", Value: "bad"}).Error
	ClearCache()

	if got := GetFloat64("f1"); got != 0.5 {
		t.Fatalf("expected 0.5, got %v", got)
	}
	if got := GetFloat64("f2"); got != 0 {
		t.Fatalf("expected 0 on parse error, got %v", got)
	}
}
