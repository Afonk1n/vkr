-- 0001_init_schema.down.sql
-- Drop initial schema (reverse order of creation)

DROP TABLE IF EXISTS album_likes;
DROP TABLE IF EXISTS track_likes;
DROP TABLE IF EXISTS review_likes;
DROP TABLE IF EXISTS reviews;
DROP TYPE IF EXISTS review_status;
DROP TABLE IF EXISTS track_genres;
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS albums;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS users;

