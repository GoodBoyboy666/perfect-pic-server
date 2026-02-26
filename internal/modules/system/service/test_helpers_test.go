package service

import (
	"testing"

	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	modulerepo "perfect-pic-server/internal/modules/system/repo"
	userrepo "perfect-pic-server/internal/modules/user/repo"
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
	systemStore := modulerepo.NewSystemRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	testService = New(appService, systemStore, userStore, imageStore)
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
