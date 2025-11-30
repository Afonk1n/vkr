import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { albumsAPI, reviewsAPI, tracksAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import ReviewForm from '../components/ReviewForm';
import ReviewCard from '../components/ReviewCard';
import LikeButton from '../components/LikeButton';
import TrackCard from '../components/TrackCard';
import { getImageUrl } from '../utils/imageUtils';
import './AlbumDetailPage.css';

const AlbumDetailPage = () => {
  const { id } = useParams();
  const { isAuthenticated } = useAuth();
  const [album, setAlbum] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [tracks, setTracks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [editingReview, setEditingReview] = useState(null);
  const [coverImageError, setCoverImageError] = useState(false);

  const fetchAlbum = useCallback(async () => {
    try {
      const response = await albumsAPI.getById(id);
      setAlbum(response.data);
    } catch (err) {
      setError('Ошибка загрузки альбома');
      console.error('Error fetching album:', err);
    }
  }, [id]);

  const fetchTracks = useCallback(async () => {
    try {
      const response = await tracksAPI.getByAlbum(id);
      setTracks(response.data);
    } catch (err) {
      console.error('Error fetching tracks:', err);
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
    fetchTracks();
  }, [fetchAlbum, fetchReviews, fetchTracks]);

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

  const handleAlbumLike = async () => {
    try {
      await albumsAPI.like(album.id);
      fetchAlbum();
    } catch (err) {
      throw err;
    }
  };

  const handleAlbumUnlike = async () => {
    try {
      await albumsAPI.unlike(album.id);
      fetchAlbum();
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
            {getImageUrl(album.cover_image_path) && !coverImageError ? (
              <img 
                src={getImageUrl(album.cover_image_path)} 
                alt={album.title}
                onError={() => setCoverImageError(true)}
              />
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
            <div className="album-actions-large">
              <LikeButton
                item={album}
                itemType="album"
                onLike={handleAlbumLike}
                onUnlike={handleAlbumUnlike}
              />
            </div>
            {album.description && (
              <p className="album-description">{album.description}</p>
            )}
          </div>
        </div>

        {/* Tracks Section */}
        {tracks.length > 0 && (
          <div className="tracks-section">
            <h2 className="section-title">Треки ({tracks.length})</h2>
            <div className="tracks-list-album">
              {tracks.map((track) => (
                <div key={track.id} className="track-item-album">
                  <div className="track-item-number">{track.track_number || '-'}</div>
                  <div className="track-item-info">
                    <div className="track-item-title">{track.title}</div>
                    {track.genres && track.genres.length > 0 && (
                      <div className="track-item-genres">
                        {track.genres.map((genre) => (
                          <span key={genre.id} className="track-item-genre-badge">
                            {genre.name}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                  <div className="track-item-duration">
                    {track.duration ? `${Math.floor(track.duration / 60)}:${(track.duration % 60).toString().padStart(2, '0')}` : '-'}
                  </div>
                  <div className="track-item-link">
                    <a href={`/tracks/${track.id}`} className="track-link-button">→</a>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

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

export default AlbumDetailPage;

