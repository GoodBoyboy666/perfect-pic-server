package middleware

import (
	"net/http"
	"net/http/httptest"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestSecurityHeaders_DynamicCSP 测试内容：验证安全相关响应头会被正确设置，且 CSP 根据验证码配置动态变化。
func TestSecurityHeaders_DynamicCSP(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)

	// 定义测试用例：不同的验证码提供方对应的 CSP 关键字段
	testCases := []struct {
		name     string
		provider string
		wantCSP  []string
	}{
		{
			name:     "Geetest",
			provider: service.CaptchaProviderGeetest,
			wantCSP: []string{
				"https://*.geetest.com",
				"script-src 'self' https://*.geetest.com",
			},
		},
		{
			name:     "Turnstile",
			provider: service.CaptchaProviderTurnstile,
			wantCSP: []string{
				"https://challenges.cloudflare.com",
				"script-src 'self' https://challenges.cloudflare.com",
			},
		},
		{
			name:     "Recaptcha",
			provider: service.CaptchaProviderRecaptcha,
			wantCSP: []string{
				"https://www.google.com",
				"https://www.gstatic.com",
				"script-src 'self' https://www.google.com https://www.gstatic.com",
			},
		},
		{
			name:     "hCaptcha",
			provider: service.CaptchaProviderHcaptcha,
			wantCSP: []string{
				"https://*.hcaptcha.com",
				"https://hcaptcha.com",
				"script-src 'self' https://*.hcaptcha.com https://hcaptcha.com",
			},
		},
		{
			name:     "Image (Default)",
			provider: service.CaptchaProviderImage,
			wantCSP: []string{
				"default-src 'self'",
				// 默认只允许同源脚本
				"script-src 'self';",
				"frame-ancestors 'none'",
			},
		},
		{
			name:     "Disabled",
			provider: service.CaptchaProviderDisabled,
			wantCSP: []string{
				"default-src 'self'",
				"script-src 'self';",
				"frame-ancestors 'none'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. 更新数据库中的 captcha_provider 设置
			// 注意：setupTestDB 已经初始化了数据库
			setting := model.Setting{
				Key:   consts.ConfigCaptchaProvider,
				Value: tc.provider,
			}
			// 使用 Save 更新或插入，确保 setting 存在
			// 因为 Key 是主键或唯一索引
			if err := db.Save(&setting).Error; err != nil {
				t.Fatalf("Failed to save setting: %v", err)
			}

			// 2. 清除 service 层的缓存，强制下一次读取从 DB 获取最新值
			testService.ClearCache()

			// 3. 构建 Gin Engine 并应用中间件
			r := gin.New()
			r.Use(SecurityHeaders(testService))
			r.GET("/ping", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// 4. 发起请求
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			r.ServeHTTP(w, req)

			// 5. 验证 CSP Header
			csp := w.Header().Get("Content-Security-Policy")

			// Debug info if needed
			// t.Logf("Provider: %s, CSP: %s", tc.provider, csp)

			for _, want := range tc.wantCSP {
				// 简单的字符串包含检查
				if !strings.Contains(csp, want) {
					t.Errorf("[%s] CSP header missing expected content.\nWant substring: %q\nGot CSP: %s", tc.name, want, csp)
				}
			}

			// 6. 验证其他静态安全头
			if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("[%s] X-Content-Type-Options = %q, want nosniff", tc.name, got)
			}
			if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
				t.Errorf("[%s] X-Frame-Options = %q, want DENY", tc.name, got)
			}
		})
	}
}
