package handler

import (
	"testing"

	authservice "perfect-pic-server/internal/modules/auth/service"
	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	userservice "perfect-pic-server/internal/modules/user/service"
	platformservice "perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var (
	testService *authservice.Service
	testHandler *Handler
)

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	userStore := userrepo.NewUserRepository(gdb)
	imageStore := imagerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	userSvc := userservice.New(appService, userStore, imageStore)
	testService = authservice.New(appService, userStore, userSvc)
	testHandler = New(testService)
	testService.ClearCache()
	return gdb
}
