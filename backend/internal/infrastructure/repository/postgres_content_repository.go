package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// postgresContentRepository PostgreSQL ile ContentRepository implementasyonu
type postgresContentRepository struct {
	db *sql.DB
}

// NewPostgresContentRepository yeni bir PostgreSQL content repository oluşturur
func NewPostgresContentRepository(db *sql.DB) port.ContentRepository {
	return &postgresContentRepository{db: db}
}

// Create yeni bir içerik oluşturur
func (r *postgresContentRepository) Create(ctx context.Context, content *entity.Content) error {
	query := `
		INSERT INTO contents (provider_id, provider_content_id, title, description, content_type, published_at, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		content.ProviderID,
		content.ProviderContentID,
		content.Title,
		content.Description,
		content.ContentType,
		content.PublishedAt,
		content.RawData,
	).Scan(&content.ID, &content.CreatedAt, &content.UpdatedAt)

	return err
}

// Update mevcut bir içeriği günceller
func (r *postgresContentRepository) Update(ctx context.Context, content *entity.Content) error {
	query := `
		UPDATE contents
		SET title = $1, description = $2, content_type = $3, published_at = $4, raw_data = $5
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		content.Title,
		content.Description,
		content.ContentType,
		content.PublishedAt,
		content.RawData,
		content.ID,
	).Scan(&content.UpdatedAt)

	return err
}

// FindByID ID'ye göre içerik getirir
func (r *postgresContentRepository) FindByID(ctx context.Context, id int64) (*entity.Content, error) {
	query := `
		SELECT 
			c.id, c.provider_id, c.provider_content_id, c.title, c.description,
			c.content_type, c.published_at, c.created_at, c.updated_at, c.raw_data,
			cs.id, cs.views, cs.likes, cs.reading_time, cs.reactions, cs.updated_at,
			csc.id, csc.base_score, csc.type_weight, csc.recency_score, 
			csc.engagement_score, csc.final_score, csc.calculated_at
		FROM contents c
		LEFT JOIN content_stats cs ON c.id = cs.content_id
		LEFT JOIN content_scores csc ON c.id = csc.content_id
		WHERE c.id = $1 AND c.deleted = 0
	`

	content := &entity.Content{
		Stats: &entity.ContentStats{},
		Score: &entity.ContentScore{},
	}

	var statsID, scoreID sql.NullInt64
	var statsUpdatedAt, scoreCalculatedAt sql.NullTime
	var rawData sql.NullString
	
	// Stats fields - can be NULL
	var views sql.NullInt64
	var likes sql.NullInt32
	var readingTime sql.NullInt32
	var reactions sql.NullInt32
	
	// Score fields - can be NULL
	var baseScore, typeWeight, recencyScore, engagementScore, finalScore sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&content.ID, &content.ProviderID, &content.ProviderContentID,
		&content.Title, &content.Description, &content.ContentType,
		&content.PublishedAt, &content.CreatedAt, &content.UpdatedAt, &rawData,
		&statsID, &views, &likes, &readingTime, &reactions, &statsUpdatedAt,
		&scoreID, &baseScore, &typeWeight, &recencyScore, &engagementScore,
		&finalScore, &scoreCalculatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to find content: %w", err)
	}

	// Handle raw data
	if rawData.Valid {
		content.RawData = rawData.String
	}

	// Handle stats - only set if exists
	if statsID.Valid {
		content.Stats.ID = statsID.Int64
		content.Stats.ContentID = content.ID
		content.Stats.Views = views.Int64
		content.Stats.Likes = int32(likes.Int32)
		content.Stats.ReadingTime = int32(readingTime.Int32)
		content.Stats.Reactions = int32(reactions.Int32)
		if statsUpdatedAt.Valid {
			content.Stats.UpdatedAt = statsUpdatedAt.Time
		}
	} else {
		content.Stats = nil
	}

	// Handle score - only set if exists
	if scoreID.Valid {
		content.Score.ID = scoreID.Int64
		content.Score.ContentID = content.ID
		content.Score.BaseScore = baseScore.Float64
		content.Score.TypeWeight = typeWeight.Float64
		content.Score.RecencyScore = recencyScore.Float64
		content.Score.EngagementScore = engagementScore.Float64
		content.Score.FinalScore = finalScore.Float64
		if scoreCalculatedAt.Valid {
			content.Score.CalculatedAt = scoreCalculatedAt.Time
		}
	} else {
		content.Score = nil
	}

	// Tag'leri yükle
	tags, err := r.loadTags(ctx, content.ID)
	if err == nil {
		content.Tags = tags
	}

	return content, nil
}

// Upsert içerik varsa günceller, yoksa ekler
func (r *postgresContentRepository) Upsert(ctx context.Context, content *entity.Content) error {
	query := `
		INSERT INTO contents (provider_id, provider_content_id, title, description, content_type, published_at, raw_data, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0)
		ON CONFLICT (provider_id, provider_content_id)
		DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			content_type = EXCLUDED.content_type,
			published_at = EXCLUDED.published_at,
			raw_data = EXCLUDED.raw_data,
			deleted = 0
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		content.ProviderID,
		content.ProviderContentID,
		content.Title,
		content.Description,
		content.ContentType,
		content.PublishedAt,
		content.RawData,
	).Scan(&content.ID, &content.CreatedAt, &content.UpdatedAt)

	return err
}

// Search arama parametrelerine göre içerikleri getirir
func (r *postgresContentRepository) Search(ctx context.Context, params port.SearchParams) ([]*entity.Content, int64, error) {
	// Arama kısmını oluştur (FROM + JOIN'ler)
	fromParts := `
		FROM contents c
		LEFT JOIN content_stats cs ON c.id = cs.content_id
		LEFT JOIN content_scores csc ON c.id = csc.content_id
		WHERE c.deleted = 0
	`

	// FTS (Full-Text Search) için vector ve query tanımla
	var searchVector string
	var searchQuery string

	// Başlık (A) ve Tagler (B) ağırlıklı vector oluştur
	searchVector = `(
		setweight(to_tsvector('english', COALESCE(c.title, '')), 'A') ||
		setweight(to_tsvector('english', COALESCE((
			SELECT string_agg(t.name, ' ') 
			FROM content_tags ct 
			JOIN tags t ON ct.tag_id = t.id 
			WHERE ct.content_id = c.id
		), '')), 'B')
	)`

	var args []interface{}
	argCount := 0

	// Arama sorgusunu FTS formatına getir (Prefix matching için :* ekle)
	whereClause := ""
	if params.Query != "" {
		argCount++
		// Özel karakterleri temizle (syntax hatasını önlemek için)
		cleaner := func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}

		words := strings.Fields(params.Query)
		var ftsWords []string
		for _, w := range words {
			cleanWord := strings.Map(cleaner, w)
			if cleanWord != "" {
				ftsWords = append(ftsWords, cleanWord+":*")
			}
		}

		if len(ftsWords) > 0 {
			searchQuery = strings.Join(ftsWords, " & ")
			args = append(args, searchQuery)
			whereClause += fmt.Sprintf(" AND %s @@ to_tsquery('english', $%d)", searchVector, argCount)
		} else {
			// Eğer tüm kelimeler temizlendiyse query'yi boşalt
			params.Query = ""
			argCount--
		}
	}

	// İçerik türü filtresi
	if params.ContentType != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND c.content_type = $%d", argCount)
		args = append(args, params.ContentType)
	}

	// Alakalılık (relevance) skorunu hesapla
	relevanceExpr := "0.0"
	if params.Query != "" {
		// ts_rank_cd (Cover Density) kullanarak kelime yoğunluğuna göre puanlıyoruz
		// {D-weight, C-weight, B-weight, A-weight} -> {0.1, 0.2, 0.5, 1.0}
		// A (Title) = 1.0, B (Tags) = 0.2 olarak ağırlıklandırıyoruz
		relevanceExpr = fmt.Sprintf("ts_rank_cd('{0.1, 0.2, 0.4, 1.0}', %s, to_tsquery('english', $1))", searchVector)
	}

	// Toplam kayıt sayısını al
	countQuery := "SELECT COUNT(*) " + fromParts + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sıralama
	orderBy := " ORDER BY "
	if params.SortBy == "relevance" && params.Query != "" {
		orderBy += "relevance_score DESC, c.published_at DESC"
	} else {
		// Varsayılan: popularity
		orderBy += "csc.final_score DESC NULLS LAST, c.published_at DESC"
	}

	// Pagination
	argCount++
	limit := params.PageSize
	offset := (params.Page - 1) * params.PageSize
	pagination := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	// Ana query
	selectQuery := fmt.Sprintf(`
		SELECT 
			c.id, c.provider_id, c.provider_content_id, c.title, c.description,
			c.content_type, c.published_at, c.created_at, c.updated_at, c.raw_data,
			cs.id, cs.views, cs.likes, cs.reading_time, cs.reactions, cs.updated_at,
			csc.id, csc.base_score, csc.type_weight, csc.recency_score,
			csc.engagement_score, csc.final_score, csc.calculated_at,
			%s as relevance_score
	`, relevanceExpr) + fromParts + whereClause + orderBy + pagination

	// Arama logu (debug için)
	log.Printf("Arama yapılıyor: Query=%s, Sort=%s, Page=%d", params.Query, params.SortBy, params.Page)
	// log.Printf("SQL: %s", selectQuery)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contents []*entity.Content
	for rows.Next() {
		content := &entity.Content{
			Stats: &entity.ContentStats{},
			Score: &entity.ContentScore{},
		}

		var statsID, scoreID sql.NullInt64
		var statsUpdatedAt, scoreCalculatedAt sql.NullTime
		var relevanceScore float64
		var rawData sql.NullString

		err := rows.Scan(
			&content.ID, &content.ProviderID, &content.ProviderContentID,
			&content.Title, &content.Description, &content.ContentType,
			&content.PublishedAt, &content.CreatedAt, &content.UpdatedAt, &rawData,
			&statsID, &content.Stats.Views, &content.Stats.Likes,
			&content.Stats.ReadingTime, &content.Stats.Reactions, &statsUpdatedAt,
			&scoreID, &content.Score.BaseScore, &content.Score.TypeWeight,
			&content.Score.RecencyScore, &content.Score.EngagementScore,
			&content.Score.FinalScore, &scoreCalculatedAt,
			&relevanceScore,
		)
		if err != nil {
			return nil, 0, err
		}

		content.RelevanceScore = relevanceScore
		if rawData.Valid {
			content.RawData = rawData.String
		}

		// Stats ve Score null kontrolü
		if !statsID.Valid {
			content.Stats = nil
		} else {
			content.Stats.ID = statsID.Int64
			content.Stats.ContentID = content.ID
			if statsUpdatedAt.Valid {
				content.Stats.UpdatedAt = statsUpdatedAt.Time
			}
		}

		if !scoreID.Valid {
			content.Score = nil
		} else {
			content.Score.ID = scoreID.Int64
			content.Score.ContentID = content.ID
			if scoreCalculatedAt.Valid {
				content.Score.CalculatedAt = scoreCalculatedAt.Time
			}
		}

		// Tag'leri yükle
		tags, err := r.loadTags(ctx, content.ID)
		if err == nil {
			content.Tags = tags
		}

		contents = append(contents, content)
	}

	return contents, total, rows.Err()
}

// CreateOrUpdateStats içerik istatistiklerini oluşturur veya günceller
func (r *postgresContentRepository) CreateOrUpdateStats(ctx context.Context, stats *entity.ContentStats) error {
	query := `
		INSERT INTO content_stats (content_id, views, likes, reading_time, reactions)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (content_id)
		DO UPDATE SET
			views = EXCLUDED.views,
			likes = EXCLUDED.likes,
			reading_time = EXCLUDED.reading_time,
			reactions = EXCLUDED.reactions
		RETURNING id, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		stats.ContentID,
		stats.Views,
		stats.Likes,
		stats.ReadingTime,
		stats.Reactions,
	).Scan(&stats.ID, &stats.UpdatedAt)

	return err
}

// CreateOrUpdateScore içerik skorunu oluşturur veya günceller
func (r *postgresContentRepository) CreateOrUpdateScore(ctx context.Context, score *entity.ContentScore) error {
	query := `
		INSERT INTO content_scores (content_id, base_score, type_weight, recency_score, engagement_score, final_score)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (content_id)
		DO UPDATE SET
			base_score = EXCLUDED.base_score,
			type_weight = EXCLUDED.type_weight,
			recency_score = EXCLUDED.recency_score,
			engagement_score = EXCLUDED.engagement_score,
			final_score = EXCLUDED.final_score,
			calculated_at = CURRENT_TIMESTAMP
		RETURNING id, calculated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		score.ContentID,
		score.BaseScore,
		score.TypeWeight,
		score.RecencyScore,
		score.EngagementScore,
		score.FinalScore,
	).Scan(&score.ID, &score.CalculatedAt)

	return err
}

// AddTags içeriğe etiketler ekler
func (r *postgresContentRepository) AddTags(ctx context.Context, contentID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Her tag için
	for _, tagName := range tags {
		// Tag'i oluştur veya mevcut olanı al
		var tagID int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO tags (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, strings.ToLower(strings.TrimSpace(tagName))).Scan(&tagID)
		if err != nil {
			return err
		}

		// Content-tag ilişkisini oluştur
		_, err = tx.ExecContext(ctx, `
			INSERT INTO content_tags (content_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, contentID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// MarkStaleContentsAsDeleted güncellenmeyen içerikleri silinmiş olarak işaretler
func (r *postgresContentRepository) MarkStaleContentsAsDeleted(ctx context.Context, providerID int64, threshold time.Time) error {
	query := `
		UPDATE contents 
		SET deleted = 1, updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $1 AND updated_at < $2 AND deleted = 0
	`
	
	result, err := r.db.ExecContext(ctx, query, providerID, threshold)
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("%d stale contents marked as deleted for provider %d", rowsAffected, providerID)
	}
	
	return nil
}

// loadTags içeriğin tag'lerini yükler (yardımcı fonksiyon)
func (r *postgresContentRepository) loadTags(ctx context.Context, contentID int64) ([]entity.Tag, error) {
	query := `
		SELECT t.id, t.name, t.created_at
		FROM tags t
		INNER JOIN content_tags ct ON t.id = ct.tag_id
		WHERE ct.content_id = $1
		ORDER BY t.name
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []entity.Tag
	for rows.Next() {
		var tag entity.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}
