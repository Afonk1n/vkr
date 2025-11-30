import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { usersAPI, reviewsAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import './ProfilePage.css';

const ProfilePage = () => {
  const { user, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchUserReviews = useCallback(async () => {
    if (!user) return;
    
    setLoading(true);
    try {
      const response = await usersAPI.getUserReviews(user.id);
      setReviews(response.data.reviews);
    } catch (err) {
      setError('Ошибка загрузки рецензий');
      console.error('Error fetching reviews:', err);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    fetchUserReviews();
  }, [isAuthenticated, navigate, fetchUserReviews]);

  const handleEditReview = (review) => {
    navigate(`/albums/${review.album_id}`);
  };

  const handleDeleteReview = async (reviewId) => {
    if (window.confirm('Вы уверены, что хотите удалить эту рецензию?')) {
      try {
        await reviewsAPI.delete(reviewId);
        fetchUserReviews();
      } catch (err) {
        alert('Ошибка при удалении рецензии');
        console.error('Error deleting review:', err);
      }
    }
  };

  if (!isAuthenticated || !user) {
    return null;
  }

  return (
    <div className="container">
      <div className="profile-page">
        <div className="profile-header">
          <h1>Профиль пользователя</h1>
          <div className="profile-info">
            <p><strong>Имя пользователя:</strong> {user.username}</p>
            <p><strong>Email:</strong> {user.email}</p>
            {user.is_admin && (
              <p className="admin-badge">Администратор</p>
            )}
          </div>
        </div>

        <div className="profile-reviews">
          <h2>Мои рецензии ({reviews.length})</h2>
          {error && <div className="error-message">{error}</div>}
          {loading ? (
            <div className="loading">Загрузка...</div>
          ) : reviews.length === 0 ? (
            <div className="empty-state">У вас пока нет рецензий</div>
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

export default ProfilePage;

