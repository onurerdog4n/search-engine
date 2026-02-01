package entity

import "time"

// ContentType içerik türünü temsil eder (video veya article)
type ContentType string

const (
	ContentTypeVideo   ContentType = "video"
	ContentTypeArticle ContentType = "article"
)

// Content ana içerik entity'si
type Content struct {
	ID                int64         `json:"id"`
	ProviderID        int64         `json:"provider_id"`
	ProviderContentID string        `json:"provider_content_id"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	ContentType       ContentType   `json:"content_type"`
	PublishedAt       time.Time     `json:"published_at"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	Stats             *ContentStats `json:"stats,omitempty"`
	Score             *ContentScore `json:"score,omitempty"`
	Tags              []Tag         `json:"tags,omitempty"`
	RelevanceScore    float64       `json:"relevance_score,omitempty"`
	RawData           string        `json:"raw_data,omitempty"` // Provider'dan gelen ham veri
	Deleted           bool          `json:"deleted"`
}

// ContentStats içerik istatistiklerini tutar
type ContentStats struct {
	ID          int64     `json:"id"`
	ContentID   int64     `json:"content_id"`
	Views       int64     `json:"views"`
	Likes       int32     `json:"likes"`
	ReadingTime int32     `json:"reading_time"` // dakika cinsinden
	Reactions   int32     `json:"reactions"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContentScore içerik skorlama bilgilerini tutar
type ContentScore struct {
	ID              int64     `json:"id"`
	ContentID       int64     `json:"content_id"`
	BaseScore       float64   `json:"base_score"`
	TypeWeight      float64   `json:"type_weight"`
	RecencyScore    float64   `json:"recency_score"`
	EngagementScore float64   `json:"engagement_score"`
	FinalScore      float64   `json:"final_score"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

// Tag içerik etiketlerini temsil eder
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Provider veri sağlayıcı bilgilerini tutar
type Provider struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Format    string    `json:"format"` // "json" veya "xml"
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderSyncLog senkronizasyon loglarını tutar
type ProviderSyncLog struct {
	ID           int64      `json:"id"`
	ProviderID   int64      `json:"provider_id"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Status       string     `json:"status"` // "success", "failed", "running"
	ItemsSynced  int32      `json:"items_synced"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// NormalizedContent provider'lardan gelen veriyi normalize edilmiş formatta tutar
type NormalizedContent struct {
	ExternalID  string       `json:"external_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	ContentType ContentType  `json:"content_type"`
	PublishedAt time.Time    `json:"published_at"`
	Stats       ContentStats `json:"stats"`
	Tags        []string     `json:"tags"`
	RawData     string       `json:"raw_data"`
}
