package handler

import (
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"
)

type compositeHandler struct {
	*AuthHandler
	*UserHandler
	*ImageHandler
	*SystemHandler
	*SettingsHandler
}

var (
	testService *service.Service
	testUserSvc *service.Service
	testHandler *compositeHandler
)

func setupTestDB(t *testing.T) {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)

	testService = service.NewAppService(userStore, imageStore, settingStore, systemStore)
	testUserSvc = testService
	if err := testService.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	testService.ClearCache()

	testHandler = &compositeHandler{
		AuthHandler:     NewAuthHandler(testService),
		UserHandler:     NewUserHandler(testService),
		ImageHandler:    NewImageHandler(testService),
		SystemHandler:   NewSystemHandler(testService),
		SettingsHandler: NewSettingsHandler(testService),
	}
}
