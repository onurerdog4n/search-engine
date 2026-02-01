package provider

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestJSONProvider_Normalize(t *testing.T) {
	// Setup
	prov := &entity.Provider{ID: 1, Name: "Test Provider"}
	p := &jsonProvider{provider: prov}

	t.Run("Should correctly normalize valid JSON video content", func(t *testing.T) {
		raw := JSONContent{
			ID:    "video-123",
			Title: "Go Tutorial",
			Type:  "video",
			Metrics: JSONMetrics{
				Views:    1000,
				Likes:    500,
				Duration: "10m",
			},
			PublishedAt: "2024-01-01T12:00:00Z",
			Tags:        []string{"go", "tutorial"},
		}
		
		// Simulate raw data string
		rawDataBytes, _ := json.Marshal(raw)
		rawData := string(rawDataBytes)

		normalized, err := p.normalize(raw, rawData)
		assert.NoError(t, err)
		assert.NotNil(t, normalized)

		assert.Equal(t, "video-123", normalized.ExternalID)
		assert.Equal(t, "Go Tutorial", normalized.Title)
		assert.Equal(t, entity.ContentTypeVideo, normalized.ContentType)
		assert.Equal(t, int64(1000), normalized.Stats.Views)
		assert.Equal(t, int32(500), normalized.Stats.Likes)
		expectedTime, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
		assert.Equal(t, expectedTime, normalized.PublishedAt)
		assert.Equal(t, rawData, normalized.RawData) // Verify RawData storage
	})

	t.Run("Should return error for invalid date format", func(t *testing.T) {
		raw := JSONContent{
			ID:          "video-123",
			Title:       "Test",
			Type:        "video",
			PublishedAt: "invalid-date",
		}
		
		rawData := `{"id":"video-123"}`

		_, err := p.normalize(raw, rawData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tarih parse hatası")
	})

	t.Run("Should return error for unknown content type", func(t *testing.T) {
		raw := JSONContent{
			ID:          "123",
			Title:       "Test",
			Type:        "unknown",
			PublishedAt: "2024-01-01T12:00:00Z",
		}
		
		rawData := `{"type":"unknown"}`

		_, err := p.normalize(raw, rawData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "geçersiz içerik türü")
	})
}
