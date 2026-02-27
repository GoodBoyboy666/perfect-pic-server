package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/service"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips sync.Map
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type redisFallbackLogState struct {
	mu         sync.Mutex
	degraded   bool
	lastWarnAt time.Time
}

const (
	defaultSensitiveOperationInterval = 2 * time.Minute
	redisFallbackLogInterval          = 1 * time.Minute
)

var redisFallbackLogStates sync.Map

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		r: r,
		b: b,
	}

	go i.cleanupLoop()

	return i
}

func getRedisFallbackLogState(scope string) *redisFallbackLogState {
	state, _ := redisFallbackLogStates.LoadOrStore(scope, &redisFallbackLogState{})
	typedState, ok := state.(*redisFallbackLogState)
	if !ok {
		fallbackState := &redisFallbackLogState{}
		redisFallbackLogStates.Store(scope, fallbackState)
		return fallbackState
	}
	return typedState
}

func logRedisFallbackDegraded(scope string, err error) {
	logRedisFallbackDegradedWithTarget(scope, "内存限流", err)
}

func logRedisFallbackDegradedWithTarget(scope string, fallbackTarget string, err error) {
	state := getRedisFallbackLogState(scope)
	now := time.Now()

	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.degraded {
		state.degraded = true
		state.lastWarnAt = now
		log.Printf(
			"⚠️ Redis %s 检查失败，已降级到%s（后续每 %s 最多记录一次）: %v",
			scope,
			fallbackTarget,
			redisFallbackLogInterval,
			err,
		)
		return
	}

	if now.Sub(state.lastWarnAt) >= redisFallbackLogInterval {
		state.lastWarnAt = now
		log.Printf("⚠️ Redis %s 仍不可用，继续使用%s: %v", scope, fallbackTarget, err)
	}
}

func logRedisFallbackRecovered(scope string) {
	logRedisFallbackRecoveredWithTarget(scope, "Redis 限流")
}

func logRedisFallbackRecoveredWithTarget(scope string, recoveredTarget string) {
	state := getRedisFallbackLogState(scope)

	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.degraded {
		return
	}

	state.degraded = false
	state.lastWarnAt = time.Time{}
	log.Printf("✅ Redis %s 已恢复，切回%s", scope, recoveredTarget)
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	if v, ok := i.ips.Load(ip); ok {
		if c, ok := v.(*client); ok {
			c.lastSeen = time.Now()
			return c.limiter
		}
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Double check
	if v, ok := i.ips.Load(ip); ok {
		if c, ok := v.(*client); ok {
			c.lastSeen = time.Now()
			return c.limiter
		}
	}

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips.Store(ip, &client{limiter: limiter, lastSeen: time.Now()})

	return limiter
}

func (i *IPRateLimiter) cleanupLoop() {
	for {
		time.Sleep(1 * time.Minute)
		i.ips.Range(func(key, value interface{}) bool {
			client, ok := value.(*client)
			if !ok {
				i.ips.Delete(key)
				return true
			}
			if time.Since(client.lastSeen) > 3*time.Minute {
				i.ips.Delete(key)
			}
			return true
		})
	}
}

// RateLimitMiddleware 按“每秒速率 + 突发容量”进行限流（令牌桶）。
// rpsKey/burstKey 分别对应配置中的 RPS 和 Burst。
func RateLimitMiddleware(appService *service.Service, rpsKey string, burstKey string) gin.HandlerFunc {
	// 每个中间件实例共用一个 IPRateLimiter，并按 IP 复用 limiter。
	// 这样可以避免每次请求都创建新 limiter。
	var limiter *IPRateLimiter
	var once sync.Once

	return func(c *gin.Context) {
		// 检查总开关
		if !appService.GetBool(consts.ConfigRateLimitEnabled) {
			c.Next()
			return
		}

		// 获取当前配置
		currentRPS := appService.GetFloat64(rpsKey)
		currentBurst := appService.GetInt(burstKey)

		// 初始化 Limiter
		once.Do(func() {
			limiter = NewIPRateLimiter(rate.Limit(currentRPS), currentBurst)
		})

		// 获取 IP 对应的 limiter
		ip := c.ClientIP()

		if redisClient := service.GetRedisClient(); redisClient != nil {
			allowed, err := allowByRedisRateLimit(redisClient, "rate", rpsKey, burstKey, ip, currentRPS, currentBurst)
			if err == nil {
				logRedisFallbackRecovered("令牌桶限流")
				if !allowed {
					c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
					c.Abort()
					return
				}
				c.Next()
				return
			}
			logRedisFallbackDegraded("令牌桶限流", err)
		}

		l := limiter.getLimiter(ip)

		// 动态更新 limit 和 burst (如果配置发生变更)
		if l.Limit() != rate.Limit(currentRPS) {
			l.SetLimit(rate.Limit(currentRPS))
		}
		if l.Burst() != currentBurst {
			l.SetBurst(currentBurst)
		}

		if !l.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IntervalRateMiddleware 按数据库配置的最小调用间隔进行限流。
// intervalKey 对应设置项，值为秒数（int），例如 120 表示 2 分钟。
func IntervalRateMiddleware(appService *service.Service, intervalKey string) gin.HandlerFunc {
	// 每个中间件实例维护自己的访问时间表，并通过 sync.Once 确保清理协程只启动一次。
	var requestTimes sync.Map
	var cleanupOnce sync.Once

	startCleanupLoop := func() {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				now := time.Now()
				interval := getIntervalBySettingKey(appService, intervalKey)
				requestTimes.Range(func(key, value interface{}) bool {
					t, ok := value.(time.Time)
					if !ok {
						requestTimes.Delete(key)
						return true
					}
					// 清理较久未访问的记录（至少超过 2*interval 且超过 5 分钟）。
					if now.Sub(t) > interval*2 && now.Sub(t) > 5*time.Minute {
						requestTimes.Delete(key)
					}
					return true
				})
			}
		}()
	}

	return func(c *gin.Context) {
		// 检查是否开启敏感操作限流
		if !appService.GetBool(consts.ConfigEnableSensitiveRateLimit) {
			c.Next()
			return
		}

		cleanupOnce.Do(startCleanupLoop)

		interval := getIntervalBySettingKey(appService, intervalKey)

		ip := c.ClientIP()

		if redisClient := service.GetRedisClient(); redisClient != nil {
			ok, err := allowByRedisInterval(redisClient, intervalKey, ip, interval)
			if err == nil {
				logRedisFallbackRecovered("间隔限流")
				if !ok {
					c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("操作过于频繁，请等待 %v 后再试", interval)})
					c.Abort()
					return
				}
				c.Next()
				return
			}
			logRedisFallbackDegraded("间隔限流", err)
		}

		val, ok := requestTimes.Load(ip)
		if ok {
			if t, ok := val.(time.Time); ok {
				if time.Since(t) < interval {
					c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("操作过于频繁，请等待 %v 后再试", interval)})
					c.Abort()
					return
				}
			}
		}

		requestTimes.Store(ip, time.Now())
		c.Next()
	}
}

func getIntervalBySettingKey(appService *service.Service, intervalKey string) time.Duration {
	seconds := appService.GetInt(intervalKey)
	if seconds <= 0 {
		return defaultSensitiveOperationInterval
	}
	return time.Duration(seconds) * time.Second
}

func allowByRedisInterval(client *redis.Client, namespace, ip string, interval time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := service.RedisKey("middleware", namespace, ip)
	result, err := client.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  interval,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return result == "OK", nil
}

func allowByRedisRateLimit(client *redis.Client, namespace, rpsKey, burstKey, ip string, rps float64, burst int) (bool, error) {
	if rps <= 0 || burst <= 0 {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().Unix()
	window := int64(1)
	if rps < 1 {
		window = int64(1 / rps)
		if window < 1 {
			window = 1
		}
	}
	bucket := now / window
	key := service.RedisKey("middleware", namespace, rpsKey, burstKey, ip, strconv.FormatInt(bucket, 10))

	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		expire := time.Duration(window)*time.Second + 2*time.Second
		if expireErr := client.Expire(ctx, key, expire).Err(); expireErr != nil {
			return false, expireErr
		}
	}

	if count > int64(burst) {
		return false, nil
	}

	return true, nil
}
