package router

import (
	"testing"

	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"

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
	appService := service.NewAppService(userStore, imageStore, settingStore, systemStore)
	appService.ClearCache()
	authHandler := handler.NewAuthHandler(appService)
	systemHandler := handler.NewSystemHandler(appService)
	settingsHandler := handler.NewSettingsHandler(appService)
	userHandler := handler.NewUserHandler(appService)
	imageHandler := handler.NewImageHandler(appService)
	rt := NewRouter(authHandler, systemHandler, settingsHandler, userHandler, imageHandler, appService)

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
