package router

import (
	adminhandler "perfect-pic-server/internal/handler/admin"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(api *gin.RouterGroup, h *adminhandler.Handler) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth())
	adminGroup.Use(middleware.UserStatusCheck())
	adminGroup.Use(middleware.AdminCheck())

	adminGroup.GET("/stats", h.GetServerStats)

	adminGroup.GET("/settings", h.GetSettings)
	adminGroup.PATCH("/settings", h.UpdateSettings)
	adminGroup.POST("/email/test", h.SendTestEmail)

	adminGroup.GET("/users", h.GetUserList)
	adminGroup.GET("/users/:id", h.GetUserDetail)
	adminGroup.POST("/users", h.CreateUser)
	adminGroup.PATCH("/users/:id", h.UpdateUser)
	adminGroup.POST("/users/:id/avatar", h.UpdateUserAvatar)
	adminGroup.DELETE("/users/:id/avatar", h.RemoveUserAvatar)
	adminGroup.DELETE("/users/:id", h.DeleteUser)

	adminGroup.GET("/images", h.GetImageList)
	adminGroup.DELETE("/images/batch", h.BatchDeleteImages)
	adminGroup.DELETE("/images/:id", h.DeleteImage)
}
