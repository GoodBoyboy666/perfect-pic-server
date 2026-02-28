package router

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(
	api *gin.RouterGroup,
	systemHandler *handler.SystemHandler,
	settingsHandler *handler.SettingsHandler,
	userHandler *handler.UserHandler,
	imageHandler *handler.ImageHandler,
	dbConfig *config.DBConfig,
) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth())
	adminGroup.Use(middleware.UserStatusCheck())
	adminGroup.Use(middleware.AdminCheck())
	bodyLimit := middleware.BodyLimitMiddleware(dbConfig)
	uploadBodyLimit := middleware.UploadBodyLimitMiddleware(dbConfig)

	adminGroup.GET("/stats", systemHandler.GetServerStats)

	adminGroup.GET("/settings", settingsHandler.GetSettings)
	adminGroup.PATCH("/settings", bodyLimit, settingsHandler.UpdateSettings)
	adminGroup.POST("/email/test", bodyLimit, settingsHandler.SendTestEmail)

	adminGroup.GET("/users", userHandler.GetUserList)
	adminGroup.GET("/users/:id", userHandler.GetUserDetail)
	adminGroup.POST("/users", bodyLimit, userHandler.CreateUser)
	adminGroup.PATCH("/users/:id", bodyLimit, userHandler.UpdateUser)
	adminGroup.DELETE("/users/:id/avatar", userHandler.RemoveUserAvatar)
	adminGroup.DELETE("/users/:id", userHandler.DeleteUser)

	adminGroup.POST("/users/:id/avatar", uploadBodyLimit, userHandler.UpdateUserAvatar)

	adminGroup.GET("/images", imageHandler.GetImageList)
	adminGroup.DELETE("/images/batch", bodyLimit, imageHandler.BatchDeleteImages)
	adminGroup.DELETE("/images/:id", imageHandler.DeleteImage)
}
