package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// MockProviderClient
type mockProviderClient struct {
	contents []*entity.NormalizedContent
}

func (m *mockProviderClient) FetchContents(ctx context.Context) ([]*entity.NormalizedContent, error) {
	return m.contents, nil
}
func (m *mockProviderClient) GetProviderInfo() *entity.Provider {
	return &entity.Provider{ID: 1, Name: "Test Provider"}
}

// MockContentRepository
type mockContentRepository struct {
	port.ContentRepository // Embed interface to skip implementing all methods
	markedDeleted          bool
	providerID             int64
	threshold              time.Time
}

func (m *mockContentRepository) Upsert(ctx context.Context, content *entity.Content) error {
	return nil
}
func (m *mockContentRepository) CreateOrUpdateStats(ctx context.Context, stats *entity.ContentStats) error {
	return nil
}
func (m *mockContentRepository) CreateOrUpdateScore(ctx context.Context, score *entity.ContentScore) error {
	return nil
}
func (m *mockContentRepository) AddTags(ctx context.Context, contentID int64, tags []string) error {
	return nil
}
func (m *mockContentRepository) MarkStaleContentsAsDeleted(ctx context.Context, providerID int64, threshold time.Time) error {
	m.markedDeleted = true
	m.providerID = providerID
	m.threshold = threshold
	return nil
}

// MockScoringService
type mockScoringService struct{}

func (m *mockScoringService) CalculateScore(content *entity.Content) (*entity.ContentScore, error) {
	return &entity.ContentScore{}, nil
}

// MockCacheRepository
type mockCacheRepository struct {
	port.CacheRepository
	clearCalled bool
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
func (m *mockCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}
func (m *mockCacheRepository) Clear(ctx context.Context) error {
	m.clearCalled = true
	return nil
}

func TestSyncProviderContentsUseCase_Execute_SoftDelete(t *testing.T) {
	// 1. Setup
	mockClient := &mockProviderClient{
		contents: []*entity.NormalizedContent{}, // No content returned (simulating deletion)
	}
	mockRepo := &mockContentRepository{}
	mockScoring := &mockScoringService{}
	mockCache := &mockCacheRepository{}

	useCase := NewSyncProviderContentsUseCase(
		[]port.ProviderClient{mockClient},
		mockRepo,
		mockScoring,
		mockCache,
	)

	// 2. Execute
	ctx := context.Background()
	startTime := time.Now()
	err := useCase.Execute(ctx)

	// 3. Verify
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !mockRepo.markedDeleted {
		t.Error("MarkStaleContentsAsDeleted was NOT called")
	}

	if !mockCache.clearCalled {
		t.Error("Cache.Clear was NOT called")
	}

	if mockRepo.providerID != 1 {
		t.Errorf("Expected ProviderID 1, got %d", mockRepo.providerID)
	}

	if mockRepo.threshold.Before(startTime) {
		t.Error("Threshold time should be after test start time")
	}
}
