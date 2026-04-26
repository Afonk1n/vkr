import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { albumsAPI, reviewsAPI, tracksAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import ReviewForm from '../components/ReviewForm';
import ReviewCard from '../components/ReviewCard';
import LikeButton from '../components/LikeButton';
import AverageScoreBadge from '../components/AverageScoreBadge';
import { getImageUrl } from '../utils/imageUtils';
import './AlbumDetailPage.css';

const formatDuration = (duration) => (
  duration ? `${Math.floor(duration / 60)}:${(duration % 60).toString().padStart(2, '0')}` : '-'
);

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
    setCoverImageError(false);
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
      fetchAlbum();
      fetchTracks();
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
        fetchTracks();
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
              <div className="album-cover-placeholder-large">♪</div>
            )}
          </div>
          <div className="album-info-large">
            <h1 className="album-title-large">{album.title}</h1>
            <p className="album-artist-large">
              <Link to={`/artists/${encodeURIComponent(album.artist)}`} className="album-artist-link">
                {album.artist}
              </Link>
            </p>
            {album.genre && <span className="album-genre-large">{album.genre.name}</span>}
            <AverageScoreBadge source={album} reviews={reviews} className="album-average-score" />
            <div className="album-actions-large">
              <LikeButton
                item={album}
                itemType="album"
                onLike={handleAlbumLike}
                onUnlike={handleAlbumUnlike}
              />
            </div>
            {album.description && <p className="album-description">{album.description}</p>}
          </div>
        </div>

        {tracks.length > 0 && (
          <div className="tracks-section">
            <h2 className="section-title">Треки ({tracks.length})</h2>
            <div className="tracks-list-album">
              {tracks.map((track) => {
                const trackCoverUrl = track.cover_image_path ? getImageUrl(track.cover_image_path) : null;
                const albumCoverUrl = album.cover_image_path ? getImageUrl(album.cover_image_path) : null;
                const coverUrl = trackCoverUrl || albumCoverUrl;

                return (
                  <Link key={track.id} to={`/tracks/${track.id}`} className="track-item-album">
                    <div className="track-item-number">{track.track_number || '-'}</div>
                    <div className="track-item-cover">
                      {coverUrl ? (
                        <img
                          src={coverUrl}
                          alt={track.title}
                          onError={(e) => {
                            e.currentTarget.style.display = 'none';
                            e.currentTarget.nextSibling.style.display = 'flex';
                          }}
                        />
                      ) : null}
                      <div className="track-item-cover-placeholder" style={{ display: coverUrl ? 'none' : 'flex' }}>
                        ♪
                      </div>
                    </div>
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
                    <AverageScoreBadge source={track} size="small" className="track-item-score" />
                    <div className="track-item-duration">{formatDuration(track.duration)}</div>
                  </Link>
                );
              })}
            </div>
          </div>
        )}

        <div className="reviews-section">
          <div className="reviews-header">
            <h2>Рецензии ({reviews.length})</h2>
            {isAuthenticated && !showReviewForm && (
              <button onClick={() => setShowReviewForm(true)} className="btn-edit">
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
