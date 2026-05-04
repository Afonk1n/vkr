import React, { useState } from 'react';
import { calculateFinalScore, formatScore, convertMultiplierToAtmosphere, convertAtmosphereToMultiplier } from '../utils/ratingCalculator';
import { REVIEW_CRITERIA } from '../utils/ratingMeta';
import './ReviewForm.css';

const criterionDescriptions = {
  rhymes:
    'Текст с учетом жанра: образность, рифмовка, смысл и точность формулировок.',
  structure:
    'Ритм, драматургия частей, цельность композиции и развитие материала.',
  implementation:
    'Исполнение, продакшн, сведение и уверенность работы внутри выбранного стиля.',
  individuality:
    'Узнаваемость, харизма, манера исполнения и способность увлечь слушателя.',
  atmosphere:
    'Общее ощущение от релиза: насколько цельно передано настроение и эмоция.',
};

const baseCriteria = [
  { key: 'rhymes', label: 'Рифмы / образы', state: 'ratingRhymes' },
  { key: 'structure', label: 'Структура / ритмика', state: 'ratingStructure' },
  { key: 'implementation', label: 'Реализация стиля', state: 'ratingImplementation' },
  { key: 'individuality', label: 'Индивидуальность / харизма', state: 'ratingIndividuality' },
];

const InfoHint = ({ text }) => (
  <span className="review-form-info" tabIndex="0" aria-label={text}>
    i
    <span className="review-form-info-panel">{text}</span>
  </span>
);

const ScoreSlider = ({ id, label, value, onChange, hint }) => (
  <div className="review-score-control">
    <div className="review-score-control-top">
      <label htmlFor={id}>
        {label}
        <InfoHint text={hint} />
      </label>
      <strong>{value}</strong>
    </div>
    <input
      type="range"
      id={id}
      min="1"
      max="10"
      value={value}
      onChange={(e) => onChange(parseInt(e.target.value, 10))}
      style={{ '--value': `${((value - 1) / 9) * 100}%` }}
    />
    <div className="review-score-scale" aria-hidden="true">
      <span>1</span>
      <span>5</span>
      <span>10</span>
    </div>
  </div>
);

const ReviewForm = ({ albumId, trackId, onSubmit, initialData, onCancel }) => {
  const [ratingRhymes, setRatingRhymes] = useState(initialData?.rating_rhymes || 5);
  const [ratingStructure, setRatingStructure] = useState(initialData?.rating_structure || 5);
  const [ratingImplementation, setRatingImplementation] = useState(initialData?.rating_implementation || 5);
  const [ratingIndividuality, setRatingIndividuality] = useState(initialData?.rating_individuality || 5);
  // Convert multiplier to rating for display (if initialData has multiplier)
  const initialAtmosphere = initialData?.atmosphere_multiplier 
    ? convertMultiplierToAtmosphere(initialData.atmosphere_multiplier)
    : 5;
  const [atmosphereRating, setAtmosphereRating] = useState(initialAtmosphere);
  const [hasText, setHasText] = useState(!!initialData?.text);
  const [text, setText] = useState(initialData?.text || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const finalScore = calculateFinalScore(
    ratingRhymes,
    ratingStructure,
    ratingImplementation,
    ratingIndividuality,
    atmosphereRating
  );
  const baseSum = ratingRhymes + ratingStructure + ratingImplementation + ratingIndividuality;
  const stateValues = {
    ratingRhymes,
    ratingStructure,
    ratingImplementation,
    ratingIndividuality,
  };
  const stateSetters = {
    ratingRhymes: setRatingRhymes,
    ratingStructure: setRatingStructure,
    ratingImplementation: setRatingImplementation,
    ratingIndividuality: setRatingIndividuality,
  };
  const scorePreviewValues = {
    rhymes: ratingRhymes,
    structure: ratingStructure,
    implementation: ratingImplementation,
    individuality: ratingIndividuality,
    atmosphere: atmosphereRating,
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // Check if user is authenticated
      const userId = localStorage.getItem('userId');
      if (!userId) {
        setError('Необходимо войти в систему для создания рецензии');
        setLoading(false);
        return;
      }

      // Validate that either albumId or trackId is provided
      if (!albumId && !trackId) {
        setError('Необходимо указать альбом или трек для рецензии');
        setLoading(false);
        return;
      }

      // Convert atmosphere rating to multiplier before sending
      const atmosphereMultiplier = convertAtmosphereToMultiplier(atmosphereRating);
      const reviewData = {
        rating_rhymes: ratingRhymes,
        rating_structure: ratingStructure,
        rating_implementation: ratingImplementation,
        rating_individuality: ratingIndividuality,
        atmosphere_rating: atmosphereRating,
        atmosphere_multiplier: atmosphereMultiplier,
        text: hasText ? text : '',
      };
      
      // Ensure IDs are numbers
      if (albumId) {
        reviewData.album_id = typeof albumId === 'string' ? parseInt(albumId, 10) : albumId;
      } else if (trackId) {
        reviewData.track_id = typeof trackId === 'string' ? parseInt(trackId, 10) : trackId;
      }
      
      await onSubmit(reviewData);
    } catch (err) {
      console.error('Error submitting review:', err);
      console.error('Error response:', err.response);
      
      let errorMessage = 'Ошибка при сохранении рецензии';
      
      if (err.response?.data) {
        if (err.response.data.message) {
          errorMessage = err.response.data.message;
        } else if (err.response.data.error) {
          errorMessage = err.response.data.error;
        }
      } else if (err.message) {
        errorMessage = err.message;
      }
      
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="review-form-container">
      <div className="review-form-head">
        <div>
          <p className="review-form-eyebrow">Рецензия</p>
          <h3>{initialData ? 'Редактировать рецензию' : 'Добавить рецензию'}</h3>
        </div>
        <div className="review-form-score-pill">
          <span>Итог</span>
          <strong>{formatScore(finalScore)}</strong>
        </div>
      </div>
      {error && <div className="error-message">{error}</div>}
      <form onSubmit={handleSubmit} className="review-form">
        <div className="rating-section">
          <div className="review-form-rating-layout">
            <div className="review-form-card base-ratings-group">
              <div className="group-header">
                <div>
                  <span className="review-form-section-kicker">Критерии</span>
                  <h4>Основная оценка</h4>
                </div>
                <span className="group-summary">Сумма {baseSum}</span>
              </div>
              <div className="base-ratings-grid">
                {baseCriteria.map((criterion) => (
                  <ScoreSlider
                    key={criterion.key}
                    id={`rating-${criterion.key}`}
                    label={criterion.label}
                    value={stateValues[criterion.state]}
                    onChange={stateSetters[criterion.state]}
                    hint={criterionDescriptions[criterion.key]}
                  />
                ))}
              </div>
            </div>

            <aside className="review-form-summary">
              <div className="review-form-summary-score">
                <span>Итоговый балл</span>
                <strong>{formatScore(finalScore)}</strong>
              </div>
              <div className="review-form-score-grid">
                {REVIEW_CRITERIA.map((criterion) => (
                  <span key={criterion.key}>
                    <small>{criterion.short}</small>
                    <b>{scorePreviewValues[criterion.key]}</b>
                  </span>
                ))}
              </div>
              <p>
                Итог складывается из базовых критериев и общего вайба. Детали видны в карточке после публикации.
              </p>
            </aside>
          </div>

          <div className="review-form-card multiplier-group">
            <div className="group-header">
              <div>
                <span className="review-form-section-kicker">Общее впечатление</span>
                <h4>Атмосфера / вайб</h4>
              </div>
              <span className="group-summary">Вайб {atmosphereRating}</span>
            </div>
            <ScoreSlider
              id="atmosphere"
              label="Атмосфера / вайб"
              value={atmosphereRating}
              onChange={setAtmosphereRating}
              hint={criterionDescriptions.atmosphere}
            />
          </div>
        </div>

        <div className="text-section review-form-card">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={hasText}
              onChange={(e) => setHasText(e.target.checked)}
            />
            <span>
              Добавить текстовую рецензию
              <small>Текстовая рецензия отправляется на модерацию.</small>
            </span>
          </label>
          
          {hasText && (
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Напишите вашу рецензию..."
              rows={6}
              className="review-textarea"
            />
          )}
        </div>

        <div className="form-actions">
          {onCancel && (
            <button type="button" onClick={onCancel} className="btn-cancel">
              Отмена
            </button>
          )}
          <button type="submit" className="btn-submit" disabled={loading}>
            {loading ? 'Сохранение...' : initialData ? 'Сохранить' : 'Отправить'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default ReviewForm;

