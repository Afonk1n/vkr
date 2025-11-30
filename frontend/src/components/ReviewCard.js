import React from 'react';
import { useAuth } from '../context/AuthContext';
import { formatScore, convertMultiplierToAtmosphere } from '../utils/ratingCalculator';
import './ReviewCard.css';

const ReviewCard = ({ review, onEdit, onDelete }) => {
  const { user, isAdmin } = useAuth();
  const canEdit = user && (user.id === review.user_id || isAdmin);

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
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

  return (
    <div className="review-card">
      <div className="review-header">
        <div className="review-author">
          <strong>{review.user?.username || 'Неизвестный пользователь'}</strong>
          <span className="review-date">{formatDate(review.created_at)}</span>
        </div>
        {canEdit && (
          <div className="review-actions">
            {onEdit && (
              <button onClick={() => onEdit(review)} className="btn-edit">
                Редактировать
              </button>
            )}
            {onDelete && (
              <button onClick={() => onDelete(review.id)} className="btn-delete">
                Удалить
              </button>
            )}
          </div>
        )}
      </div>

      <div className="review-ratings">
        <div className="rating-item">
          <span>Рифмы/Образы:</span>
          <strong>{review.rating_rhymes}/10</strong>
        </div>
        <div className="rating-item">
          <span>Структура/Ритмика:</span>
          <strong>{review.rating_structure}/10</strong>
        </div>
        <div className="rating-item">
          <span>Реализация стиля:</span>
          <strong>{review.rating_implementation}/10</strong>
        </div>
        <div className="rating-item">
          <span>Индивидуальность/Харизма:</span>
          <strong>{review.rating_individuality}/10</strong>
        </div>
        <div className="rating-item">
          <span>Атмосфера/Вайб:</span>
          <strong>{convertMultiplierToAtmosphere(review.atmosphere_multiplier)}/10</strong>
        </div>
      </div>

      <div className="review-score">
        Итоговый балл: <strong>{formatScore(review.final_score)}</strong>
      </div>

      {review.text && review.status === 'approved' && (
        <div className="review-text">{review.text}</div>
      )}

      {review.status !== 'approved' && (
        <div className="review-status">
          {getStatusBadge(review.status)}
          {review.status === 'pending' && (
            <span className="review-note">
              Текстовая рецензия будет опубликована после модерации
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default ReviewCard;

