package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onurerdog4n/search-engine/internal/application/usecase"
	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
	"github.com/onurerdog4n/search-engine/internal/domain/service"
	"github.com/onurerdog4n/search-engine/internal/transport/middleware"
)

// Mock repository for testing
type mockContentRepository struct {
	searchFunc func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error)
}

func (m *mockContentRepository) Search(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockContentRepository) FindByID(ctx context.Context, id int64) (*entity.Content, error) {
	return nil, nil
}

func (m *mockContentRepository) Create(ctx context.Context, content *entity.Content) error {
	return nil
}

func (m *mockContentRepository) Update(ctx context.Context, content *entity.Content) error {
	return nil
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
	return nil
}

// Mock cache for testing
type mockCache struct {
	getFunc func(ctx context.Context, key string) ([]byte, error)
	setFunc func(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return nil, errors.New("not found")
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockCache) InvalidatePattern(ctx context.Context, pattern string) error {
	return nil
}

func TestSearchHandler_HandleSearch(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		mockRepo := &mockContentRepository{
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

		mockCacheRepo := &mockCache{}
		searchUseCase := usecase.NewSearchContentsUseCase(mockRepo, mockCacheRepo, 60*time.Second)
		handler := NewSearchHandler(searchUseCase)

		req := httptest.NewRequest("GET", "/api/v1/search?query=test&page=1&page_size=20", nil)
		w := httptest.NewRecorder()

		handler.HandleSearch(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result usecase.SearchResult
		err := json.NewDecoder(w.Body).Decode(&result)
		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "Test Content", result.Items[0].Title)
		assert.Equal(t, int64(1), result.Pagination.TotalItems)
	})

	t.Run("search with type filter", func(t *testing.T) {
		mockRepo := &mockContentRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				assert.Equal(t, entity.ContentTypeVideo, params.ContentType)
				return []*entity.Content{}, 0, nil
			},
		}

		mockCacheRepo := &mockCache{}
		searchUseCase := usecase.NewSearchContentsUseCase(mockRepo, mockCacheRepo, 60*time.Second)
		handler := NewSearchHandler(searchUseCase)

		req := httptest.NewRequest("GET", "/api/v1/search?query=test&type=video", nil)
		w := httptest.NewRecorder()

		handler.HandleSearch(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("search with sort parameter", func(t *testing.T) {
		mockRepo := &mockContentRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				assert.Equal(t, "relevance", params.SortBy)
				return []*entity.Content{}, 0, nil
			},
		}

		mockCacheRepo := &mockCache{}
		searchUseCase := usecase.NewSearchContentsUseCase(mockRepo, mockCacheRepo, 60*time.Second)
		handler := NewSearchHandler(searchUseCase)

		req := httptest.NewRequest("GET", "/api/v1/search?query=test&sort=relevance", nil)
		w := httptest.NewRecorder()

		handler.HandleSearch(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("pagination parameters", func(t *testing.T) {
		mockRepo := &mockContentRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				assert.Equal(t, 2, params.Page)
				assert.Equal(t, 10, params.PageSize)
				return []*entity.Content{}, 0, nil
			},
		}

		mockCacheRepo := &mockCache{}
		searchUseCase := usecase.NewSearchContentsUseCase(mockRepo, mockCacheRepo, 60*time.Second)
		handler := NewSearchHandler(searchUseCase)

		req := httptest.NewRequest("GET", "/api/v1/search?query=test&page=2&page_size=10", nil)
		w := httptest.NewRecorder()

		handler.HandleSearch(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("default parameters", func(t *testing.T) {
		mockRepo := &mockContentRepository{
			searchFunc: func(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
				assert.Equal(t, 1, params.Page)
				assert.Equal(t, 20, params.PageSize)
				assert.Equal(t, "popularity", params.SortBy)
				return []*entity.Content{}, 0, nil
			},
		}

		mockCacheRepo := &mockCache{}
		searchUseCase := usecase.NewSearchContentsUseCase(mockRepo, mockCacheRepo, 60*time.Second)
		handler := NewSearchHandler(searchUseCase)

		req := httptest.NewRequest("GET", "/api/v1/search?query=test", nil)
		w := httptest.NewRecorder()

		handler.HandleSearch(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthHandler_HandleHealth(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestSyncHandler_HandleSync(t *testing.T) {
	// Mock sync use case
	mockProviders := []port.ProviderClient{}
	mockRepo := &mockContentRepository{}
	mockCacheRepo := &mockCache{}
	mockScoringService := service.NewScoringService(service.ScoringRules{
		VideoTypeWeight:   1.5,
		ArticleTypeWeight: 1.0,
	})

	syncUseCase := usecase.NewSyncProviderContentsUseCase(
		mockProviders,
		mockRepo,
		mockScoringService,
		mockCacheRepo,
	)

	handler := NewSyncHandler(syncUseCase)

	req := httptest.NewRequest("POST", "/api/v1/admin/sync", nil)
	w := httptest.NewRecorder()

	handler.HandleSync(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Senkronizasyon başlatıldı", response["message"])
	assert.Equal(t, "running", response["status"])
}

func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.CORS(handler).ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}
