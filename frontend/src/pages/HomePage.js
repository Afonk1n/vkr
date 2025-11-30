import React, { useState, useEffect, useCallback } from 'react';
import { albumsAPI } from '../services/api';
import AlbumCard from '../components/AlbumCard';
import Filters from '../components/Filters';
import './HomePage.css';

const HomePage = () => {
  const [albums, setAlbums] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filters, setFilters] = useState({
    genre_id: null,
    search: '',
    sort_by: 'created_at',
    sort_order: 'desc',
    page: 1,
    page_size: 20,
  });
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

  useEffect(() => {
    fetchAlbums();
  }, [fetchAlbums]);

  const handleFilterChange = (newFilters) => {
    setFilters({ ...newFilters, page: 1, page_size: 20 });
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
      <Filters onFilterChange={handleFilterChange} filters={filters} />
      {error && <div className="error-message">{error}</div>}
      {albums.length === 0 ? (
        <div className="empty-state">Альбомы не найдены</div>
      ) : (
        <>
          <div className="albums-grid">
            {albums.map((album) => (
              <AlbumCard key={album.id} album={album} />
            ))}
          </div>
          {pagination.total > pagination.page_size && (
            <div className="pagination-info">
              Показано {albums.length} из {pagination.total} альбомов
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default HomePage;

