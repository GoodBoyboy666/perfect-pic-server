package handler

import (
	"testing"

	"perfect-pic-server/internal/config"
	pkgmail "perfect-pic-server/internal/pkg/email"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"
	adminuc "perfect-pic-server/internal/usecase/admin"
	appuc "perfect-pic-server/internal/usecase/app"

	"gorm.io/gorm"
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
	testGormDB  *gorm.DB
)

func setupTestDB(t *testing.T) {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	testGormDB = gdb
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)
	passkeyStore := repository.NewPasskeyRepository(gdb)

	dbConfig := config.NewDBConfig(settingStore)
	authService := service.NewAuthService(dbConfig)
	userService := service.NewUserService(userStore, dbConfig, nil)
	imageService := service.NewImageService(imageStore, dbConfig)
	emailService := service.NewEmailService(dbConfig, pkgmail.NewMailer())
	captchaService := service.NewCaptchaService(dbConfig)
	initService := service.NewInitService(systemStore, dbConfig)
	passkeyService := service.NewPasskeyService(passkeyStore, dbConfig, nil)
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
