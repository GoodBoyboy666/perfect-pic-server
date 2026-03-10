package middleware

import (
	"net/http"
	"net/http/httptest"
	"perfect-pic-server/internal/pkg/cache"
	"perfect-pic-server/internal/pkg/jwt"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"testing"
	"time"

	"perfect-pic-server/internal/model"

	"github.com/gin-gonic/gin"
)

func buildTestJWT() *jwt.JWT {
	return jwt.NewJWT(&jwt.Config{
		JWTSecret: []byte("test_jwt_secret"),
		Duration:  time.Hour,
	})
}

func buildTestStatusCache() *cache.Store {
	return cache.NewStore(nil, &cache.Config{Prefix: "test"})
}

// 测试内容：验证缺少 Authorization 头时返回 401。
func TestJWTAuth_MissingHeaderUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := buildTestJWT()
	authMiddleware := NewAuthMiddleware(jwtService, nil)

	r := gin.New()
	r.GET("/x", authMiddleware.JWTAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际为 %d", w.Code)
	}
}

// 测试内容：验证有效登录令牌会在上下文中设置用户信息。
func TestJWTAuth_ValidTokenSetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := buildTestJWT()
	authMiddleware := NewAuthMiddleware(jwtService, nil)

	r := gin.New()
	r.GET("/x", authMiddleware.JWTAuth(), func(c *gin.Context) {
		id, _ := c.Get("id")
		username, _ := c.Get("username")
		if id != uint(1) || username != "alice" {
			c.JSON(500, gin.H{"bad": true})
			return
		}
		if _, exists := c.Get("admin"); exists {
			c.JSON(500, gin.H{"bad": "unexpected admin context"})
			return
		}
		c.Status(http.StatusOK)
	})

	token, err := jwtService.GenerateLoginToken(1, "alice", true)
	if err != nil {
		t.Fatalf("GenerateLoginToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d", w.Code)
	}
}

// 测试内容：验证被禁用用户状态会被拦截并返回 403。
func TestUserStatusCheck_BannedForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)
	resetStatusCache()
	statusCache := buildTestStatusCache()
	userStore := repository.NewUserRepository(gdb)
	userService := service.NewUserService(userStore, testService, statusCache, buildTestJWT())
	authMiddleware := NewAuthMiddleware(buildTestJWT(), userService)

	u := model.User{Username: "alice", Password: "x", Status: 2, Email: "a@example.com"}
	if err := testGormDB.Create(&u).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	r := gin.New()
	r.GET("/x",
		func(c *gin.Context) { c.Set("id", u.ID); c.Next() },
		authMiddleware.UserStatusCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("期望 403，实际为 %d", w.Code)
	}
}

// 测试内容：验证正常用户状态通过检查并返回 200。
func TestUserStatusCheck_NormalOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)
	resetStatusCache()
	statusCache := buildTestStatusCache()
	userStore := repository.NewUserRepository(gdb)
	userService := service.NewUserService(userStore, testService, statusCache, buildTestJWT())
	authMiddleware := NewAuthMiddleware(buildTestJWT(), userService)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = testGormDB.Create(&u).Error

	r := gin.New()
	r.GET("/x",
		func(c *gin.Context) { c.Set("id", u.ID); c.Next() },
		authMiddleware.UserStatusCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d", w.Code)
	}
}

// 测试内容：验证缺失 id、类型错误、未找到用户与禁用状态的错误分支处理。
func TestUserStatusCheck_ErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)
	resetStatusCache()
	statusCache := buildTestStatusCache()
	userStore := repository.NewUserRepository(gdb)
	userService := service.NewUserService(userStore, testService, statusCache, buildTestJWT())
	authMiddleware := NewAuthMiddleware(buildTestJWT(), userService)

	// 缺少 id
	r1 := gin.New()
	r1.GET("/x", authMiddleware.UserStatusCheck(), func(c *gin.Context) { c.Status(http.StatusOK) })
	w1 := httptest.NewRecorder()
	r1.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w1.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际为 %d", w1.Code)
	}

	// id 类型错误
	r2 := gin.New()
	r2.GET("/x",
		func(c *gin.Context) { c.Set("id", "bad"); c.Next() },
		authMiddleware.UserStatusCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际为 %d", w2.Code)
	}

	// 用户未找到
	r3 := gin.New()
	r3.GET("/x",
		func(c *gin.Context) { c.Set("id", uint(999)); c.Next() },
		authMiddleware.UserStatusCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w3 := httptest.NewRecorder()
	r3.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w3.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际为 %d", w3.Code)
	}

	// 状态 3 已禁用
	u := model.User{Username: "d", Password: "x", Status: 3, Email: "d@example.com"}
	_ = testGormDB.Create(&u).Error
	r4 := gin.New()
	r4.GET("/x",
		func(c *gin.Context) { c.Set("id", u.ID); c.Next() },
		authMiddleware.UserStatusCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w4 := httptest.NewRecorder()
	r4.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w4.Code != http.StatusForbidden {
		t.Fatalf("期望 403，实际为 %d", w4.Code)
	}
}

// 测试内容：验证管理员校验基于数据库权限（非 JWT claim），并支持管理员通过。
func TestAdminCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)
	resetStatusCache()
	statusCache := buildTestStatusCache()
	userStore := repository.NewUserRepository(gdb)
	userService := service.NewUserService(userStore, testService, statusCache, buildTestJWT())
	authMiddleware := NewAuthMiddleware(buildTestJWT(), userService)

	normalUser := model.User{Username: "normal_user", Password: "x", Status: 1, Email: "normal@example.com", Admin: false}
	if err := testGormDB.Create(&normalUser).Error; err != nil {
		t.Fatalf("create normal user failed: %v", err)
	}
	adminUser := model.User{Username: "admin_user", Password: "x", Status: 1, Email: "admin@example.com", Admin: true}
	if err := testGormDB.Create(&adminUser).Error; err != nil {
		t.Fatalf("create admin user failed: %v", err)
	}

	// 非管理员返回 403。
	r := gin.New()
	r.GET("/admin",
		func(c *gin.Context) { c.Set("id", normalUser.ID); c.Next() },
		authMiddleware.AdminCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("期望 403 for non-admin(db)，实际为 %d", w.Code)
	}

	// 管理员返回 200。
	r2 := gin.New()
	r2.GET("/admin",
		func(c *gin.Context) { c.Set("id", adminUser.ID); c.Next() },
		authMiddleware.AdminCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("期望 200 for admin(db)，实际为 %d", w2.Code)
	}

	// 用户不存在时返回 401（与 UserStatusCheck 一致）。
	r3 := gin.New()
	r3.GET("/admin",
		func(c *gin.Context) { c.Set("id", uint(999999)); c.Next() },
		authMiddleware.AdminCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w3 := httptest.NewRecorder()
	r3.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if w3.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401 for missing user，实际为 %d", w3.Code)
	}

	// 用户服务未初始化时返回 500。
	r4 := gin.New()
	r4.GET("/admin",
		func(c *gin.Context) { c.Set("id", adminUser.ID); c.Next() },
		NewAuthMiddleware(buildTestJWT(), nil).AdminCheck(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	w4 := httptest.NewRecorder()
	r4.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if w4.Code != http.StatusInternalServerError {
		t.Fatalf("期望 500 for nil userService，实际为 %d", w4.Code)
	}
}

// 测试内容：验证清除用户状态缓存会移除本地缓存条目。
func TestClearUserStatusCache_RemovesLocalCache(t *testing.T) {
	gdb := setupTestDB(t)
	resetStatusCache()
	statusCache := buildTestStatusCache()
	userStore := repository.NewUserRepository(gdb)
	userService := service.NewUserService(userStore, testService, statusCache, buildTestJWT())

	key := statusCache.RedisKey("auth", "user_status", "1")
	statusCache.Set(key, "2", time.Minute)
	userService.ClearUserStatusCache(uint(1))
	if _, ok := statusCache.Get(key); ok {
		t.Fatalf("期望缓存条目被移除")
	}
}
