import React, { useState, useEffect, useCallback } from 'react';
import AlbumCard from '../components/AlbumCard';
import TrackCard from '../components/TrackCard';
import ReviewCardSmall from '../components/ReviewCardSmall';
import { albumsAPI, tracksAPI, reviewsAPI } from '../services/api';
import './HomePage.css';

const LATEST_ALBUMS = 5;
const BEST_ALBUMS = 5;
const RATED_POOL = 24;
const HIDDEN_GEMS = 5;
const POPULAR_TRACKS = 8;
const POPULAR_REVIEWS = 6;
const ARTIST_PICKS = 6;

const TopsPage = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [latestAlbums, setLatestAlbums] = useState([]);
  const [bestAlbums, setBestAlbums] = useState([]);
  const [hiddenGems, setHiddenGems] = useState([]);
  const [popularTracks, setPopularTracks] = useState([]);
  const [popularReviews, setPopularReviews] = useState([]);
  const [artistPicks, setArtistPicks] = useState([]);

  const loadTops = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [latestRes, ratedRes, tracksRes, popularRevRes, artistRes] = await Promise.all([
        albumsAPI.getAll({ sort_by: 'release_date', sort_order: 'desc', page_size: LATEST_ALBUMS, page: 1 }),
        albumsAPI.getAll({ sort_by: 'average_rating', sort_order: 'desc', page_size: RATED_POOL, page: 1 }),
        tracksAPI.getPopular({ limit: POPULAR_TRACKS }),
        reviewsAPI.getPopular({ limit: POPULAR_REVIEWS }),
        reviewsAPI.getAll({ page_size: 30, sort_by: 'created_at', sort_order: 'desc' }),
      ]);

      const latest = Array.isArray(latestRes.data?.albums) ? latestRes.data.albums : [];
      setLatestAlbums(latest);

      const rated = (Array.isArray(ratedRes.data?.albums) ? ratedRes.data.albums : [])
        .filter((a) => Number(a.average_rating) > 0);
      setBestAlbums(rated.slice(0, BEST_ALBUMS));
      // Скрытые жемчужины: хороший балл, но меньше всего лайков (недооценённое).
      // Относительно, а не по жёсткому порогу — иначе при «налайканных» данных пусто.
      const bestIds = new Set(rated.slice(0, BEST_ALBUMS).map((a) => a.id));
      setHiddenGems(
        rated
          .filter((a) => Number(a.average_rating) >= 65 && !bestIds.has(a.id))
          .sort((a, b) => (a.likes?.length || 0) - (b.likes?.length || 0))
          .slice(0, HIDDEN_GEMS)
      );

      setPopularTracks(Array.isArray(tracksRes.data) ? tracksRes.data : []);
      setPopularReviews(Array.isArray(popularRevRes.data) ? popularRevRes.data : []);

      const reviews = Array.isArray(artistRes.data?.reviews) ? artistRes.data.reviews : [];
      setArtistPicks(reviews.filter((r) => r.has_artist_mark).slice(0, ARTIST_PICKS));
    } catch (e) {
      console.error(e);
      setError('Не удалось загрузить топы.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadTops();
  }, [loadTops]);

  const refetchPopularReviews = useCallback(() => {
    reviewsAPI
      .getPopular({ limit: POPULAR_REVIEWS })
      .then((res) => {
        if (Array.isArray(res.data)) setPopularReviews(res.data);
      })
      .catch(() => {});
  }, []);

  return (
    <div className="container">
      <header className="feed-page-intro">
        <p className="feed-page-lead">
          Лучшее по оценкам, недооценённые жемчужины, свежие релизы и самое популярное за сутки.
        </p>
      </header>

      {error && <div className="error-message">{error}</div>}

      {loading ? (
        <div className="home-skeleton-wrap" aria-busy="true">
          <div className="home-skeleton home-skeleton--wide" />
          <div className="home-skeleton" />
        </div>
      ) : (
        <>
          <section className="home-section">
            <h2 className="section-title">Лучшие по оценке</h2>
            {bestAlbums.length === 0 ? (
              <div className="empty-state empty-state--soft">Пока нет оценённых альбомов</div>
            ) : (
              <div className="albums-grid">
                {bestAlbums.map((album) => (
                  <AlbumCard key={album.id} album={album} />
                ))}
              </div>
            )}
          </section>

          {hiddenGems.length > 0 && (
            <section className="home-section">
              <h2 className="section-title">Скрытые жемчужины</h2>
              <p className="feed-page-lead" style={{ marginTop: '-0.5rem' }}>
                Высокие оценки, но мало лайков — стоит послушать.
              </p>
              <div className="albums-grid">
                {hiddenGems.map((album) => (
                  <AlbumCard key={album.id} album={album} />
                ))}
              </div>
            </section>
          )}

          <section className="home-section">
            <h2 className="section-title">Последние релизы</h2>
            {latestAlbums.length === 0 ? (
              <div className="empty-state empty-state--soft">Нет альбомов в каталоге</div>
            ) : (
              <div className="albums-grid">
                {latestAlbums.map((album) => (
                  <AlbumCard key={album.id} album={album} />
                ))}
              </div>
            )}
          </section>

          <section className="home-section">
            <h2 className="section-title">Топ треков за сутки</h2>
            {popularTracks.length === 0 ? (
              <div className="empty-state empty-state--soft">За последние сутки нет активности по лайкам</div>
            ) : (
              <div className="tracks-list">
                {popularTracks.map((track) => (
                  <TrackCard key={track.id} track={track} onUpdate={loadTops} />
                ))}
              </div>
            )}
          </section>

          <section className="home-section">
            <h2 className="section-title">Топ рецензий за сутки</h2>
            {popularReviews.length === 0 ? (
              <div className="empty-state empty-state--soft">Нет популярных рецензий за сутки</div>
            ) : (
              <div className="reviews-grid-popular">
                {popularReviews.map((review) => (
                  <ReviewCardSmall key={review.id} review={review} onUpdate={refetchPopularReviews} />
                ))}
              </div>
            )}
          </section>

          {artistPicks.length > 0 && (
            <section className="home-section">
              <h2 className="section-title">Выбор артистов</h2>
              <p className="feed-page-lead" style={{ marginTop: '-0.5rem' }}>
                Рецензии, отмеченные верифицированными артистами.
              </p>
              <div className="reviews-grid-popular">
                {artistPicks.map((review) => (
                  <ReviewCardSmall key={review.id} review={review} />
                ))}
              </div>
            </section>
          )}
        </>
      )}
    </div>
  );
};

export default TopsPage;
