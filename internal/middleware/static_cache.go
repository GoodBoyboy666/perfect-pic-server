package middleware

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"

	"github.com/gin-gonic/gin"
)

type StaticCacheMiddleware struct {
	dbConfig *config.DBConfig
}

// StaticCacheMiddleware 为静态资源添加 Cache-Control 头
// 缓存策略由 ConfigStaticCacheControl 配置决定
func (m *StaticCacheMiddleware) StaticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cc := m.dbConfig.GetString(consts.ConfigStaticCacheControl)
		if cc != "" {
			c.Header("Cache-Control", cc)
		}
		c.Next()
	}
}
