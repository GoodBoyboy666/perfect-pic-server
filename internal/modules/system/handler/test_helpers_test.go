package handler

import (
	"testing"

	imagerepo "perfect-pic-server/internal/modules/image/repo"
	imageservice "perfect-pic-server/internal/modules/image/service"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	modulerepo "perfect-pic-server/internal/modules/system/repo"
	systemservice "perfect-pic-server/internal/modules/system/service"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	userservice "perfect-pic-server/internal/modules/user/service"
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
	userStore := userrepo.NewUserRepository(gdb)
	imageStore := imagerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	systemStore := modulerepo.NewSystemRepository(gdb)
	testService = platformservice.NewAppService(settingStore)
	userService := userservice.New(testService, userStore)
	imageService := imageservice.New(testService, userService, imageStore)
	userService.SetImageService(imageService)
	testHandler = New(systemservice.New(testService, systemStore, userService, imageService))
	testService.ClearCache()
	return gdb
}
