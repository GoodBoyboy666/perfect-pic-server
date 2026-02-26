package handler

import (
	"testing"

	authservice "perfect-pic-server/internal/modules/auth/service"
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
	testUserSvc *userservice.Service
)

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	userStore := userrepo.NewUserRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	appService := platformservice.NewAppService(settingStore)
	testUserSvc = userservice.New(appService, userStore)
	testService = authservice.New(appService, testUserSvc)
	testHandler = New(testService)
	testService.ClearCache()
	return gdb
}
