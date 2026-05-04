ALTER TABLE users DROP COLUMN IF EXISTS favorite_album_ids;
ALTER TABLE users DROP COLUMN IF EXISTS favorite_artists;
ALTER TABLE users DROP COLUMN IF EXISTS favorite_track_ids;
ALTER TABLE users DROP COLUMN IF EXISTS preferences_manual;
ALTER TABLE users DROP COLUMN IF EXISTS is_verified_artist;
