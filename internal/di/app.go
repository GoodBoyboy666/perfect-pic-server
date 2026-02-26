package di

import (
	"perfect-pic-server/internal/modules"
	"perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/router"
)

type Application struct {
	Modules *modules.AppModules
	Router  *router.Router
	Service *service.AppService
}

func NewApplication(m *modules.AppModules, r *router.Router, s *service.AppService) *Application {
	return &Application{
		Modules: m,
		Router:  r,
		Service: s,
	}
}
