import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { usersAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import { getImageUrl } from '../utils/imageUtils';
import './UserProfilePage.css';

const UserProfilePage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchUserData();
    fetchUserReviews();
  }, [id]);

  const fetchUserData = async () => {
    try {
      const response = await usersAPI.getById(id);
      setUser(response.data);
    } catch (err) {
      setError('Пользователь не найден');
      console.error('Error fetching user:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchUserReviews = async () => {
    try {
      const response = await usersAPI.getUserReviews(id);
      setReviews(response.data.reviews);
    } catch (err) {
      console.error('Error fetching reviews:', err);
    }
  };

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="container">
        <div className="error-message">{error || 'Пользователь не найден'}</div>
        <button onClick={() => navigate(-1)} className="btn-back">
          Назад
        </button>
      </div>
    );
  }

  const socialLinks = user?.social_links 
    ? (typeof user.social_links === 'string' 
        ? JSON.parse(user.social_links) 
        : user.social_links)
    : {};

  return (
    <div className="container">
      <div className="user-profile-page">
        <button onClick={() => navigate(-1)} className="btn-back">
          ← Назад
        </button>
        
        <div className="profile-header-card">
          <div className="profile-avatar-section">
            {user?.avatar_path && getImageUrl(user.avatar_path) ? (
              <img 
                src={getImageUrl(user.avatar_path)} 
                alt={user.username}
                className="profile-avatar"
              />
            ) : (
              <div className="profile-avatar-placeholder">
                {user?.username?.charAt(0).toUpperCase() || 'U'}
              </div>
            )}
          </div>
          <div className="profile-info-section">
            <h1 className="profile-username">{user?.username || 'Пользователь'}</h1>
            {user?.email && (
              <p className="profile-email">{user.email}</p>
            )}
            {user?.is_admin && (
              <span className="admin-badge">Администратор</span>
            )}
            {user?.created_at && (
              <p className="profile-joined">
                Присоединился: {new Date(user.created_at).toLocaleDateString('ru-RU')}
              </p>
            )}
            {user?.bio && (
              <div className="profile-bio">
                <p>{user.bio}</p>
              </div>
            )}
            {(socialLinks.vk || socialLinks.telegram || socialLinks.instagram) && (
              <div className="profile-social-links">
                {socialLinks.vk && (
                  <a href={socialLinks.vk} target="_blank" rel="noopener noreferrer" className="social-link">
                    VK
                  </a>
                )}
                {socialLinks.telegram && (
                  <a href={`https://t.me/${socialLinks.telegram.replace('@', '')}`} target="_blank" rel="noopener noreferrer" className="social-link">
                    Telegram
                  </a>
                )}
                {socialLinks.instagram && (
                  <a href={socialLinks.instagram} target="_blank" rel="noopener noreferrer" className="social-link">
                    Instagram
                  </a>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="profile-reviews">
          <h2>Рецензии пользователя ({reviews.length})</h2>
          {loading ? (
            <div className="loading">Загрузка...</div>
          ) : reviews.length === 0 ? (
            <div className="empty-state">У пользователя пока нет рецензий</div>
          ) : (
            <div className="reviews-list">
              {reviews.map((review) => (
                <ReviewCard
                  key={review.id}
                  review={review}
                  hideLike={false}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default UserProfilePage;

