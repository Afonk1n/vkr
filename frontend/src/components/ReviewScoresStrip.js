import React, { useMemo } from 'react';
import { formatScore, convertMultiplierToAtmosphere } from '../utils/ratingCalculator';
import { REVIEW_CRITERIA, getReviewScoreValues, formatMultiplierShort } from '../utils/ratingMeta';
import './ReviewScoresStrip.css';

const valueForCriterion = (key, v) => {
  if (key === 'atmosphere') return convertMultiplierToAtmosphere(v.mult) || 0;
  switch (key) {
    case 'rhymes':
      return v.rh;
    case 'structure':
      return v.st;
    case 'implementation':
      return v.impl;
    case 'individuality':
      return v.ind;
    default:
      return 0;
  }
};

const ReviewScoresStrip = ({ review, size = 'default', className = '' }) => {
  const v = useMemo(() => getReviewScoreValues(review), [review]);
  const finalShown = formatScore(review?.final_score) || '0';
  const multDisplay = formatMultiplierShort(v.mult);

  const stopInnerClick = (e) => {
    e.stopPropagation();
  };

  return (
    <div
      className={`review-scores-strip review-scores-strip--${size} ${className}`.trim()}
      onClick={stopInnerClick}
      role="group"
      aria-label="Оценки рецензии"
    >
      <div className="review-scores-strip-inner">
        <div className="review-scores-strip-total" title="Итоговый балл по формуле сайта">
          {finalShown}
        </div>
        <div className="review-scores-strip-cells">
          {REVIEW_CRITERIA.map((c) => {
            const val = valueForCriterion(c.key, v);
            const tip = `${c.title}\n${c.hint}`;
            return (
              <span key={c.key} className="review-scores-strip-cell" title={tip}>
                <span className="review-scores-strip-cell-abbr">{c.short}</span>
                <span className="review-scores-strip-cell-val">{val}</span>
              </span>
            );
          })}
        </div>
        <details className="review-scores-strip-details" onClick={stopInnerClick}>
          <summary
            className="review-scores-strip-details-btn"
            title="Как считается итог"
            aria-label="Подробнее про формулу оценки"
            onClick={stopInnerClick}
          >
            ?
          </summary>
          <div className="review-scores-strip-panel" onClick={stopInnerClick}>
            <div className="review-scores-strip-formula" aria-hidden>
              <span className="review-scores-strip-formula-bracket">(</span>
              <span className="review-scores-strip-formula-nums">
                {v.rh}+{v.st}+{v.impl}+{v.ind}
              </span>
              <span className="review-scores-strip-formula-bracket">)</span>
              <span className="review-scores-strip-formula-op">×</span>
              <span className="review-scores-strip-formula-k">1.4</span>
              <span className="review-scores-strip-formula-op">×</span>
              <span className="review-scores-strip-formula-m" title="Множитель из оценки атмосферы 1–10">
                {multDisplay}
              </span>
              <span className="review-scores-strip-formula-op">≈</span>
              <span className="review-scores-strip-formula-res">{finalShown}</span>
            </div>
            <p className="review-scores-strip-formula-caption">
              Сумма четырёх базовых критериев умножается на коэффициент <strong>1.4</strong> и на{' '}
              <strong>множитель атмосферы</strong> (от шкалы 1–10 нелинейно, макс. около 1.607). Результат
              округляется до целого.
            </p>
            <ul className="review-scores-strip-legend">
              {REVIEW_CRITERIA.map((c) => (
                <li key={c.key}>
                  <strong>{c.short}</strong> — {c.title}: {c.hint}
                </li>
              ))}
            </ul>
          </div>
        </details>
      </div>
    </div>
  );
};

export default ReviewScoresStrip;
