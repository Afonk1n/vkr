import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { usersAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import ReviewCard from '../components/ReviewCard';
import FollowButton from '../components/FollowButton';
import ProfileDashboard from '../components/ProfileDashboard';
import './UserProfilePage.css';

const UserProfilePage = () => {
  const { id } = useParams();
  const { user: me, isAuthenticated } = useAuth();
  const [user, setUser] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [likedReviews, setLikedReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [reviewsLoading, setReviewsLoading] = useState(true);
  const [likedReviewsLoading, setLikedReviewsLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchUserData = useCallback(async () => {
    setLoading(true);
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
    setReviewsLoading(true);
    try {
      const response = await usersAPI.getUserReviews(id);
      setReviews(response.data.reviews || []);
    } catch (err) {
      console.error('Error fetching reviews:', err);
    } finally {
      setReviewsLoading(false);
    }
  }, [id]);

  const fetchLikedReviews = useCallback(async () => {
    setLikedReviewsLoading(true);
    try {
      const response = await usersAPI.getLikedReviews(id);
      setLikedReviews(response.data.reviews || []);
    } catch (err) {
      console.error('Error fetching liked reviews:', err);
    } finally {
      setLikedReviewsLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchUserData();
    fetchUserReviews();
    fetchLikedReviews();
  }, [fetchUserData, fetchUserReviews, fetchLikedReviews]);

  const handleLikedReviewRemoved = useCallback((reviewId) => {
    setLikedReviews((current) => current.filter((review) => review.id !== reviewId));
  }, []);

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

  const targetId = Number(id);
  const isOwnProfile = Boolean(isAuthenticated && me && me.id === targetId);

  return (
    <div className="container">
      <div className="user-profile-page">
        <ProfileDashboard
          profileUser={user}
          reviews={reviews}
          reviewsLoading={reviewsLoading}
          likedReviews={likedReviews}
          likedReviewsLoading={likedReviewsLoading}
          isOwner={isOwnProfile}
          onPreferencesUpdate={(favorites) => setUser((prev) => ({ ...prev, ...favorites }))}
          emptyReviewsText="У пользователя пока нет рецензий"
          followSlot={!isOwnProfile && isAuthenticated ? (
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
          ) : (
            <div className="profile-follow-meta">
              <span>{user.followers_count ?? 0} подписчиков</span>
              <span className="profile-follow-dot">·</span>
              <span>{user.following_count ?? 0} подписок</span>
            </div>
          )}
          renderReviews={() => (
            <div className="reviews-list">
              {reviews.map((review) => (
                <ReviewCard
                  key={review.id}
                  review={review}
                  hideLike={false}
                  onUnlikeComplete={handleLikedReviewRemoved}
                />
              ))}
            </div>
          )}
          renderLikedReviews={() => (
            <div className="reviews-list">
              {likedReviews.map((review) => (
                <ReviewCard
                  key={review.id}
                  review={review}
                  hideLike={false}
                />
              ))}
            </div>
          )}
        />
      </div>
    </div>
  );
};

export default UserProfilePage;
