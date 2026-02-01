package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/onurerdog4n/search-engine/internal/application/usecase"
	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
	"github.com/onurerdog4n/search-engine/internal/domain/service"
	"github.com/onurerdog4n/search-engine/internal/infrastructure/cache"
	"github.com/onurerdog4n/search-engine/internal/infrastructure/config"
	"github.com/onurerdog4n/search-engine/internal/infrastructure/logger"
	"github.com/onurerdog4n/search-engine/internal/infrastructure/provider"
	"github.com/onurerdog4n/search-engine/internal/infrastructure/repository"
	transportHttp "github.com/onurerdog4n/search-engine/internal/transport/http"
	"github.com/onurerdog4n/search-engine/internal/transport/middleware"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize logger
	err = logger.InitGlobalLogger(logger.Config{
		Level:      cfg.Logger.Level,
		Encoding:   cfg.Logger.Encoding,
		OutputPath: cfg.Logger.OutputPath,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.GetLogger().Sync()

	logger.Info("Starting search engine server", zap.String("version", "1.0.0"))

	// 3. Database connection with pooling
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Database ping failed", zap.Error(err))
	}
	logger.Info("Database connection established")

	// 4. Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.URL,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("Redis connection failed", zap.Error(err))
	}
	logger.Info("Redis connection established")

	// 5. Repositories oluştur
	contentRepo := repository.NewPostgresContentRepository(db)
	cacheRepo := cache.NewRedisCache(rdb)

	// 6. Services
	scoringService := service.NewScoringService(service.ScoringRules{
		VideoTypeWeight:   1.5,
		ArticleTypeWeight: 1.0,
	})

	// 7. Provider clients
	providerClients := createProviderClients(db)
	logger.Info("Provider clients created", zap.Int("count", len(providerClients)))

	// 8. Use cases
	// 8. Use cases
	searchUseCase := usecase.NewSearchContentsUseCase(
		contentRepo,
		cacheRepo,
		time.Duration(cfg.Cache.TTLSeconds)*time.Second,
	)

	syncUseCase := usecase.NewSyncProviderContentsUseCase(
		providerClients,
		contentRepo,
		scoringService,
		cacheRepo,
	)

	// 9. İlk senkronizasyonu başlat
	log.Println("İlk provider senkronizasyonu başlatılıyor...")
	go syncUseCase.Execute(ctx)

	// 10. Periyodik senkronizasyon scheduler'ı başlat
	startSyncScheduler(syncUseCase, cfg.Sync.IntervalSeconds)

	// 11. HTTP handlers oluştur
	searchHandler := transportHttp.NewSearchHandler(searchUseCase)
	syncHandler := transportHttp.NewSyncHandler(syncUseCase)
	healthHandler := transportHttp.NewHealthHandler(db, rdb)

	// 12. Router setup
	r := mux.NewRouter()

	// Global middleware'ler
	r.Use(middleware.CORS)
	r.Use(middleware.Logging)

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Rate limiter (search endpoint için)
	rateLimiter := middleware.NewRateLimiter(cfg.Server.RateLimitPerMinute)
	rateLimiter.CleanupOldLimiters()

	// Public endpoints
	api.HandleFunc("/search", searchHandler.HandleSearch).Methods("GET", "OPTIONS")
	api.HandleFunc("/health", healthHandler.HandleHealth).Methods("GET")

	// Admin endpoints (rate limit yok)
	api.HandleFunc("/admin/sync", syncHandler.HandleSync).Methods("POST", "OPTIONS")

	// Rate limiter'ı search endpoint'ine ekle
	searchRoute := api.NewRoute().Path("/search").Methods("GET")
	searchRoute.Handler(rateLimiter.Middleware(http.HandlerFunc(searchHandler.HandleSearch)))

	// 13. Server'ı başlat
	addr := ":" + cfg.Server.Port
	log.Printf("🚀 Server başlatılıyor: http://localhost%s", addr)
	log.Printf("   - Health check: http://localhost%s/api/v1/health", addr)
	log.Printf("   - Search: http://localhost%s/api/v1/search?query=go", addr)
	log.Printf("   - Admin sync: http://localhost%s/api/v1/admin/sync", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server başlatma hatası: %v", err)
	}
}

// createProviderClients database'den provider'ları okuyup client'ları oluşturur
func createProviderClients(db *sql.DB) []port.ProviderClient {
	// Provider'ları database'den oku
	rows, err := db.Query("SELECT id, name, url, format FROM providers WHERE is_active = true")
	if err != nil {
		log.Printf("Provider'lar okunamadı: %v", err)
		return nil
	}
	defer rows.Close()

	var clients []port.ProviderClient

	for rows.Next() {
		var p entity.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Format); err != nil {
			log.Printf("Provider scan hatası: %v", err)
			continue
		}

		// Format'a göre uygun client oluştur
		var client port.ProviderClient

		switch p.Format {
		case "json":
			client = provider.NewJSONProvider(&p, p.URL)
		case "xml":
			client = provider.NewXMLProvider(&p, p.URL)
		default:
			log.Printf("Bilinmeyen provider formatı: %s", p.Format)
			continue
		}

		clients = append(clients, client)
	}

	return clients
}

// startSyncScheduler periyodik senkronizasyon scheduler'ını başlatır
func startSyncScheduler(syncUseCase *usecase.SyncProviderContentsUseCase, intervalSeconds int) {
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	go func() {
		for range ticker.C {
			log.Println("Periyodik senkronizasyon başlatılıyor...")
			ctx := context.Background()
			if err := syncUseCase.Execute(ctx); err != nil {
				log.Printf("Periyodik senkronizasyon hatası: %v", err)
			}
		}
	}()
	log.Printf("✓ Periyodik senkronizasyon scheduler başlatıldı (%d saniye aralıkla)", intervalSeconds)
}
