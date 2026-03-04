//go:build wireinject
// +build wireinject

package di

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/pkg/cache"
	"perfect-pic-server/internal/pkg/database"
	pkgmail "perfect-pic-server/internal/pkg/email"
	jwtpkg "perfect-pic-server/internal/pkg/jwt"
	"perfect-pic-server/internal/pkg/ratelimit"
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
		config.NewStaticConfig,
		config.NewRedisClientConfig,
		config.NewJWTConfig,
		config.NewCacheConfig,
		config.NewDBConnectionConfig,
		database.NewGormDB,
		redis.NewRedisClient,
		jwtpkg.NewJWT,
		cache.NewStore,
		ratelimit.NewTokenBucketLimiter,
		ratelimit.NewIntervalLimiter,
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
		pkgmail.NewMailer,
		service.NewInitService,
		service.NewPasskeyService,
		admin.NewUserManageUseCase,
		admin.NewSettingsUseCase,
		admin.NewStatUseCase,
		app.NewAuthUseCase,
		app.NewUserUseCase,
		app.NewImageUseCase,
		app.NewPasskeyUseCase,
		middleware.NewAuthMiddleware,
		middleware.NewRateLimitMiddleware,
		middleware.NewBodyLimitConfig,
		middleware.NewSecurityHeadersMiddleware,
		middleware.NewStaticCacheMiddleware,
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
