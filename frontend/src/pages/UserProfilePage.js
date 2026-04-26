import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { usersAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import ReviewCard from '../components/ReviewCard';
import BadgeList from '../components/BadgeList';
import FavoriteAlbums from '../components/FavoriteAlbums';
import GenreRadar from '../components/GenreRadar';
import FollowButton from '../components/FollowButton';
import { getImageUrl } from '../utils/imageUtils';
import './UserProfilePage.css';

const UserProfilePage = () => {
  const { id } = useParams();
  const { user: me, isAuthenticated } = useAuth();
  const [user, setUser] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchUserData = useCallback(async () => {
    try {
      const response = await usersAPI.getById(id);
      setUser(response.data);
    } catch (err) {
      setError('Пользователь не найден');
      console.error('Error fetching user:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  const fetchUserReviews = useCallback(async () => {
    try {
      const response = await usersAPI.getUserReviews(id);
      setReviews(response.data.reviews);
    } catch (err) {
      console.error('Error fetching reviews:', err);
    }
  }, [id]);

  useEffect(() => {
    fetchUserData();
    fetchUserReviews();
  }, [fetchUserData, fetchUserReviews]);

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
      </div>
    );
  }

  const socialLinks = user?.social_links
    ? (typeof user.social_links === 'string'
        ? JSON.parse(user.social_links)
        : user.social_links)
    : {};

  const targetId = Number(id);
  const isOwnProfile = Boolean(isAuthenticated && me && me.id === targetId);

  return (
    <div className="container">
      <div className="user-profile-page">
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
            <h1 className="profile-username">
              {user?.username || 'Пользователь'}
              {user?.is_verified_artist && (
                <span className="verified-badge" title="Верифицированный артист">✓</span>
              )}
            </h1>
            {isOwnProfile && user?.email && (
              <p className="profile-email">{user.email}</p>
            )}
            <div className="profile-follow-row">
              <div className="profile-follow-meta">
                <span>{user.followers_count ?? 0} подписчиков</span>
                <span className="profile-follow-dot">·</span>
                <span>{user.following_count ?? 0} подписок</span>
              </div>
              {!isOwnProfile && isAuthenticated && (
                <FollowButton
                  userId={user.id}
                  initialFollowing={user.is_following}
                  onChange={(following) => {
                    setUser((prev) => ({
                      ...prev,
                      is_following: following,
                      followers_count: Math.max(0, (prev.followers_count ?? 0) + (following ? 1 : -1)),
                    }));
                  }}
                />
              )}
            </div>
            {user?.is_admin && (
              <span className="admin-badge">Администратор</span>
            )}
            {user?.badges && user.badges.length > 0 && (
              <BadgeList badges={user.badges} profileContext="other" />
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

            {user?.stats && (
              <div className="profile-stats">
                <div className="stat-item">
                  <span className="stat-value">{user.stats.total_reviews}</span>
                  <span className="stat-label">рецензий</span>
                </div>
                <div className="stat-item">
                  <span className="stat-value">{user.stats.avg_score || '—'}</span>
                  <span className="stat-label">ср. оценка</span>
                </div>
                <div className="stat-item">
                  <span className="stat-value">{user.stats.total_likes_received}</span>
                  <span className="stat-label">лайков</span>
                </div>
                {user.stats.top_genre && (
                  <div className="stat-item">
                    <span className="stat-value stat-genre">{user.stats.top_genre}</span>
                    <span className="stat-label">топ жанр</span>
                  </div>
                )}
              </div>
            )}
          </div>

          {user?.genre_stats && user.genre_stats.length >= 3 && (
            <div className="profile-radar-section">
              <h3 className="radar-title">Жанровый профиль</h3>
              <GenreRadar genreStats={user.genre_stats} />
            </div>
          )}
        </div>

        <FavoriteAlbums albums={user?.favorite_albums} isOwner={false} userId={user.id} />

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

