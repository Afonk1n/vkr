-- 0001_init_schema.up.sql
-- Initial schema for music review site

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    avatar_path TEXT,
    bio TEXT,
    social_links JSONB,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);

CREATE TABLE genres (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_genres_deleted_at ON genres (deleted_at);

CREATE TABLE albums (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    genre_id INTEGER NOT NULL REFERENCES genres(id),
    cover_image_path TEXT,
    release_date DATE,
    description TEXT,
    average_rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_albums_genre_id ON albums (genre_id);
CREATE INDEX idx_albums_deleted_at ON albums (deleted_at);

CREATE TABLE tracks (
    id SERIAL PRIMARY KEY,
    album_id INTEGER NOT NULL REFERENCES albums(id),
    title TEXT NOT NULL,
    duration INTEGER,
    track_number INTEGER,
    cover_image_path TEXT,
    average_rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tracks_album_id ON tracks (album_id);
CREATE INDEX idx_tracks_deleted_at ON tracks (deleted_at);

CREATE TABLE track_genres (
    id SERIAL PRIMARY KEY,
    track_id INTEGER NOT NULL REFERENCES tracks(id),
    genre_id INTEGER NOT NULL REFERENCES genres(id)
);

CREATE INDEX idx_track_genres_track_id ON track_genres (track_id);
CREATE INDEX idx_track_genres_genre_id ON track_genres (genre_id);
CREATE UNIQUE INDEX ux_track_genres_track_genre ON track_genres (track_id, genre_id);

CREATE TYPE review_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    album_id INTEGER REFERENCES albums(id),
    track_id INTEGER REFERENCES tracks(id),
    text TEXT,
    rating_rhymes INTEGER NOT NULL CHECK (rating_rhymes >= 1 AND rating_rhymes <= 10),
    rating_structure INTEGER NOT NULL CHECK (rating_structure >= 1 AND rating_structure <= 10),
    rating_implementation INTEGER NOT NULL CHECK (rating_implementation >= 1 AND rating_implementation <= 10),
    rating_individuality INTEGER NOT NULL CHECK (rating_individuality >= 1 AND rating_individuality <= 10),
    atmosphere_multiplier DOUBLE PRECISION NOT NULL CHECK (atmosphere_multiplier >= 1.0000 AND atmosphere_multiplier <= 1.6072),
    final_score DOUBLE PRECISION NOT NULL,
    status review_status NOT NULL DEFAULT 'pending',
    moderated_by INTEGER REFERENCES users(id),
    moderated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_reviews_user_id ON reviews (user_id);
CREATE INDEX idx_reviews_album_id ON reviews (album_id);
CREATE INDEX idx_reviews_track_id ON reviews (track_id);
CREATE INDEX idx_reviews_status ON reviews (status);
CREATE INDEX idx_reviews_deleted_at ON reviews (deleted_at);

CREATE TABLE review_likes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    review_id INTEGER NOT NULL REFERENCES reviews(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_review_likes_user_review ON review_likes (user_id, review_id);
CREATE INDEX idx_review_likes_deleted_at ON review_likes (deleted_at);

CREATE TABLE track_likes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    track_id INTEGER NOT NULL REFERENCES tracks(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_track_likes_user_track ON track_likes (user_id, track_id);
CREATE INDEX idx_track_likes_deleted_at ON track_likes (deleted_at);

CREATE TABLE album_likes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    album_id INTEGER NOT NULL REFERENCES albums(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_album_likes_user_album ON album_likes (user_id, album_id);
CREATE INDEX idx_album_likes_deleted_at ON album_likes (deleted_at);

