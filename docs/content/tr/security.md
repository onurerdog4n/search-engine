# Security & Best Practices

## 🔒 Genel Bakış

Bu proje, **production-ready** bir uygulama olarak kapsamlı güvenlik önlemleri içerir. OWASP Top 10 ve modern güvenlik best practice'lerine uygun olarak geliştirilmiştir.

## 🛡️ Güvenlik Katmanları

### 1. Input Validation

**Dosya**: `internal/infrastructure/validation/validator.go`

Tüm kullanıcı girdileri validate edilir.

```go
type SearchParamsValidator struct {
    Query       string `validate:"max=200"`
    ContentType string `validate:"omitempty,oneof=video article"`
    SortBy      string `validate:"omitempty,oneof=relevance popularity date"`
    Page        int    `validate:"min=1,max=1000"`
    PageSize    int    `validate:"min=1,max=100"`
}

func ValidateSearchParams(params port.SearchParams) error {
    validator := SearchParamsValidator{
        Query:       params.Query,
        ContentType: string(params.ContentType),
        SortBy:      params.SortBy,
        Page:        params.Page,
        PageSize:    params.PageSize,
    }
    
    validate := validator.New()
    return validate.Struct(validator)
}
```

**Validation Rules**:
- ✅ Query max 200 karakter
- ✅ ContentType sadece "video" veya "article"
- ✅ SortBy sadece "relevance", "popularity" veya "date"
- ✅ Page minimum 1, maximum 1000
- ✅ PageSize minimum 1, maximum 100

### 2. SQL Injection Prevention

**Parameterized Queries** kullanılır, string concatenation yapılmaz.

```go
// ❌ ASLA BÖYLE YAPMA (SQL Injection riski!)
query := fmt.Sprintf("SELECT * FROM contents WHERE title = '%s'", userInput)

// ✅ DOĞRU (Parameterized query)
query := "SELECT * FROM contents WHERE title = $1"
db.QueryContext(ctx, query, userInput)
```

**Full-Text Search Sanitization**:
```go
func sanitizeQuery(query string) string {
    // FTS özel karakterlerini escape et
    replacer := strings.NewReplacer(
        "&", "",
        "|", "",
        "!", "",
        "(", "",
        ")", "",
        ":", "",
        "*", "",
    )
    return replacer.Replace(query)
}
```

### 3. Rate Limiting

**Dosya**: `internal/transport/middleware/rate_limiter.go`

DDoS ve brute-force saldırılarına karşı koruma.

```go
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     int // requests per minute
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // IP adresini al (X-Forwarded-For header'ı kontrol et)
        ip := getClientIP(r)
        
        // Rate limiter'ı al veya oluştur
        limiter := rl.getLimiter(ip)
        
        // Rate limit kontrolü
        if !limiter.Allow() {
            // Metrics
            metrics.RateLimitExceeded.WithLabelValues(r.URL.Path).Inc()
            
            // Response headers
            w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.rate))
            w.Header().Set("X-RateLimit-Remaining", "0")
            w.Header().Set("Retry-After", "60")
            
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func getClientIP(r *http.Request) string {
    // X-Forwarded-For header'ını kontrol et (proxy/load balancer)
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }
    
    // X-Real-IP header'ını kontrol et
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }
    
    // RemoteAddr kullan
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
```

**Rate Limit Headers**:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1643673600
Retry-After: 60
```

### 4. CORS Configuration

**Dosya**: `internal/transport/middleware/cors.go`

Cross-Origin Resource Sharing güvenli şekilde yapılandırılmıştır.

```go
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allowed origins (production'da whitelist kullan)
        allowedOrigins := []string{
            "http://localhost:3000",
            "https://yourdomain.com",
        }
        
        origin := r.Header.Get("Origin")
        if isAllowedOrigin(origin, allowedOrigins) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }
        
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Max-Age", "3600")
        
        // Preflight request
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 5. Security Headers

**Dosya**: `internal/transport/middleware/security.go`

OWASP önerilen security header'ları eklenir.

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // XSS Protection
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        // HTTPS enforcement (production)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // Content Security Policy
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        
        // Referrer Policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Permissions Policy
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        next.ServeHTTP(w, r)
    })
}
```

### 6. Dependency Security

**GitHub Actions Workflow**: `.github/workflows/security.yml`

Otomatik dependency scanning ve vulnerability detection.

```yaml
name: Security Scan

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * 0'  # Her Pazar gece

jobs:
  dependency-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt json -out gosec-report.json ./...'
      
      - name: Run Nancy (Dependency Scanner)
        run: |
          go install github.com/sonatype-nexus-community/nancy@latest
          go list -json -m all | nancy sleuth
      
      - name: Run Trivy (Container Scanner)
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
```

**Audit Script**: `scripts/audit-deps.sh`

```bash
#!/bin/bash

echo "🔍 Checking for known vulnerabilities..."

# Go vulnerability check
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Dependency audit
echo ""
echo "📦 Auditing dependencies..."
go list -json -m all | nancy sleuth

# Outdated dependencies
echo ""
echo "📅 Checking for outdated dependencies..."
go list -u -m all

echo ""
echo "✅ Security audit complete!"
```

### 7. Environment Variables

**Dosya**: `internal/infrastructure/config/config.go`

Hassas bilgiler environment variable'larda saklanır.

```go
type Config struct {
    DatabaseURL         string `validate:"required,url"`
    RedisURL            string `validate:"required"`
    Port                string `validate:"required,number"`
    SyncIntervalSeconds int    `validate:"min=60,max=86400"`
    RateLimitPerMinute  int    `validate:"min=1,max=10000"`
    CacheTTLSeconds     int    `validate:"min=1,max=3600"`
    Environment         string `validate:"oneof=development staging production"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{
        DatabaseURL:         os.Getenv("DATABASE_URL"),
        RedisURL:            os.Getenv("REDIS_URL"),
        Port:                getEnv("PORT", "8080"),
        SyncIntervalSeconds: getEnvAsInt("SYNC_INTERVAL", 3600),
        RateLimitPerMinute:  getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
        CacheTTLSeconds:     getEnvAsInt("CACHE_TTL_SECONDS", 60),
        Environment:         getEnv("ENVIRONMENT", "development"),
    }
    
    // Validate config
    validate := validator.New()
    if err := validate.Struct(cfg); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return cfg, nil
}
```

**❌ ASLA YAPMA**:
```go
// Hardcoded credentials
db, _ := sql.Open("postgres", "postgres://user:password@localhost/db")

// Hardcoded API keys
apiKey := "sk-1234567890abcdef"
```

**✅ DOĞRU**:
```go
// Environment variables
dbURL := os.Getenv("DATABASE_URL")
db, _ := sql.Open("postgres", dbURL)

apiKey := os.Getenv("API_KEY")
```

### 8. Error Handling

**Dosya**: `internal/domain/port/errors.go`

Hata mesajlarında hassas bilgi sızdırılmaz.

```go
// ❌ Kötü (Internal details exposed)
return fmt.Errorf("database error: %v", err)

// ✅ İyi (Generic error message)
logger.Error("Database error", zap.Error(err))
return errors.New("internal server error")
```

**Custom Error Types**:
```go
var (
    ErrContentNotFound     = errors.New("content not found")
    ErrInvalidInput        = errors.New("invalid input")
    ErrRateLimitExceeded   = errors.New("rate limit exceeded")
    ErrInternalServerError = errors.New("internal server error")
)

func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
    // ...
    
    if err != nil {
        // Log internal error
        logger.Error("Search failed", zap.Error(err))
        
        // Return generic error to user
        http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
        return
    }
}
```

### 9. Database Security

**Connection Pooling**:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(5 * time.Minute)
```

**Prepared Statements**:
```go
// Prepared statement (SQL injection safe)
stmt, err := db.Prepare("SELECT * FROM contents WHERE id = $1")
defer stmt.Close()

row := stmt.QueryRow(id)
```

**Least Privilege Principle**:
```sql
-- Database user sadece gerekli izinlere sahip
GRANT SELECT, INSERT, UPDATE ON contents TO search_engine_user;
GRANT SELECT, INSERT, UPDATE ON content_stats TO search_engine_user;

-- Admin işlemleri için ayrı user
GRANT ALL PRIVILEGES ON ALL TABLES TO search_engine_admin;
```

### 10. Redis Security

**Password Protection**:
```bash
# redis.conf
requirepass your-strong-password-here
```

**Connection**:
```go
rdb := redis.NewClient(&redis.Options{
    Addr:     os.Getenv("REDIS_URL"),
    Password: os.Getenv("REDIS_PASSWORD"),
    DB:       0,
    
    // TLS (production)
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
})
```

## 🔐 Authentication & Authorization (Future)

### JWT Token (Opsiyonel)

```go
type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

func GenerateToken(userID, role string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            Issuer:    "search-engine",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, err
}
```

### API Key Authentication (Opsiyonel)

```go
func APIKeyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        
        if apiKey == "" {
            http.Error(w, "API key required", http.StatusUnauthorized)
            return
        }
        
        // Validate API key (database lookup)
        if !isValidAPIKey(apiKey) {
            http.Error(w, "Invalid API key", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## 🚨 Security Checklist

### Development

- [x] Input validation tüm endpoint'lerde
- [x] Parameterized queries (SQL injection prevention)
- [x] Rate limiting
- [x] CORS configuration
- [x] Security headers
- [x] Error handling (no sensitive data leakage)
- [x] Environment variables for secrets
- [x] Dependency scanning
- [x] Code security scanning (Gosec)

### Production

- [ ] HTTPS enforcement (TLS/SSL)
- [ ] Database encryption at rest
- [ ] Redis password protection
- [ ] Secrets management (HashiCorp Vault, AWS Secrets Manager)
- [ ] Network security (VPC, Security Groups)
- [ ] DDoS protection (Cloudflare, AWS Shield)
- [ ] Web Application Firewall (WAF)
- [ ] Regular security audits
- [ ] Incident response plan
- [ ] Backup and disaster recovery

## 📚 Security Tools

### Static Analysis

```bash
# Gosec - Go security checker
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# Go vet - Built-in static analysis
go vet ./...

# Staticcheck - Advanced static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

### Dependency Scanning

```bash
# Govulncheck - Official Go vulnerability scanner
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Nancy - Sonatype dependency scanner
go install github.com/sonatype-nexus-community/nancy@latest
go list -json -m all | nancy sleuth
```

### Container Scanning

```bash
# Trivy - Container vulnerability scanner
trivy image your-image:tag

# Grype - Container and filesystem scanner
grype dir:.
```

## 🔍 Security Monitoring

### Logging Security Events

```go
// Failed authentication
logger.Warn("Authentication failed",
    zap.String("ip", clientIP),
    zap.String("user_agent", userAgent),
)

// Rate limit exceeded
logger.Warn("Rate limit exceeded",
    zap.String("ip", clientIP),
    zap.String("endpoint", r.URL.Path),
)

// Suspicious activity
logger.Error("Suspicious activity detected",
    zap.String("ip", clientIP),
    zap.String("pattern", "SQL injection attempt"),
    zap.String("payload", sanitizedPayload),
)
```

### Metrics

```go
// Security metrics
security_events_total{type="rate_limit_exceeded"}
security_events_total{type="invalid_input"}
security_events_total{type="auth_failed"}
```

## 📖 OWASP Top 10 Coverage

1. **Broken Access Control** ✅
   - Rate limiting
   - Input validation

2. **Cryptographic Failures** ✅
   - HTTPS enforcement (production)
   - Secure password storage (future)

3. **Injection** ✅
   - Parameterized queries
   - Input sanitization

4. **Insecure Design** ✅
   - Clean Architecture
   - Security by design

5. **Security Misconfiguration** ✅
   - Security headers
   - Secure defaults

6. **Vulnerable Components** ✅
   - Dependency scanning
   - Regular updates

7. **Authentication Failures** ⚠️
   - JWT ready (future implementation)

8. **Software and Data Integrity** ✅
   - Code signing (future)
   - Dependency verification

9. **Logging and Monitoring** ✅
   - Structured logging
   - Security event logging

10. **Server-Side Request Forgery** ✅
    - URL validation
    - Whitelist approach

## 🎯 Best Practices

### 1. Principle of Least Privilege

```go
// Database user sadece gerekli izinlere sahip
// Admin işlemleri için ayrı user kullan
```

### 2. Defense in Depth

```
User Request
    ↓
Rate Limiter (Layer 1)
    ↓
Input Validation (Layer 2)
    ↓
Authentication (Layer 3)
    ↓
Authorization (Layer 4)
    ↓
Business Logic (Layer 5)
    ↓
Database (Layer 6)
```

### 3. Fail Securely

```go
// Hata durumunda güvenli default
if err != nil {
    logger.Error("Error occurred", zap.Error(err))
    return ErrInternalServerError  // Generic error
}
```

### 4. Regular Updates

```bash
# Dependencies'leri düzenli güncelle
go get -u ./...
go mod tidy

# Security scan
./scripts/audit-deps.sh
```

## 📚 Kaynaklar

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://github.com/OWASP/Go-SCP)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
