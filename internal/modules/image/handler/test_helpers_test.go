package handler

import (
	"testing"

	modulerepo "perfect-pic-server/internal/modules/image/repo"
	imageservice "perfect-pic-server/internal/modules/image/service"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
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
	imageStore := modulerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	testService = platformservice.NewAppService(settingStore)
	userService := userservice.New(testService, userStore)
	testHandler = New(imageservice.New(testService, userService, imageStore))
	testService.ClearCache()
	return gdb
}
