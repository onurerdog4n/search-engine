package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
	"golang.org/x/time/rate"
)

// jsonProvider JSON formatındaki provider client implementasyonu
type jsonProvider struct {
	provider *entity.Provider
	apiURL   string
	limiter  *rate.Limiter
}

// JSONContent JSON dosyasındaki içerik yapısı
type JSONContent struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Metrics     JSONMetrics `json:"metrics"`
	PublishedAt string      `json:"published_at"`
	Tags        []string    `json:"tags"`
}

// JSONMetrics JSON'daki metrics yapısı
type JSONMetrics struct {
	Views       int64  `json:"views"`
	Likes       int32  `json:"likes"`
	Duration    string `json:"duration,omitempty"`     // Video için
	ReadingTime int32  `json:"reading_time,omitempty"` // Article için
	Reactions   int32  `json:"reactions,omitempty"`    // Article için
}

// JSONResponse JSON dosyasının root yapısı
type JSONResponse struct {
	Contents   []JSONContent `json:"contents"`
	Pagination struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"pagination"`
}

// NewJSONProvider yeni bir JSON provider client oluşturur
func NewJSONProvider(provider *entity.Provider, apiURL string) port.ProviderClient {
	// Rate Limiter: Saniyede 1 istek (Burst 1)
	return &jsonProvider{
		provider: provider,
		apiURL:   apiURL,
		limiter:  rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// FetchContents Mock API'den içerikleri sayfalar halinde çeker ve normalize eder
func (p *jsonProvider) FetchContents(ctx context.Context) ([]*entity.NormalizedContent, error) {
	var allNormalized []*entity.NormalizedContent
	var page int = 1
	var totalPages int = 1 // En az bir sayfa var varsayıyoruz

	for page <= totalPages {
		// Rate Limiter bekleme
		if err := p.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter hatası: %w", err)
		}

		// Mock API'den sayfayı çek
		var resp *http.Response
		var err error
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			url := fmt.Sprintf("%s?page=%d", p.apiURL, page)
			resp, err = http.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			log.Printf("JSON API retry %d/%d (Page %d): %v", i+1, maxRetries, page, err)
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff benzeri
		}

		if err != nil {
			return nil, fmt.Errorf("JSON API isteği başarısız: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("JSON API hata döndü: %d", resp.StatusCode)
		}

		// Body'i oku (Raw Data için)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("response body okuma hatası: %w", err)
		}

		// JSON parse et
		var response JSONResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			return nil, fmt.Errorf("JSON parse hatası: %w", err)
		}

		// İlk sayfada toplam sayfa sayısını hesapla
		if page == 1 && response.Pagination.PerPage > 0 {
			totalPages = (response.Pagination.Total + response.Pagination.PerPage - 1) / response.Pagination.PerPage
		}

		if len(allNormalized) >= response.Pagination.Total || len(allNormalized) >= 1000 { // Güvenlik sınırı 1000
			break
		}

		// Normalize et
		for _, raw := range response.Contents {
			// Bu içerik için raw datayı tekrar marshal ediyoruz (bireysel raw data saklamak istiyorsak)
			// Veya tüm sayfanın raw datasını saklayamayız çünkü context item bazlı. 
			// En doğrusu item'a ait raw datayı saklamak.
			itemRawBytes, _ := json.Marshal(raw)
			
			content, err := p.normalize(raw, string(itemRawBytes))
			if err != nil {
				continue
			}
			allNormalized = append(allNormalized, content)
		}

		page++
	}

	return allNormalized, nil
}

// GetProviderInfo provider bilgilerini döner
func (p *jsonProvider) GetProviderInfo() *entity.Provider {
	return p.provider
}

// normalize JSON içeriğini NormalizedContent'e dönüştürür
func (p *jsonProvider) normalize(raw JSONContent, rawData string) (*entity.NormalizedContent, error) {
	// Tarih parse et
	publishedAt, err := time.Parse(time.RFC3339, raw.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("tarih parse hatası: %w", err)
	}

	// İçerik türünü belirle
	var contentType entity.ContentType
	switch raw.Type {
	case "video":
		contentType = entity.ContentTypeVideo
	case "article":
		contentType = entity.ContentTypeArticle
	default:
		return nil, fmt.Errorf("geçersiz içerik türü: %s", raw.Type)
	}

	// Normalize et
	return &entity.NormalizedContent{
		ExternalID:  raw.ID,
		Title:       raw.Title,
		Description: "", // JSON'da description yok
		ContentType: contentType,
		PublishedAt: publishedAt,
		Stats: entity.ContentStats{
			Views:       raw.Metrics.Views,
			Likes:       raw.Metrics.Likes,
			ReadingTime: raw.Metrics.ReadingTime,
			Reactions:   raw.Metrics.Reactions,
		},
		Tags:    raw.Tags,
		RawData: rawData,
	}, nil
}
