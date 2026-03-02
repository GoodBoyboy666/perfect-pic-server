//go:build wireinject
// +build wireinject

package di

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/handler"
	pkgcaptcha "perfect-pic-server/internal/pkg/captcha"
	"perfect-pic-server/internal/pkg/redis"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/router"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/usecase/admin"
	"perfect-pic-server/internal/usecase/app"

	"github.com/google/wire"
)

func InitializeApplication() (*Application, error) {
	wire.Build(
		db.NewGormDB,
		redis.NewRedisClient,
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
		pkgcaptcha.NewCaptcha,
		service.NewInitService,
		service.NewPasskeyService,
		admin.NewUserManageUseCase,
		admin.NewSettingsUseCase,
		admin.NewStatUseCase,
		app.NewAuthUseCase,
		app.NewUserUseCase,
		app.NewImageUseCase,
		app.NewPasskeyUseCase,
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
