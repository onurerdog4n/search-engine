---
title: Kurulum
description: Sistem kurulum ve yapılandırma kılavuzu
navigation: true
---

# 📦 Kurulum ve Çalıştırma

## Docker Compose ile Kurulum (Önerilen)

En hızlı ve kolay kurulum yöntemi.

### Gereksinimler

- Docker 20.10+
- Docker Compose 2.0+

### Adımlar

```bash
# Repository'yi klonla
git clone <repository-url>
cd project-search

# Tüm servisleri başlat
docker-compose up --build

# Detached mode (arka planda)
docker-compose up --build -d
```

### Servisler

| Servis | Port | URL |
|--------|------|-----|
| **Backend** | 8080 | http://localhost:8080 |
| **Frontend** | 3000 | http://localhost:3000 |
| **PostgreSQL** | 5432 | localhost:5432 |
| **Redis** | 6379 | localhost:6379 |
| **Mock API** | 8081 | http://localhost:8081 |

### Logları İzleme

```bash
# Tüm servislerin logları
docker-compose logs -f

# Sadece backend logları
docker-compose logs -f backend

# Sadece database logları
docker-compose logs -f postgres
```

### Durdurma ve Temizleme

```bash
# Servisleri durdur
docker-compose down

# Volume'leri de sil (database verileri)
docker-compose down -v

# Image'leri de sil
docker-compose down --rmi all
```

## Manuel Kurulum

### 1. PostgreSQL Kurulumu

#### macOS
```bash
# Homebrew ile
brew install postgresql@16
brew services start postgresql@16
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install postgresql-16
sudo systemctl start postgresql
```

#### Database Oluşturma

```bash
# PostgreSQL'e bağlan
psql postgres

# Database oluştur
CREATE DATABASE search_engine;

# Kullanıcı oluştur (opsiyonel)
CREATE USER search_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE search_engine TO search_user;
```

#### Migration'ları Çalıştır

```bash
cd backend

# Migration dosyalarını çalıştır
psql -U postgres -d search_engine -f migrations/001_create_tables.up.sql
psql -U postgres -d search_engine -f migrations/002_add_raw_data.up.sql
psql -U postgres -d search_engine -f migrations/003_add_deleted_column.up.sql
```

### 2. Redis Kurulumu

#### macOS
```bash
brew install redis
brew services start redis
```

#### Linux
```bash
sudo apt install redis-server
sudo systemctl start redis-server
```

#### Test Et

```bash
redis-cli ping
# Çıktı: PONG
```

### 3. Backend Kurulumu

#### Gereksinimler

- Go 1.21 veya üzeri

#### Kurulum

```bash
cd backend

# Go modüllerini indir
go mod download

# Environment değişkenlerini ayarla
cp .env.example .env

# .env dosyasını düzenle
nano .env
```

#### Environment Variables

```.env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/search_engine?sslmode=disable

# Redis
REDIS_URL=localhost:6379

# Server
PORT=8080

# Senkronizasyon (saniye cinsinden)
SYNC_INTERVAL=3600  # 1 saat

# Rate Limiting
RATE_LIMIT_PER_MINUTE=60

# Cache
CACHE_TTL_SECONDS=60
```

#### Çalıştırma

```bash
# Development mode
go run cmd/server/main.go

# Production build
go build -o bin/server cmd/server/main.go
./bin/server
```

### 4. Mock API Kurulumu (Opsiyonel)

Test için mock provider API'leri.

```bash
cd backend/mock-api

# Çalıştır
go run main.go

# Mock API: http://localhost:8081
```

**Endpoints:**
- `http://localhost:8081/provider-1?page=1` (JSON)
- `http://localhost:8081/provider-2?page=1` (XML)

### 5. Frontend Kurulumu (Opsiyonel)

#### Gereksinimler

- Node.js 18+
- npm veya yarn

#### Kurulum

```bash
cd frontend

# Dependencies
npm install

# Environment ayarla
echo "NUXT_PUBLIC_API_BASE=http://localhost:8080" > .env

# Development mode
npm run dev

# Production build
npm run build
npm run preview
```

## Doğrulama

### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

**Beklenen Çıktı:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T14:30:00Z"
}
```

### Arama Testi

```bash
curl "http://localhost:8080/api/v1/search?query=golang"
```

### Manuel Senkronizasyon

```bash
curl -X POST http://localhost:8080/api/v1/admin/sync
```

## Troubleshooting

### PostgreSQL Bağlantı Hatası

```
Error: connection refused
```

**Çözüm:**
```bash
# PostgreSQL çalışıyor mu kontrol et
pg_isready

# Servisi başlat
brew services start postgresql@16
```

### Redis Bağlantı Hatası

```
Error: dial tcp 127.0.0.1:6379: connect: connection refused
```

**Çözüm:**
```bash
# Redis çalışıyor mu kontrol et
redis-cli ping

# Servisi başlat
brew services start redis
```

### Migration Hataları

```
Error: relation "contents" already exists
```

**Çözüm:**
```bash
# Database'i sıfırdan oluştur
psql -U postgres -d search_engine -f migrations/001_create_tables.down.sql
psql -U postgres -d search_engine -f migrations/001_create_tables.up.sql
```

### Port Zaten Kullanımda

```
Error: bind: address already in use
```

**Çözüm:**
```bash
# Port'u kullanan process'i bul
lsof -i :8080

# Process'i öldür
kill -9 <PID>
```

## Production Deployment

### Environment Variables

Production için `.env` dosyasını güncelle:

```env
# Güvenli password kullan
DATABASE_URL=postgres://user:STRONG_PASSWORD@db-host:5432/search_engine?sslmode=require

# Redis authentication ekle
REDIS_URL=redis://:PASSWORD@redis-host:6379

# Production port
PORT=8080

# Rate limiting artır
RATE_LIMIT_PER_MINUTE=120
```

### Systemd Service (Linux)

```ini
[Unit]
Description=Search Engine Backend
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/search-engine
ExecStart=/opt/search-engine/bin/server
Restart=always
Environment="DATABASE_URL=..."
Environment="REDIS_URL=..."

[Install]
WantedBy=multi-user.target
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### SSL/TLS (Let's Encrypt)

```bash
# Certbot ile SSL
sudo certbot --nginx -d api.example.com
```

## Monitoring

### Logs

```bash
# Backend logs
tail -f /var/log/search-engine/backend.log

# PostgreSQL logs
tail -f /var/log/postgresql/postgresql-16-main.log

# Redis logs
tail -f /var/log/redis/redis-server.log
```

### Metrics

Prometheus endpoint'i eklemek için:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

## Backup ve Recovery

### Database Backup

```bash
# Backup oluştur
pg_dump -U postgres search_engine > backup.sql

# Backup'ı geri yükle
psql -U postgres search_engine < backup.sql
```

### Automated Backups

```bash
# Cron job ekle
0 2 * * * pg_dump -U postgres search_engine > /backups/search_engine_$(date +\%Y\%m\%d).sql
```
