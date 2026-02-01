# Monitoring, Logging ve Observability

## 📊 Genel Bakış

Bu proje, **production-ready** bir uygulama olarak kapsamlı **monitoring**, **logging** ve **observability** altyapısına sahiptir. Sistem sağlığını izlemek, performans sorunlarını tespit etmek ve hata ayıklamak için modern araçlar kullanılır.

## 🎯 Observability Pillars

### 1. Metrics (Metrikler)
- **Prometheus** ile sistem metrikleri toplama
- **Grafana** ile görselleştirme (opsiyonel)
- HTTP request metrikleri, cache hit/miss oranları, database query süreleri

### 2. Logs (Loglar)
- **Structured logging** with **Zap**
- JSON formatında loglar
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID tracking ile distributed tracing

### 3. Traces (İzleme)
- Request ID ile end-to-end request tracking
- Middleware chain'de request flow takibi

## 📈 Prometheus Metrics

### Metrics Endpoint

```
GET /metrics
```

**Örnek Response**:
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/search",status="200"} 1523

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/v1/search",le="0.005"} 1234
http_request_duration_seconds_bucket{method="GET",path="/api/v1/search",le="0.01"} 1456
http_request_duration_seconds_bucket{method="GET",path="/api/v1/search",le="0.025"} 1500
```

### Toplanan Metrikler

#### HTTP Metrikleri

**Dosya**: `internal/infrastructure/metrics/prometheus.go`

```go
// Request sayısı (method, path, status code ile)
http_requests_total{method="GET", path="/api/v1/search", status="200"}

// Request süresi (histogram)
http_request_duration_seconds{method="GET", path="/api/v1/search"}

// Response boyutu
http_response_size_bytes{method="GET", path="/api/v1/search"}

// Aktif request sayısı
http_requests_in_flight{method="GET", path="/api/v1/search"}
```

#### Cache Metrikleri

```go
// Cache hit sayısı
cache_hits_total{cache_type="redis"}

// Cache miss sayısı
cache_misses_total{cache_type="redis"}

// Cache hit oranı (calculated metric)
cache_hit_ratio = cache_hits_total / (cache_hits_total + cache_misses_total)
```

#### Database Metrikleri

```go
// Database query süresi
db_query_duration_seconds{operation="search"}

// Database connection pool
db_connections_active
db_connections_idle
db_connections_max
```

#### Rate Limiter Metrikleri

```go
// Rate limit aşımları
rate_limit_exceeded_total{endpoint="/api/v1/search"}

// Rate limit requests
rate_limit_requests_total{endpoint="/api/v1/search"}
```

### Metrics Middleware

**Dosya**: `internal/transport/middleware/metrics.go`

```go
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Request sayacını artır
        metrics.HTTPRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Inc()
        
        // In-flight request sayacını artır
        metrics.HTTPRequestsInFlight.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Inc()
        defer metrics.HTTPRequestsInFlight.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Dec()
        
        // Response writer wrapper
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Next handler'ı çağır
        next.ServeHTTP(rw, r)
        
        // Duration'ı kaydet
        duration := time.Since(start).Seconds()
        metrics.HTTPRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Observe(duration)
        
        // Status code ile request sayacını güncelle
        metrics.HTTPRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(rw.statusCode),
        ).Inc()
    })
}
```

## 📝 Structured Logging

### Zap Logger

**Dosya**: `internal/infrastructure/logger/zap.go`

**Özellikler**:
- JSON formatında structured logging
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Automatic field extraction (timestamp, caller, stack trace)
- High performance (zero-allocation)

**Initialization**:
```go
func NewLogger(env string) (*zap.Logger, error) {
    var config zap.Config
    
    if env == "production" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
    }
    
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    return config.Build()
}
```

**Usage**:
```go
logger.Info("Search request received",
    zap.String("request_id", requestID),
    zap.String("query", query),
    zap.Int("page", page),
    zap.Duration("duration", duration),
)

logger.Error("Database error",
    zap.String("request_id", requestID),
    zap.Error(err),
    zap.String("operation", "search"),
)
```

### Log Levels

#### DEBUG
Geliştirme sırasında detaylı bilgi için kullanılır.
```go
logger.Debug("Cache lookup",
    zap.String("key", cacheKey),
    zap.Bool("found", found),
)
```

#### INFO
Normal operasyonel bilgiler.
```go
logger.Info("Provider sync completed",
    zap.String("provider", providerName),
    zap.Int("contents", contentCount),
    zap.Duration("duration", duration),
)
```

#### WARN
Potansiyel sorunlar ama uygulama çalışmaya devam ediyor.
```go
logger.Warn("Cache miss",
    zap.String("key", cacheKey),
    zap.String("reason", "expired"),
)
```

#### ERROR
Hata durumları ama uygulama recover edebiliyor.
```go
logger.Error("Failed to fetch provider contents",
    zap.String("provider", providerName),
    zap.Error(err),
)
```

#### FATAL
Kritik hatalar, uygulama kapanıyor.
```go
logger.Fatal("Database connection failed",
    zap.Error(err),
)
```

### Log Format

**Development**:
```
2026-01-31T19:00:00.000Z    INFO    Search request received
    request_id: abc-123
    query: golang
    page: 1
    duration: 45ms
```

**Production (JSON)**:
```json
{
  "level": "info",
  "timestamp": "2026-01-31T19:00:00.000Z",
  "caller": "http/handlers.go:45",
  "msg": "Search request received",
  "request_id": "abc-123",
  "query": "golang",
  "page": 1,
  "duration": 0.045
}
```

## 🔍 Request Tracking

### Request ID Middleware

**Dosya**: `internal/transport/middleware/request_id.go`

Her request'e unique bir ID atanır ve tüm log'larda kullanılır.

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Request ID'yi header'dan al veya oluştur
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Response header'a ekle
        w.Header().Set("X-Request-ID", requestID)
        
        // Context'e ekle
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        
        // Next handler'ı çağır
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Usage in Logs**:
```go
requestID := r.Context().Value("request_id").(string)

logger.Info("Processing request",
    zap.String("request_id", requestID),
    zap.String("path", r.URL.Path),
)
```

### Logging Middleware

**Dosya**: `internal/transport/middleware/logging.go`

Her HTTP request'i loglar.

```go
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Request ID'yi al
            requestID := r.Context().Value("request_id").(string)
            
            // Request'i logla
            logger.Info("Request started",
                zap.String("request_id", requestID),
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.String("remote_addr", r.RemoteAddr),
                zap.String("user_agent", r.UserAgent()),
            )
            
            // Response writer wrapper
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            
            // Next handler'ı çağır
            next.ServeHTTP(rw, r)
            
            // Response'u logla
            duration := time.Since(start)
            logger.Info("Request completed",
                zap.String("request_id", requestID),
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.Int("status", rw.statusCode),
                zap.Duration("duration", duration),
                zap.Int("response_size", rw.bytesWritten),
            )
        })
    }
}
```

## 🏥 Health Checks

### Health Endpoint

```
GET /api/v1/health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-31T19:00:00Z",
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 2.5
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 0.8
    }
  },
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

**Implementation**:
```go
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    response := HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   "1.0.0",
        Checks:    make(map[string]HealthCheck),
    }
    
    // Database health check
    dbStart := time.Now()
    if err := h.db.PingContext(ctx); err != nil {
        response.Status = "unhealthy"
        response.Checks["database"] = HealthCheck{
            Status: "unhealthy",
            Error:  err.Error(),
        }
    } else {
        response.Checks["database"] = HealthCheck{
            Status:     "healthy",
            LatencyMs:  time.Since(dbStart).Milliseconds(),
        }
    }
    
    // Redis health check
    redisStart := time.Now()
    if err := h.redis.Ping(ctx).Err(); err != nil {
        response.Status = "unhealthy"
        response.Checks["redis"] = HealthCheck{
            Status: "unhealthy",
            Error:  err.Error(),
        }
    } else {
        response.Checks["redis"] = HealthCheck{
            Status:     "healthy",
            LatencyMs:  time.Since(redisStart).Milliseconds(),
        }
    }
    
    // HTTP status code
    statusCode := http.StatusOK
    if response.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

## 🔧 Profiling

### pprof Integration

**Dosya**: `cmd/server/main.go`

```go
import _ "net/http/pprof"

func main() {
    // pprof otomatik olarak /debug/pprof/* endpoint'lerini ekler
    
    // Server'ı başlat
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

### Available Endpoints

```
GET /debug/pprof/          - Index page
GET /debug/pprof/heap      - Heap profiling
GET /debug/pprof/goroutine - Goroutine profiling
GET /debug/pprof/profile   - CPU profiling (30s)
GET /debug/pprof/trace     - Execution trace
```

### Usage

```bash
# Heap profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# CPU profiling (30 seconds)
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Goroutine profiling
go tool pprof http://localhost:8080/debug/pprof/goroutine

# Interactive mode
(pprof) top10
(pprof) list functionName
(pprof) web
```

## 📊 Grafana Dashboard (Opsiyonel)

### Prometheus + Grafana Setup

**docker-compose.yml**:
```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
  
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
```

**prometheus.yml**:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'search-engine'
    static_configs:
      - targets: ['backend:8080']
```

### Grafana Panels

#### 1. Request Rate
```promql
rate(http_requests_total[5m])
```

#### 2. Error Rate
```promql
rate(http_requests_total{status=~"5.."}[5m])
```

#### 3. Request Duration (p95)
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

#### 4. Cache Hit Ratio
```promql
cache_hits_total / (cache_hits_total + cache_misses_total)
```

## 🚨 Alerting (Opsiyonel)

### Prometheus Alerts

**alerts.yml**:
```yaml
groups:
  - name: search_engine_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} requests/sec"
      
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "P95 latency is {{ $value }} seconds"
      
      - alert: LowCacheHitRatio
        expr: cache_hits_total / (cache_hits_total + cache_misses_total) < 0.7
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit ratio"
          description: "Cache hit ratio is {{ $value }}"
```

## 📚 Best Practices

### 1. Structured Logging

```go
// ❌ Kötü
log.Printf("User %s searched for %s", userID, query)

// ✅ İyi
logger.Info("Search performed",
    zap.String("user_id", userID),
    zap.String("query", query),
    zap.Int("results", len(results)),
)
```

### 2. Log Levels

```go
// DEBUG: Geliştirme detayları
logger.Debug("Cache key generated", zap.String("key", key))

// INFO: Normal operasyonlar
logger.Info("Request processed", zap.Duration("duration", duration))

// WARN: Potansiyel sorunlar
logger.Warn("Slow query detected", zap.Duration("duration", duration))

// ERROR: Hatalar
logger.Error("Database error", zap.Error(err))

// FATAL: Kritik hatalar
logger.Fatal("Failed to start server", zap.Error(err))
```

### 3. Metrics Naming

```
# Pattern: <namespace>_<subsystem>_<name>_<unit>
http_requests_total
http_request_duration_seconds
cache_hits_total
db_query_duration_seconds
```

### 4. Request Context

```go
// Request ID'yi context'e ekle
ctx := context.WithValue(r.Context(), "request_id", requestID)

// Tüm fonksiyonlarda kullan
func ProcessRequest(ctx context.Context) {
    requestID := ctx.Value("request_id").(string)
    logger.Info("Processing", zap.String("request_id", requestID))
}
```

## 🔍 Troubleshooting

### High Memory Usage

```bash
# Heap profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Top memory consumers
(pprof) top10

# Detailed analysis
(pprof) list functionName
```

### High CPU Usage

```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Top CPU consumers
(pprof) top10
```

### Goroutine Leaks

```bash
# Goroutine profiling
go tool pprof http://localhost:8080/debug/pprof/goroutine

# Check goroutine count
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

## 📈 Performance Monitoring

### Key Metrics to Monitor

1. **Request Rate**: Requests per second
2. **Error Rate**: Errors per second
3. **Latency**: P50, P95, P99 response times
4. **Cache Hit Ratio**: Cache effectiveness
5. **Database Connection Pool**: Active/idle connections
6. **Goroutine Count**: Potential leaks
7. **Memory Usage**: Heap allocation

### SLO/SLA Targets

- **Availability**: 99.9% uptime
- **Latency**: P95 < 100ms, P99 < 500ms
- **Error Rate**: < 0.1%
- **Cache Hit Ratio**: > 80%
