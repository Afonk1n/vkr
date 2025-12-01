import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { albumsAPI } from '../services/api';
import AlbumCard from '../components/AlbumCard';
import './ArtistPage.css';

const ArtistPage = () => {
  const { name } = useParams();
  const navigate = useNavigate();
  const [albums, setAlbums] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [artistName, setArtistName] = useState('');

  useEffect(() => {
    fetchAlbums();
  }, [name]);

  const fetchAlbums = async () => {
    try {
      setLoading(true);
      const decodedName = decodeURIComponent(name);
      setArtistName(decodedName);
      const response = await albumsAPI.getByArtist(decodedName);
      setAlbums(response.data.albums || []);
    } catch (err) {
      setError('Ошибка загрузки альбомов артиста');
      console.error('Error fetching albums:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  if (error || albums.length === 0) {
    return (
      <div className="container">
        <button onClick={() => navigate(-1)} className="btn-back">
          ← Назад
        </button>
        <div className="error-message">
          {error || `Альбомы артиста "${artistName || name}" не найдены`}
        </div>
      </div>
    );
  }

  // Calculate statistics
  const totalAlbums = albums.length;
  const avgRating = albums.length > 0
    ? Math.round(albums.reduce((sum, album) => sum + (album.average_rating || 0), 0) / albums.length)
    : 0;
  const totalLikes = albums.reduce((sum, album) => sum + (album.likes?.length || 0), 0);

  return (
    <div className="container">
      <div className="artist-page">
        <button onClick={() => navigate(-1)} className="btn-back">
          ← Назад
        </button>

        <div className="artist-header">
          <h1 className="artist-name">{artistName || name}</h1>
          <div className="artist-stats">
            <div className="artist-stat-item">
              <span className="stat-label">Альбомов:</span>
              <span className="stat-value">{totalAlbums}</span>
            </div>
            {avgRating > 0 && (
              <div className="artist-stat-item">
                <span className="stat-label">Средний рейтинг:</span>
                <span className="stat-value">⭐ {avgRating}</span>
              </div>
            )}
            {totalLikes > 0 && (
              <div className="artist-stat-item">
                <span className="stat-label">Лайков:</span>
                <span className="stat-value">❤️ {totalLikes}</span>
              </div>
            )}
          </div>
        </div>

        <div className="artist-albums-section">
          <h2 className="section-title">Альбомы ({totalAlbums})</h2>
          <div className="albums-grid">
            {albums.map((album) => (
              <AlbumCard key={album.id} album={album} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ArtistPage;

