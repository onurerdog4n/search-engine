package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// Mock repository for testing
type mockSearchRepository struct {
	searchFunc func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error)
}

func (m *mockSearchRepository) Search(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockSearchRepository) FindByID(ctx context.Context, id int64) (*entity.Content, error) {
	return nil, nil
}

func (m *mockSearchRepository) Create(ctx context.Context, content *entity.Content) error {
	return nil
}

func (m *mockSearchRepository) Update(ctx context.Context, content *entity.Content) error {
	return nil
}

func (m *mockSearchRepository) Upsert(ctx context.Context, content *entity.Content) error {
	return nil
}

func (m *mockSearchRepository) CreateOrUpdateStats(ctx context.Context, stats *entity.ContentStats) error {
	return nil
}

func (m *mockSearchRepository) CreateOrUpdateScore(ctx context.Context, score *entity.ContentScore) error {
	return nil
}

func (m *mockSearchRepository) AddTags(ctx context.Context, contentID int64, tags []string) error {
	return nil
}

func (m *mockSearchRepository) MarkStaleContentsAsDeleted(ctx context.Context, providerID int64, threshold time.Time) error {
	return nil
}

// Mock cache for testing
type mockSearchCache struct {
	storage map[string][]byte
	getFunc func(ctx context.Context, key string) ([]byte, error)
	setFunc func(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

func newMockSearchCache() *mockSearchCache {
	return &mockSearchCache{
		storage: make(map[string][]byte),
	}
}

func (m *mockSearchCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	if val, ok := m.storage[key]; ok {
		return val, nil
	}
	return nil, errors.New("not found")
}

func (m *mockSearchCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	m.storage[key] = value
	return nil
}

func (m *mockSearchCache) Delete(ctx context.Context, key string) error {
	delete(m.storage, key)
	return nil
}

func (m *mockSearchCache) InvalidatePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *mockSearchCache) Clear(ctx context.Context) error {
	m.storage = make(map[string][]byte)
	return nil
}

func TestSearchContentsUseCase_Execute(t *testing.T) {
	t.Run("successful search without cache", func(t *testing.T) {
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				contents := []*entity.Content{
					{
						ID:          1,
						Title:       "Test Content",
						ContentType: entity.ContentTypeVideo,
					},
				}
				return contents, 1, nil
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:    "test",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		result, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "Test Content", result.Items[0].Title)
		assert.Equal(t, int64(1), result.Pagination.TotalItems)
	})

	t.Run("cache hit", func(t *testing.T) {
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				t.Fatal("Repository should not be called on cache hit")
				return nil, 0, nil
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:    "test",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		// First call - populate cache
		mockRepo.searchFunc = func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
			return []*entity.Content{{ID: 1, Title: "Cached Content"}}, 1, nil
		}
		result1, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)

		// Second call - should use cache
		mockRepo.searchFunc = func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
			t.Fatal("Repository should not be called on cache hit")
			return nil, 0, nil
		}
		result2, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)

		assert.Equal(t, result1.Items[0].Title, result2.Items[0].Title)
	})

	t.Run("parameter validation - invalid sort", func(t *testing.T) {
		mockRepo := &mockSearchRepository{}
		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:    "test",
			SortBy:   "invalid",
			Page:     1,
			PageSize: 20,
		}

		_, err := useCase.Execute(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "geçersiz sıralama kriteri")
	})

	t.Run("parameter validation - invalid content type", func(t *testing.T) {
		mockRepo := &mockSearchRepository{}
		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:       "test",
			ContentType: "invalid",
			SortBy:      "popularity",
			Page:        1,
			PageSize:    20,
		}

		_, err := useCase.Execute(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "geçersiz içerik türü")
	})

	t.Run("parameter defaults", func(t *testing.T) {
		var capturedParams port.SearchParams
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				capturedParams = params
				return []*entity.Content{}, 0, nil
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query: "test",
			// No SortBy, Page, PageSize specified
		}

		_, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)

		assert.Equal(t, "popularity", capturedParams.SortBy)
		assert.Equal(t, 1, capturedParams.Page)
		assert.Equal(t, 20, capturedParams.PageSize)
	})

	t.Run("page size limits", func(t *testing.T) {
		var capturedParams port.SearchParams
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				capturedParams = params
				return []*entity.Content{}, 0, nil
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		// Test max limit
		params := port.SearchParams{
			Query:    "test",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 100, // Should be capped at 50
		}

		_, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, 50, capturedParams.PageSize)
	})

	t.Run("empty results", func(t *testing.T) {
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				return nil, 0, nil
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:    "nonexistent",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		result, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)
		assert.Len(t, result.Items, 0)
		assert.Equal(t, int64(0), result.Pagination.TotalItems)
	})

	t.Run("pagination calculation", func(t *testing.T) {
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				return []*entity.Content{{ID: 1}}, 100, nil // 100 total items
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:    "test",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		result, err := useCase.Execute(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(100), result.Pagination.TotalItems)
		assert.Equal(t, int64(5), result.Pagination.TotalPages) // 100 / 20 = 5
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &mockSearchRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				return nil, 0, errors.New("database error")
			},
		}

		mockCache := newMockSearchCache()
		useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

		params := port.SearchParams{
			Query:    "test",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		_, err := useCase.Execute(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "arama hatası")
	})
}

func TestSearchContentsUseCase_CacheKeyGeneration(t *testing.T) {
	mockRepo := &mockSearchRepository{
		searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
			return []*entity.Content{}, 0, nil
		},
	}

	mockCache := newMockSearchCache()
	useCase := NewSearchContentsUseCase(mockRepo, mockCache, 60*time.Second)

	// Execute with same parameters twice
	params := port.SearchParams{
		Query:    "test",
		SortBy:   "popularity",
		Page:     1,
		PageSize: 20,
	}

	_, err := useCase.Execute(context.Background(), params)
	require.NoError(t, err)

	// Cache should have one entry
	assert.Len(t, mockCache.storage, 1)

	// Different parameters should generate different cache key
	params.Page = 2
	_, err = useCase.Execute(context.Background(), params)
	require.NoError(t, err)

	// Cache should have two entries
	assert.Len(t, mockCache.storage, 2)
}
