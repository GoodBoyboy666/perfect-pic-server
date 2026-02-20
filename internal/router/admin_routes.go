package router

import (
	adminhandler "perfect-pic-server/internal/handler/admin"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(api *gin.RouterGroup) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth())
	adminGroup.Use(middleware.UserStatusCheck())
	adminGroup.Use(middleware.AdminCheck())

	adminGroup.GET("/stats", adminhandler.GetServerStats)

	adminGroup.GET("/settings", adminhandler.GetSettings)
	adminGroup.PATCH("/settings", adminhandler.UpdateSettings)
	adminGroup.POST("/email/test", adminhandler.SendTestEmail)

	adminGroup.GET("/users", adminhandler.GetUserList)
	adminGroup.GET("/users/:id", adminhandler.GetUserDetail)
	adminGroup.POST("/users", adminhandler.CreateUser)
	adminGroup.PATCH("/users/:id", adminhandler.UpdateUser)
	adminGroup.POST("/users/:id/avatar", adminhandler.UpdateUserAvatar)
	adminGroup.DELETE("/users/:id/avatar", adminhandler.RemoveUserAvatar)
	adminGroup.DELETE("/users/:id", adminhandler.DeleteUser)

	adminGroup.GET("/images", adminhandler.GetImageList)
	adminGroup.DELETE("/images/batch", adminhandler.BatchDeleteImages)
	adminGroup.DELETE("/images/:id", adminhandler.DeleteImage)
}
