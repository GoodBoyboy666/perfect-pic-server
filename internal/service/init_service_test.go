package service

import (
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func TestInitializeSystemAndIsSystemInitialized(t *testing.T) {
	setupTestDB(t)

	// Create default settings rows so InitializeSystem's Update() hits real rows.
	InitializeSettings()
	ClearCache()

	if IsSystemInitialized() {
		t.Fatalf("expected system to be uninitialized when allow_init is true")
	}

	payload := InitPayload{
		Username:        "admin",
		Password:        "abc12345",
		SiteName:        "MySite",
		SiteDescription: "Desc",
	}
	if err := InitializeSystem(payload); err != nil {
		t.Fatalf("InitializeSystem: %v", err)
	}

	if !IsSystemInitialized() {
		t.Fatalf("expected system to be initialized after InitializeSystem")
	}

	var s model.Setting
	_ = db.DB.Where("key = ?", consts.ConfigSiteName).First(&s).Error
	if s.Value != "MySite" {
		t.Fatalf("expected site name updated, got %q", s.Value)
	}

	var u model.User
	if err := db.DB.Where("username = ?", "admin").First(&u).Error; err != nil {
		t.Fatalf("expected admin user created: %v", err)
	}
	if !u.Admin {
		t.Fatalf("expected admin flag true")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("abc12345")) != nil {
		t.Fatalf("expected password to be hashed and match")
	}
}
