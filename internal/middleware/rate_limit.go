package middleware

import (
	"fmt"
	"net/http"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/pkg/ratelimit"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultSensitiveOperationInterval = 2 * time.Minute
)

type RateLimitMiddleware struct {
	dbConfig           *config.DBConfig
	tokenBucketLimiter *ratelimit.TokenBucketLimiter
	intervalLimiter    *ratelimit.IntervalLimiter
}

func NewRateLimitMiddleware(
	dbConfig *config.DBConfig,
	tokenBucketLimiter *ratelimit.TokenBucketLimiter,
	intervalLimiter *ratelimit.IntervalLimiter,
) *RateLimitMiddleware {
	if tokenBucketLimiter == nil {
		tokenBucketLimiter = ratelimit.NewTokenBucketLimiter(nil)
	}
	if intervalLimiter == nil {
		intervalLimiter = ratelimit.NewIntervalLimiter(nil)
	}
	return &RateLimitMiddleware{
		dbConfig:           dbConfig,
		tokenBucketLimiter: tokenBucketLimiter,
		intervalLimiter:    intervalLimiter,
	}
}

// RateLimit 按“每秒速率 + 突发容量”进行限流（令牌桶）。
// rpsKey/burstKey 分别对应配置中的 RPS 和 Burst。
func (m *RateLimitMiddleware) RateLimit(rpsKey string, burstKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.dbConfig.GetBool(consts.ConfigRateLimitEnabled) {
			c.Next()
			return
		}

		currentRPS := m.dbConfig.GetFloat64(rpsKey)
		currentBurst := m.dbConfig.GetInt(burstKey)
		ip := c.ClientIP()

		if !m.tokenBucketLimiter.Allow(ip, "rate", rpsKey, burstKey, currentRPS, currentBurst) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IntervalRate 按数据库配置的最小调用间隔进行限流。
// intervalKey 对应设置项，值为秒数（int），例如 120 表示 2 分钟。
func (m *RateLimitMiddleware) IntervalRate(intervalKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.dbConfig.GetBool(consts.ConfigEnableSensitiveRateLimit) {
			c.Next()
			return
		}

		interval := getIntervalBySettingKey(m.dbConfig, intervalKey)
		ip := c.ClientIP()

		if !m.intervalLimiter.Allow(ip, intervalKey, interval) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("操作过于频繁，请等待 %v 后再试", interval)})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getIntervalBySettingKey(dbConfig *config.DBConfig, intervalKey string) time.Duration {
	seconds := dbConfig.GetInt(intervalKey)
	if seconds <= 0 {
		return defaultSensitiveOperationInterval
	}
	return time.Duration(seconds) * time.Second
}
