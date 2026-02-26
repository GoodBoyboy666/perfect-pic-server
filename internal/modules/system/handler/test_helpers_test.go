package handler

import (
	"testing"

	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	modulerepo "perfect-pic-server/internal/modules/system/repo"
	systemservice "perfect-pic-server/internal/modules/system/service"
	userrepo "perfect-pic-server/internal/modules/user/repo"
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
	testHandler = New(systemservice.New(testService, systemStore, userStore, imageStore))
	testService.ClearCache()
	return gdb
}
