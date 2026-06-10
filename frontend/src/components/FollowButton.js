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
    const prev = following;
    const next = !prev;

    // Оптимистично: статус и счётчик подписчиков (через onChange) меняются сразу.
    setFollowing(next);
    onChange?.(next);
    setBusy(true);
    try {
      if (next) {
        await usersAPI.follow(userId);
      } else {
        await usersAPI.unfollow(userId);
      }
    } catch (e) {
      console.error(e);
      // Откат при ошибке.
      setFollowing(prev);
      onChange?.(prev);
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
