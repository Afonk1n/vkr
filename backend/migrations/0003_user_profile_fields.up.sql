ALTER TABLE users ADD COLUMN IF NOT EXISTS favorite_album_ids text NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS favorite_artists text NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS favorite_track_ids text NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferences_manual boolean NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified_artist boolean NOT NULL DEFAULT false;
