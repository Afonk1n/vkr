import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { reviewsAPI } from '../services/api';
import { getImageUrl } from '../utils/imageUtils';
import LikeButton from './LikeButton';
import ReviewScoresStrip from './ReviewScoresStrip';
import './ReviewCard.css';

const ReviewCard = ({ review, onEdit, onDelete, onUpdate, moderationActions, hideLike }) => {
  const { user, isAdmin } = useAuth();
  const canEdit = user && (user.id === review.user_id || isAdmin);
  const [avatarError, setAvatarError] = useState(false);

  const handleLike = async () => {
    try {
      await reviewsAPI.like(review.id);
      if (onUpdate) onUpdate();
    } catch (err) {
      throw err;
    }
  };

  const handleUnlike = async () => {
    try {
      await reviewsAPI.unlike(review.id);
      if (onUpdate) onUpdate();
    } catch (err) {
      throw err;
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  };

  const getStatusBadge = (status) => {
    const statusMap = {
      pending: { text: 'На модерации', class: 'status-pending' },
      approved: { text: 'Одобрена', class: 'status-approved' },
      rejected: { text: 'Отклонена', class: 'status-rejected' },
    };
    const statusInfo = statusMap[status] || { text: status, class: '' };
    return <span className={`status-badge ${statusInfo.class}`}>{statusInfo.text}</span>;
  };

  // Извлекаем первую строку как заголовок (если она короткая)
  const getReviewTitle = (text) => {
    if (!text) return null;
    const lines = text.split('\n');
    const firstLine = lines[0].trim();
    // Если первая строка короткая (менее 50 символов) и есть вторая строка, считаем её заголовком
    if (firstLine.length > 0 && firstLine.length < 50 && lines.length > 1) {
      return firstLine;
    }
    return null;
  };

  const getReviewContent = (text) => {
    if (!text) return '';
    const title = getReviewTitle(text);
    if (title) {
      const lines = text.split('\n');
      return lines.slice(1).join('\n').trim();
    }
    return text;
  };

  const reviewTitle = getReviewTitle(review.text);
  const reviewContent = getReviewContent(review.text);

  return (
    <div className="review-card">
      <div className="review-header-compact">
        <div className="review-author-compact">
          {review.user?.id ? (
            <Link to={`/users/${review.user.id}`} className="review-author-avatar-link">
              {(review.user?.avatar_path && getImageUrl(review.user.avatar_path) && !avatarError) ? (
                <img
                  src={getImageUrl(review.user.avatar_path)}
                  alt={review.user?.username || 'Пользователь'}
                  className="review-author-avatar"
                  onError={() => setAvatarError(true)}
                />
              ) : (
                <div className="review-author-avatar-placeholder">
                  {(review.user?.username || 'П')[0].toUpperCase()}
                </div>
              )}
            </Link>
          ) : (
            <>
              {(review.user?.avatar_path && getImageUrl(review.user.avatar_path) && !avatarError) ? (
                <img
                  src={getImageUrl(review.user.avatar_path)}
                  alt={review.user?.username || 'Пользователь'}
                  className="review-author-avatar"
                  onError={() => setAvatarError(true)}
                />
              ) : (
                <div className="review-author-avatar-placeholder">
                  {(review.user?.username || 'П')[0].toUpperCase()}
                </div>
              )}
            </>
          )}
          <div className="review-author-text-compact">
            <div className="review-author-name-row">
              {review.user?.id ? (
                <Link to={`/users/${review.user.id}`} className="review-author-name-link">
                  <strong>{review.user?.username || 'Неизвестный пользователь'}</strong>
                </Link>
              ) : (
                <strong>{review.user?.username || 'Неизвестный пользователь'}</strong>
              )}
              {review.status !== 'approved' && (
                <span className="review-status-inline">
                  {getStatusBadge(review.status)}
                </span>
              )}
            </div>
            <span className="review-date-compact-header">{formatDate(review.created_at)}</span>

            {/* Информация об альбоме или треке - перемещена сюда */}
            {(review.album || review.track) && (
              <div className="review-item-info-blocks">
                {review.album ? (
                  <>
                    <span className="review-info-badge">
                      Артист: <Link to={`/artists/${encodeURIComponent(review.album.artist)}`} className="review-info-link">{review.album.artist}</Link>
                    </span>
                    <span className="review-info-badge">
                      Альбом: <Link to={`/albums/${review.album.id}`} className="review-info-link">{review.album.title}</Link>
                    </span>
                  </>
                ) : review.track ? (
                  <>
                    <span className="review-info-badge">
                      Артист: <Link to={`/artists/${encodeURIComponent(review.track.album?.artist || 'Неизвестный артист')}`} className="review-info-link">{review.track.album?.artist || 'Неизвестный артист'}</Link>
                    </span>
                    <span className="review-info-badge">
                      Альбом: <Link to={`/albums/${review.track.album?.id || review.track.album_id}`} className="review-info-link">{review.track.album?.title || 'Неизвестный альбом'}</Link>
                    </span>
                    <span className="review-info-badge">
                      Трек: <Link to={`/tracks/${review.track.id}`} className="review-info-link">{review.track.title}</Link>
                    </span>
                  </>
                ) : null}
              </div>
            )}
          </div>
        </div>

        <div className="review-scores-compact">
          <ReviewScoresStrip review={review} size="default" />
        </div>
      </div>

      {review.text && (review.status === 'approved' || isAdmin || moderationActions) && (
        <div className="review-content-compact">
          {reviewTitle && (
            <h3 className="review-title-compact">{reviewTitle}</h3>
          )}
          <div className="review-text-compact">{reviewContent}</div>
        </div>
      )}

      {/* Footer с кнопками - показывается всегда, если не moderationActions */}
      {!moderationActions && (
        <div className="review-footer-compact">
          <div className="review-footer-left">
            {!hideLike && (review.status === 'approved' || isAdmin || review.text) && (
              <LikeButton
                item={review}
                itemType="review"
                onLike={handleLike}
                onUnlike={handleUnlike}
              />
            )}
          </div>
          <div className="review-footer-right">
            {canEdit && (
              <div className="review-actions-compact">
                {onEdit && (
                  <button onClick={() => onEdit(review)} className="btn-edit-small" title="Редактировать">
                    ✎
                  </button>
                )}
                {onDelete && (
                  <button onClick={() => onDelete(review.id)} className="btn-delete-small" title="Удалить">
                    ×
                  </button>
                )}
              </div>
            )}
          </div>
        </div>
      )}

      {!review.text && review.status === 'pending' && !isAdmin && (
        <div className="review-note-compact">
          Текстовая рецензия будет опубликована после модерации
        </div>
      )}

      {moderationActions && (
        <div className="review-moderation-actions">
          {moderationActions}
        </div>
      )}

    </div>
  );
};

export default ReviewCard;
