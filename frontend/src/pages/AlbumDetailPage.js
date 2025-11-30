import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { albumsAPI, reviewsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import ReviewForm from '../components/ReviewForm';
import ReviewCard from '../components/ReviewCard';
import './AlbumDetailPage.css';

const AlbumDetailPage = () => {
  const { id } = useParams();
  const { isAuthenticated } = useAuth();
  const [album, setAlbum] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [editingReview, setEditingReview] = useState(null);

  const fetchAlbum = useCallback(async () => {
    try {
      const response = await albumsAPI.getById(id);
      setAlbum(response.data);
    } catch (err) {
      setError('Ошибка загрузки альбома');
      console.error('Error fetching album:', err);
    }
  }, [id]);

  const fetchReviews = useCallback(async () => {
    try {
      const response = await reviewsAPI.getAll({ album_id: id });
      setReviews(response.data.reviews);
    } catch (err) {
      console.error('Error fetching reviews:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchAlbum();
    fetchReviews();
  }, [fetchAlbum, fetchReviews]);

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
      fetchAlbum(); // Refresh album to update average rating
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
        fetchAlbum();
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

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  if (error || !album) {
    return (
      <div className="container">
        <div className="error-message">{error || 'Альбом не найден'}</div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="album-detail">
        <div className="album-header">
          <div className="album-cover-large">
            {album.cover_image_path ? (
              <img src={album.cover_image_path} alt={album.title} />
            ) : (
              <div className="album-cover-placeholder-large">🎵</div>
            )}
          </div>
          <div className="album-info-large">
            <h1 className="album-title-large">{album.title}</h1>
            <p className="album-artist-large">{album.artist}</p>
            {album.genre && (
              <span className="album-genre-large">{album.genre.name}</span>
            )}
            {album.average_rating > 0 && (
              <div className="album-rating-large">
                ⭐ Средний рейтинг: {Math.round(album.average_rating)}
              </div>
            )}
            {album.description && (
              <p className="album-description">{album.description}</p>
            )}
          </div>
        </div>

        <div className="reviews-section">
          <div className="reviews-header">
            <h2>Рецензии ({reviews.length})</h2>
            {isAuthenticated && !showReviewForm && (
              <button
                onClick={() => setShowReviewForm(true)}
                className="btn-primary"
              >
                Добавить рецензию
              </button>
            )}
          </div>

          {showReviewForm && (
            <ReviewForm
              albumId={album.id}
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
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AlbumDetailPage;

