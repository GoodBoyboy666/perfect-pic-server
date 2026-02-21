package service

import (
	"testing"

	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *AppService

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	testService = NewAppService(repository.NewRepositories(
		repository.NewUserRepository(gdb),
		repository.NewImageRepository(gdb),
		repository.NewSettingRepository(gdb),
		repository.NewSystemRepository(gdb),
	))
	testService.ClearCache()
	return gdb
}

func mustTestService(t *testing.T) *AppService {
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
