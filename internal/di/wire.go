//go:build wireinject
// +build wireinject

package di

import (
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/router"
	"perfect-pic-server/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func InitializeApplication(gormDB *gorm.DB) (*Application, error) {
	wire.Build(
		repository.NewUserRepository,
		repository.NewImageRepository,
		repository.NewSettingRepository,
		repository.NewSystemRepository,
		service.NewAppService,
		handler.NewAuthHandler,
		handler.NewSystemHandler,
		handler.NewSettingsHandler,
		handler.NewUserHandler,
		handler.NewImageHandler,
		router.NewRouter,
		NewApplication,
	)
	return nil, nil
}
