package image

import (
	"perfect-pic-server/internal/modules/image/handler"
	"perfect-pic-server/internal/modules/image/repo"
	"perfect-pic-server/internal/modules/image/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func New(appService *platformservice.AppService, userService service.UserService, imageStore repo.ImageStore) *Module {
	moduleService := service.New(appService, userService, imageStore)
	moduleHandler := handler.New(moduleService)

	return &Module{
		Service: moduleService,
		Handler: moduleHandler,
	}
}
