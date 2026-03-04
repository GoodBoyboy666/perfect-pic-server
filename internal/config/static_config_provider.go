package config

import (
	"perfect-pic-server/internal/pkg/cache"
	"perfect-pic-server/internal/pkg/database"
	jwtpkg "perfect-pic-server/internal/pkg/jwt"
	redispkg "perfect-pic-server/internal/pkg/redis"
	"time"
)

// NewStaticConfig 提供静态配置快照给 DI。
func NewStaticConfig() *Config {
	cfg := get()
	return &cfg
}

func NewCacheConfig(cfg *Config) *cache.Config {
	return &cache.Config{
		Prefix: cfg.Redis.Prefix,
	}
}

func NewRedisClientConfig(cfg *Config) *redispkg.Config {
	return &redispkg.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
}

func NewJWTConfig(cfg *Config) *jwtpkg.Config {
	return &jwtpkg.Config{
		JWTSecret: []byte(cfg.JWT.Secret),
		Duration:  time.Duration(cfg.JWT.ExpirationHours),
	}
}

func NewDBConnectionConfig(cfg *Config) *database.DbConnectionConfig {
	return &database.DbConnectionConfig{
		Type:     cfg.Database.Type,
		Filename: cfg.Database.Filename,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
		SSL:      cfg.Database.SSL,
	}
}
