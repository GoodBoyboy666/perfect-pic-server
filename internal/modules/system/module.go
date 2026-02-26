package system

import (
	"perfect-pic-server/internal/modules/system/handler"
	"perfect-pic-server/internal/modules/system/repo"
	"perfect-pic-server/internal/modules/system/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func New(
	appService *platformservice.AppService,
	systemStore repo.SystemStore,
	userService service.UserService,
	imageService service.ImageService,
) *Module {
	moduleService := service.New(appService, systemStore, userService, imageService)
	moduleHandler := handler.New(moduleService)

	return &Module{
		Service: moduleService,
		Handler: moduleHandler,
	}
}
