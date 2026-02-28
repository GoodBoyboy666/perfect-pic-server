package app

import (
	"os"
	"path/filepath"
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/testutils"
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
