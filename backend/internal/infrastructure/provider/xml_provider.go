package provider

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
	"golang.org/x/time/rate"
)

// xmlProvider XML formatındaki provider client implementasyonu
type xmlProvider struct {
	provider *entity.Provider
	apiURL   string
	limiter  *rate.Limiter
}

// XMLItem XML dosyasındaki içerik yapısı
type XMLItem struct {
	ID         string   `xml:"id"`
	Title      string   `xml:"headline"`
	Type       string   `xml:"type"`
	Stats      XMLStats `xml:"stats"`
	PubDate    string   `xml:"publication_date"`
	Categories struct {
		Category []string `xml:"category"`
	} `xml:"categories"`
}

// XMLStats XML'deki stats yapısı
type XMLStats struct {
	Views       int64 `xml:"views"`
	Likes       int32 `xml:"likes"`
	ReadingTime int32 `xml:"reading_time"` // Article için
	Reactions   int32 `xml:"reactions"`    // Article için
}

// XMLResponse XML dosyasının root yapısı
type XMLResponse struct {
	XMLName xml.Name `xml:"feed"`
	Items   struct {
		Items []XMLItem `xml:"item"`
	} `xml:"items"`
	Meta struct {
		Total   int `xml:"total_count"`
		Page    int `xml:"current_page"`
		PerPage int `xml:"items_per_page"`
	} `xml:"meta"`
}

// NewXMLProvider yeni bir XML provider client oluşturur
func NewXMLProvider(provider *entity.Provider, apiURL string) port.ProviderClient {
	// Rate Limiter: Saniyede 1 istek (Burst 1)
	return &xmlProvider{
		provider: provider,
		apiURL:   apiURL,
		limiter:  rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// FetchContents Mock API'den içerikleri sayfalar halinde çeker ve normalize eder
func (p *xmlProvider) FetchContents(ctx context.Context) ([]*entity.NormalizedContent, error) {
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
			log.Printf("XML API retry %d/%d (Page %d): %v", i+1, maxRetries, page, err)
			time.Sleep(time.Second * time.Duration(i+1))
		}

		if err != nil {
			return nil, fmt.Errorf("XML API isteği başarısız: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("XML API hata döndü: %d", resp.StatusCode)
		}

		// Body'i oku (Raw Data için)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("response body okuma hatası: %w", err)
		}

		// XML parse et
		var response XMLResponse
		if err := xml.Unmarshal(bodyBytes, &response); err != nil {
			return nil, fmt.Errorf("XML parse hatası: %w", err)
		}

		// İlk sayfada toplam sayfa sayısını hesapla
		if page == 1 && response.Meta.PerPage > 0 {
			totalPages = (response.Meta.Total + response.Meta.PerPage - 1) / response.Meta.PerPage
		}

		if len(allNormalized) >= response.Meta.Total || len(allNormalized) >= 1000 {
			break
		}

		// Normalize et
		for _, raw := range response.Items.Items {
			// Item'a özel raw datayı elde etmek için tekrar marshal ediyoruz
			itemRawBytes, _ := xml.Marshal(raw)

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
func (p *xmlProvider) GetProviderInfo() *entity.Provider {
	return p.provider
}

// normalize XML içeriğini NormalizedContent'e dönüştürür
func (p *xmlProvider) normalize(raw XMLItem, rawData string) (*entity.NormalizedContent, error) {
	if raw.ID == "" {
		return nil, fmt.Errorf("ID eksik")
	}

	// Tarih parse et
	publishedAt, err := time.Parse(time.RFC3339, raw.PubDate)
	if err != nil {
		// YYYY-MM-DD formatını dene
		publishedAt, err = time.Parse("2006-01-02", raw.PubDate)
		if err != nil {
			return nil, fmt.Errorf("tarih parse hatası (%s): %w", raw.PubDate, err)
		}
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
		Description: "",
		ContentType: contentType,
		PublishedAt: publishedAt,
		Stats: entity.ContentStats{
			Views:       raw.Stats.Views,
			Likes:       raw.Stats.Likes,
			ReadingTime: raw.Stats.ReadingTime,
			Reactions:   raw.Stats.Reactions,
		},
		Tags:    raw.Categories.Category,
		RawData: rawData,
	}, nil
}
