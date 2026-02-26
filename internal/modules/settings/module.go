package settings

import (
	"perfect-pic-server/internal/modules/settings/handler"
	"perfect-pic-server/internal/modules/settings/repo"
	"perfect-pic-server/internal/modules/settings/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Module struct {
	Service *service.Service
	Handler *handler.Handler
}

func New(appService *platformservice.AppService, settingStore repo.SettingStore) *Module {
	moduleService := service.New(appService, settingStore)
	moduleHandler := handler.New(moduleService)

	return &Module{
		Service: moduleService,
		Handler: moduleHandler,
	}
}
