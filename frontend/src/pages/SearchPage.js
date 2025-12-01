import React, { useState, useEffect } from 'react';
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
        // Check if we have any filters
        const hasFilters = searchQuery.trim() || selectedGenres.length > 0;
        
        let albums = [];
        
        if (hasFilters) {
          // If filters are applied, get filtered albums
          const params = {
            page: 1,
            page_size: 100, // Get more albums to have more tracks
            sort_by: sortBy,
            sort_order: sortOrder,
          };

          if (searchQuery.trim()) {
            params.search = searchQuery.trim();
          }

          // Use first selected genre (API supports only one genre_id at a time)
          if (selectedGenres.length > 0) {
            params.genre_id = selectedGenres[0];
          }

          try {
            const albumsResponse = await albumsAPI.getAll(params);
            if (isCancelled) return;
            albums = albumsResponse.data?.albums || albumsResponse.data || [];
          } catch (apiErr) {
            console.error('[SearchPage] Error fetching albums:', apiErr);
            throw apiErr;
          }
        } else {
          // If no filters, get all albums sorted by created_at DESC
          const params = {
            page: 1,
            page_size: 100, // Get all albums
            sort_by: 'created_at',
            sort_order: 'desc',
          };
          try {
            const albumsResponse = await albumsAPI.getAll(params);
            if (isCancelled) return;
            albums = albumsResponse.data?.albums || albumsResponse.data || [];
          } catch (apiErr) {
            console.error('[SearchPage] Error fetching albums:', apiErr);
            throw apiErr;
          }
        }
        
        if (albums.length === 0) {
          setTracks([]);
          setPagination({
            total: 0,
            page: currentPage,
            page_size: pageSize,
          });
          setLoading(false);
          return;
        }
        
        // Get all tracks from these albums
        const trackPromises = albums.map(async (album) => {
          try {
            const tracksResponse = await tracksAPI.getByAlbum(album.id);
            const albumTracks = Array.isArray(tracksResponse.data) ? tracksResponse.data : [];
            return albumTracks.map(track => ({ ...track, album }));
          } catch (err) {
            console.error(`[SearchPage] Error fetching tracks for album ${album.id}:`, err);
            return [];
          }
        });

        const trackArrays = await Promise.all(trackPromises);
        if (isCancelled) return;
        
        // Flatten all tracks
        let allTracks = trackArrays.flat();
        
        if (allTracks.length === 0) {
          setTracks([]);
          setPagination({
            total: 0,
            page: currentPage,
            page_size: pageSize,
          });
          setLoading(false);
          return;
        }
        
        // Sort tracks based on sortBy and sortOrder
        allTracks.sort((a, b) => {
          let valueA, valueB;
          let isString = false;
          
          switch (sortBy) {
            case 'created_at': {
              const dateA = a.created_at ? new Date(a.created_at) : new Date(0);
              const dateB = b.created_at ? new Date(b.created_at) : new Date(0);
              valueA = isNaN(dateA.getTime()) ? 0 : dateA.getTime();
              valueB = isNaN(dateB.getTime()) ? 0 : dateB.getTime();
              break;
            }
            case 'release_date': {
              const releaseDateA = a.album?.release_date || a.created_at;
              const releaseDateB = b.album?.release_date || b.created_at;
              const dateA = releaseDateA ? new Date(releaseDateA) : new Date(0);
              const dateB = releaseDateB ? new Date(releaseDateB) : new Date(0);
              valueA = isNaN(dateA.getTime()) ? 0 : dateA.getTime();
              valueB = isNaN(dateB.getTime()) ? 0 : dateB.getTime();
              break;
            }
            case 'title':
              valueA = (a.title || '').toLowerCase();
              valueB = (b.title || '').toLowerCase();
              isString = true;
              break;
            case 'average_rating':
              valueA = Number(a.average_rating) || 0;
              valueB = Number(b.average_rating) || 0;
              break;
            default: {
              const dateA = a.created_at ? new Date(a.created_at) : new Date(0);
              const dateB = b.created_at ? new Date(b.created_at) : new Date(0);
              valueA = isNaN(dateA.getTime()) ? 0 : dateA.getTime();
              valueB = isNaN(dateB.getTime()) ? 0 : dateB.getTime();
              break;
            }
          }
          
          if (isString) {
            // String comparison
            if (sortOrder === 'asc') {
              return valueA.localeCompare(valueB);
            } else {
              return valueB.localeCompare(valueA);
            }
          } else {
            // Number/Date comparison
            if (sortOrder === 'asc') {
              return valueA - valueB;
            } else {
              return valueB - valueA;
            }
          }
        });
        
        // Apply pagination to tracks
        const startIndex = (currentPage - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        const paginatedTracks = allTracks.slice(startIndex, endIndex);
        
        // Ensure all tracks have required fields
        const validTracks = paginatedTracks.filter(track => track && track.id);
        setTracks(validTracks);
        setPagination({
          total: allTracks.length,
          page: currentPage,
          page_size: pageSize,
        });
      } catch (err) {
        if (isCancelled) return;
        setError('Ошибка загрузки треков');
        console.error('[SearchPage] Error fetching tracks:', err);
        console.error('[SearchPage] Error details:', err.response?.data || err.message);
        setTracks([]);
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
    const [newSortBy, newSortOrder] = e.target.value.split('_');
    setSortBy(newSortBy);
    setSortOrder(newSortOrder);
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

