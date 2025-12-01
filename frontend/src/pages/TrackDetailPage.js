import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { tracksAPI, reviewsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import ReviewForm from '../components/ReviewForm';
import ReviewCard from '../components/ReviewCard';
import LikeButton from '../components/LikeButton';
import { getImageUrl } from '../utils/imageUtils';
import './TrackDetailPage.css';

const TrackDetailPage = () => {
  const { id } = useParams();
  const { isAuthenticated } = useAuth();
  const [track, setTrack] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [editingReview, setEditingReview] = useState(null);
  const [coverImageError, setCoverImageError] = useState(false);

  const fetchTrack = useCallback(async () => {
    try {
      const response = await tracksAPI.getById(id);
      setTrack(response.data);
    } catch (err) {
      setError('Ошибка загрузки трека');
      console.error('Error fetching track:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  const fetchReviews = useCallback(async () => {
    try {
      const response = await reviewsAPI.getAll({ track_id: id });
      setReviews(response.data.reviews);
    } catch (err) {
      console.error('Error fetching reviews:', err);
    }
  }, [id]);

  useEffect(() => {
    fetchTrack();
    fetchReviews();
    setCoverImageError(false); // Reset error when track changes
  }, [fetchTrack, fetchReviews]);

  const handleReviewSubmit = async (reviewData) => {
    try {
      if (editingReview) {
        await reviewsAPI.update(editingReview.id, reviewData);
      } else {
        await reviewsAPI.create(reviewData);
      }
      setShowReviewForm(false);
      setEditingReview(null);
      fetchReviews();
    } catch (err) {
      throw err;
    }
  };

  const handleEditReview = (review) => {
    setEditingReview(review);
    setShowReviewForm(true);
  };

  const handleDeleteReview = async (reviewId) => {
    if (window.confirm('Вы уверены, что хотите удалить эту рецензию?')) {
      try {
        await reviewsAPI.delete(reviewId);
        fetchReviews();
      } catch (err) {
        alert('Ошибка при удалении рецензии');
        console.error('Error deleting review:', err);
      }
    }
  };

  const handleCancelEdit = () => {
    setEditingReview(null);
    setShowReviewForm(false);
  };

  const handleTrackLike = async () => {
    try {
      await tracksAPI.like(track.id);
      fetchTrack();
    } catch (err) {
      throw err;
    }
  };

  const handleTrackUnlike = async () => {
    try {
      await tracksAPI.unlike(track.id);
      fetchTrack();
    } catch (err) {
      throw err;
    }
  };

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  if (error || !track) {
    return (
      <div className="container">
        <div className="error-message">{error || 'Трек не найден'}</div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="track-detail">
        <div className="track-header">
          <div className="track-cover-large">
            {getImageUrl(track.cover_image_path || track.album?.cover_image_path) && !coverImageError ? (
              <img 
                src={getImageUrl(track.cover_image_path || track.album?.cover_image_path)} 
                alt={track.album?.title || track.title}
                onError={() => setCoverImageError(true)}
              />
            ) : null}
            <div className="track-cover-placeholder-large" style={{ display: (!getImageUrl(track.cover_image_path) && !getImageUrl(track.album?.cover_image_path)) || coverImageError ? 'flex' : 'none' }}>
              🎵
            </div>
          </div>
          <div className="track-info-large">
            <h1 className="track-title-large">{track.title}</h1>
            {track.album && (
              <div className="track-album-info">
                <span className="track-album-label">Альбом:</span>{' '}
                <Link to={`/albums/${track.album.id}`} className="track-album-link">
                  {track.album.title}
                </Link>
                {track.album.artist && (
                  <>
                    {' • '}
                    <Link 
                      to={`/artists/${encodeURIComponent(track.album.artist)}`} 
                      className="track-artist-link"
                    >
                      {track.album.artist}
                    </Link>
                  </>
                )}
              </div>
            )}
            {track.duration && (
              <div className="track-duration">
                Длительность: {Math.floor(track.duration / 60)}:{(track.duration % 60).toString().padStart(2, '0')}
              </div>
            )}
            {track.track_number && (
              <div className="track-number">
                Номер трека: {track.track_number}
              </div>
            )}
            {track.genres && track.genres.length > 0 && (
              <div className="track-genres">
                <span className="track-genres-label">Жанры:</span>
                <div className="track-genres-list">
                  {track.genres.map((genre) => (
                    <span key={genre.id} className="track-genre-badge">
                      {genre.name}
                    </span>
                  ))}
                </div>
              </div>
            )}
            <div className="track-actions-large">
              <LikeButton
                item={track}
                itemType="track"
                onLike={handleTrackLike}
                onUnlike={handleTrackUnlike}
              />
            </div>
          </div>
        </div>

        {track.album && (
          <div className="track-album-section">
            <h2 className="section-title">Альбом</h2>
            <Link to={`/albums/${track.album.id}`} className="album-link-card">
              {getImageUrl(track.album.cover_image_path) && (
                <img
                  src={getImageUrl(track.album.cover_image_path)}
                  alt={track.album.title}
                  className="album-link-image"
                />
              )}
              <div className="album-link-info">
                <h3>{track.album.title}</h3>
                <p>{track.album.artist}</p>
                {track.album.genre && (
                  <span className="album-link-genre">{track.album.genre.name}</span>
                )}
              </div>
            </Link>
          </div>
        )}

        <div className="reviews-section">
          <div className="reviews-header">
            <h2>Рецензии ({reviews.length})</h2>
            {isAuthenticated && !showReviewForm && (
              <button
                onClick={() => setShowReviewForm(true)}
                className="btn-edit"
              >
                Добавить рецензию
              </button>
            )}
          </div>

          {showReviewForm && (
            <ReviewForm
              trackId={track.id}
              onSubmit={handleReviewSubmit}
              initialData={editingReview}
              onCancel={editingReview ? handleCancelEdit : () => setShowReviewForm(false)}
            />
          )}

          {reviews.length === 0 ? (
            <div className="empty-state">Пока нет рецензий</div>
          ) : (
            <div className="reviews-list">
              {reviews.map((review) => (
                <ReviewCard
                  key={review.id}
                  review={review}
                  onEdit={handleEditReview}
                  onDelete={handleDeleteReview}
                  onUpdate={fetchReviews}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TrackDetailPage;

