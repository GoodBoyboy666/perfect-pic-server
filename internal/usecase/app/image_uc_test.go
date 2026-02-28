package app

import (
	"os"
	"path/filepath"
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/testutils"
	"strconv"
	"strings"
	"testing"
)

func TestImageUseCase_ProcessImageUpload_QuotaExceeded(t *testing.T) {
	f := setupAppFixture(t)
	chdirForTest(t, t.TempDir())

	quota := int64(1)
	u := model.User{
		Username:     "alice",
		Password:     "x",
		Status:       1,
		Email:        "alice@example.com",
		StorageQuota: &quota,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	fh := mustFileHeader(t, "a.png", testutils.MinimalPNG())
	_, _, err := f.imageUC.ProcessImageUpload(fh, u.ID)
	if serviceErr := assertServiceErrorCode(t, err, common.ErrorCodeForbidden); !strings.Contains(serviceErr.Message, "存储空间不足") {
		t.Fatalf("expected quota exceeded message, got: %q", serviceErr.Message)
	}
}

func TestImageUseCase_ProcessImageUpload_Success(t *testing.T) {
	f := setupAppFixture(t)
	chdirForTest(t, t.TempDir())

	u := model.User{
		Username: "alice",
		Password: "x",
		Status:   1,
		Email:    "alice@example.com",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	fh := mustFileHeader(t, "a.png", testutils.MinimalPNG())
	img, url, err := f.imageUC.ProcessImageUpload(fh, u.ID)
	if err != nil {
		t.Fatalf("ProcessImageUpload failed: %v", err)
	}
	if img == nil || img.ID == 0 {
		t.Fatalf("expected created image record")
	}
	if !strings.HasPrefix(url, "/imgs/") {
		t.Fatalf("expected /imgs prefix, got: %q", url)
	}

	full := filepath.Join("uploads", "imgs", filepath.FromSlash(img.Path))
	if _, err := os.Stat(full); err != nil {
		t.Fatalf("expected uploaded file exists: %v", err)
	}
}

func TestImageUseCase_UpdateAndRemoveUserAvatar_Success(t *testing.T) {
	f := setupAppFixture(t)
	chdirForTest(t, t.TempDir())

	u := model.User{
		Username: "alice",
		Password: "x",
		Status:   1,
		Email:    "alice@example.com",
		Avatar:   "old.png",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	oldPath := filepath.Join("uploads", "avatars", strconv.FormatUint(uint64(u.ID), 10), "old.png")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0755); err != nil {
		t.Fatalf("create old avatar dir failed: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("x"), 0644); err != nil {
		t.Fatalf("write old avatar failed: %v", err)
	}

	fh := mustFileHeader(t, "a.png", testutils.MinimalPNG())
	newName, err := f.imageUC.UpdateUserAvatar(&u, fh)
	if err != nil {
		t.Fatalf("UpdateUserAvatar failed: %v", err)
	}
	if newName == "" {
		t.Fatalf("expected non-empty new avatar name")
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expected old avatar removed, err=%v", err)
	}

	var got model.User
	if err := db.DB.First(&got, u.ID).Error; err != nil {
		t.Fatalf("reload user failed: %v", err)
	}
	if got.Avatar == "" {
		t.Fatalf("expected avatar stored in db")
	}
	newPath := filepath.Join("uploads", "avatars", strconv.FormatUint(uint64(u.ID), 10), got.Avatar)
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("expected new avatar exists: %v", err)
	}

	if err := f.imageUC.RemoveUserAvatar(&got); err != nil {
		t.Fatalf("RemoveUserAvatar failed: %v", err)
	}

	var got2 model.User
	if err := db.DB.First(&got2, u.ID).Error; err != nil {
		t.Fatalf("reload user2 failed: %v", err)
	}
	if got2.Avatar != "" {
		t.Fatalf("expected avatar cleared, got %q", got2.Avatar)
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Fatalf("expected avatar file removed, err=%v", err)
	}
}

func TestImageUseCase_RemoveUserAvatar_NoAvatarNoop(t *testing.T) {
	f := setupAppFixture(t)
	chdirForTest(t, t.TempDir())

	u := model.User{
		Username: "alice",
		Password: "x",
		Status:   1,
		Email:    "alice@example.com",
		Avatar:   "",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	if err := f.imageUC.RemoveUserAvatar(&u); err != nil {
		t.Fatalf("RemoveUserAvatar no-avatar should be nil, got: %v", err)
	}
}
