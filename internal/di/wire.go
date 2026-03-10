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
		config.StaticConfigSet,
		config.NewDBConfig,
		database.NewGormDB,
		redis.NewRedisClient,
		jwtpkg.NewJWT,
		cache.NewStore,
		ratelimit.RateLimiter,
		pkgmail.NewMailer,
		repository.RepoSet,
		service.ServiceSet,
		admin.AdminUseCaseSet,
		app.UseCaseSet,
		middleware.MiddlewareSet,
		handler.HandlerSet,
		router.NewRouter,
		NewApplication,
	)
	return nil, nil
}
