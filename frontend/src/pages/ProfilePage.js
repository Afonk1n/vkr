import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { usersAPI, reviewsAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import ProfileEditForm from '../components/ProfileEditForm';
import BadgeList from '../components/BadgeList';
import { getImageUrl } from '../utils/imageUtils';
import './ProfilePage.css';

const ProfilePage = () => {
  const { user, isAuthenticated, updateUser } = useAuth();
  const navigate = useNavigate();
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [currentUser, setCurrentUser] = useState(user);

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
    fetchCurrentUser();
  }, [isAuthenticated, navigate, fetchUserReviews]);

  const fetchCurrentUser = async () => {
    if (!user) return;
    try {
      const response = await usersAPI.getById(user.id);
      setCurrentUser(response.data);
    } catch (err) {
      console.error('Error fetching user:', err);
    }
  };

  const handleSaveProfile = async (profileData) => {
    try {
      const response = await usersAPI.update(user.id, profileData);
      const updatedUser = response.data;
      setCurrentUser(updatedUser);
      setIsEditing(false);
      
      // Update auth context without re-login
      // Only re-login if password was actually changed
      if (profileData.password && profileData.password.trim() !== '') {
        // If password was changed, we need to re-login with new password
        // But we don't have the new password here, so we'll just update user data
        // The user will need to log in again on next session if password changed
        if (updateUser) {
          updateUser(updatedUser);
        }
      } else {
        // No password change, just update user data
        if (updateUser) {
          updateUser(updatedUser);
        }
      }
      
      alert('Профиль успешно обновлен');
    } catch (err) {
      alert('Ошибка при обновлении профиля');
      console.error('Error updating profile:', err);
    }
  };

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

  const socialLinks = currentUser?.social_links 
    ? (typeof currentUser.social_links === 'string' 
        ? JSON.parse(currentUser.social_links) 
        : currentUser.social_links)
    : {};

  return (
    <div className="container">
      <div className="profile-page">
        {isEditing ? (
          <ProfileEditForm
            user={currentUser}
            onSave={handleSaveProfile}
            onCancel={() => setIsEditing(false)}
            updateUser={updateUser}
          />
        ) : (
          <>
            <div className="profile-header-card">
              <div className="profile-avatar-section">
                {currentUser?.avatar_path && getImageUrl(currentUser.avatar_path) ? (
                  <img 
                    src={getImageUrl(currentUser.avatar_path)} 
                    alt={currentUser.username}
                    className="profile-avatar"
                  />
                ) : (
                  <div className="profile-avatar-placeholder">
                    {currentUser?.username?.charAt(0).toUpperCase() || 'U'}
                  </div>
                )}
                <button 
                  className="btn-edit-profile"
                  onClick={() => setIsEditing(true)}
                >
                  Редактировать профиль
                </button>
              </div>
              <div className="profile-info-section">
                <h1 className="profile-username">{currentUser?.username || user?.username || 'Пользователь'}</h1>
                {currentUser?.email && (
                  <p className="profile-email">{currentUser.email}</p>
                )}
                {currentUser?.is_admin && (
                  <span className="admin-badge">Администратор</span>
                )}
                {currentUser?.badges && currentUser.badges.length > 0 && (
                  <BadgeList badges={currentUser.badges} />
                )}
                {currentUser?.created_at && (
                  <p className="profile-joined">
                    Присоединился: {new Date(currentUser.created_at).toLocaleDateString('ru-RU')}
                  </p>
                )}
                {currentUser?.bio && (
                  <div className="profile-bio">
                    <p>{currentUser.bio}</p>
                  </div>
                )}
                {(socialLinks.vk || socialLinks.telegram || socialLinks.max) && (
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
                    {socialLinks.max && (
                      <a href={socialLinks.max.startsWith('http') ? socialLinks.max : `https://max.ru/${socialLinks.max.replace('@', '')}`} target="_blank" rel="noopener noreferrer" className="social-link">
                        MAX
                      </a>
                    )}
                  </div>
                )}
              </div>
            </div>
          </>
        )}

        {!isEditing && (
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
        )}
      </div>
    </div>
  );
};

export default ProfilePage;

