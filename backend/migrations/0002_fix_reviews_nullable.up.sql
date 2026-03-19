-- 0002_fix_reviews_nullable.up.sql
-- Ensure album_id and track_id in reviews are nullable (align with models and fixReviewsTableConstraints)

ALTER TABLE reviews ALTER COLUMN album_id DROP NOT NULL;
ALTER TABLE reviews ALTER COLUMN track_id DROP NOT NULL;

