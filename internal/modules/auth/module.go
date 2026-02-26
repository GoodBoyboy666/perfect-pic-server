package auth

import (
	"perfect-pic-server/internal/modules/auth/handler"
	"perfect-pic-server/internal/modules/auth/service"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	userservice "perfect-pic-server/internal/modules/user/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func New(appService *platformservice.AppService, userStore userrepo.UserStore, userService *userservice.Service) *Module {
	moduleService := service.New(appService, userStore, userService)
	moduleHandler := handler.New(moduleService)

	return &Module{
		Service: moduleService,
		Handler: moduleHandler,
	}
}
