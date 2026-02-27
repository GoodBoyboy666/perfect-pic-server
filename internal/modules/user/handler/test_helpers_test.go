package handler

import (
	"testing"

	authservice "perfect-pic-server/internal/modules/auth/service"
	imagerepo "perfect-pic-server/internal/modules/image/repo"
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
	imageStore := imagerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	testService = platformservice.NewAppService(settingStore)
	userSvc := userservice.New(testService, userStore)
	imageSvc := imageservice.New(testService, userSvc, imageStore)
	userSvc.SetImageService(imageSvc)
	authSvc := authservice.New(testService, userSvc)
	testHandler = New(userSvc, authSvc, imageSvc)
	testService.ClearCache()
	return gdb
}
