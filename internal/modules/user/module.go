package user

import (
	"perfect-pic-server/internal/modules/user/handler"
	"perfect-pic-server/internal/modules/user/repo"
	"perfect-pic-server/internal/modules/user/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func New(appService *platformservice.AppService, userStore repo.UserStore, imageStore repo.ImageStore) *Module {
	moduleService := service.New(appService, userStore, imageStore)

	return &Module{
		Service: moduleService,
		Handler: nil,
	}
}
