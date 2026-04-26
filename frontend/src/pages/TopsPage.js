import React, { useState, useEffect, useCallback } from 'react';
import AlbumCard from '../components/AlbumCard';
import TrackCard from '../components/TrackCard';
import ReviewCardSmall from '../components/ReviewCardSmall';
import { albumsAPI, tracksAPI, reviewsAPI } from '../services/api';
import './HomePage.css';

const LATEST_ALBUMS = 5;
const POPULAR_TRACKS = 8;
const POPULAR_REVIEWS = 6;

const TopsPage = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [latestAlbums, setLatestAlbums] = useState([]);
  const [popularTracks, setPopularTracks] = useState([]);
  const [popularReviews, setPopularReviews] = useState([]);

  const loadTops = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [albumsRes, tracksRes, popularRevRes] = await Promise.all([
        albumsAPI.getAll({
          sort_by: 'release_date',
          sort_order: 'desc',
          page_size: LATEST_ALBUMS,
          page: 1,
        }),
        tracksAPI.getPopular({ limit: POPULAR_TRACKS }),
        reviewsAPI.getPopular({ limit: POPULAR_REVIEWS }),
      ]);
      const rawAlbums = albumsRes.data?.albums ?? albumsRes.data;
      setLatestAlbums(Array.isArray(rawAlbums) ? rawAlbums : []);
      setPopularTracks(Array.isArray(tracksRes.data) ? tracksRes.data : []);
      setPopularReviews(Array.isArray(popularRevRes.data) ? popularRevRes.data : []);
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
          Релизы по дате и самое популярное за последние сутки по лайкам.
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
                  <ReviewCardSmall
                    key={review.id}
                    review={review}
                    onUpdate={refetchPopularReviews}
                  />
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  );
};

export default TopsPage;
