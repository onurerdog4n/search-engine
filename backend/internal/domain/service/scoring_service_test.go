package service

import (
	"testing"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestScoringService_CalculateScore(t *testing.T) {
	// Initialize service with default rules
	rules := ScoringRules{
		VideoTypeWeight:   1.5,
		ArticleTypeWeight: 1.0,
	}
	service := NewScoringService(rules)

	t.Run("Should return nil if stats are missing", func(t *testing.T) {
		content := &entity.Content{
			ID:          1,
			Stats:       nil,
			ContentType: entity.ContentTypeVideo,
		}

		score, err := service.CalculateScore(content)
		assert.NoError(t, err)
		assert.Nil(t, score)
	})

	t.Run("Should calculate correct score for newly published VIDEO", func(t *testing.T) {
		// New video (published < 7 days ago which gives +5 Recency)
		// Views: 10000 -> 10 points
		// Likes: 500 -> 5 points
		// Base = 15
		// Type Weight = 1.5
		// Weighted Base = 22.5
		// Engagement = (500/10000) * 10 = 0.05 * 10 = 0.5
		// Recency = 5
		// Final = 22.5 + 5 + 0.5 = 28.0

		now := time.Now()
		content := &entity.Content{
			ID:          1,
			ContentType: entity.ContentTypeVideo,
			PublishedAt: now.Add(-24 * time.Hour), // 1 day old
			Stats: &entity.ContentStats{
				Views: 10000,
				Likes: 500,
			},
		}

		score, err := service.CalculateScore(content)
		assert.NoError(t, err)
		assert.NotNil(t, score)

		assert.Equal(t, 15.0, score.BaseScore)
		assert.Equal(t, 5.0, score.RecencyScore)
		assert.Equal(t, 0.5, score.EngagementScore)
		assert.Equal(t, 28.0, score.FinalScore)
	})

	t.Run("Should calculate correct score for old ARTICLE", func(t *testing.T) {
		// Old article (published > 90 days ago which gives 0 Recency)
		// ReadingTime: 5 -> 5 points
		// Reactions: 100 -> 2 points (100/50)
		// Base = 7
		// Type Weight = 1.0
		// Weighted Base = 7.0
		// Engagement = (100/5) * 5 = 20 * 5 = 100
		// Recency = 0
		// Final = 7 + 0 + 100 = 107

		now := time.Now()
		content := &entity.Content{
			ID:          2,
			ContentType: entity.ContentTypeArticle,
			PublishedAt: now.Add(-100 * 24 * time.Hour), // 100 days old
			Stats: &entity.ContentStats{
				ReadingTime: 5,
				Reactions:   100,
			},
		}

		score, err := service.CalculateScore(content)
		assert.NoError(t, err)
		assert.NotNil(t, score)

		assert.Equal(t, 7.0, score.BaseScore)
		assert.Equal(t, 0.0, score.RecencyScore)
		assert.Equal(t, 100.0, score.EngagementScore)
		assert.Equal(t, 107.0, score.FinalScore)
	})

	t.Run("Should handle zero division in engagement score", func(t *testing.T) {
		content := &entity.Content{
			ID:          3,
			ContentType: entity.ContentTypeVideo,
			PublishedAt: time.Now(),
			Stats: &entity.ContentStats{
				Views: 0,
				Likes: 10,
			},
		}

		score, err := service.CalculateScore(content)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, score.EngagementScore)
	})
}

func TestScoringService_RecencyScore(t *testing.T) {
	service := NewScoringService(ScoringRules{}) // use defaults
	// We need to access private method or test via public API. 
	// Since calculateRecencyScore is private, we test it via CalculateScore or make it public if needed.
	// But usually, we test behavior through public API. We already covered extremes in previous tests.
	// Let's add specific ranges via public API.

	tests := []struct {
		name          string
		daysOld       int
		expectedScore float64
	}{
		{"1 week old", 7, 5.0},
		{"1 month old", 20, 3.0},
		{"3 months old", 60, 1.0},
		{"Very old", 91, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := &entity.Content{
				Stats:       &entity.ContentStats{}, // valid stats needed
				PublishedAt: time.Now().Add(time.Duration(-tt.daysOld) * 24 * time.Hour),
			}
			score, _ := service.CalculateScore(content)
			assert.Equal(t, tt.expectedScore, score.RecencyScore)
		})
	}
}
