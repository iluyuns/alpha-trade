package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存实现（实现 port.Cache 接口）
type RedisCache struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisCache 创建 Redis 缓存
func NewRedisCache(client *redis.Client, keyPrefix string) *RedisCache {
	return &RedisCache{
		client:    client,
		keyPrefix: keyPrefix,
	}
}

func (c *RedisCache) prefixKey(key string) string {
	if c.keyPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.keyPrefix, key)
}

// Get 读取缓存
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, c.prefixKey(key)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache miss: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("cache get failed: %w", err)
	}
	return val, nil
}

// Set 写入缓存
func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, c.prefixKey(key), value, ttl).Err()
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefixKey(key)).Err()
}

// Exists 检查缓存是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	return n > 0, err
}
