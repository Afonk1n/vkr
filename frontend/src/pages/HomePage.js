import React, { useState, useEffect, useCallback } from 'react';
import { albumsAPI, tracksAPI, reviewsAPI } from '../services/api';
import { useFilters } from '../context/FilterContext';
import AlbumCard from '../components/AlbumCard';
import TrackCard from '../components/TrackCard';
import ReviewCardSmall from '../components/ReviewCardSmall';
import Filters from '../components/Filters';
import { mockAlbums, mockTracks, mockReviews } from '../data/mockData';
import './HomePage.css';

// ВРЕМЕННЫЙ ФЛАГ ДЛЯ ДЕМОНСТРАЦИИ БЕЗ BACKEND
const USE_MOCK_DATA = true;

// Проверка импорта моковых данных
console.log('Mock data check:', {
  mockAlbums: mockAlbums?.length || 0,
  mockTracks: mockTracks?.length || 0,
  mockReviews: mockReviews?.length || 0
});

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

  // Загрузка моковых данных
  const loadMockData = () => {
    console.log('Loading mock data...', { 
      mockAlbumsCount: mockAlbums?.length || 0, 
      mockTracksCount: mockTracks?.length || 0, 
      mockReviewsCount: mockReviews?.length || 0 
    });
    
    if (!mockAlbums || mockAlbums.length === 0) {
      console.error('Mock albums are empty or undefined!');
      setLoading(false);
      return;
    }
    
    setLoading(true);
    
    // Загружаем данные сразу, без задержки
    const latest = mockAlbums.slice(0, 5);
    const popular = mockTracks.slice(0, 8);
    const reviews = mockReviews.slice(0, 6);
    
    console.log('Setting mock data:', { 
      latestCount: latest.length, 
      popularCount: popular.length, 
      reviewsCount: reviews.length 
    });
    
    setLatestAlbums(latest);
    setPopularTracks(popular);
    setPopularReviews(reviews);
    setAlbums(mockAlbums);
    setPagination({
      total: mockAlbums.length,
      page: 1,
      page_size: 20,
    });
    setLoading(false);
    
    console.log('Mock data loaded successfully');
  };

  const fetchAlbums = useCallback(async () => {
    // В моковом режиме полностью отключаем API запросы
    if (USE_MOCK_DATA) {
      console.log('fetchAlbums: skipping API call in mock mode');
      return;
    }
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
  }, [USE_MOCK_DATA ? null : filters]); // Отключаем зависимость в моковом режиме

  const fetchPopularTracks = async () => {
    // В моковом режиме полностью отключаем API запросы
    if (USE_MOCK_DATA) {
      console.log('fetchPopularTracks: skipping API call in mock mode');
      return;
    }
    try {
      const response = await tracksAPI.getPopular({ limit: 8 });
      setPopularTracks(response.data);
    } catch (err) {
      console.error('Error fetching popular tracks:', err);
    }
  };

  const fetchLatestAlbums = async () => {
    // В моковом режиме полностью отключаем API запросы
    if (USE_MOCK_DATA) {
      console.log('fetchLatestAlbums: skipping API call in mock mode');
      return;
    }
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
    // В моковом режиме полностью отключаем API запросы
    if (USE_MOCK_DATA) {
      console.log('fetchPopularReviews: skipping API call in mock mode');
      return;
    }
    try {
      const response = await reviewsAPI.getPopular({ limit: 6 });
      console.log('Popular reviews fetched:', response.data);
      setPopularReviews(response.data || []);
    } catch (err) {
      console.error('Error fetching popular reviews:', err);
      setPopularReviews([]);
    }
  };

  useEffect(() => {
    console.log('HomePage useEffect, USE_MOCK_DATA:', USE_MOCK_DATA);
    if (USE_MOCK_DATA) {
      // В моковом режиме загружаем данные сразу, без API запросов
      loadMockData();
    } else {
      // Только если НЕ моковый режим - делаем API запросы
      fetchAlbums();
      fetchPopularTracks();
      fetchLatestAlbums();
      fetchPopularReviews();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Отключаем обновление при изменении filters в моковом режиме
  useEffect(() => {
    if (!USE_MOCK_DATA && filters) {
      fetchAlbums();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [USE_MOCK_DATA ? null : filters]);

  const handleFilterChange = (newFilters) => {
    updateFilters({ ...newFilters, page: 1, page_size: 20 });
  };

  // Отладочная информация
  useEffect(() => {
    console.log('HomePage state:', {
      loading,
      latestAlbums: latestAlbums.length,
      popularTracks: popularTracks.length,
      popularReviews: popularReviews.length,
      USE_MOCK_DATA
    });
  }, [loading, latestAlbums, popularTracks, popularReviews]);

  if (loading && latestAlbums.length === 0 && popularTracks.length === 0) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  // Отладочная информация для рендеринга
  console.log('Rendering HomePage with:', {
    loading,
    latestAlbums: latestAlbums.length,
    popularTracks: popularTracks.length,
    popularReviews: popularReviews.length
  });

  return (
    <div className="container">
      <h1 className="page-title">Актуальное</h1>
      
      {/* Latest Albums Section */}
      {latestAlbums && latestAlbums.length > 0 ? (
        <section className="home-section">
          <h2 className="section-title">Последние релизы</h2>
          <div className="albums-grid">
            {latestAlbums.map((album) => (
              <AlbumCard key={album.id} album={album} />
            ))}
          </div>
        </section>
      ) : (
        !loading && <div>Нет альбомов для отображения</div>
      )}

      {/* Popular Tracks Section */}
      {popularTracks && popularTracks.length > 0 ? (
        <section className="home-section">
          <h2 className="section-title">Топ треков за сутки</h2>
          <div className="tracks-list">
            {popularTracks.map((track) => (
              <TrackCard key={track.id} track={track} onUpdate={fetchPopularTracks} />
            ))}
          </div>
        </section>
      ) : (
        !loading && <div>Нет треков для отображения</div>
      )}

      {/* Popular Reviews Section */}
      {popularReviews && popularReviews.length > 0 ? (
        (() => {
          const validReviews = popularReviews
            .filter(review => review && review.album_id && review.album)
            .slice(0, 6);
          
          if (validReviews.length === 0) {
            return null;
          }

          return (
            <section className="home-section">
              <h2 className="section-title">Топ рецензий за сутки</h2>
              <div className="reviews-grid-popular">
                {validReviews.map((review) => (
                  <ReviewCardSmall key={`review-${review.id}`} review={review} onUpdate={fetchPopularReviews} />
                ))}
              </div>
            </section>
          );
        })()
      ) : (
        !loading && <div>Нет рецензий для отображения</div>
      )}
    </div>
  );
};

export default HomePage;

