import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { usersAPI, reviewsAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import ProfileEditForm from '../components/ProfileEditForm';
import ProfileDashboard from '../components/ProfileDashboard';
import './ProfilePage.css';

const ProfilePage = () => {
  const { user, isAuthenticated, updateUser, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const [reviews, setReviews] = useState([]);
  const [likedReviews, setLikedReviews] = useState([]);
  const [reviewsLoading, setReviewsLoading] = useState(true);
  const [likedReviewsLoading, setLikedReviewsLoading] = useState(true);
  const [error, setError] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [currentUser, setCurrentUser] = useState(user);

  const fetchUserReviews = useCallback(async () => {
    if (!user) return;

    setReviewsLoading(true);
    try {
      const response = await usersAPI.getUserReviews(user.id);
      setReviews(response.data.reviews || []);
    } catch (err) {
      setError('Ошибка загрузки рецензий');
      console.error('Error fetching reviews:', err);
    } finally {
      setReviewsLoading(false);
    }
  }, [user]);

  const fetchCurrentUser = useCallback(async () => {
    if (!user) return;
    try {
      const response = await usersAPI.getById(user.id);
      setCurrentUser(response.data);
    } catch (err) {
      console.error('Error fetching user:', err);
    }
  }, [user]);

  const fetchLikedReviews = useCallback(async () => {
    if (!user) return;

    setLikedReviewsLoading(true);
    try {
      const response = await usersAPI.getLikedReviews(user.id);
      setLikedReviews(response.data.reviews || []);
    } catch (err) {
      console.error('Error fetching liked reviews:', err);
    } finally {
      setLikedReviewsLoading(false);
    }
  }, [user]);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    fetchUserReviews();
    fetchCurrentUser();
    fetchLikedReviews();
  }, [authLoading, isAuthenticated, navigate, fetchUserReviews, fetchCurrentUser, fetchLikedReviews]);

  const handleSaveProfile = async (profileData) => {
    try {
      const response = await usersAPI.update(user.id, profileData);
      const updatedUser = response.data;
      setCurrentUser(updatedUser);
      setIsEditing(false);

      if (updateUser) {
        updateUser(updatedUser);
      }
    } catch (err) {
      console.error('Error updating profile:', err);
      throw new Error(err.response?.data?.message || 'Ошибка при обновлении профиля');
    }
  };

  const handleEditReview = (review) => {
    navigate(`/albums/${review.album_id}`);
  };

  const handleDeleteReview = async (reviewId) => {
    try {
      await reviewsAPI.delete(reviewId);
      fetchUserReviews();
    } catch (err) {
      setError('Ошибка при удалении рецензии');
      console.error('Error deleting review:', err);
    }
  };

  const handleLikedReviewRemoved = useCallback((reviewId) => {
    setLikedReviews((current) => current.filter((review) => review.id !== reviewId));
  }, []);

  if (!isAuthenticated || !user) {
    return null;
  }

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
          <ProfileDashboard
            profileUser={currentUser}
            reviews={reviews}
            reviewsLoading={reviewsLoading}
            reviewsError={error}
            likedReviews={likedReviews}
            likedReviewsLoading={likedReviewsLoading}
            isOwner
            onEditProfile={() => setIsEditing(true)}
            onPreferencesUpdate={(favorites) => setCurrentUser((prev) => ({ ...prev, ...favorites }))}
            emptyReviewsText="У вас пока нет рецензий"
            renderReviews={() => (
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
            renderLikedReviews={() => (
              <div className="reviews-list">
                {likedReviews.map((review) => (
                  <ReviewCard
                    key={review.id}
                    review={review}
                    hideLike={false}
                    onUnlikeComplete={handleLikedReviewRemoved}
                  />
                ))}
              </div>
            )}
          />
        )}
      </div>
    </div>
  );
};

export default ProfilePage;
