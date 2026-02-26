package service

import (
	"testing"

	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	userservice "perfect-pic-server/internal/modules/user/service"
	platformservice "perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *Service

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	userStore := userrepo.NewUserRepository(gdb)
	imageStore := imagerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	userSvc := userservice.New(appService, userStore, imageStore)
	testService = New(appService, userStore, userSvc)
	testService.ClearCache()
	return gdb
}

func mustTestService(t *testing.T) *Service {
	t.Helper()
	if testService == nil {
		setupTestDB(t)
	}
	return testService
}

func resetPasskeySessionStore() {
	passkeySessionStore.Range(func(key, value any) bool {
		passkeySessionStore.Delete(key)
		return true
	})
}

func resetPasswordResetStore() {
	// Token store state belongs to user module service; tests regenerate tokens per case.
}
