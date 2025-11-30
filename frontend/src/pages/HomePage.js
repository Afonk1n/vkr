import React, { useState, useEffect, useCallback } from 'react';
import { albumsAPI, tracksAPI, reviewsAPI } from '../services/api';
import { useFilters } from '../context/FilterContext';
import AlbumCard from '../components/AlbumCard';
import TrackCard from '../components/TrackCard';
import Filters from '../components/Filters';
import './HomePage.css';

const HomePage = () => {
  const [albums, setAlbums] = useState([]);
  const [popularTracks, setPopularTracks] = useState([]);
  const [latestAlbums, setLatestAlbums] = useState([]);
  const [popularReviews, setPopularReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { filters, updateFilters } = useFilters();
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    page_size: 20,
  });

  const fetchAlbums = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params = {
        ...filters,
        genre_id: filters.genre_id || undefined,
        search: filters.search || undefined,
      };
      const response = await albumsAPI.getAll(params);
      setAlbums(response.data.albums);
      setPagination({
        total: response.data.total,
        page: response.data.page,
        page_size: response.data.page_size,
      });
    } catch (err) {
      setError('Ошибка загрузки альбомов');
      console.error('Error fetching albums:', err);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  const fetchPopularTracks = async () => {
    try {
      const response = await tracksAPI.getPopular({ limit: 8 });
      setPopularTracks(response.data);
    } catch (err) {
      console.error('Error fetching popular tracks:', err);
    }
  };

  const fetchLatestAlbums = async () => {
    try {
      const response = await albumsAPI.getAll({
        sort_by: 'created_at',
        sort_order: 'desc',
        page_size: 5,
      });
      setLatestAlbums(response.data.albums);
    } catch (err) {
      console.error('Error fetching latest albums:', err);
    }
  };

  const fetchPopularReviews = async () => {
    try {
      const response = await reviewsAPI.getPopular({ limit: 6 });
      setPopularReviews(response.data);
    } catch (err) {
      console.error('Error fetching popular reviews:', err);
    }
  };

  useEffect(() => {
    fetchAlbums();
    fetchPopularTracks();
    fetchLatestAlbums();
    fetchPopularReviews();
  }, [fetchAlbums]);

  const handleFilterChange = (newFilters) => {
    updateFilters({ ...newFilters, page: 1, page_size: 20 });
  };

  if (loading && albums.length === 0) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="container">
      <h1 className="page-title">Музыкальные альбомы</h1>
      
      {/* Latest Albums Section */}
      {latestAlbums.length > 0 && (
        <section className="home-section">
          <h2 className="section-title">Последние релизы</h2>
          <div className="albums-grid">
            {latestAlbums.map((album) => (
              <AlbumCard key={album.id} album={album} />
            ))}
          </div>
        </section>
      )}

      {/* Popular Tracks Section - moved here */}
      {popularTracks.length > 0 && (
        <section className="home-section">
          <h2 className="section-title">Самые залайканные треки за последние сутки</h2>
          <div className="tracks-list">
            {popularTracks.map((track) => (
              <TrackCard key={track.id} track={track} onUpdate={fetchPopularTracks} />
            ))}
          </div>
        </section>
      )}

      {/* Popular Reviews Section */}
      {popularReviews.length > 0 && (
        <section className="home-section">
          <h2 className="section-title">Популярные рецензии за последние сутки</h2>
          <div className="albums-grid">
            {popularReviews
              .filter(review => review.album) // Только рецензии с альбомами
              .map((review) => (
                <AlbumCard key={`review-${review.id}`} album={review.album} />
              ))}
          </div>
        </section>
      )}
    </div>
  );
};

export default HomePage;

