import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { searchAPI } from '../services/api';
import { getImageUrl } from '../utils/imageUtils';
import './SearchBar.css';

const SearchBar = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState({ albums: [], tracks: [] });
  const [showResults, setShowResults] = useState(false);
  const [loading, setLoading] = useState(false);
  const searchRef = useRef(null);
  const resultsRef = useRef(null);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (
        searchRef.current &&
        !searchRef.current.contains(event.target) &&
        resultsRef.current &&
        !resultsRef.current.contains(event.target)
      ) {
        setShowResults(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  useEffect(() => {
    const search = async () => {
      if (query.trim().length < 2) {
        setResults({ albums: [], tracks: [] });
        setShowResults(false);
        return;
      }

      setLoading(true);
      try {
        const response = await searchAPI.search(query);
        setResults(response.data);
        setShowResults(true);
      } catch (error) {
        console.error('Search error:', error);
        setResults({ albums: [], tracks: [] });
      } finally {
        setLoading(false);
      }
    };

    const timeoutId = setTimeout(search, 300);
    return () => clearTimeout(timeoutId);
  }, [query]);

  const handleAlbumClick = (albumId) => {
    navigate(`/albums/${albumId}`);
    setShowResults(false);
    setQuery('');
  };

  const handleTrackClick = (track) => {
    navigate(`/albums/${track.album_id}`);
    setShowResults(false);
    setQuery('');
  };

  const hasResults = results.albums.length > 0 || results.tracks.length > 0;
  
  const handleFiltersClick = () => {
    navigate('/search');
  };

  return (
    <>
      <div className="search-container" ref={searchRef}>
        <div className="search-bar">
          <input
            type="text"
            className="search-input"
            placeholder="Поиск альбомов и треков..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => query.length >= 2 && setShowResults(true)}
          />
          <button
            className="search-filter-btn"
            onClick={handleFiltersClick}
            title="Фильтры"
          >
            🔍 Фильтры
          </button>
        </div>
      {showResults && (hasResults || loading) && (
        <div className="search-results" ref={resultsRef}>
          {loading ? (
            <div className="search-loading">Поиск...</div>
          ) : (
            <>
              {results.albums.length > 0 && (
                <div className="search-section">
                  <div className="search-section-title">Альбомы</div>
                  {results.albums.map((album) => (
                    <div
                      key={album.id}
                      className="search-result-item"
                      onClick={() => handleAlbumClick(album.id)}
                    >
                      {getImageUrl(album.cover_image_path) && (
                        <img
                          src={getImageUrl(album.cover_image_path)}
                          alt={album.title}
                          className="search-result-image"
                        />
                      )}
                      <div className="search-result-info">
                        <div className="search-result-title">{album.title}</div>
                        <div className="search-result-subtitle">{album.artist}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {results.tracks.length > 0 && (
                <div className="search-section">
                  <div className="search-section-title">Треки</div>
                  {results.tracks.map((track) => (
                    <div
                      key={track.id}
                      className="search-result-item"
                      onClick={() => handleTrackClick(track)}
                    >
                      {getImageUrl(track.cover_image_path) && (
                        <img
                          src={getImageUrl(track.cover_image_path)}
                          alt={track.title}
                          className="search-result-image"
                        />
                      )}
                      <div className="search-result-info">
                        <div className="search-result-title">{track.title}</div>
                        <div className="search-result-subtitle">
                          {track.album_title} • {track.artist}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {!hasResults && query.length >= 2 && (
                <div className="search-no-results">Ничего не найдено</div>
              )}
            </>
          )}
        </div>
      )}
      </div>
    </>
  );
};

export default SearchBar;

