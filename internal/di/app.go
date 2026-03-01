package di

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/router"
)

type Application struct {
	Router   *router.Router
	DbConfig *config.DBConfig
}

func NewApplication(r *router.Router, dbConfig *config.DBConfig) *Application {
	return &Application{
		Router:   r,
		DbConfig: dbConfig,
	}
}
