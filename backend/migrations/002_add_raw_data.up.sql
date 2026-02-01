-- contents tablosuna raw_data sütunu ekle
ALTER TABLE contents ADD COLUMN IF NOT EXISTS raw_data TEXT;
