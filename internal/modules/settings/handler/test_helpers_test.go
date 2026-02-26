package handler

import (
	"testing"

	modulerepo "perfect-pic-server/internal/modules/settings/repo"
	settingsservice "perfect-pic-server/internal/modules/settings/service"
	platformservice "perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var (
	testService *platformservice.AppService
	testHandler *Handler
)

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	settingStore := modulerepo.NewSettingRepository(gdb)
	testService = platformservice.NewAppService(settingStore)
	testHandler = New(settingsservice.New(testService, settingStore))
	testService.ClearCache()
	return gdb
}
