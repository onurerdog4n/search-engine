package service

import (
	"math"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
)

// ScoringService skorlama işlemlerini yönetir
type ScoringService interface {
	CalculateScore(content *entity.Content) (*entity.ContentScore, error)
}

// scoringService ScoringService interface'inin implementasyonu
type scoringService struct {
	rules ScoringRules
}

// ScoringRules skorlama kurallarını tutar
type ScoringRules struct {
	VideoTypeWeight   float64 // Video içerikler için katsayı (varsayılan: 1.5)
	ArticleTypeWeight float64 // Makale içerikler için katsayı (varsayılan: 1.0)
}

// NewScoringService yeni bir ScoringService oluşturur
func NewScoringService(rules ScoringRules) ScoringService {
	// Varsayılan değerleri ayarla
	if rules.VideoTypeWeight == 0 {
		rules.VideoTypeWeight = 1.5
	}
	if rules.ArticleTypeWeight == 0 {
		rules.ArticleTypeWeight = 1.0
	}

	return &scoringService{
		rules: rules,
	}
}

// CalculateScore içerik için skor hesaplar
// Formül: (BaseScore × TypeWeight) + RecencyScore + EngagementScore
func (s *scoringService) CalculateScore(content *entity.Content) (*entity.ContentScore, error) {
	if content.Stats == nil {
		return nil, nil
	}

	score := &entity.ContentScore{
		ContentID:    content.ID,
		CalculatedAt: time.Now(),
	}

	// Base score hesaplama
	if content.ContentType == entity.ContentTypeVideo {
		// Video için: views/1000 + likes/100
		score.BaseScore = float64(content.Stats.Views)/1000.0 + float64(content.Stats.Likes)/100.0
		score.TypeWeight = s.rules.VideoTypeWeight
	} else {
		// Makale için: reading_time + reactions/50
		score.BaseScore = float64(content.Stats.ReadingTime) + float64(content.Stats.Reactions)/50.0
		score.TypeWeight = s.rules.ArticleTypeWeight
	}

	// Güncellik skoru hesaplama
	score.RecencyScore = s.calculateRecencyScore(content.PublishedAt)

	// Etkileşim skoru hesaplama
	score.EngagementScore = s.calculateEngagementScore(content)

	// Final skor hesaplama
	score.FinalScore = (score.BaseScore * score.TypeWeight) + score.RecencyScore + score.EngagementScore

	// Skorları 2 ondalık basamağa yuvarla
	score.BaseScore = math.Round(score.BaseScore*100) / 100
	score.RecencyScore = math.Round(score.RecencyScore*100) / 100
	score.EngagementScore = math.Round(score.EngagementScore*100) / 100
	score.FinalScore = math.Round(score.FinalScore*100) / 100

	return score, nil
}

// calculateRecencyScore yayın tarihine göre güncellik skoru hesaplar
// 1 hafta içinde: +5
// 1 ay içinde: +3
// 3 ay içinde: +1
// Daha eski: 0
func (s *scoringService) calculateRecencyScore(publishedAt time.Time) float64 {
	now := time.Now()
	duration := now.Sub(publishedAt)

	switch {
	case duration <= 7*24*time.Hour: // 1 hafta
		return 5.0
	case duration <= 30*24*time.Hour: // 1 ay
		return 3.0
	case duration <= 90*24*time.Hour: // 3 ay
		return 1.0
	default:
		return 0.0
	}
}

// calculateEngagementScore içerik türüne göre etkileşim skoru hesaplar
// Video için: (likes/views) × 10
// Makale için: (reactions/reading_time) × 5
func (s *scoringService) calculateEngagementScore(content *entity.Content) float64 {
	if content.Stats == nil {
		return 0.0
	}

	if content.ContentType == entity.ContentTypeVideo {
		// Video için etkileşim
		if content.Stats.Views == 0 {
			return 0.0
		}
		return (float64(content.Stats.Likes) / float64(content.Stats.Views)) * 10.0
	} else {
		// Makale için etkileşim
		if content.Stats.ReadingTime == 0 {
			return 0.0
		}
		return (float64(content.Stats.Reactions) / float64(content.Stats.ReadingTime)) * 5.0
	}
}
