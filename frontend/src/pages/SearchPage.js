import React, { useState, useEffect, useCallback } from 'react';
import { albumsAPI, genresAPI, tracksAPI } from '../services/api';
import TrackCard from '../components/TrackCard';
import './SearchPage.css';

const SearchPage = () => {
  const [tracks, setTracks] = useState([]);
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

  // Fetch tracks when filters change
  const fetchTracks = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      // Get albums with filters first
      const params = {
        page: pagination.page,
        page_size: pagination.page_size,
        sort_by: sortBy,
        sort_order: sortOrder,
      };

      if (searchQuery.trim()) {
        params.search = searchQuery.trim();
      }

      if (selectedGenres.length > 0) {
        params.genre_id = selectedGenres[0];
      }

      const albumsResponse = await albumsAPI.getAll(params);
      const albums = albumsResponse.data.albums || [];
      
      // Get all tracks from these albums
      const allTracks = [];
      for (const album of albums) {
        try {
          const tracksResponse = await tracksAPI.getByAlbum(album.id);
          const albumTracks = Array.isArray(tracksResponse.data) ? tracksResponse.data : [];
          allTracks.push(...albumTracks.map(track => ({ ...track, album })));
        } catch (err) {
          console.error(`Error fetching tracks for album ${album.id}:`, err);
        }
      }
      
      setTracks(allTracks);
      setPagination({
        total: allTracks.length,
        page: albumsResponse.data.page || 1,
        page_size: albumsResponse.data.page_size || 20,
      });
    } catch (err) {
      setError('Ошибка загрузки треков');
      console.error('Error fetching tracks:', err);
    } finally {
      setLoading(false);
    }
  }, [selectedGenres, searchQuery, sortBy, sortOrder, pagination.page, pagination.page_size]);

  useEffect(() => {
    fetchTracks();
  }, [fetchTracks]);

  const handleGenreToggle = (genreId) => {
    setSelectedGenres((prev) => {
      if (prev.includes(genreId)) {
        return prev.filter((id) => id !== genreId);
      } else {
        return [...prev, genreId];
      }
    });
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleSortChange = (e) => {
    const [newSortBy, newSortOrder] = e.target.value.split('_');
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
              <div className="genres-buttons">
                {genres.map((genre) => (
                  <button
                    key={genre.id}
                    className={`genre-button ${selectedGenres.includes(genre.id) ? 'active' : ''}`}
                    onClick={() => handleGenreToggle(genre.id)}
                  >
                    {genre.name}
                  </button>
                ))}
              </div>
            </div>

            <div className="filters-section">
              <h3 className="filters-section-title">Сортировка</h3>
              <select
                className="sort-select"
                value={`${sortBy}_${sortOrder}`}
                onChange={handleSortChange}
              >
                <option value="created_at_desc">По дате добавления (новые)</option>
                <option value="created_at_asc">По дате добавления (старые)</option>
                <option value="release_date_desc">По дате выхода (новые)</option>
                <option value="release_date_asc">По дате выхода (старые)</option>
                <option value="title_asc">По алфавиту (А-Я)</option>
                <option value="title_desc">По алфавиту (Я-А)</option>
                <option value="average_rating_desc">По рейтингу (высокий)</option>
                <option value="average_rating_asc">По рейтингу (низкий)</option>
              </select>
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
                    Найдено треков: {pagination.total}
                  </p>
                </div>

                {tracks.length === 0 ? (
                  <div className="empty-state">
                    Треки не найдены. Попробуйте изменить фильтры.
                  </div>
                ) : (
                  <>
                    <div className="tracks-grid">
                      {tracks.map((track) => (
                        <TrackCard key={track.id} track={track} />
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

