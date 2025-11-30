import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { albumsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { getImageUrl } from '../utils/imageUtils';
import './AlbumCard.css';

const AlbumCard = ({ album, onUpdate }) => {
  const [imageError, setImageError] = useState(false);
  const handleLike = async () => {
    try {
      await albumsAPI.like(album.id);
      if (onUpdate) onUpdate();
    } catch (err) {
      throw err;
    }
  };

  const handleUnlike = async () => {
    try {
      await albumsAPI.unlike(album.id);
      if (onUpdate) onUpdate();
    } catch (err) {
      throw err;
    }
  };

  const { user } = useAuth();
  const [likeCount, setLikeCount] = useState(album.likes?.length || 0);
  const [isLiked, setIsLiked] = useState(false);

  useEffect(() => {
    setLikeCount(album.likes?.length || 0);
    if (user && album.likes) {
      setIsLiked(album.likes.some(like => like.user_id === user.id));
    }
  }, [album.likes, user]);

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

  const imageUrl = getImageUrl(album.cover_image_path);

  return (
    <Link to={`/albums/${album.id}`} className="album-card">
      <div className="album-cover">
        {imageUrl && !imageError ? (
          <img 
            src={imageUrl} 
            alt={album.title}
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="album-cover-placeholder">🎵</div>
        )}
        {/* Лайки и счетчики прямо на обложке, как в Figma */}
        <div className="album-cover-overlay">
          <div className="album-stats">
            {likeCount > 0 && (
              <div className="album-stat-item">
                <span className="stat-icon">❤️</span>
                <span className="stat-count">{likeCount}</span>
              </div>
            )}
            {album.average_rating > 0 && (
              <div className="album-stat-item">
                <span className="stat-icon">⭐</span>
                <span className="stat-count">{Math.round(album.average_rating)}</span>
              </div>
            )}
          </div>
          <button
            className={`album-like-button ${isLiked ? 'liked' : ''}`}
            onClick={handleLikeClick}
            title="Поставить лайк"
          >
            {isLiked ? '❤️' : '🤍'}
          </button>
        </div>
      </div>
      <div className="album-info">
        <h3 className="album-title">{album.title}</h3>
        <p className="album-artist">{album.artist}</p>
        {album.genre && (
          <span className="album-genre">{album.genre.name}</span>
        )}
      </div>
    </Link>
  );
};

export default AlbumCard;

