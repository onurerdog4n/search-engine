package usecase

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// SearchContentsUseCase arama use case'i
type SearchContentsUseCase struct {
	contentRepo port.ContentRepository
	cache       port.CacheRepository
	cacheTTL    time.Duration
}

// SearchResult arama sonucu yapısı
type SearchResult struct {
	Items      []*entity.Content `json:"items"`
	Pagination Pagination        `json:"pagination"`
}

// Pagination sayfalama bilgileri
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

// NewSearchContentsUseCase yeni bir arama use case oluşturur
func NewSearchContentsUseCase(
	contentRepo port.ContentRepository,
	cache port.CacheRepository,
	cacheTTL time.Duration,
) *SearchContentsUseCase {
	return &SearchContentsUseCase{
		contentRepo: contentRepo,
		cache:       cache,
		cacheTTL:    cacheTTL,
	}
}

// Execute arama işlemini gerçekleştirir
func (uc *SearchContentsUseCase) Execute(ctx context.Context, params port.SearchParams) (*SearchResult, error) {
	// 1. Parametreleri validate et
	if err := uc.validateParams(&params); err != nil {
		return nil, err
	}

	// 2. Cache key oluştur
	cacheKey := uc.generateCacheKey(params)

	// 3. Cache'den kontrol et
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var result SearchResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	// 4. Database'den ara
	contents, total, err := uc.contentRepo.Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("arama hatası: %w", err)
	}

	// 5. Sonucu hazırla
	if contents == nil {
		contents = make([]*entity.Content, 0)
	}
	result := &SearchResult{
		Items: contents,
		Pagination: Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: (total + int64(params.PageSize) - 1) / int64(params.PageSize),
		},
	}

	// 6. Cache'e kaydet
	if data, err := json.Marshal(result); err == nil {
		// Cache hatası kritik değil, loglanabilir ama devam edilir
		_ = uc.cache.Set(ctx, cacheKey, data, uc.cacheTTL)
	}

	return result, nil
}

// validateParams arama parametrelerini validate eder
func (uc *SearchContentsUseCase) validateParams(params *port.SearchParams) error {
	// Query artık zorunlu değil (keşfet özelliği için)

	// Page minimum 1
	if params.Page < 1 {
		params.Page = 1
	}

	// PageSize varsayılan ve maksimum kontrol
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 50 {
		params.PageSize = 50
	}

	// SortBy varsayılan değer
	if params.SortBy == "" {
		params.SortBy = "popularity"
	}

	// SortBy geçerli değer kontrolü
	if params.SortBy != "popularity" && params.SortBy != "relevance" {
		return fmt.Errorf("geçersiz sıralama kriteri: %s (popularity veya relevance olmalı)", params.SortBy)
	}

	// ContentType geçerli değer kontrolü (boş olabilir)
	if params.ContentType != "" &&
		params.ContentType != entity.ContentTypeVideo &&
		params.ContentType != entity.ContentTypeArticle {
		return fmt.Errorf("geçersiz içerik türü: %s", params.ContentType)
	}

	return nil
}

// generateCacheKey arama parametrelerinden cache key oluşturur
func (uc *SearchContentsUseCase) generateCacheKey(params port.SearchParams) string {
	// Parametreleri string'e çevir ve hash'le
	key := fmt.Sprintf("search:%s:%s:%s:%d:%d",
		params.Query,
		params.ContentType,
		params.SortBy,
		params.Page,
		params.PageSize,
	)

	// MD5 hash ile kısalt
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("search:%x", hash)
}
