package service

import (
	"testing"

	modulerepo "perfect-pic-server/internal/modules/image/repo"
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
	imageStore := modulerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	userService := userservice.New(appService, userStore)
	testService = New(appService, userService, imageStore)
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
