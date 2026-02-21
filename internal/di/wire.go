//go:build wireinject
// +build wireinject

package di

import (
	"perfect-pic-server/internal/handler"
	adminhandler "perfect-pic-server/internal/handler/admin"
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
		repository.NewRepositories,
		service.NewAppService,
		handler.NewHandler,
		adminhandler.NewHandler,
		router.NewRouter,
		NewApplication,
	)
	return nil, nil
}
