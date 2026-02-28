package router

import (
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"
	adminuc "perfect-pic-server/internal/usecase/admin"
	appuc "perfect-pic-server/internal/usecase/app"

	"github.com/gin-gonic/gin"
)

// 测试内容：验证核心 API 路由被正确注册。
func TestInitRouter_RegistersCoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := testutils.SetupDB(t)
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)
	passkeyStore := repository.NewPasskeyRepository(gdb)

	dbConfig := config.NewDBConfig(settingStore)
	if err := dbConfig.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	dbConfig.ClearCache()

	authService := service.NewAuthService(dbConfig)
	captchaService := service.NewCaptchaService(dbConfig)
	userService := service.NewUserService(userStore, dbConfig)
	imageService := service.NewImageService(imageStore, dbConfig)
	emailService := service.NewEmailService(dbConfig)
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

	authHandler := handler.NewAuthHandler(authService, captchaService, authUseCase, initService, dbConfig, passkeyUseCase)
	systemHandler := handler.NewSystemHandler(initService, statUseCase, dbConfig, userService)
	settingsHandler := handler.NewSettingsHandler(settingsService, settingsUseCase)
	userHandler := handler.NewUserHandler(userService, userUseCase, userManageUseCase, imageService, imageUseCase, authService, passkeyService, passkeyUseCase)
	imageHandler := handler.NewImageHandler(imageService, imageUseCase)
	rt := NewRouter(authHandler, systemHandler, settingsHandler, userHandler, imageHandler, dbConfig)

	r := gin.New()
	rt.Init(r)

	type wantRoute struct {
		method string
		path   string
	}
	wants := []wantRoute{
		{method: "GET", path: "/api/ping"},
		{method: "POST", path: "/api/login"},
		{method: "POST", path: "/api/register"},
		{method: "POST", path: "/api/auth/passkey/login/start"},
		{method: "POST", path: "/api/auth/passkey/login/finish"},
		{method: "GET", path: "/api/user/passkeys"},
		{method: "PATCH", path: "/api/user/passkeys/:id/name"},
		{method: "DELETE", path: "/api/user/passkeys/:id"},
		{method: "POST", path: "/api/user/passkeys/register/start"},
		{method: "POST", path: "/api/user/passkeys/register/finish"},
		{method: "GET", path: "/api/user/ping"},
		{method: "GET", path: "/api/admin/stats"},
	}

	have := make(map[string]bool)
	for _, rt := range r.Routes() {
		have[rt.Method+" "+rt.Path] = true
	}

	for _, w := range wants {
		if !have[w.method+" "+w.path] {
			t.Fatalf("缺少路由: %s %s", w.method, w.path)
		}
	}
}
