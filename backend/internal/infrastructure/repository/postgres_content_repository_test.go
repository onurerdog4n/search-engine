package repository

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
	"github.com/onurerdog4n/search-engine/internal/testutil"
)

func TestPostgresContentRepository_Upsert(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")

	t.Run("insert new content", func(t *testing.T) {
		content := &entity.Content{
			ProviderID:        provider.ID,
			ProviderContentID: "test-123",
			Title:             "Test Content",
			Description:       "Test Description",
			ContentType:       entity.ContentTypeVideo,
			PublishedAt:       time.Now(),
			RawData:           `{"test": "data"}`,
		}

		err := repo.Upsert(context.Background(), content)
		require.NoError(t, err)
		assert.NotZero(t, content.ID)
		assert.NotZero(t, content.CreatedAt)
	})

	t.Run("update existing content", func(t *testing.T) {
		content := &entity.Content{
			ProviderID:        provider.ID,
			ProviderContentID: "test-456",
			Title:             "Original Title",
			Description:       "Original Description",
			ContentType:       entity.ContentTypeVideo,
			PublishedAt:       time.Now(),
		}

		// Insert
		err := repo.Upsert(context.Background(), content)
		require.NoError(t, err)
		originalID := content.ID

		// Update
		content.Title = "Updated Title"
		content.Description = "Updated Description"
		err = repo.Upsert(context.Background(), content)
		require.NoError(t, err)

		// ID should remain the same
		assert.Equal(t, originalID, content.ID)

		// Verify update
		found, err := repo.FindByID(context.Background(), content.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", found.Title)
		assert.Equal(t, "Updated Description", found.Description)
	})
}

func TestPostgresContentRepository_Search(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")

	// Create test contents
	content1 := testutil.CreateTestContentWithScore(t, db, provider.ID, 150.0)
	content1.Title = "Golang Tutorial for Beginners"
	repo.Upsert(context.Background(), content1)

	content2 := testutil.CreateTestContentWithScore(t, db, provider.ID, 100.0)
	content2.Title = "Python Programming Guide"
	content2.ContentType = entity.ContentTypeArticle
	repo.Upsert(context.Background(), content2)

	content3 := testutil.CreateTestContentWithScore(t, db, provider.ID, 200.0)
	content3.Title = "Advanced Golang Patterns"
	repo.Upsert(context.Background(), content3)

	// Add tags
	golangTag := testutil.CreateTestTag(t, db, "golang")
	pythonTag := testutil.CreateTestTag(t, db, "python")
	testutil.AddTagToContent(t, db, content1.ID, golangTag.ID)
	testutil.AddTagToContent(t, db, content3.ID, golangTag.ID)
	testutil.AddTagToContent(t, db, content2.ID, pythonTag.ID)

	t.Run("search by query", func(t *testing.T) {
		params := port.SearchParams{
			Query:    "golang",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		results, total, err := repo.Search(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)

		// Should be sorted by score DESC
		assert.Equal(t, "Advanced Golang Patterns", results[0].Title)
		assert.Equal(t, "Golang Tutorial for Beginners", results[1].Title)
	})

	t.Run("filter by content type", func(t *testing.T) {
		params := port.SearchParams{
			ContentType: entity.ContentTypeVideo,
			SortBy:      "popularity",
			Page:        1,
			PageSize:    20,
		}

		results, total, err := repo.Search(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // content1 and content3 are videos
		assert.Len(t, results, 2)
	})

	t.Run("pagination", func(t *testing.T) {
		params := port.SearchParams{
			SortBy:   "popularity",
			Page:     1,
			PageSize: 2,
		}

		results, total, err := repo.Search(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 2)

		// Page 2
		params.Page = 2
		results, total, err = repo.Search(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 1)
	})

	t.Run("sort by relevance", func(t *testing.T) {
		params := port.SearchParams{
			Query:    "golang",
			SortBy:   "relevance",
			Page:     1,
			PageSize: 20,
		}

		results, total, err := repo.Search(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
		// Results should have relevance scores
		assert.Greater(t, results[0].RelevanceScore, 0.0)
	})

	t.Run("empty query returns all", func(t *testing.T) {
		params := port.SearchParams{
			Query:    "",
			SortBy:   "popularity",
			Page:     1,
			PageSize: 20,
		}

		results, total, err := repo.Search(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
	})
}

func TestPostgresContentRepository_CreateOrUpdateStats(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")
	content := testutil.CreateTestContent(t, db, provider.ID, entity.ContentTypeVideo)

	t.Run("create stats", func(t *testing.T) {
		stats := &entity.ContentStats{
			ContentID:   content.ID,
			Views:       10000,
			Likes:       500,
			ReadingTime: 0,
			Reactions:   0,
		}

		err := repo.CreateOrUpdateStats(context.Background(), stats)
		require.NoError(t, err)
		assert.NotZero(t, stats.ID)
		assert.NotZero(t, stats.UpdatedAt)
	})

	t.Run("update stats", func(t *testing.T) {
		stats := &entity.ContentStats{
			ContentID:   content.ID,
			Views:       20000,
			Likes:       1000,
			ReadingTime: 0,
			Reactions:   0,
		}

		err := repo.CreateOrUpdateStats(context.Background(), stats)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(context.Background(), content.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(20000), found.Stats.Views)
		assert.Equal(t, int32(1000), found.Stats.Likes)
	})
}

func TestPostgresContentRepository_CreateOrUpdateScore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")
	content := testutil.CreateTestContent(t, db, provider.ID, entity.ContentTypeVideo)

	t.Run("create score", func(t *testing.T) {
		score := &entity.ContentScore{
			ContentID:       content.ID,
			BaseScore:       100.0,
			TypeWeight:      1.5,
			RecencyScore:    5.0,
			EngagementScore: 0.5,
			FinalScore:      155.5,
		}

		err := repo.CreateOrUpdateScore(context.Background(), score)
		require.NoError(t, err)
		assert.NotZero(t, score.ID)
		assert.NotZero(t, score.CalculatedAt)
	})

	t.Run("update score", func(t *testing.T) {
		score := &entity.ContentScore{
			ContentID:       content.ID,
			BaseScore:       120.0,
			TypeWeight:      1.5,
			RecencyScore:    3.0,
			EngagementScore: 0.7,
			FinalScore:      183.7,
		}

		err := repo.CreateOrUpdateScore(context.Background(), score)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(context.Background(), content.ID)
		require.NoError(t, err)
		assert.Equal(t, 183.7, found.Score.FinalScore)
	})
}

func TestPostgresContentRepository_AddTags(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")
	content := testutil.CreateTestContent(t, db, provider.ID, entity.ContentTypeVideo)

	t.Run("add tags", func(t *testing.T) {
		tags := []string{"golang", "tutorial", "beginner"}

		err := repo.AddTags(context.Background(), content.ID, tags)
		require.NoError(t, err)

		// Verify tags
		found, err := repo.FindByID(context.Background(), content.ID)
		require.NoError(t, err)
		assert.Len(t, found.Tags, 3)
	})

	t.Run("add duplicate tags", func(t *testing.T) {
		tags := []string{"golang", "advanced"}

		err := repo.AddTags(context.Background(), content.ID, tags)
		require.NoError(t, err)

		// Should have 4 tags total (golang not duplicated)
		found, err := repo.FindByID(context.Background(), content.ID)
		require.NoError(t, err)
		assert.Len(t, found.Tags, 4)
	})
}

func TestPostgresContentRepository_MarkStaleContentsAsDeleted(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")

	// Create old content
	oldContent := testutil.CreateTestContent(t, db, provider.ID, entity.ContentTypeVideo)
	
	// Temporarily disable trigger to set old updated_at
	_, err := db.Exec("ALTER TABLE contents DISABLE TRIGGER update_contents_updated_at")
	require.NoError(t, err)
	defer db.Exec("ALTER TABLE contents ENABLE TRIGGER update_contents_updated_at")
	
	// Manually set updated_at to old time (2 hours ago) using UTC
	oldTime := time.Now().UTC().Add(-2 * time.Hour)
	_, err = db.Exec("UPDATE contents SET updated_at = $1 WHERE id = $2", 
		oldTime, oldContent.ID)
	require.NoError(t, err)
	
	// Re-enable trigger
	_, err = db.Exec("ALTER TABLE contents ENABLE TRIGGER update_contents_updated_at")
	require.NoError(t, err)
	
	// Verify the update worked
	var verifyTime time.Time
	err = db.QueryRow("SELECT updated_at FROM contents WHERE id = $1", oldContent.ID).Scan(&verifyTime)
	require.NoError(t, err)
	require.True(t, verifyTime.Before(time.Now().UTC().Add(-90*time.Minute)), 
		"Old content updated_at should be at least 90 minutes old, got: %v", verifyTime)

	// Create new content
	newContent := testutil.CreateTestContent(t, db, provider.ID, entity.ContentTypeVideo)

	t.Run("mark stale contents", func(t *testing.T) {
		// Threshold: 1 hour ago (content older than this should be marked as deleted)
		threshold := time.Now().UTC().Add(-1 * time.Hour)
		
		err := repo.MarkStaleContentsAsDeleted(context.Background(), provider.ID, threshold)
		require.NoError(t, err)

		// Verify old content is marked as deleted by checking database directly
		var deleted int
		err = db.QueryRow("SELECT deleted FROM contents WHERE id = $1", oldContent.ID).Scan(&deleted)
		require.NoError(t, err)
		assert.Equal(t, 1, deleted, "Old content should be marked as deleted")

		// FindByID should not find deleted content
		found, err := repo.FindByID(context.Background(), oldContent.ID)
		assert.Error(t, err, "Should not find deleted content")
		assert.Nil(t, found)

		// New content should still exist and not be deleted
		found, err = repo.FindByID(context.Background(), newContent.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		
		// Verify new content is not deleted
		err = db.QueryRow("SELECT deleted FROM contents WHERE id = $1", newContent.ID).Scan(&deleted)
		require.NoError(t, err)
		assert.Equal(t, 0, deleted, "New content should not be marked as deleted")
	})
}

func TestPostgresContentRepository_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewPostgresContentRepository(db)
	provider := testutil.CreateTestProvider(t, db, "Test Provider", "json")

	t.Run("find existing content", func(t *testing.T) {
		content := testutil.CreateTestContentWithScore(t, db, provider.ID, 150.0)

		found, err := repo.FindByID(context.Background(), content.ID)
		require.NoError(t, err)
		assert.Equal(t, content.ID, found.ID)
		assert.NotNil(t, found.Stats)
		assert.NotNil(t, found.Score)
	})

	t.Run("find non-existing content", func(t *testing.T) {
		found, err := repo.FindByID(context.Background(), 99999)
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}
