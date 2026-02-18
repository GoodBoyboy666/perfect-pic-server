package middleware

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 添加安全相关的 HTTP 响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止浏览器猜测内容类型
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持 (Clickjacking)
		c.Header("X-Frame-Options", "DENY")

		// Content Security Policy (CSP)
		// 动态根据当前的验证码服务商配置 CSP
		var (
			imgSrc     = []string{"'self'", "data:", "blob:"}
			styleSrc   = []string{"'self'", "'unsafe-inline'"}
			scriptSrc  = []string{"'self'"}
			connectSrc = []string{"'self'"}
			frameSrc   = []string{"'self'"}
		)

		provider := strings.ToLower(service.GetString(consts.ConfigCaptchaProvider))

		switch provider {
		case service.CaptchaProviderGeetest:
			geetest := "https://*.geetest.com"
			imgSrc = append(imgSrc, geetest)
			styleSrc = append(styleSrc, geetest)
			scriptSrc = append(scriptSrc, geetest)
			connectSrc = append(connectSrc, geetest)
			frameSrc = append(frameSrc, geetest)

		case service.CaptchaProviderTurnstile:
			turnstile := "https://challenges.cloudflare.com"
			scriptSrc = append(scriptSrc, turnstile)
			connectSrc = append(connectSrc, turnstile)
			frameSrc = append(frameSrc, turnstile)

		case service.CaptchaProviderRecaptcha:
			google := "https://www.google.com"
			gstatic := "https://www.gstatic.com"
			imgSrc = append(imgSrc, google, gstatic)
			scriptSrc = append(scriptSrc, google, gstatic)
			connectSrc = append(connectSrc, google)
			frameSrc = append(frameSrc, google)

		case service.CaptchaProviderHcaptcha:
			hcaptcha := "https://*.hcaptcha.com"
			hcaptchaMain := "https://hcaptcha.com"
			imgSrc = append(imgSrc, hcaptcha, hcaptchaMain)
			styleSrc = append(styleSrc, hcaptcha, hcaptchaMain)
			scriptSrc = append(scriptSrc, hcaptcha, hcaptchaMain)
			connectSrc = append(connectSrc, hcaptcha, hcaptchaMain)
			frameSrc = append(frameSrc, hcaptcha, hcaptchaMain)
		}

		csp := "default-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none'; " +
			"img-src " + strings.Join(imgSrc, " ") + "; " +
			"style-src " + strings.Join(styleSrc, " ") + "; " +
			"script-src " + strings.Join(scriptSrc, " ") + "; " +
			"connect-src " + strings.Join(connectSrc, " ") + "; " +
			"frame-src " + strings.Join(frameSrc, " ") + ";"

		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}
