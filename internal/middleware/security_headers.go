package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders 添加安全相关的 HTTP 响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止浏览器猜测内容类型
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持 (Clickjacking)
		c.Header("X-Frame-Options", "DENY")

		// Content Security Policy (CSP)
		// 限制资源加载来源，防止 XSS
		// default-src 'self': 默认只允许加载同源资源
		// img-src 'self' data: blob:: 允许加载同源图片以及 data: 和 blob: 协议的图片
		// style-src 'self' 'unsafe-inline': 允许同源样式和内联样式 (很多前端框架需要)
		// script-src 'self': 只允许同源脚本
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self';")

		c.Next()
	}
}
