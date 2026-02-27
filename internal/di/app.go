package di

import (
	"perfect-pic-server/internal/router"
	"perfect-pic-server/internal/service"
)

type Application struct {
	Router  *router.Router
	Service *service.Service
}

func NewApplication(r *router.Router, s *service.Service) *Application {
	return &Application{
		Router:  r,
		Service: s,
	}
}
