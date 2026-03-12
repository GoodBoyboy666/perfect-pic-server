package handler

import (
	"perfect-pic-server/internal/consts"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/pkg/cache"
	pkgmail "perfect-pic-server/internal/pkg/email"
	jwtpkg "perfect-pic-server/internal/pkg/jwt"
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
	t.Setenv("PERFECT_PIC_SMTP_HOST", "127.0.0.1")
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	testGormDB = gdb
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)
	passkeyStore := repository.NewPasskeyRepository(gdb)

	dbConfig := config.NewDBConfig(settingStore)
	staticConfig := config.NewStaticConfig()
	tokenService := jwtpkg.NewJWT(config.NewJWTConfig(staticConfig))
	cacheStore := cache.NewStore(nil, config.NewCacheConfig(staticConfig))

	authService := service.NewAuthService(dbConfig, tokenService)
	userService := service.NewUserService(userStore, dbConfig, cacheStore, tokenService)
	imageService := service.NewImageService(imageStore, dbConfig, staticConfig)
	emailService := service.NewEmailService(dbConfig, pkgmail.NewMailer(), staticConfig)
	captchaService := service.NewCaptchaService(dbConfig)
	initService := service.NewInitService(systemStore, dbConfig)
	passkeyService := service.NewPasskeyService(passkeyStore, dbConfig, cacheStore)
	settingsService := service.NewSettingsService(settingStore, dbConfig)

	authUseCase := appuc.NewAuthUseCase(authService, userStore, userService, emailService, initService, dbConfig)
	userUseCase := appuc.NewUserUseCase(authService, userService, userStore, emailService, dbConfig)
	imageUseCase := appuc.NewImageUseCase(imageService, userService, userStore, staticConfig, dbConfig)
	passkeyUseCase := appuc.NewPasskeyUseCase(passkeyService, passkeyStore, authService, userStore)
	userManageUseCase := adminuc.NewUserManageUseCase(userService, imageService, passkeyService)
	settingsUseCase := adminuc.NewSettingsUseCase(emailService)
	statUseCase := adminuc.NewStatUseCase(imageStore, userStore)

	testService = dbConfig
	testUserSvc = userService
	if err := testService.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	if err := settingStore.UpdateSettings([]repository.UpdateSettingItem{{
		Key:   consts.ConfigEnableSMTP,
		Value: "true",
	}}, ""); err != nil {
		t.Fatalf("enable smtp for tests failed: %v", err)
	}
	testService.ClearCache()

	testHandler = &compositeHandler{
		AuthHandler:     NewAuthHandler(authService, captchaService, authUseCase, initService, dbConfig, passkeyUseCase),
		UserHandler:     NewUserHandler(userService, userUseCase, userManageUseCase, imageService, imageUseCase, authService, passkeyService, passkeyUseCase),
		ImageHandler:    NewImageHandler(imageService, imageUseCase),
		SystemHandler:   NewSystemHandler(initService, statUseCase, dbConfig, staticConfig, userService),
		SettingsHandler: NewSettingsHandler(settingsService, settingsUseCase),
	}
}
