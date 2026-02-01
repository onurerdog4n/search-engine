package usecase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
	"github.com/onurerdog4n/search-engine/internal/domain/service"
)

// SyncProviderContentsUseCase provider senkronizasyon use case'i
type SyncProviderContentsUseCase struct {
	providerClients []port.ProviderClient
	contentRepo     port.ContentRepository
	scoringService  service.ScoringService
	cache           port.CacheRepository
}

// NewSyncProviderContentsUseCase yeni bir sync use case oluşturur
func NewSyncProviderContentsUseCase(
	providerClients []port.ProviderClient,
	contentRepo port.ContentRepository,
	scoringService service.ScoringService,
	cache port.CacheRepository,
) *SyncProviderContentsUseCase {
	return &SyncProviderContentsUseCase{
		providerClients: providerClients,
		contentRepo:     contentRepo,
		scoringService:  scoringService,
		cache:           cache,
	}
}

// Execute tüm provider'lardan veri çeker ve senkronize eder
func (uc *SyncProviderContentsUseCase) Execute(ctx context.Context) error {
	log.Println("Provider senkronizasyonu başlatılıyor...")

	var wg sync.WaitGroup
	// Her provider için senkronizasyon yap
	for _, client := range uc.providerClients {
		wg.Add(1)
		go func(c port.ProviderClient) {
			defer wg.Done()
			if err := uc.syncProvider(ctx, c); err != nil {
				log.Printf("Provider senkronizasyon hatası (%s): %v",
					c.GetProviderInfo().Name, err)
			}
		}(client)
	}

	wg.Wait()
	
	// Cache'i temizle (Invalidation)
	if err := uc.cache.Clear(ctx); err != nil {
		log.Printf("Cache temizleme hatası: %v", err)
	}

	log.Println("Provider senkronizasyonu tamamlandı")
	return nil
}

// syncProvider tek bir provider'ı senkronize eder
func (uc *SyncProviderContentsUseCase) syncProvider(ctx context.Context, client port.ProviderClient) error {
	provider := client.GetProviderInfo()
	log.Printf("Provider senkronizasyonu başlıyor: %s", provider.Name)

	startTime := time.Now()
	syncedCount := 0

	// 1. Provider'dan içerikleri çek
	normalized, err := client.FetchContents(ctx)
	if err != nil {
		return fmt.Errorf("içerikler çekilemedi: %w", err)
	}

	log.Printf("%s provider'ından %d içerik çekildi", provider.Name, len(normalized))

	// 2. Her içerik için işlem yap
	for _, nc := range normalized {
		if err := uc.processContent(ctx, provider.ID, nc); err != nil {
			log.Printf("İçerik işleme hatası (ID: %s): %v", nc.ExternalID, err)
			continue
		}
		syncedCount++
	}

	// 3. Silinmiş olanları işaretle (Soft Delete)
	if err := uc.contentRepo.MarkStaleContentsAsDeleted(ctx, provider.ID, startTime); err != nil {
		log.Printf("Silinmiş içerikleri işaretleme hatası (%s): %v", provider.Name, err)
	}

	duration := time.Since(startTime)
	log.Printf("Provider senkronizasyonu tamamlandı: %s (%d içerik, %v)",
		provider.Name, syncedCount, duration)

	return nil
}

// processContent tek bir içeriği işler (upsert + stats + score + tags)
func (uc *SyncProviderContentsUseCase) processContent(
	ctx context.Context,
	providerID int64,
	nc *entity.NormalizedContent,
) error {
	// 1. Content entity'sini oluştur
	content := &entity.Content{
		ProviderID:        providerID,
		ProviderContentID: nc.ExternalID,
		Title:             nc.Title,
		Description:       nc.Description,
		ContentType:       nc.ContentType,
		PublishedAt:       nc.PublishedAt,
	}

	// 2. Upsert yap (varsa güncelle, yoksa ekle)
	if err := uc.contentRepo.Upsert(ctx, content); err != nil {
		return fmt.Errorf("upsert hatası: %w", err)
	}

	// 3. Stats oluştur/güncelle
	stats := &entity.ContentStats{
		ContentID:   content.ID,
		Views:       nc.Stats.Views,
		Likes:       nc.Stats.Likes,
		ReadingTime: nc.Stats.ReadingTime,
		Reactions:   nc.Stats.Reactions,
	}

	if err := uc.contentRepo.CreateOrUpdateStats(ctx, stats); err != nil {
		return fmt.Errorf("stats hatası: %w", err)
	}

	// Stats'ı content'e ekle (skorlama için gerekli)
	content.Stats = stats

	// 4. Skor hesapla ve kaydet
	score, err := uc.scoringService.CalculateScore(content)
	if err != nil {
		return fmt.Errorf("skor hesaplama hatası: %w", err)
	}

	if score != nil {
		score.ContentID = content.ID
		if err := uc.contentRepo.CreateOrUpdateScore(ctx, score); err != nil {
			return fmt.Errorf("skor kaydetme hatası: %w", err)
		}
	}

	// 5. Tag'leri ekle
	if len(nc.Tags) > 0 {
		if err := uc.contentRepo.AddTags(ctx, content.ID, nc.Tags); err != nil {
			// Tag hatası kritik değil, logla ve devam et
			log.Printf("Tag ekleme hatası (Content ID: %d): %v", content.ID, err)
		}
	}

	return nil
}

// ExecuteAsync senkronizasyonu arka planda başlatır
func (uc *SyncProviderContentsUseCase) ExecuteAsync() {
	go func() {
		ctx := context.Background()
		if err := uc.Execute(ctx); err != nil {
			log.Printf("Async senkronizasyon hatası: %v", err)
		}
	}()
}
