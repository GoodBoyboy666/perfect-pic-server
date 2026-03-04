package middleware

import (
	"net/http"
	"perfect-pic-server/internal/pkg/jwt"
	"perfect-pic-server/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwt         *jwt.JWT
	userService *service.UserService
}

func NewAuthMiddleware(jwt *jwt.JWT, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		jwt:         jwt,
		userService: userService,
	}
}

func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.jwt == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "认证组件未初始化"})
			c.Abort()
			return
		}

		// 获取请求头 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要认证才能访问"})
			c.Abort()
			return
		}

		// 检查格式是否为 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式错误"})
			c.Abort()
			return
		}

		//解析 Token
		claims, err := m.jwt.ParseLoginToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			c.Abort()
			return
		}

		c.Set("id", claims.ID)
		c.Set("username", claims.Username)
		c.Set("admin", claims.Admin)
		c.Next()
	}
}

// UserStatusCheck 检查用户状态是否被封禁
//
//nolint:gocyclo
func (m *AuthMiddleware) UserStatusCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.userService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户服务未初始化"})
			c.Abort()
			return
		}

		userID, exists := c.Get("id")
		if !exists {
			// 如果没有上下文中的 id，说明 JWT 中间件可能未执行或失败但未 Abort（理论上不可能），或者顺序不对
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
			c.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID类型"})
			c.Abort()
			return
		}

		currentStatus, err := m.userService.GetUserStatus(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		if currentStatus == 2 {
			c.JSON(http.StatusForbidden, gin.H{"error": "账号已被封禁"})
			c.Abort()
			return
		}
		if currentStatus == 3 {
			c.JSON(http.StatusForbidden, gin.H{"error": "账号已停用"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exist := c.Get("admin")
		isAdmin, ok := value.(bool)
		if !exist || !ok || !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限才能访问"})
			c.Abort()
			return
		}
		c.Next()
	}
}
