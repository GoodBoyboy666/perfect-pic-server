package service

import (
	"testing"

	imagerepo "perfect-pic-server/internal/modules/image/repo"
	imageservice "perfect-pic-server/internal/modules/image/service"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	modulerepo "perfect-pic-server/internal/modules/system/repo"
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
	systemStore := modulerepo.NewSystemRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	userService := userservice.New(appService, userStore)
	imageService := imageservice.New(appService, userService, imageStore)
	userService.SetImageService(imageService)
	testService = New(appService, systemStore, userService, imageService)
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
