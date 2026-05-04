import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { reviewsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { getImageUrl } from '../utils/imageUtils';
import ReviewScoresStrip from './ReviewScoresStrip';
import './ReviewCardSmall.css';

const ReviewCardSmall = ({ review, onUpdate }) => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [likeCount, setLikeCount] = useState(review.likes?.length || 0);
  const [isLiked, setIsLiked] = useState(false);
  const hasArtistMark = review.has_artist_mark || (review.artist_mark_usernames || []).length > 0 ||
    (review.likes || []).some((like) => like.user?.is_verified_artist);

  useEffect(() => {
    setLikeCount(review.likes?.length || 0);
    if (user && review.likes) {
      setIsLiked(review.likes.some((like) => like.user_id === user.id));
    }
  }, [review.likes, user]);

  const handleClick = () => {
    if (review.album_id) {
      navigate(`/albums/${review.album_id}`);
    } else if (review.track_id) {
      navigate(`/tracks/${review.track_id}`);
    }
  };

  const handleLike = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (process.env.REACT_APP_USE_MOCK === 'true' || !process.env.REACT_APP_API_URL) {
      setLikeCount((prev) => prev + 1);
      setIsLiked(true);
      return;
    }
    try {
      await reviewsAPI.like(review.id);
      setLikeCount((prev) => prev + 1);
      setIsLiked(true);
      if (onUpdate) onUpdate();
    } catch (err) {
      console.error('Error liking review:', err);
    }
  };

  const handleUnlike = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (process.env.REACT_APP_USE_MOCK === 'true' || !process.env.REACT_APP_API_URL) {
      setLikeCount((prev) => Math.max(0, prev - 1));
      setIsLiked(false);
      return;
    }
    try {
      await reviewsAPI.unlike(review.id);
      setLikeCount((prev) => Math.max(0, prev - 1));
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
          {review.user?.id ? (
            <Link
              to={`/users/${review.user.id}`}
              className="review-card-small-user-link"
              onClick={(e) => e.stopPropagation()}
            >
              {review.user?.username || 'Пользователь'}
            </Link>
          ) : (
            review.user?.username || 'Пользователь'
          )}
        </div>
        <div className="review-card-small-scores">
          <ReviewScoresStrip review={review} size="small" />
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
              <div className="review-card-small-image-placeholder">♪</div>
            )}
            <div className="review-card-small-cover-overlay">
              {likeCount > 0 && (
                <div className="review-card-small-stat-item">
                  <span className="stat-icon">❤</span>
                  <span className="stat-count">{likeCount}</span>
                </div>
              )}
              <button
                className={`review-card-small-like-button ${isLiked ? 'liked' : ''}`}
                onClick={isLiked ? handleUnlike : handleLike}
                title={isLiked ? 'Убрать лайк' : 'Поставить лайк'}
              >
                {isLiked ? '❤' : '♡'}
              </button>
            </div>
          </div>
          <div className="review-card-small-album-info">
            <div className="review-card-small-album-title">{review.album.title}</div>
            <div className="review-card-small-album-artist">{review.album.artist}</div>
            {hasArtistMark && (
              <div className="review-card-small-artist-mark">
                <span>★</span>
                Отмечено артистом
              </div>
            )}
          </div>
        </div>
      )}
      {review.text && (
        <div className="review-card-small-text">
          {review.text.length > 100 ? `${review.text.substring(0, 100)}...` : review.text}
        </div>
      )}
    </div>
  );
};

export default ReviewCardSmall;
