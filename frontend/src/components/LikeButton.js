import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import './LikeButton.css';

const LikeButton = ({ item, itemType, onLike, onUnlike }) => {
  const { isAuthenticated, user } = useAuth();
  const [liked, setLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (item) {
      // Initialize like count from item
      const count = item.likes?.length || 0;
      setLikeCount(count);
      
      // Check if current user has liked this item
      if (user && item.likes) {
        setLiked(item.likes.some(like => like.user_id === user.id));
      } else {
        setLiked(false);
      }
    }
  }, [item, user]);

  const handleLike = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (!isAuthenticated) {
      alert('Войдите, чтобы ставить лайки');
      return;
    }

    if (loading) return;

    setLoading(true);
    try {
      if (liked) {
        await onUnlike();
        setLiked(false);
        setLikeCount(prev => Math.max(0, prev - 1));
      } else {
        await onLike();
        setLiked(true);
        setLikeCount(prev => prev + 1);
      }
    } catch (err) {
      console.error('Error toggling like:', err);
      // Revert on error
      setLiked(!liked);
      setLikeCount(prev => liked ? prev + 1 : Math.max(0, prev - 1));
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      className={`like-button ${liked ? 'liked' : ''} ${loading ? 'loading' : ''}`}
      onClick={handleLike}
      disabled={loading || !isAuthenticated}
      title={isAuthenticated ? (liked ? 'Убрать лайк' : 'Поставить лайк') : 'Войдите, чтобы ставить лайки'}
    >
      <span className="like-icon">{liked ? '❤️' : '🤍'}</span>
      {likeCount > 0 && <span className="like-count">{likeCount}</span>}
    </button>
  );
};

export default LikeButton;

