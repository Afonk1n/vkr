-- 0002_fix_reviews_nullable.down.sql
-- Optional rollback: make album_id and track_id NOT NULL again (use with caution)

ALTER TABLE reviews ALTER COLUMN album_id SET NOT NULL;
ALTER TABLE reviews ALTER COLUMN track_id SET NOT NULL;

