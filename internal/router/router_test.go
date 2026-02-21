package router

import (
	"testing"

	"perfect-pic-server/internal/handler"
	adminhandler "perfect-pic-server/internal/handler/admin"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"

	"github.com/gin-gonic/gin"
)

// 测试内容：验证核心 API 路由被正确注册。
func TestInitRouter_RegistersCoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := testutils.SetupDB(t)
	appService := service.NewAppService(repository.NewRepositories(
		repository.NewUserRepository(gdb),
		repository.NewImageRepository(gdb),
		repository.NewSettingRepository(gdb),
		repository.NewSystemRepository(gdb),
	))
	appService.ClearCache()
	h := handler.NewHandler(appService)
	ah := adminhandler.NewHandler(appService)
	rt := NewRouter(h, ah, appService)

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
		{method: "GET", path: "/api/user/passkeys"},
		{method: "POST", path: "/api/user/passkeys/register/start"},
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
