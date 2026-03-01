package admin

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"runtime"
	"testing"
)

func TestStatUseCase_AdminGetServerStats(t *testing.T) {
	f := setupAdminFixture(t)

	u1 := model.User{Username: "u1", Password: "x", Status: 1, Email: "u1@example.com"}
	u2 := model.User{Username: "u2", Password: "x", Status: 1, Email: "u2@example.com"}
	if err := db.DB.Create(&u1).Error; err != nil {
		t.Fatalf("create user1 failed: %v", err)
	}
	if err := db.DB.Create(&u2).Error; err != nil {
		t.Fatalf("create user2 failed: %v", err)
	}

	img := model.Image{
		Filename:   "a.png",
		Path:       "2026/01/01/a.png",
		Size:       123,
		MimeType:   ".png",
		UploadedAt: 1,
		UserID:     u1.ID,
		Width:      1,
		Height:     1,
	}
	if err := db.DB.Create(&img).Error; err != nil {
		t.Fatalf("create image failed: %v", err)
	}

	stats, err := f.statUC.AdminGetServerStats()
	if err != nil {
		t.Fatalf("AdminGetServerStats failed: %v", err)
	}

	if stats.ImageCount != 1 {
		t.Fatalf("expected image_count=1, got %d", stats.ImageCount)
	}
	if stats.StorageUsage != 123 {
		t.Fatalf("expected storage_usage=123, got %d", stats.StorageUsage)
	}
	if stats.UserCount != 2 {
		t.Fatalf("expected user_count=2, got %d", stats.UserCount)
	}
	if stats.SystemInfo.OS != runtime.GOOS || stats.SystemInfo.Arch != runtime.GOARCH {
		t.Fatalf("unexpected system info: %+v", stats.SystemInfo)
	}
	if stats.SystemInfo.NumCPU <= 0 || stats.SystemInfo.NumGoroutine <= 0 {
		t.Fatalf("expected positive system counters: %+v", stats.SystemInfo)
	}
}
