ALTER TABLE contents ADD COLUMN IF NOT EXISTS deleted INTEGER DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_contents_deleted ON contents(deleted);
