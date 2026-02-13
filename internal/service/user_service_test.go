package service

import (
	"os"
	"path/filepath"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

func TestGenerateAndVerifyForgetPasswordToken_OneTimeUse(t *testing.T) {
	setupTestDB(t)
	resetPasswordResetStore()

	token, err := GenerateForgetPasswordToken(42)
	if err != nil {
		t.Fatalf("GenerateForgetPasswordToken: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("expected 64-char hex token, got len=%d token=%q", len(token), token)
	}

	uid, ok := VerifyForgetPasswordToken(token)
	if !ok || uid != 42 {
		t.Fatalf("expected valid token for uid=42, got uid=%d ok=%v", uid, ok)
	}

	uid2, ok2 := VerifyForgetPasswordToken(token)
	if ok2 || uid2 != 0 {
		t.Fatalf("expected one-time use token to be invalid on second use, got uid=%d ok=%v", uid2, ok2)
	}
}

func TestDeleteUserFiles_RemovesAvatarDirAndImages(t *testing.T) {
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldwd) }()

	userID := uint(7)

	// Create avatar dir with a file.
	avatarDir := filepath.Join("uploads", "avatars", "7")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		t.Fatalf("mkdir avatar dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(avatarDir, "a.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("write avatar file: %v", err)
	}

	// Create an image record and a physical file.
	imgRel := filepath.ToSlash(filepath.Join("2026", "02", "13", "x.png"))
	imgLocal := filepath.FromSlash(imgRel)
	imgFile := filepath.Join("uploads", "imgs", imgLocal)
	if err := os.MkdirAll(filepath.Dir(imgFile), 0755); err != nil {
		t.Fatalf("mkdir img dir: %v", err)
	}
	if err := os.WriteFile(imgFile, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatalf("write img file: %v", err)
	}

	if err := db.DB.Create(&model.Image{
		Filename:   "x.png",
		Path:       imgRel,
		Size:       4,
		MimeType:   ".png",
		UploadedAt: 1,
		UserID:     userID,
		Width:      1,
		Height:     1,
	}).Error; err != nil {
		t.Fatalf("create image record: %v", err)
	}

	if err := DeleteUserFiles(userID); err != nil {
		t.Fatalf("DeleteUserFiles: %v", err)
	}

	if _, err := os.Stat(avatarDir); !os.IsNotExist(err) {
		t.Fatalf("expected avatar dir to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(imgFile); !os.IsNotExist(err) {
		t.Fatalf("expected image file to be removed, stat err=%v", err)
	}
}

func TestGetSystemDefaultStorageQuota(t *testing.T) {
	setupTestDB(t)

	// When setting is missing, it should fall back to default value from DefaultSettings.
	if got := GetSystemDefaultStorageQuota(); got <= 0 {
		t.Fatalf("expected positive default quota, got %d", got)
	}

	// Override to a custom value.
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigDefaultStorageQuota, Value: "123"}).Error
	ClearCache()
	if got := GetSystemDefaultStorageQuota(); got != 123 {
		t.Fatalf("expected 123, got %d", got)
	}
}
