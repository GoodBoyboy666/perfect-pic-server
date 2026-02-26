package service

import (
	"testing"

	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	modulerepo "perfect-pic-server/internal/modules/user/repo"
	platformservice "perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *Service

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	userStore := modulerepo.NewUserRepository(gdb)
	imageStore := imagerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	testService = New(appService, userStore, imageStore)
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

func resetPasswordResetStore() {
	passwordResetStore.Range(func(key, value any) bool {
		passwordResetStore.Delete(key)
		return true
	})
	passwordResetTokenStore.Range(func(key, value any) bool {
		passwordResetTokenStore.Delete(key)
		return true
	})
}

func resetEmailChangeStore() {
	emailChangeStore.Range(func(key, value any) bool {
		emailChangeStore.Delete(key)
		return true
	})
	emailChangeTokenStore.Range(func(key, value any) bool {
		emailChangeTokenStore.Delete(key)
		return true
	})
}
