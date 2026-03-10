package middleware

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/pkg/jwt"
	"perfect-pic-server/internal/pkg/ratelimit"
	"perfect-pic-server/internal/service"

	"github.com/google/wire"
)

func NewAuthMiddleware(jwt *jwt.JWT, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		jwt:         jwt,
		userService: userService,
	}
}

func NewBodyLimitConfig(dbConfig *config.DBConfig) *BodyLimitMiddleware {
	return &BodyLimitMiddleware{dbConfig: dbConfig}
}

func NewRateLimitMiddleware(
	dbConfig *config.DBConfig,
	tokenBucketLimiter *ratelimit.TokenBucketLimiter,
	intervalLimiter *ratelimit.IntervalLimiter,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		dbConfig:           dbConfig,
		tokenBucketLimiter: tokenBucketLimiter,
		intervalLimiter:    intervalLimiter,
	}
}

func NewSecurityHeadersMiddleware(dbConfig *config.DBConfig) *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{dbConfig: dbConfig}
}

func NewStaticCacheMiddleware(dbConfig *config.DBConfig) *StaticCacheMiddleware {
	return &StaticCacheMiddleware{dbConfig: dbConfig}
}

var MiddlewareSet = wire.NewSet(
	NewAuthMiddleware,
	NewBodyLimitConfig,
	NewRateLimitMiddleware,
	NewSecurityHeadersMiddleware,
	NewStaticCacheMiddleware,
)
