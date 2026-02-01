package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/onurerdog4n/search-engine/internal/application/usecase"
	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// SearchHandler arama HTTP handler'ı
type SearchHandler struct {
	searchUseCase *usecase.SearchContentsUseCase
}

// NewSearchHandler yeni bir search handler oluşturur
func NewSearchHandler(searchUseCase *usecase.SearchContentsUseCase) *SearchHandler {
	return &SearchHandler{
		searchUseCase: searchUseCase,
	}
}

// HandleSearch arama isteğini işler
// GET /api/v1/search?query=go&type=video&sort=popularity&page=1&page_size=20
func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	// 1. Query parametrelerini al
	query := r.URL.Query().Get("query")
	// Query artık zorunlu değil, boş ise tüm sonuçlar döner

	contentType := r.URL.Query().Get("type")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "popularity"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	// 2. Search params oluştur
	params := port.SearchParams{
		Query:       query,
		ContentType: entity.ContentType(contentType),
		SortBy:      sortBy,
		Page:        page,
		PageSize:    pageSize,
	}

	// 3. Use case'i çalıştır
	result, err := h.searchUseCase.Execute(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Başarılı response döndür
	respondJSON(w, http.StatusOK, result)
}

// SyncHandler senkronizasyon HTTP handler'ı
type SyncHandler struct {
	syncUseCase *usecase.SyncProviderContentsUseCase
}

// NewSyncHandler yeni bir sync handler oluşturur
func NewSyncHandler(syncUseCase *usecase.SyncProviderContentsUseCase) *SyncHandler {
	return &SyncHandler{
		syncUseCase: syncUseCase,
	}
}

// HandleSync senkronizasyon isteğini işler
// POST /api/v1/admin/sync
func (h *SyncHandler) HandleSync(w http.ResponseWriter, r *http.Request) {
	// Arka planda senkronizasyonu başlat
	h.syncUseCase.ExecuteAsync()

	// Hemen response döndür
	respondJSON(w, http.StatusAccepted, map[string]string{
		"message": "Senkronizasyon başlatıldı",
		"status":  "running",
	})
}

// HealthHandler health check HTTP handler'ı
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthHandler yeni bir health handler oluşturur
func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// HandleHealth health check isteğini işler
// GET /api/v1/health
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"services":  make(map[string]string),
	}

	services := health["services"].(map[string]string)

	// Check database
	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			services["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			services["database"] = "healthy"
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			services["redis"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			services["redis"] = "healthy"
		}
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	respondJSON(w, statusCode, health)
}

// respondJSON JSON response döndürür
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError hata response döndürür
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}
