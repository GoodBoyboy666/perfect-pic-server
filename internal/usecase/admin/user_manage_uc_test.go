package admin

import (
	"errors"
	"os"
	"path/filepath"
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"strconv"
	"testing"

	"gorm.io/gorm"
)

func TestUserManageUseCase_AdminDeleteUser_SoftDelete(t *testing.T) {
	f := setupAdminFixture(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "alice@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	if err := f.userManageUC.AdminDeleteUser(u.ID, false); err != nil {
		t.Fatalf("AdminDeleteUser soft failed: %v", err)
	}

	var got model.User
	if err := db.DB.Unscoped().First(&got, u.ID).Error; err != nil {
		t.Fatalf("load deleted user failed: %v", err)
	}
	if got.Status != 3 {
		t.Fatalf("expected status=3, got %d", got.Status)
	}
	if got.Username == "alice" || got.Email == "alice@example.com" {
		t.Fatalf("expected rewritten unique fields, got username=%q email=%q", got.Username, got.Email)
	}
}

func TestUserManageUseCase_AdminDeleteUser_NotFound(t *testing.T) {
	f := setupAdminFixture(t)

	err := f.userManageUC.AdminDeleteUser(999, false)
	assertServiceErrorCode(t, err, common.ErrorCodeNotFound)
}

func TestUserManageUseCase_AdminDeleteUser_HardDeleteCleansDataAndFiles(t *testing.T) {
	f := setupAdminFixture(t)
	chdirForTest(t, t.TempDir())

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "alice@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	avatarDir := filepath.Join("uploads", "avatars", strconv.FormatUint(uint64(u.ID), 10))
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		t.Fatalf("create avatar dir failed: %v", err)
	}
	avatarFile := filepath.Join(avatarDir, "a.txt")
	if err := os.WriteFile(avatarFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write avatar file failed: %v", err)
	}

	imgRel := "2026/02/13/a.png"
	imgFile := filepath.Join("uploads", "imgs", filepath.FromSlash(imgRel))
	if err := os.MkdirAll(filepath.Dir(imgFile), 0755); err != nil {
		t.Fatalf("create image dir failed: %v", err)
	}
	if err := os.WriteFile(imgFile, []byte("img"), 0644); err != nil {
		t.Fatalf("write image file failed: %v", err)
	}

	img := model.Image{
		Filename:   "a.png",
		Path:       imgRel,
		Size:       3,
		MimeType:   ".png",
		UploadedAt: 1,
		UserID:     u.ID,
		Width:      1,
		Height:     1,
	}
	if err := db.DB.Create(&img).Error; err != nil {
		t.Fatalf("create image record failed: %v", err)
	}

	passkey := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_admin_delete",
		Credential:   `{"id":"cred_admin_delete"}`,
	}
	if err := db.DB.Create(&passkey).Error; err != nil {
		t.Fatalf("create passkey failed: %v", err)
	}

	if err := f.userManageUC.AdminDeleteUser(u.ID, true); err != nil {
		t.Fatalf("AdminDeleteUser hard failed: %v", err)
	}

	if err := db.DB.Unscoped().First(&model.User{}, u.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected user hard-deleted, got err=%v", err)
	}

	var imgCount int64
	if err := db.DB.Unscoped().Model(&model.Image{}).Where("user_id = ?", u.ID).Count(&imgCount).Error; err != nil {
		t.Fatalf("count images failed: %v", err)
	}
	if imgCount != 0 {
		t.Fatalf("expected 0 images after hard delete, got %d", imgCount)
	}

	var passkeyCount int64
	if err := db.DB.Unscoped().Model(&model.PasskeyCredential{}).Where("user_id = ?", u.ID).Count(&passkeyCount).Error; err != nil {
		t.Fatalf("count passkeys failed: %v", err)
	}
	if passkeyCount != 0 {
		t.Fatalf("expected 0 passkeys after hard delete, got %d", passkeyCount)
	}

	if _, err := os.Stat(avatarDir); !os.IsNotExist(err) {
		t.Fatalf("expected avatar dir removed, err=%v", err)
	}
	if _, err := os.Stat(imgFile); !os.IsNotExist(err) {
		t.Fatalf("expected image file removed, err=%v", err)
	}
}
