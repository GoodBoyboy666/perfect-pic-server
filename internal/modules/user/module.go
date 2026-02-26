package user

import (
	authservice "perfect-pic-server/internal/modules/auth/service"
	imageservice "perfect-pic-server/internal/modules/image/service"
	"perfect-pic-server/internal/modules/user/handler"
	"perfect-pic-server/internal/modules/user/repo"
	"perfect-pic-server/internal/modules/user/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func NewService(appService *platformservice.AppService, userStore repo.UserStore, imageStore repo.ImageStore) *service.Service {
	return service.New(appService, userStore, imageStore)
}

func New(moduleService *service.Service, authService *authservice.Service, imageService *imageservice.Service) *Module {
	return &Module{
		Service: moduleService,
		Handler: handler.New(moduleService, authService, imageService),
	}
}
