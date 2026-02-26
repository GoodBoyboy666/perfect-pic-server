package service

import (
	"testing"

	modulerepo "perfect-pic-server/internal/modules/settings/repo"
	platformservice "perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *Service

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	settingStore := modulerepo.NewSettingRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	testService = New(appService, settingStore)
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
