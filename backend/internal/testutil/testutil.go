package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Use test database URL from environment or default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/search_engine?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Clean up tables before tests
	CleanupTestDB(t, db)

	return db
}

// TeardownTestDB closes the database connection and cleans up
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	CleanupTestDB(t, db)
	db.Close()
}

// CleanupTestDB removes all test data
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"content_tags",
		"content_scores",
		"content_stats",
		"contents",
		"tags",
		"provider_sync_logs",
		"providers",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
	}
}

// CreateTestProvider creates a test provider
func CreateTestProvider(t *testing.T, db *sql.DB, name, format string) *entity.Provider {
	t.Helper()

	provider := &entity.Provider{
		Name:     name,
		URL:      "http://test-api:8081/test",
		Format:   format,
		IsActive: true,
	}

	err := db.QueryRow(`
		INSERT INTO providers (name, url, format, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, provider.Name, provider.URL, provider.Format, provider.IsActive).
		Scan(&provider.ID, &provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		t.Fatalf("Failed to create test provider: %v", err)
	}

	return provider
}

// CreateTestContent creates a test content
func CreateTestContent(t *testing.T, db *sql.DB, providerID int64, contentType entity.ContentType) *entity.Content {
	t.Helper()

	content := &entity.Content{
		ProviderID:        providerID,
		ProviderContentID: fmt.Sprintf("test-content-%d", time.Now().UnixNano()),
		Title:             "Test Content",
		Description:       "Test Description",
		ContentType:       contentType,
		PublishedAt:       time.Now().Add(-24 * time.Hour),
		RawData:           `{"test": "data"}`,
	}

	err := db.QueryRow(`
		INSERT INTO contents (provider_id, provider_content_id, title, description, content_type, published_at, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, content.ProviderID, content.ProviderContentID, content.Title, content.Description,
		content.ContentType, content.PublishedAt, content.RawData).
		Scan(&content.ID, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		t.Fatalf("Failed to create test content: %v", err)
	}

	return content
}

// CreateTestContentWithStats creates a test content with stats
func CreateTestContentWithStats(t *testing.T, db *sql.DB, providerID int64, contentType entity.ContentType, views int64, likes int32) *entity.Content {
	t.Helper()

	content := CreateTestContent(t, db, providerID, contentType)

	// Add stats
	stats := &entity.ContentStats{
		ContentID:   content.ID,
		Views:       views,
		Likes:       likes,
		ReadingTime: 5,
		Reactions:   10,
	}

	err := db.QueryRow(`
		INSERT INTO content_stats (content_id, views, likes, reading_time, reactions)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, updated_at
	`, stats.ContentID, stats.Views, stats.Likes, stats.ReadingTime, stats.Reactions).
		Scan(&stats.ID, &stats.UpdatedAt)

	if err != nil {
		t.Fatalf("Failed to create test stats: %v", err)
	}

	content.Stats = stats
	return content
}

// CreateTestContentWithScore creates a test content with score
func CreateTestContentWithScore(t *testing.T, db *sql.DB, providerID int64, finalScore float64) *entity.Content {
	t.Helper()

	content := CreateTestContentWithStats(t, db, providerID, entity.ContentTypeVideo, 100000, 5000)

	// Add score
	score := &entity.ContentScore{
		ContentID:       content.ID,
		BaseScore:       100.0,
		TypeWeight:      1.5,
		RecencyScore:    5.0,
		EngagementScore: 0.5,
		FinalScore:      finalScore,
	}

	err := db.QueryRow(`
		INSERT INTO content_scores (content_id, base_score, type_weight, recency_score, engagement_score, final_score)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, calculated_at
	`, score.ContentID, score.BaseScore, score.TypeWeight, score.RecencyScore,
		score.EngagementScore, score.FinalScore).
		Scan(&score.ID, &score.CalculatedAt)

	if err != nil {
		t.Fatalf("Failed to create test score: %v", err)
	}

	content.Score = score
	return content
}

// CreateTestTag creates a test tag
func CreateTestTag(t *testing.T, db *sql.DB, name string) *entity.Tag {
	t.Helper()

	tag := &entity.Tag{
		Name: name,
	}

	err := db.QueryRow(`
		INSERT INTO tags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, created_at
	`, tag.Name).Scan(&tag.ID, &tag.CreatedAt)

	if err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}

	return tag
}

// AddTagToContent associates a tag with content
func AddTagToContent(t *testing.T, db *sql.DB, contentID, tagID int64) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO content_tags (content_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, contentID, tagID)

	if err != nil {
		t.Fatalf("Failed to add tag to content: %v", err)
	}
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if condition() {
				return true
			}
		}
	}
}
