package di

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/router"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Application struct {
	Router                *router.Router
	DbConfig              *config.DBConfig
	GormDB                *gorm.DB
	RedisDB               *redis.Client
	StaticConfig          *config.Config
	StaticCacheMiddleware *middleware.StaticCacheMiddleware
}

func NewApplication(r *router.Router, dbConfig *config.DBConfig, gormDB *gorm.DB, redisDB *redis.Client, staticConfig *config.Config, staticCacheMiddleware *middleware.StaticCacheMiddleware) *Application {
	return &Application{
		Router:                r,
		DbConfig:              dbConfig,
		GormDB:                gormDB,
		RedisDB:               redisDB,
		StaticConfig:          staticConfig,
		StaticCacheMiddleware: staticCacheMiddleware,
	}
}
