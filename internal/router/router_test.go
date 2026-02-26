package router

import (
	"testing"

	"perfect-pic-server/internal/modules"
	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	systemrepo "perfect-pic-server/internal/modules/system/repo"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	"perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"github.com/gin-gonic/gin"
)

// 测试内容：验证核心 API 路由被正确注册。
func TestInitRouter_RegistersCoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := testutils.SetupDB(t)
	userStore := userrepo.NewUserRepository(gdb)
	imageStore := imagerepo.NewImageRepository(gdb)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	systemStore := systemrepo.NewSystemRepository(gdb)
	appService := service.NewAppService(settingStore)
	appService.ClearCache()
	appModules := modules.New(appService, userStore, imageStore, settingStore, systemStore)
	rt := NewRouter(appModules, appService)

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
