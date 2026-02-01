package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/onurerdog4n/search-engine/internal/infrastructure/metrics"
)

// RateLimiter rate limiting middleware'i
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter yeni bir rate limiter oluşturur
// requestsPerMinute: dakikada izin verilen istek sayısı
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	r := rate.Limit(float64(requestsPerMinute) / 60.0) // Saniye başına rate
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    requestsPerMinute,
	}
}

// getRealIP gets the real IP address from request
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

// Middleware rate limiting middleware'ini döndürür
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get real IP address
		ip := getRealIP(r)

		// Limiter'ı al veya oluştur
		limiter := rl.getLimiter(ip)

		// Rate limit kontrolü
		if !limiter.Allow() {
			// Record metrics
			metrics.RecordRateLimitExceeded(r.URL.Path)

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", string(rune(rl.burst)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit aşıldı. Lütfen 60 saniye sonra tekrar deneyin."}`))
			return
		}

		// Add rate limit headers
		w.Header().Set("X-RateLimit-Limit", string(rune(rl.burst)))
		// Note: Remaining count would require tracking, simplified here

		// İsteği işle
		next.ServeHTTP(w, r)
	})
}

// getLimiter IP için limiter döndürür veya oluşturur
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

// CleanupOldLimiters eski limiter'ları temizler (opsiyonel, memory leak önlemek için)
func (rl *RateLimiter) CleanupOldLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			// Basit cleanup: tüm limiter'ları sıfırla
			// Production'da daha sofistike bir cleanup yapılabilir
			rl.limiters = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
		}
	}()
}
