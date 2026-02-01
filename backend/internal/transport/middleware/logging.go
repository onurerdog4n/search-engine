package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/onurerdog4n/search-engine/internal/infrastructure/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// Logging middleware logs HTTP requests with structured logging
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get request ID from context
		requestID := GetRequestID(r.Context())

		// Wrap response writer
		wrapped := newResponseWriter(w)

		// Log request
		log := logger.GetLogger().WithRequestID(requestID)
		log.Info("incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log response
		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.Int("bytes", wrapped.written),
		}

		if wrapped.statusCode >= 500 {
			log.Error("request completed with server error", fields...)
		} else if wrapped.statusCode >= 400 {
			log.Warn("request completed with client error", fields...)
		} else {
			log.Info("request completed", fields...)
		}
	})
}
