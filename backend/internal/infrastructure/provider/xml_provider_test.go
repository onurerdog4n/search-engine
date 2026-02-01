package provider

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestXMLProvider_Normalize(t *testing.T) {
	// Setup
	prov := &entity.Provider{ID: 2, Name: "XML Provider"}
	p := &xmlProvider{provider: prov}

	t.Run("Should correctly normalize valid XML article content", func(t *testing.T) {
		raw := XMLItem{
			ID:    "article-123",
			Title: "Go Tips",
			Type:  "article",
			Stats: XMLStats{
				Views:       2000,
				Likes:       100,
				ReadingTime: 5,
				Reactions:   50,
			},
			PubDate: "2024-01-01T15:30:00Z",
		}
		
		// Simulate raw data
		rawDataBytes, _ := xml.Marshal(raw)
		rawData := string(rawDataBytes)

		normalized, err := p.normalize(raw, rawData)
		assert.NoError(t, err)
		assert.NotNil(t, normalized)

		assert.Equal(t, "article-123", normalized.ExternalID)
		assert.Equal(t, "Go Tips", normalized.Title)
		assert.Equal(t, entity.ContentTypeArticle, normalized.ContentType)
		assert.Equal(t, int32(5), normalized.Stats.ReadingTime)
		assert.Equal(t, int32(50), normalized.Stats.Reactions)
		expectedTime, _ := time.Parse(time.RFC3339, "2024-01-01T15:30:00Z")
		assert.Equal(t, expectedTime, normalized.PublishedAt)
		assert.Equal(t, rawData, normalized.RawData) // Verify RawData
	})

	t.Run("Should return error for missing ID", func(t *testing.T) {
		raw := XMLItem{
			ID:      "",
			Title:   "No ID",
			Type:    "article",
			PubDate: "2024-01-01T15:30:00Z",
		}
		
		rawData := `<item><title>No ID</title></item>`

		_, err := p.normalize(raw, rawData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID eksik")
	})
}
