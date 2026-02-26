package auth

import (
	"perfect-pic-server/internal/modules/auth/handler"
	"perfect-pic-server/internal/modules/auth/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func New(appService *platformservice.AppService, userService service.UserService) *Module {
	moduleService := service.New(appService, userService)
	moduleHandler := handler.New(moduleService)

	return &Module{
		Service: moduleService,
		Handler: moduleHandler,
	}
}
