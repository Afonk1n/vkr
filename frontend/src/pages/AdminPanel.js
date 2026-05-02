import React, { useCallback, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { reviewsAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import './AdminPanel.css';

const AdminPanel = () => {
  const { isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();
  const [pendingReviews, setPendingReviews] = useState([]);
  const [status, setStatus] = useState('pending');
  const [stats, setStats] = useState({ pending: 0, approved: 0, rejected: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchReviews = useCallback(async (nextStatus = status) => {
    setLoading(true);
    setError('');
    try {
      const [current, pending, approved, rejected] = await Promise.all([
        reviewsAPI.getAll({ status: nextStatus, page_size: 30 }),
        reviewsAPI.getAll({ status: 'pending', page_size: 1 }),
        reviewsAPI.getAll({ status: 'approved', page_size: 1 }),
        reviewsAPI.getAll({ status: 'rejected', page_size: 1 }),
      ]);
      setPendingReviews(current.data.reviews || []);
      setStats({
        pending: pending.data.total || 0,
        approved: approved.data.total || 0,
        rejected: rejected.data.total || 0,
      });
    } catch (err) {
      setError('Ошибка загрузки рецензий');
      console.error('Error fetching reviews:', err);
    } finally {
      setLoading(false);
    }
  }, [status]);

  useEffect(() => {
    if (!isAuthenticated || !isAdmin) {
      navigate('/feed');
      return;
    }
    fetchReviews(status);
  }, [isAuthenticated, isAdmin, navigate, status, fetchReviews]);

  const handleApprove = async (reviewId) => {
    try {
      await reviewsAPI.approve(reviewId);
      fetchReviews(status);
    } catch (err) {
      setError('Ошибка при одобрении рецензии');
      console.error('Error approving review:', err);
    }
  };

  const handleReject = async (reviewId) => {
    try {
      await reviewsAPI.reject(reviewId);
      fetchReviews(status);
    } catch (err) {
      setError('Ошибка при отклонении рецензии');
      console.error('Error rejecting review:', err);
    }
  };

  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return (
    <div className="container">
      <div className="admin-panel">
        <div className="admin-header">
          <div>
            <span className="admin-kicker">Модерация</span>
            <h1>Панель администратора</h1>
            <p>Проверка рецензий, быстрые статусы и аккуратное принятие решений.</p>
          </div>
        </div>

        <div className="admin-stats-grid" aria-label="Статусы рецензий">
          <button type="button" className={`admin-stat-card ${status === 'pending' ? 'admin-stat-card--active' : ''}`} onClick={() => setStatus('pending')}>
            <span>На проверке</span>
            <strong>{stats.pending}</strong>
          </button>
          <button type="button" className={`admin-stat-card ${status === 'approved' ? 'admin-stat-card--active' : ''}`} onClick={() => setStatus('approved')}>
            <span>Одобрено</span>
            <strong>{stats.approved}</strong>
          </button>
          <button type="button" className={`admin-stat-card ${status === 'rejected' ? 'admin-stat-card--active' : ''}`} onClick={() => setStatus('rejected')}>
            <span>Отклонено</span>
            <strong>{stats.rejected}</strong>
          </button>
        </div>

        {error && <div className="error-message">{error}</div>}

        {loading ? (
          <div className="loading">Загрузка...</div>
        ) : pendingReviews.length === 0 ? (
          <div className="empty-state admin-empty-state">
            {status === 'pending' ? 'Нет рецензий на модерации' : 'В этом статусе пока пусто'}
          </div>
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
                      Одобрить
                    </button>
                    <button
                      onClick={() => handleReject(review.id)}
                      className="btn-reject"
                      title="Отклонить"
                    >
                      Отклонить
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
