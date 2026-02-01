package port

import (
	"context"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
)

// ProviderClient veri sağlayıcılardan içerik çekmek için interface
type ProviderClient interface {
	// FetchContents provider'dan tüm içerikleri çeker ve normalize eder
	FetchContents(ctx context.Context) ([]*entity.NormalizedContent, error)

	// GetProviderInfo provider bilgilerini döner
	GetProviderInfo() *entity.Provider
}
