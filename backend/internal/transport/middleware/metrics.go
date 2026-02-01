package middleware

import (
	"net/http"
	"time"

	"github.com/onurerdog4n/search-engine/internal/infrastructure/metrics"
)

// Metrics middleware collects Prometheus metrics
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := newResponseWriter(w)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		metrics.RecordHTTPRequest(
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}
