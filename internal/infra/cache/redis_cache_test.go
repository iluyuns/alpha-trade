package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   14,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	_ = client.FlushDB(ctx)
	return client
}

func TestRedisCache_SetGet(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	if err := cache.Set(ctx, "k1", "v1", 10*time.Second); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "v1" {
		t.Errorf("got %s, want v1", val)
	}
}

func TestRedisCache_Miss(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected cache miss error")
	}
}

func TestRedisCache_Delete(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	_ = cache.Set(ctx, "del-key", "val", 10*time.Second)
	_ = cache.Delete(ctx, "del-key")

	exists, _ := cache.Exists(ctx, "del-key")
	if exists {
		t.Error("key should be deleted")
	}
}

func TestRedisCache_TTL(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	_ = cache.Set(ctx, "ttl-key", "val", 1*time.Second)
	time.Sleep(2 * time.Second)

	_, err := cache.Get(ctx, "ttl-key")
	if err == nil {
		t.Error("expected cache miss after TTL expiry")
	}
}
