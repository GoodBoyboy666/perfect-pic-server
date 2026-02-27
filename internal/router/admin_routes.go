package router

import (
	"perfect-pic-server/internal/middleware"
	imagehandler "perfect-pic-server/internal/modules/image/handler"
	settingshandler "perfect-pic-server/internal/modules/settings/handler"
	systemhandler "perfect-pic-server/internal/modules/system/handler"
	userhandler "perfect-pic-server/internal/modules/user/handler"
	"perfect-pic-server/internal/platform/service"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(
	api *gin.RouterGroup,
	systemHandler *systemhandler.Handler,
	settingsHandler *settingshandler.Handler,
	userHandler *userhandler.Handler,
	imageHandler *imagehandler.Handler,
	appService *service.AppService,
) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth())
	adminGroup.Use(middleware.UserStatusCheck())
	adminGroup.Use(middleware.AdminCheck())
	bodyLimit := middleware.BodyLimitMiddleware(appService)
	uploadBodyLimit := middleware.UploadBodyLimitMiddleware(appService)

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
