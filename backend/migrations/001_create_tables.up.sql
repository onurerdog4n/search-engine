-- Provider tablosu: Veri sağlayıcı bilgilerini tutar
CREATE TABLE IF NOT EXISTS providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    format VARCHAR(10) NOT NULL CHECK (format IN ('json', 'xml')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Contents tablosu: Ana içerik bilgilerini tutar
CREATE TABLE IF NOT EXISTS contents (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    provider_content_id VARCHAR(100) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('video', 'article')),
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, provider_content_id)
);

-- Content stats tablosu: İçerik istatistiklerini tutar
CREATE TABLE IF NOT EXISTS content_stats (
    id SERIAL PRIMARY KEY,
    content_id INTEGER NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    views BIGINT DEFAULT 0,
    likes INTEGER DEFAULT 0,
    reading_time INTEGER DEFAULT 0,
    reactions INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(content_id)
);

-- Content scores tablosu: İçerik skorlarını tutar
CREATE TABLE IF NOT EXISTS content_scores (
    id SERIAL PRIMARY KEY,
    content_id INTEGER NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    base_score DECIMAL(10,2) DEFAULT 0,
    type_weight DECIMAL(5,2) DEFAULT 0,
    recency_score DECIMAL(5,2) DEFAULT 0,
    engagement_score DECIMAL(10,2) DEFAULT 0,
    final_score DECIMAL(10,2) DEFAULT 0,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(content_id)
);

-- Tags tablosu: Etiketleri tutar
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Content tags tablosu: İçerik-etiket ilişkisini tutar (many-to-many)
CREATE TABLE IF NOT EXISTS content_tags (
    content_id INTEGER NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(content_id, tag_id)
);

-- Provider sync logs tablosu: Senkronizasyon loglarını tutar
CREATE TABLE IF NOT EXISTS provider_sync_logs (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'running')),
    items_synced INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Scoring rules tablosu: Skorlama kurallarını JSON formatında tutar
CREATE TABLE IF NOT EXISTS scoring_rules (
    id SERIAL PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL UNIQUE,
    rule_value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- İndeksler: Performans optimizasyonu için

-- Contents tablosu indeksleri
CREATE INDEX IF NOT EXISTS idx_contents_title ON contents USING GIN (to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_contents_title_pattern ON contents (title text_pattern_ops);
CREATE INDEX IF NOT EXISTS idx_contents_type ON contents(content_type);
CREATE INDEX IF NOT EXISTS idx_contents_published ON contents(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_contents_provider ON contents(provider_id);

-- Content stats indeksleri
CREATE INDEX IF NOT EXISTS idx_stats_views ON content_stats(views DESC);
CREATE INDEX IF NOT EXISTS idx_stats_content_id ON content_stats(content_id);

-- Content scores indeksleri
CREATE INDEX IF NOT EXISTS idx_scores_final ON content_scores(final_score DESC);
CREATE INDEX IF NOT EXISTS idx_scores_content_id ON content_scores(content_id);

-- Tags indeksleri
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- Content tags indeksleri
CREATE INDEX IF NOT EXISTS idx_content_tags_content ON content_tags(content_id);
CREATE INDEX IF NOT EXISTS idx_content_tags_tag ON content_tags(tag_id);

-- Provider sync logs indeksleri
CREATE INDEX IF NOT EXISTS idx_sync_logs_provider ON provider_sync_logs(provider_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_started ON provider_sync_logs(started_at DESC);

-- Trigger: updated_at otomatik güncelleme
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_providers_updated_at BEFORE UPDATE ON providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_contents_updated_at BEFORE UPDATE ON contents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_content_stats_updated_at BEFORE UPDATE ON content_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Başlangıç verileri: Provider'ları ekle
INSERT INTO providers (name, url, format, is_active) VALUES
    ('Provider 1 (JSON)', 'http://mock-api:8081/provider-1', 'json', true),
    ('Provider 2 (XML)', 'http://mock-api:8081/provider-2', 'xml', true)
ON CONFLICT DO NOTHING;

-- Başlangıç verileri: Skorlama kuralları
INSERT INTO scoring_rules (rule_name, rule_value, description) VALUES
    ('video_type_weight', '1.5', 'Video içerikler için tür katsayısı'),
    ('article_type_weight', '1.0', 'Makale içerikler için tür katsayısı')
ON CONFLICT (rule_name) DO NOTHING;
