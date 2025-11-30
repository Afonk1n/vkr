import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { reviewsAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import './AdminPanel.css';

const AdminPanel = () => {
  const { isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();
  const [pendingReviews, setPendingReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!isAuthenticated || !isAdmin) {
      navigate('/');
      return;
    }
    fetchPendingReviews();
  }, [isAuthenticated, isAdmin, navigate]);

  const fetchPendingReviews = async () => {
    setLoading(true);
    try {
      const response = await reviewsAPI.getAll({ status: 'pending' });
      setPendingReviews(response.data.reviews);
    } catch (err) {
      setError('Ошибка загрузки рецензий');
      console.error('Error fetching reviews:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (reviewId) => {
    try {
      await reviewsAPI.approve(reviewId);
      fetchPendingReviews();
    } catch (err) {
      alert('Ошибка при одобрении рецензии');
      console.error('Error approving review:', err);
    }
  };

  const handleReject = async (reviewId) => {
    if (window.confirm('Вы уверены, что хотите отклонить эту рецензию?')) {
      try {
        await reviewsAPI.reject(reviewId);
        fetchPendingReviews();
      } catch (err) {
        alert('Ошибка при отклонении рецензии');
        console.error('Error rejecting review:', err);
      }
    }
  };

  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return (
    <div className="container">
      <div className="admin-panel">
        <div className="admin-header">
          <h1>Панель администратора</h1>
          <p>Рецензии на модерации</p>
        </div>

        {error && <div className="error-message">{error}</div>}
        
        {loading ? (
          <div className="loading">Загрузка...</div>
        ) : pendingReviews.length === 0 ? (
          <div className="empty-state">Нет рецензий на модерации</div>
        ) : (
          <div className="reviews-list">
            {pendingReviews.map((review) => (
              <ReviewCard 
                key={review.id} 
                review={review}
                hideLike={true}
                moderationActions={
                  <div className="moderation-actions">
                    <button
                      onClick={() => handleApprove(review.id)}
                      className="btn-approve"
                      title="Одобрить"
                    >
                      👍
                    </button>
                    <button
                      onClick={() => handleReject(review.id)}
                      className="btn-reject"
                      title="Отклонить"
                    >
                      👎
                    </button>
                  </div>
                }
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminPanel;

