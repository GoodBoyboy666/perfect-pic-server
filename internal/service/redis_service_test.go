package service

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisKey_DefaultPrefix(t *testing.T) {
	// TestMain initializes config with default prefix when not set.
	got := RedisKey("a", "b")
	if got != "perfect_pic:a:b" {
		t.Fatalf("unexpected key: %q", got)
	}
}

func TestRedisKey_NoParts(t *testing.T) {
	if got := RedisKey(); got != "perfect_pic" {
		t.Fatalf("unexpected key: %q", got)
	}
}

func TestGetRedisClient_DisabledReturnsNil(t *testing.T) {
	// TestMain sets redis disabled; GetRedisClient should stay nil.
	if c := GetRedisClient(); c != nil {
		t.Fatalf("expected nil redis client when disabled")
	}
}

func TestCloseRedisClient(t *testing.T) {
	redisClient = nil
	if err := CloseRedisClient(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	redisClient = redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	if err := CloseRedisClient(); err != nil {
		t.Fatalf("CloseRedisClient: %v", err)
	}
	redisClient = nil
}
