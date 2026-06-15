import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { reviewsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { getImageUrl } from '../utils/imageUtils';
import ReviewScoresStrip from './ReviewScoresStrip';
import './ReviewCardSmall.css';

const ReviewCardSmall = ({ review }) => {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();
  const [likeCount, setLikeCount] = useState(review.likes?.length || 0);
  const [isLiked, setIsLiked] = useState(false);
  const [likeBusy, setLikeBusy] = useState(false);
  const hasArtistMark = review.has_artist_mark || (review.artist_mark_usernames || []).length > 0 ||
    (review.likes || []).some((like) => like.user?.is_verified_artist);
  const target = review.album
    ? {
      title: review.album.title,
      subtitle: review.album.artist,
      cover: review.album.cover_image_path,
      alt: review.album.title,
    }
    : review.track
      ? {
        title: review.track.title,
        subtitle: review.track.album?.artist
          ? `${review.track.album.title || 'Альбом'} · ${review.track.album.artist}`
          : review.track.album?.title || 'Трек',
        cover: review.track.cover_image_path || review.track.album?.cover_image_path,
        alt: review.track.title,
      }
      : null;

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

  // Один оптимистичный обработчик: UI меняется сразу, при ошибке — откат.
  // Полный рефетч ленты (onUpdate) больше не дёргаем на каждый лайк.
  const toggleLike = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (likeBusy) return;

    const isMock = process.env.REACT_APP_USE_MOCK === 'true' || !process.env.REACT_APP_API_URL;
    const prevLiked = isLiked;
    const prevCount = likeCount;
    const nextLiked = !prevLiked;

    setIsLiked(nextLiked);
    setLikeCount((c) => (nextLiked ? c + 1 : Math.max(0, c - 1)));
    if (isMock) return;

    setLikeBusy(true);
    try {
      if (nextLiked) {
        await reviewsAPI.like(review.id);
      } else {
        await reviewsAPI.unlike(review.id);
      }
    } catch (err) {
      console.error('Error toggling review like:', err);
      setIsLiked(prevLiked);
      setLikeCount(prevCount);
    } finally {
      setLikeBusy(false);
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
      {target && (
        <div className="review-card-small-album">
          <div className="review-card-small-cover">
            {getImageUrl(target.cover) ? (
              <img
                src={getImageUrl(target.cover)}
                alt={target.alt}
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
                onClick={toggleLike}
                onMouseDown={(e) => e.preventDefault()}
                disabled={!isAuthenticated || likeBusy}
                aria-pressed={isLiked}
                title={
                  !isAuthenticated
                    ? 'Войдите, чтобы ставить лайки'
                    : isLiked
                      ? 'Убрать лайк'
                      : 'Поставить лайк'
                }
              >
                {isLiked ? '❤' : '♡'}
              </button>
            </div>
          </div>
          <div className="review-card-small-album-info">
            <div className="review-card-small-album-title">{target.title}</div>
            <div className="review-card-small-album-artist">{target.subtitle}</div>
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
