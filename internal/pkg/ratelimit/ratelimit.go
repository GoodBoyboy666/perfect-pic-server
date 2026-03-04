package ratelimit

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

const (
	defaultSensitiveOperationInterval = 2 * time.Minute
	redisFallbackLogInterval          = 1 * time.Minute
	redisOpTimeout                    = time.Second
	defaultRedisKeyPrefix             = "perfect_pic"
)

type ipRateLimiter struct {
	ips sync.Map
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

type clientEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type redisFallbackLogState struct {
	mu         sync.Mutex
	degraded   bool
	lastWarnAt time.Time
}

type TokenBucketLimiter struct {
	local       *ipRateLimiter
	once        sync.Once
	redisClient *redis.Client
}

type IntervalLimiter struct {
	requestTimes sync.Map
	cleanupOnce  sync.Once

	redisClient *redis.Client
}

var redisFallbackLogStates sync.Map

func NewTokenBucketLimiter(redisClient *redis.Client) *TokenBucketLimiter {
	return &TokenBucketLimiter{redisClient: redisClient}
}

func NewIntervalLimiter(redisClient *redis.Client) *IntervalLimiter {
	return &IntervalLimiter{
		redisClient: redisClient,
	}
}

func (l *TokenBucketLimiter) Allow(
	ip, namespace, rpsKey, burstKey string,
	rps float64,
	burst int,
) bool {
	if namespace == "" {
		namespace = "rate"
	}

	l.once.Do(func() {
		l.local = newIPRateLimiter(rate.Limit(rps), burst)
	})

	if l.redisClient != nil {
		allowed, err := AllowByRedisRateLimit(l.redisClient, namespace, rpsKey, burstKey, ip, rps, burst)
		if err == nil {
			logRedisFallbackRecovered("令牌桶限流")
			return allowed
		}
		logRedisFallbackDegraded("令牌桶限流", err)
	}

	scopeKey := namespace + ":" + rpsKey + ":" + burstKey + ":" + ip
	localLimiter := l.local.getLimiter(scopeKey)
	if localLimiter.Limit() != rate.Limit(rps) {
		localLimiter.SetLimit(rate.Limit(rps))
	}
	if localLimiter.Burst() != burst {
		localLimiter.SetBurst(burst)
	}

	return localLimiter.Allow()
}

func (l *IntervalLimiter) Allow(ip, namespace string, interval time.Duration) bool {
	if namespace == "" {
		namespace = "interval"
	}
	if interval <= 0 {
		interval = defaultSensitiveOperationInterval
	}

	l.cleanupOnce.Do(l.startCleanupLoop)

	if l.redisClient != nil {
		ok, err := AllowByRedisInterval(l.redisClient, namespace, ip, interval)
		if err == nil {
			logRedisFallbackRecovered("间隔限流")
			return ok
		}
		logRedisFallbackDegraded("间隔限流", err)
	}

	localKey := namespace + ":" + ip
	if value, ok := l.requestTimes.Load(localKey); ok {
		lastTime, castOK := value.(time.Time)
		if !castOK {
			l.requestTimes.Delete(localKey)
		} else if time.Since(lastTime) < interval {
			return false
		}
	}

	l.requestTimes.Store(localKey, time.Now())
	return true
}

func (l *IntervalLimiter) startCleanupLoop() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			l.requestTimes.Range(func(key, value interface{}) bool {
				lastTime, ok := value.(time.Time)
				if !ok {
					l.requestTimes.Delete(key)
					return true
				}
				if now.Sub(lastTime) > 10*time.Minute {
					l.requestTimes.Delete(key)
				}
				return true
			})
		}
	}()
}

func AllowByRedisInterval(client *redis.Client, namespace, ip string, interval time.Duration) (bool, error) {
	if client == nil {
		return false, errors.New("redis client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()

	key := buildRedisKey("middleware", namespace, ip)
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

func AllowByRedisRateLimit(
	client *redis.Client,
	namespace, rpsKey, burstKey, ip string,
	rps float64,
	burst int,
) (bool, error) {
	if rps <= 0 || burst <= 0 {
		return true, nil
	}
	if client == nil {
		return false, errors.New("redis client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
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
	key := buildRedisKey("middleware", namespace, rpsKey, burstKey, ip, strconv.FormatInt(bucket, 10))
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

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	limiter := &ipRateLimiter{r: r, b: b}
	go limiter.cleanupLoop()
	return limiter
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	if value, ok := l.ips.Load(ip); ok {
		if entry, castOK := value.(*clientEntry); castOK {
			entry.lastSeen = time.Now()
			return entry.limiter
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if value, ok := l.ips.Load(ip); ok {
		if entry, castOK := value.(*clientEntry); castOK {
			entry.lastSeen = time.Now()
			return entry.limiter
		}
	}

	newLimiter := rate.NewLimiter(l.r, l.b)
	l.ips.Store(ip, &clientEntry{limiter: newLimiter, lastSeen: time.Now()})
	return newLimiter
}

func (l *ipRateLimiter) cleanupLoop() {
	for {
		time.Sleep(time.Minute)
		l.ips.Range(func(key, value interface{}) bool {
			entry, ok := value.(*clientEntry)
			if !ok {
				l.ips.Delete(key)
				return true
			}
			if time.Since(entry.lastSeen) > 3*time.Minute {
				l.ips.Delete(key)
			}
			return true
		})
	}
}

func getRedisFallbackLogState(scope string) *redisFallbackLogState {
	state, _ := redisFallbackLogStates.LoadOrStore(scope, &redisFallbackLogState{})
	if typedState, ok := state.(*redisFallbackLogState); ok {
		return typedState
	}

	fallbackState := &redisFallbackLogState{}
	redisFallbackLogStates.Store(scope, fallbackState)
	return fallbackState
}

func logRedisFallbackDegraded(scope string, err error) {
	state := getRedisFallbackLogState(scope)
	now := time.Now()

	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.degraded {
		state.degraded = true
		state.lastWarnAt = now
		log.Printf("⚠️ Redis %s 检查失败，已降级到内存限流（后续每 %s 最多记录一次）: %v", scope, redisFallbackLogInterval, err)
		return
	}

	if now.Sub(state.lastWarnAt) >= redisFallbackLogInterval {
		state.lastWarnAt = now
		log.Printf("⚠️ Redis %s 仍不可用，继续使用内存限流: %v", scope, err)
	}
}

func logRedisFallbackRecovered(scope string) {
	state := getRedisFallbackLogState(scope)

	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.degraded {
		return
	}

	state.degraded = false
	state.lastWarnAt = time.Time{}
	log.Printf("✅ Redis %s 已恢复，切回Redis 限流", scope)
}

func buildRedisKey(parts ...string) string {
	if len(parts) == 0 {
		return defaultRedisKeyPrefix
	}

	key := defaultRedisKeyPrefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}
