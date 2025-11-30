import React, { useState } from 'react';
import { calculateFinalScore, formatScore, convertMultiplierToAtmosphere, convertAtmosphereToMultiplier } from '../utils/ratingCalculator';
import './ReviewForm.css';

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
        text: hasText ? text : '',
      };
      
      // Ensure IDs are numbers
      if (albumId) {
        reviewData.album_id = typeof albumId === 'string' ? parseInt(albumId, 10) : albumId;
      } else if (trackId) {
        reviewData.track_id = typeof trackId === 'string' ? parseInt(trackId, 10) : trackId;
      }
      
      console.log('Submitting review data:', reviewData);
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
      <h3>{initialData ? 'Редактировать рецензию' : 'Добавить рецензию'}</h3>
      {error && <div className="error-message">{error}</div>}
      <form onSubmit={handleSubmit} className="review-form">
        <div className="rating-section">
          <h4>Оценка по критериям (1-10)</h4>
          
          <div className="rating-group">
            <label htmlFor="rating-rhymes">
              Рифмы / Образы: {ratingRhymes}
              <span className="tooltip">ℹ️
                <span className="tooltiptext">
                  Оценка за текст, учитывающая жанровые особенности. Для легкой фоновой поп-музыки или электронной музыки допускаются тривиальные конструкции для максимального балла. Для текстоцентричных жанров требуется сложная рифмовка и глубокий смысл.
                </span>
              </span>
            </label>
            <input
              type="range"
              id="rating-rhymes"
              min="1"
              max="10"
              value={ratingRhymes}
              onChange={(e) => setRatingRhymes(parseInt(e.target.value))}
            />
          </div>

          <div className="rating-group">
            <label htmlFor="rating-structure">
              Структура / Ритмика: {ratingStructure}
              <span className="tooltip">ℹ️
                <span className="tooltiptext">
                  Включает оценку ритмической составляющей (стихотворный ритм — мелодичность, драматургия частей, контрасты) и гармонию структуры (целостность всех частей трека, концепция альбома и расположение песен).
                </span>
              </span>
            </label>
            <input
              type="range"
              id="rating-structure"
              min="1"
              max="10"
              value={ratingStructure}
              onChange={(e) => setRatingStructure(parseInt(e.target.value))}
            />
          </div>

          <div className="rating-group">
            <label htmlFor="rating-implementation">
              Реализация стиля: {ratingImplementation}
              <span className="tooltip">ℹ️
                <span className="tooltiptext">
                  Оценивает работу исполнителя (качество вокала, речитатив, умение работать с мелодией) и работу звукорежиссера/саунд-продюсера (качество сведения, звучание инструментала).
                </span>
              </span>
            </label>
            <input
              type="range"
              id="rating-implementation"
              min="1"
              max="10"
              value={ratingImplementation}
              onChange={(e) => setRatingImplementation(parseInt(e.target.value))}
            />
          </div>

          <div className="rating-group">
            <label htmlFor="rating-individuality">
              Индивидуальность / Харизма: {ratingIndividuality}
              <span className="tooltip">ℹ️
                <span className="tooltiptext">
                  Оценивает уникальность тембра голоса, стиль исполнения, узнаваемость, а также способность артиста передать и погрузить слушателя в эмоции (верю — не верю песне).
                </span>
              </span>
            </label>
            <input
              type="range"
              id="rating-individuality"
              min="1"
              max="10"
              value={ratingIndividuality}
              onChange={(e) => setRatingIndividuality(parseInt(e.target.value))}
            />
          </div>

          <div className="rating-group">
            <label htmlFor="atmosphere">
              Атмосфера / Вайб: {atmosphereRating}
              <span className="tooltip">ℹ️
                <span className="tooltiptext">
                  Субъективная оценка, показывающая, насколько автор сумел передать атмосферу и палитру эмоций композиции. Диапазон: 1-10
                </span>
              </span>
            </label>
            <input
              type="range"
              id="atmosphere"
              min="1"
              max="10"
              value={atmosphereRating}
              onChange={(e) => setAtmosphereRating(parseInt(e.target.value))}
            />
          </div>

          <div className="final-score">
            <strong>Итоговый балл: {formatScore(finalScore)}</strong>
          </div>
        </div>

        <div className="text-section">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={hasText}
              onChange={(e) => setHasText(e.target.checked)}
            />
            Добавить текстовую рецензию (будет отправлена на модерацию)
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

