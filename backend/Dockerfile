# Build stage
FROM golang:1.21-alpine AS builder

# Gerekli paketleri yükle
RUN apk add --no-cache git

# Çalışma dizinini ayarla
WORKDIR /app

# Kaynak kodları kopyala
COPY . .

# Go modüllerini düzenle ve indir
RUN go mod tidy && go mod download && go mod verify

# Binary'yi derle
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:latest

# Gerekli paketleri yükle
RUN apk --no-cache add ca-certificates tzdata

# Çalışma dizinini ayarla
WORKDIR /app

# Binary'yi kopyala
COPY --from=builder /app/server /usr/local/bin/server

# Migrations'ları kopyala
COPY --from=builder /app/migrations ./migrations

# Port'u expose et
EXPOSE 8080

# Uygulamayı çalıştır
CMD ["server"]
