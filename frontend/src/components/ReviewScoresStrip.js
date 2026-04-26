import React, { useMemo } from 'react';
import { formatScore, convertMultiplierToAtmosphere } from '../utils/ratingCalculator';
import { REVIEW_CRITERIA, getReviewScoreValues } from '../utils/ratingMeta';
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
  const criteriaValues = REVIEW_CRITERIA.map((c) => ({
    ...c,
    value: valueForCriterion(c.key, v),
  }));

  const stopInnerClick = (e) => {
    e.stopPropagation();
  };

  return (
    <div
      className={`review-scores-strip review-scores-strip--${size} ${className}`.trim()}
      onClick={stopInnerClick}
      role="group"
      tabIndex={0}
      aria-label="Оценка рецензии"
    >
      <div className="review-scores-strip-inner">
        <div className="review-scores-strip-total">{finalShown}</div>
        <div className="review-scores-strip-cells">
          {criteriaValues.map((c) => (
            <span key={c.key} className="review-scores-strip-cell" aria-label={`${c.title}: ${c.value}`}>
              {c.value}
            </span>
          ))}
        </div>
        <div className="review-scores-strip-panel" onClick={stopInnerClick}>
          <div className="review-scores-strip-panel-title">Из чего складывается оценка</div>
          <div className="review-scores-strip-formula" aria-hidden>
            <span className="review-scores-strip-formula-bracket">(</span>
            <span className="review-scores-strip-formula-nums">
              {v.rh}+{v.st}+{v.impl}+{v.ind}
            </span>
            <span className="review-scores-strip-formula-bracket">)</span>
            <span className="review-scores-strip-formula-op">×</span>
            <span className="review-scores-strip-formula-m">вайб</span>
            <span className="review-scores-strip-formula-op">≈</span>
            <span className="review-scores-strip-formula-res">{finalShown}</span>
          </div>
          <ul className="review-scores-strip-legend">
            {criteriaValues.map((c) => (
              <li key={c.key}>
                <span>{c.title}</span>
                <strong>{c.value}</strong>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
};

export default ReviewScoresStrip;
