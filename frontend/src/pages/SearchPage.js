import React, { useState, useEffect } from 'react';
import { genresAPI, tracksAPI } from '../services/api';
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
  
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);

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
  useEffect(() => {
    let isCancelled = false;
    
    const fetchTracks = async () => {
      setLoading(true);
      setError('');
      try {
        // Build params for tracks API
        const params = {
          page: currentPage,
          page_size: pageSize,
          sort_by: sortBy,
          sort_order: sortOrder,
        };

        // Add search query if provided
        if (searchQuery.trim()) {
          params.search = searchQuery.trim();
        }

        // Add genre_ids array if any genres selected
        if (selectedGenres.length > 0) {
          params.genre_ids = selectedGenres;
        }

        // Fetch tracks directly from API
        const response = await tracksAPI.getAll(params);
        if (isCancelled) return;

        const tracksData = response.data?.tracks || response.data || [];
        const total = response.data?.total || 0;
        const page = response.data?.page || currentPage;
        const pageSizeResponse = response.data?.page_size || pageSize;

        // Ensure all tracks have album relationship loaded
        const validTracks = tracksData.filter(track => track && track.id);
        
        setTracks(validTracks);
        setPagination({
          total: total,
          page: page,
          page_size: pageSizeResponse,
        });
      } catch (err) {
        if (isCancelled) return;
        setError('Ошибка загрузки треков');
        console.error('[SearchPage] Error fetching tracks:', err);
        console.error('[SearchPage] Error details:', err.response?.data || err.message);
        setTracks([]);
        setPagination({
          total: 0,
          page: currentPage,
          page_size: pageSize,
        });
      } finally {
        if (!isCancelled) {
          setLoading(false);
        }
      }
    };

    fetchTracks();
    
    return () => {
      isCancelled = true;
    };
  }, [selectedGenres, searchQuery, sortBy, sortOrder, currentPage, pageSize]);

  const handleGenreToggle = (genreId) => {
    setSelectedGenres((prev) => {
      if (prev.includes(genreId)) {
        return prev.filter((id) => id !== genreId);
      } else {
        return [...prev, genreId];
      }
    });
    setCurrentPage(1);
  };

  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    setCurrentPage(1);
  };

  const handleSortChange = (e) => {
    const value = e.target.value;
    if (!value || !value.includes('_')) {
      console.error('Invalid sort value:', value);
      return;
    }
    const [newSortBy, newSortOrder] = value.split('_');
    
    // Validate sortBy and sortOrder
    const validSortBy = ['created_at', 'release_date', 'title', 'average_rating', 'likes_count'].includes(newSortBy) 
      ? newSortBy 
      : 'created_at';
    const validSortOrder = ['asc', 'desc'].includes(newSortOrder) 
      ? newSortOrder 
      : 'desc';
    
    setSortBy(validSortBy);
    setSortOrder(validSortOrder);
    setCurrentPage(1);
  };

  const clearFilters = () => {
    setSelectedGenres([]);
    setSearchQuery('');
    setSortBy('created_at');
    setSortOrder('desc');
    setCurrentPage(1);
  };

  const handlePageChange = (newPage) => {
    setCurrentPage(newPage);
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
                <option value="likes_count_desc">По лайкам (больше)</option>
                <option value="likes_count_asc">По лайкам (меньше)</option>
              </select>
            </div>

            <button onClick={clearFilters} className="btn-clear-filters">
              Сбросить фильтры
            </button>
          </div>

          {/* Right side with results */}
          <div className="search-results-wrapper">
            {error && (
              <div className="error-message">{error}</div>
            )}
            
            {loading ? (
              <div className="loading">Загрузка...</div>
            ) : (
              <>
                {/* Results header with count and pagination */}
                <div className="results-top-bar">
                  <div className="results-info">
                    <span className="results-count-label">Найдено треков:</span>
                    <span className="results-count-value">{pagination.total}</span>
                  </div>
                  
                  {pagination.total > pageSize && (
                    <div className="results-pagination">
                      <button
                        onClick={() => handlePageChange(currentPage - 1)}
                        disabled={currentPage === 1}
                        className="pagination-button"
                        aria-label="Предыдущая страница"
                      >
                        ←
                      </button>
                      <span className="pagination-text">
                        {currentPage} из {Math.ceil(pagination.total / pageSize)}
                      </span>
                      <button
                        onClick={() => handlePageChange(currentPage + 1)}
                        disabled={currentPage >= Math.ceil(pagination.total / pageSize)}
                        className="pagination-button"
                        aria-label="Следующая страница"
                      >
                        →
                      </button>
                    </div>
                  )}
                </div>

                {/* Tracks grid */}
                {tracks.length === 0 ? (
                  <div className="empty-state">
                    <p>Треки не найдены</p>
                    <p className="empty-state-hint">Попробуйте изменить фильтры поиска</p>
                  </div>
                ) : (
                  <div className="tracks-grid-container">
                    <div className="tracks-grid">
                      {tracks.map((track) => (
                        <TrackCard key={track.id} track={track} />
                      ))}
                    </div>
                  </div>
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

