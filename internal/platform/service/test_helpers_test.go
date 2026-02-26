package service

import (
	"testing"

	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *AppService

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	testService = NewAppService(settingStore)
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
