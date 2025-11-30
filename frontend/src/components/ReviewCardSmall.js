import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { reviewsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { getImageUrl } from '../utils/imageUtils';
import './ReviewCardSmall.css';

const ReviewCardSmall = ({ review, onUpdate }) => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [likeCount, setLikeCount] = useState(review.likes?.length || 0);
  const [isLiked, setIsLiked] = useState(false);

  useEffect(() => {
    setLikeCount(review.likes?.length || 0);
    if (user && review.likes) {
      setIsLiked(review.likes.some(like => like.user_id === user.id));
    }
  }, [review.likes, user]);

  const handleClick = () => {
    navigate(`/albums/${review.album_id}`);
  };

  const handleLike = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      await reviewsAPI.like(review.id);
      setLikeCount(prev => prev + 1);
      setIsLiked(true);
      if (onUpdate) onUpdate();
    } catch (err) {
      console.error('Error liking review:', err);
    }
  };

  const handleUnlike = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      await reviewsAPI.unlike(review.id);
      setLikeCount(prev => Math.max(0, prev - 1));
      setIsLiked(false);
      if (onUpdate) onUpdate();
    } catch (err) {
      console.error('Error unliking review:', err);
    }
  };

  return (
    <div className="review-card-small" onClick={handleClick}>
      <div className="review-card-small-header">
        <div className="review-card-small-user">
          {review.user?.username || 'Пользователь'}
        </div>
        <div className="review-card-small-score">
          {review.final_score?.toFixed(1) || '0.0'}
        </div>
      </div>
      {review.album && (
        <div className="review-card-small-album">
          <div className="review-card-small-cover">
            {getImageUrl(review.album.cover_image_path) ? (
              <img
                src={getImageUrl(review.album.cover_image_path)}
                alt={review.album.title}
                className="review-card-small-image"
              />
            ) : (
              <div className="review-card-small-image-placeholder">🎵</div>
            )}
            <div className="review-card-small-cover-overlay">
              {likeCount > 0 && (
                <div className="review-card-small-stat-item">
                  <span className="stat-icon">❤️</span>
                  <span className="stat-count">{likeCount}</span>
                </div>
              )}
              <button
                className={`review-card-small-like-button ${isLiked ? 'liked' : ''}`}
                onClick={isLiked ? handleUnlike : handleLike}
                title={isLiked ? 'Убрать лайк' : 'Поставить лайк'}
              >
                {isLiked ? '❤️' : '🤍'}
              </button>
            </div>
          </div>
          <div className="review-card-small-album-info">
            <div className="review-card-small-album-title">{review.album.title}</div>
            <div className="review-card-small-album-artist">{review.album.artist}</div>
          </div>
        </div>
      )}
      {review.text && (
        <div className="review-card-small-text">
          {review.text.length > 100 ? review.text.substring(0, 100) + '...' : review.text}
        </div>
      )}
    </div>
  );
};

export default ReviewCardSmall;

