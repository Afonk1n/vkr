ALTER TABLE users ADD COLUMN IF NOT EXISTS artist_name text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_users_artist_name ON users (artist_name);
