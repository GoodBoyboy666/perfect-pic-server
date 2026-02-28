package handler

import (
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"
	adminuc "perfect-pic-server/internal/usecase/admin"
	appuc "perfect-pic-server/internal/usecase/app"
)

type compositeHandler struct {
	*AuthHandler
	*UserHandler
	*ImageHandler
	*SystemHandler
	*SettingsHandler
}

var (
	testService *config.DBConfig
	testUserSvc *service.UserService
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
	passkeyStore := repository.NewPasskeyRepository(gdb)

	dbConfig := config.NewDBConfig(settingStore)
	authService := service.NewAuthService(dbConfig)
	userService := service.NewUserService(userStore, dbConfig)
	imageService := service.NewImageService(imageStore, dbConfig)
	emailService := service.NewEmailService(dbConfig)
	captchaService := service.NewCaptchaService(dbConfig)
	initService := service.NewInitService(systemStore, dbConfig)
	passkeyService := service.NewPasskeyService(passkeyStore, dbConfig)
	settingsService := service.NewSettingsService(settingStore, dbConfig)

	authUseCase := appuc.NewAuthUseCase(authService, userStore, userService, emailService, initService, dbConfig)
	userUseCase := appuc.NewUserUseCase(userService, userStore, emailService, dbConfig)
	imageUseCase := appuc.NewImageUseCase(imageService, userService, userStore, dbConfig)
	passkeyUseCase := appuc.NewPasskeyUseCase(passkeyService, passkeyStore, authService, userStore)
	userManageUseCase := adminuc.NewUserManageUseCase(userService, imageService, passkeyService)
	settingsUseCase := adminuc.NewSettingsUseCase(emailService)
	statUseCase := adminuc.NewStatUseCase(imageStore, userStore)

	testService = dbConfig
	testUserSvc = userService
	if err := testService.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	testService.ClearCache()

	testHandler = &compositeHandler{
		AuthHandler:     NewAuthHandler(authService, captchaService, authUseCase, initService, dbConfig, passkeyUseCase),
		UserHandler:     NewUserHandler(userService, userUseCase, userManageUseCase, imageService, imageUseCase, authService, passkeyService, passkeyUseCase),
		ImageHandler:    NewImageHandler(imageService, imageUseCase),
		SystemHandler:   NewSystemHandler(initService, statUseCase, dbConfig, userService),
		SettingsHandler: NewSettingsHandler(settingsService, settingsUseCase),
	}
}
