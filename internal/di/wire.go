//go:build wireinject
// +build wireinject

package di

import (
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/router"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/usecase/admin"
	"perfect-pic-server/internal/usecase/app"
	"perfect-pic-server/internal/config"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func InitializeApplication(gormDB *gorm.DB) (*Application, error) {
	wire.Build(
		repository.NewUserRepository,
		repository.NewImageRepository,
		repository.NewSettingRepository,
		repository.NewSystemRepository,
		repository.NewPasskeyRepository,
		config.NewDBConfig,
		service.NewUserService,
		service.NewImageService,
		service.NewSettingsService,
		service.NewAuthService,
		service.NewEmailService,
		service.NewCaptchaService,
		service.NewInitService,
		service.NewPasskeyService,
		admin.NewUserManageUseCase,
		admin.NewSettingsUseCase,
		admin.NewStatUseCase,
		admin.NewAdminUseCase,
		app.NewAuthUseCase,
		app.NewUserUseCase,
		app.NewImageUseCase,
		app.NewPasskeyUseCase,
		app.NewAppUseCase,
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
