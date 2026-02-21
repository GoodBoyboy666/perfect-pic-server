package di

import (
	"perfect-pic-server/internal/router"
	"perfect-pic-server/internal/service"
)

type Application struct {
	Router  *router.Router
	Service *service.AppService
}

func NewApplication(r *router.Router, s *service.AppService) *Application {
	return &Application{
		Router:  r,
		Service: s,
	}
}
