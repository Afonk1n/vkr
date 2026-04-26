import React, { useState, useEffect } from 'react';
import { usersAPI } from '../services/api';
import './FollowButton.css';

const FollowButton = ({ userId, initialFollowing, onChange, compact = false }) => {
  const [following, setFollowing] = useState(Boolean(initialFollowing));
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setFollowing(Boolean(initialFollowing));
  }, [initialFollowing, userId]);

  const handleClick = async () => {
    if (busy) return;
    setBusy(true);
    try {
      if (following) {
        await usersAPI.unfollow(userId);
        setFollowing(false);
        onChange?.(false);
      } else {
        await usersAPI.follow(userId);
        setFollowing(true);
        onChange?.(true);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setBusy(false);
    }
  };

  return (
    <button
      type="button"
      className={`follow-btn ${following ? 'follow-btn--following' : ''} ${compact ? 'follow-btn--compact' : ''}`}
      onClick={handleClick}
      disabled={busy}
      aria-pressed={following}
    >
      {busy ? '…' : following ? 'В подписках' : 'Подписаться'}
    </button>
  );
};

export default FollowButton;
