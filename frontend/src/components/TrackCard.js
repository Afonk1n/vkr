import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { tracksAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { getImageUrl } from '../utils/imageUtils';
import './TrackCard.css';

const TrackCard = ({ track, onUpdate }) => {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/tracks/${track.id}`);
  };

  const handleLike = async () => {
    try {
      await tracksAPI.like(track.id);
      if (onUpdate) onUpdate();
    } catch (err) {
      throw err;
    }
  };

  const handleUnlike = async () => {
    try {
      await tracksAPI.unlike(track.id);
      if (onUpdate) onUpdate();
    } catch (err) {
      throw err;
    }
  };

  const { user } = useAuth();
  const [likeCount, setLikeCount] = useState(track.likes?.length || 0);
  const [isLiked, setIsLiked] = useState(false);

  useEffect(() => {
    setLikeCount(track.likes?.length || 0);
    if (user && track.likes) {
      setIsLiked(track.likes.some(like => like.user_id === user.id));
    }
  }, [track.likes, user]);

  const handleLikeClick = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      if (isLiked) {
        await handleUnlike();
        setLikeCount(prev => Math.max(0, prev - 1));
        setIsLiked(false);
      } else {
        await handleLike();
        setLikeCount(prev => prev + 1);
        setIsLiked(true);
      }
    } catch (err) {
      console.error('Error toggling like:', err);
    }
  };

  const coverImagePath = track.cover_image_path || track.album?.cover_image_path;
  const coverImageUrl = getImageUrl(coverImagePath);
  const [imageError, setImageError] = React.useState(false);

  return (
    <div className="track-card" onClick={handleClick}>
      <div className="track-card-cover">
        {coverImageUrl && !imageError ? (
          <img
            src={coverImageUrl}
            alt={track.album?.title || track.title}
            className="track-card-image"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="track-card-image-placeholder">🎵</div>
        )}
        {/* Лайки и счетчики прямо на обложке, как в AlbumCard */}
        <div className="track-card-cover-overlay">
          <div className="track-card-stats">
            {likeCount > 0 && (
              <div className="track-card-stat-item">
                <span className="stat-icon">❤️</span>
                <span className="stat-count">{likeCount}</span>
              </div>
            )}
            {track.average_rating > 0 && (
              <div className="track-card-stat-item">
                <span className="stat-icon">⭐</span>
                <span className="stat-count">{Math.round(track.average_rating)}</span>
              </div>
            )}
          </div>
          <button
            className={`track-card-like-button ${isLiked ? 'liked' : ''}`}
            onClick={handleLikeClick}
            title="Поставить лайк"
          >
            {isLiked ? '❤️' : '🤍'}
          </button>
        </div>
      </div>
      <div className="track-card-info">
        <h3 className="track-card-title">{track.title}</h3>
        <p className="track-card-subtitle">
          <span className="track-album-label">Альбом:</span>{' '}
          <span className="track-album-title">{track.album?.title || 'Без альбома'}</span>
          {' • '}
          {track.album?.artist ? (
            <Link 
              to={`/artists/${encodeURIComponent(track.album.artist)}`} 
              className="track-artist-link"
              onClick={(e) => e.stopPropagation()}
            >
              {track.album.artist}
            </Link>
          ) : (
            <span className="track-artist-link">Неизвестный артист</span>
          )}
        </p>
        {track.genres && track.genres.length > 0 && (
          <div className="track-card-genres">
            {track.genres.map((genre) => (
              <span key={genre.id} className="track-card-genre-badge">
                {genre.name}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default TrackCard;

