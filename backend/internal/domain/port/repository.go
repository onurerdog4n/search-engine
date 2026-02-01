package port

import (
	"context"
	"errors"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
)

var (
	// ErrContentNotFound içerik bulunamadığında döner
	ErrContentNotFound = errors.New("content not found")
	// ErrDuplicateContent aynı içerik zaten varsa döner
	ErrDuplicateContent = errors.New("content already exists")
)

// ContentRepository içerik veri erişim katmanı interface'i
type ContentRepository interface {
	// Create yeni bir içerik oluşturur
	Create(ctx context.Context, content *entity.Content) error

	// Update mevcut bir içeriği günceller
	Update(ctx context.Context, content *entity.Content) error

	// FindByID ID'ye göre içerik getirir
	FindByID(ctx context.Context, id int64) (*entity.Content, error)

	// Upsert içerik varsa günceller, yoksa ekler (provider_id + provider_content_id bazlı)
	Upsert(ctx context.Context, content *entity.Content) error

	// Search arama parametrelerine göre içerikleri getirir
	Search(ctx context.Context, params SearchParams) ([]*entity.Content, int64, error)

	// CreateOrUpdateStats içerik istatistiklerini oluşturur veya günceller
	CreateOrUpdateStats(ctx context.Context, stats *entity.ContentStats) error

	// CreateOrUpdateScore içerik skorunu oluşturur veya günceller
	CreateOrUpdateScore(ctx context.Context, score *entity.ContentScore) error

	// AddTags içeriğe etiketler ekler
	AddTags(ctx context.Context, contentID int64, tags []string) error

	// MarkStaleContentsAsDeleted güncellenmeyen içerikleri silinmiş olarak işaretler
	MarkStaleContentsAsDeleted(ctx context.Context, providerID int64, threshold time.Time) error
}

// SearchParams arama parametrelerini tutar
type SearchParams struct {
	Query       string              // Arama terimi (zorunlu)
	ContentType entity.ContentType  // İçerik türü filtresi (opsiyonel)
	SortBy      string              // Sıralama kriteri: "popularity" veya "relevance"
	Page        int                 // Sayfa numarası (1'den başlar)
	PageSize    int                 // Sayfa boyutu (max 50)
}

// ProviderRepository provider veri erişim katmanı interface'i
type ProviderRepository interface {
	// FindByID ID'ye göre provider getirir
	FindByID(ctx context.Context, id int64) (*entity.Provider, error)

	// FindAll tüm aktif provider'ları getirir
	FindAll(ctx context.Context) ([]*entity.Provider, error)

	// CreateSyncLog senkronizasyon logu oluşturur
	CreateSyncLog(ctx context.Context, log *entity.ProviderSyncLog) error

	// UpdateSyncLog senkronizasyon logunu günceller
	UpdateSyncLog(ctx context.Context, log *entity.ProviderSyncLog) error
}
