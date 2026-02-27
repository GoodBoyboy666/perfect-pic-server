package service

import (
	"sync"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var (
	testService     *Service
	testUserService *Service
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)

	testService = NewAppService(userStore, imageStore, settingStore, systemStore)
	testUserService = testService

	if err := testService.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
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
	clearSyncMap(&passwordResetStore)
	clearSyncMap(&passwordResetTokenStore)
}

func resetEmailChangeStore() {
	clearSyncMap(&emailChangeStore)
	clearSyncMap(&emailChangeTokenStore)
}

func resetPasskeySessionStore() {
	clearSyncMap(&passkeySessionStore)
}

func clearSyncMap(store *sync.Map) {
	store.Range(func(key, _ interface{}) bool {
		store.Delete(key)
		return true
	})
}
