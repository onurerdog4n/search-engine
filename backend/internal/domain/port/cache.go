package port

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrCacheMiss cache'de veri bulunamadığında döner
	ErrCacheMiss = errors.New("cache miss")
)

// CacheRepository cache veri erişim katmanı interface'i
type CacheRepository interface {
	// Get cache'den veri okur
	// Key bulunamazsa ErrCacheMiss döner
	Get(ctx context.Context, key string) ([]byte, error)

	// Set cache'e veri yazar
	// TTL süresi sonunda otomatik olarak silinir
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete cache'den veri siler
	Delete(ctx context.Context, key string) error

	// Clear tüm cache'i temizler (opsiyonel, dikkatli kullanılmalı)
	Clear(ctx context.Context) error
}
