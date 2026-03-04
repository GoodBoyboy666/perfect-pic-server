package redis

import (
	"perfect-pic-server/internal/pkg/cache"
	"testing"
)

// 测试内容：验证 Redis key 使用默认前缀拼接。
func TestRedisKey_DefaultPrefix(t *testing.T) {
	store := cache.NewStore(nil, &cache.Config{Prefix: ""})
	got := store.RedisKey("a", "b")
	if got != "perfect_pic:a:b" {
		t.Fatalf("非预期 key: %q", got)
	}
}

// 测试内容：验证无参数时返回仅包含前缀的 key。
func TestRedisKey_NoParts(t *testing.T) {
	store := cache.NewStore(nil, &cache.Config{Prefix: ""})
	if got := store.RedisKey(); got != "perfect_pic" {
		t.Fatalf("非预期 key: %q", got)
	}
}

// 测试内容：验证 Redis 不可达时初始化客户端返回 nil。
func TestNewRedisClient_UnavailableReturnsNil(t *testing.T) {
	if c := NewRedisClient(&Config{Addr: "127.0.0.1:1", DB: 0}); c != nil {
		t.Fatalf("期望为 nil redis client when unavailable")
	}
}
