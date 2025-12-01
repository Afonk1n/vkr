import React, { useState, useEffect, useCallback } from 'react';
import { albumsAPI, genresAPI } from '../services/api';
import AlbumCard from '../components/AlbumCard';
import './SearchPage.css';

const SearchPage = () => {
  const [albums, setAlbums] = useState([]);
  const [genres, setGenres] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // Filters state
  const [selectedGenres, setSelectedGenres] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortOrder, setSortOrder] = useState('desc');
  
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    page_size: 20,
  });

  // Fetch genres on mount
  useEffect(() => {
    const fetchGenres = async () => {
      try {
        const response = await genresAPI.getAll();
        setGenres(response.data);
      } catch (err) {
        console.error('Error fetching genres:', err);
      }
    };
    fetchGenres();
  }, []);

  // Fetch albums when filters change
  const fetchAlbums = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params = {
        page: pagination.page,
        page_size: pagination.page_size,
        sort_by: sortBy,
        sort_order: sortOrder,
      };

      // Add search query if exists
      if (searchQuery.trim()) {
        params.search = searchQuery.trim();
      }

      // Add genre filter if any selected
      if (selectedGenres.length > 0) {
        // For now, use first selected genre (backend supports single genre_id)
        // In future, can extend backend to support multiple genres
        params.genre_id = selectedGenres[0];
      }

      const response = await albumsAPI.getAll(params);
      setAlbums(response.data.albums || []);
      setPagination({
        total: response.data.total || 0,
        page: response.data.page || 1,
        page_size: response.data.page_size || 20,
      });
    } catch (err) {
      setError('Ошибка загрузки альбомов');
      console.error('Error fetching albums:', err);
    } finally {
      setLoading(false);
    }
  }, [selectedGenres, searchQuery, sortBy, sortOrder, pagination.page, pagination.page_size]);

  useEffect(() => {
    fetchAlbums();
  }, [fetchAlbums]);

  const handleGenreToggle = (genreId) => {
    setSelectedGenres((prev) => {
      if (prev.includes(genreId)) {
        return prev.filter((id) => id !== genreId);
      } else {
        // For now, only allow one genre (backend limitation)
        // Can be extended to support multiple genres
        return [genreId];
      }
    });
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleSortChange = (newSortBy, newSortOrder) => {
    setSortBy(newSortBy);
    setSortOrder(newSortOrder);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const clearFilters = () => {
    setSelectedGenres([]);
    setSearchQuery('');
    setSortBy('created_at');
    setSortOrder('desc');
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handlePageChange = (newPage) => {
    setPagination((prev) => ({ ...prev, page: newPage }));
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div className="container">
      <div className="search-page">
        <h1 className="search-page-title">Каталог альбомов</h1>
        
        <div className="search-page-content">
          {/* Left sidebar with filters */}
          <div className="search-filters-panel">
            <div className="filters-section">
              <h3 className="filters-section-title">Поиск</h3>
              <input
                type="text"
                className="search-input"
                placeholder="Название или исполнитель..."
                value={searchQuery}
                onChange={handleSearchChange}
              />
            </div>

            <div className="filters-section">
              <h3 className="filters-section-title">Жанры</h3>
              <div className="genres-checkboxes">
                {genres.map((genre) => (
                  <label key={genre.id} className="genre-checkbox-label">
                    <input
                      type="checkbox"
                      checked={selectedGenres.includes(genre.id)}
                      onChange={() => handleGenreToggle(genre.id)}
                    />
                    <span>{genre.name}</span>
                  </label>
                ))}
              </div>
            </div>

            <div className="filters-section">
              <h3 className="filters-section-title">Сортировка</h3>
              <div className="sort-options">
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'release_date' && sortOrder === 'desc'}
                    onChange={() => handleSortChange('release_date', 'desc')}
                  />
                  <span>По дате выхода (новые)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'release_date' && sortOrder === 'asc'}
                    onChange={() => handleSortChange('release_date', 'asc')}
                  />
                  <span>По дате выхода (старые)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'title' && sortOrder === 'asc'}
                    onChange={() => handleSortChange('title', 'asc')}
                  />
                  <span>По алфавиту (А-Я)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'title' && sortOrder === 'desc'}
                    onChange={() => handleSortChange('title', 'desc')}
                  />
                  <span>По алфавиту (Я-А)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'average_rating' && sortOrder === 'desc'}
                    onChange={() => handleSortChange('average_rating', 'desc')}
                  />
                  <span>По рейтингу (высокий)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'average_rating' && sortOrder === 'asc'}
                    onChange={() => handleSortChange('average_rating', 'asc')}
                  />
                  <span>По рейтингу (низкий)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'created_at' && sortOrder === 'desc'}
                    onChange={() => handleSortChange('created_at', 'desc')}
                  />
                  <span>По дате добавления (новые)</span>
                </label>
                <label className="sort-option">
                  <input
                    type="radio"
                    name="sort"
                    checked={sortBy === 'created_at' && sortOrder === 'asc'}
                    onChange={() => handleSortChange('created_at', 'asc')}
                  />
                  <span>По дате добавления (старые)</span>
                </label>
              </div>
            </div>

            <button onClick={clearFilters} className="btn-clear-filters">
              Сбросить фильтры
            </button>
          </div>

          {/* Right side with results */}
          <div className="search-results">
            {error && <div className="error-message">{error}</div>}
            
            {loading ? (
              <div className="loading">Загрузка...</div>
            ) : (
              <>
                <div className="results-header">
                  <p className="results-count">
                    Найдено альбомов: {pagination.total}
                  </p>
                </div>

                {albums.length === 0 ? (
                  <div className="empty-state">
                    Альбомы не найдены. Попробуйте изменить фильтры.
                  </div>
                ) : (
                  <>
                    <div className="albums-grid">
                      {albums.map((album) => (
                        <AlbumCard key={album.id} album={album} />
                      ))}
                    </div>

                    {/* Pagination */}
                    {pagination.total > pagination.page_size && (
                      <div className="pagination">
                        <button
                          onClick={() => handlePageChange(pagination.page - 1)}
                          disabled={pagination.page === 1}
                          className="pagination-btn"
                        >
                          ← Назад
                        </button>
                        <span className="pagination-info">
                          Страница {pagination.page} из {Math.ceil(pagination.total / pagination.page_size)}
                        </span>
                        <button
                          onClick={() => handlePageChange(pagination.page + 1)}
                          disabled={pagination.page >= Math.ceil(pagination.total / pagination.page_size)}
                          className="pagination-btn"
                        >
                          Вперёд →
                        </button>
                      </div>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SearchPage;

