package router

import (
	"perfect-pic-server/internal/middleware"
	imagehandler "perfect-pic-server/internal/modules/image/handler"
	settingshandler "perfect-pic-server/internal/modules/settings/handler"
	systemhandler "perfect-pic-server/internal/modules/system/handler"
	userhandler "perfect-pic-server/internal/modules/user/handler"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(api *gin.RouterGroup, systemHandler *systemhandler.Handler, settingsHandler *settingshandler.Handler, userHandler *userhandler.Handler, imageHandler *imagehandler.Handler) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth())
	adminGroup.Use(middleware.UserStatusCheck())
	adminGroup.Use(middleware.AdminCheck())

	adminGroup.GET("/stats", systemHandler.GetServerStats)

	adminGroup.GET("/settings", settingsHandler.GetSettings)
	adminGroup.PATCH("/settings", settingsHandler.UpdateSettings)
	adminGroup.POST("/email/test", settingsHandler.SendTestEmail)

	adminGroup.GET("/users", userHandler.GetUserList)
	adminGroup.GET("/users/:id", userHandler.GetUserDetail)
	adminGroup.POST("/users", userHandler.CreateUser)
	adminGroup.PATCH("/users/:id", userHandler.UpdateUser)
	adminGroup.POST("/users/:id/avatar", userHandler.UpdateUserAvatar)
	adminGroup.DELETE("/users/:id/avatar", userHandler.RemoveUserAvatar)
	adminGroup.DELETE("/users/:id", userHandler.DeleteUser)

	adminGroup.GET("/images", imageHandler.GetImageList)
	adminGroup.DELETE("/images/batch", imageHandler.BatchDeleteImages)
	adminGroup.DELETE("/images/:id", imageHandler.DeleteImage)
}
