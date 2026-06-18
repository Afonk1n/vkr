DROP INDEX IF EXISTS idx_users_artist_name;
ALTER TABLE users DROP COLUMN IF EXISTS artist_name;
