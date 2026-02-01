package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// redisCache Redis ile CacheRepository implementasyonu
type redisCache struct {
	client *redis.Client
}

// NewRedisCache yeni bir Redis cache repository oluşturur
func NewRedisCache(client *redis.Client) port.CacheRepository {
	return &redisCache{client: client}
}

// Get cache'den veri okur
func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, port.ErrCacheMiss
	}
	return val, err
}

// Set cache'e veri yazar
func (c *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete cache'den veri siler
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Clear tüm cache'i temizler
func (c *redisCache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}
