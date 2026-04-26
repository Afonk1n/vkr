import React, { useMemo } from 'react';
import { REVIEW_CRITERIA } from '../utils/ratingMeta';
import './AverageScoreBadge.css';

const avg = (items, getter) => {
  if (!items?.length) return 0;
  return items.reduce((sum, item) => sum + (Number(getter(item)) || 0), 0) / items.length;
};

const fmtWhole = (value) => String(Math.round(Number(value) || 0));

const valuesFromReviews = (reviews = []) => {
  const approved = reviews.filter((r) => !r.status || r.status === 'approved');
  if (!approved.length) return null;
  const atmosphereMultiplier = avg(approved, (r) => r.atmosphere_multiplier);
  const atmosphere = 1 + (atmosphereMultiplier - 1) / (0.6072 / 9);
  return {
    final: Math.round(avg(approved, (r) => r.final_score)),
    rhymes: avg(approved, (r) => r.rating_rhymes),
    structure: avg(approved, (r) => r.rating_structure),
    implementation: avg(approved, (r) => r.rating_implementation),
    individuality: avg(approved, (r) => r.rating_individuality),
    atmosphere,
    count: approved.length,
  };
};

const valuesFromSource = (source) => {
  const final = Math.round(Number(source?.average_rating) || 0);
  if (!final) return null;
  return {
    final,
    rhymes: Number(source?.average_rating_rhymes) || 0,
    structure: Number(source?.average_rating_structure) || 0,
    implementation: Number(source?.average_rating_implementation) || 0,
    individuality: Number(source?.average_rating_individuality) || 0,
    atmosphere: Number(source?.average_atmosphere_rating) || 0,
    count: Number(source?.approved_reviews_count) || 0,
  };
};

const AverageScoreBadge = ({ source, reviews, size = 'default', className = '' }) => {
  const values = useMemo(() => valuesFromSource(source) || valuesFromReviews(reviews), [source, reviews]);

  if (!values?.final) return null;

  const criteria = [
    { key: 'rhymes', title: REVIEW_CRITERIA[0].title, value: values.rhymes },
    { key: 'structure', title: REVIEW_CRITERIA[1].title, value: values.structure },
    { key: 'implementation', title: REVIEW_CRITERIA[2].title, value: values.implementation },
    { key: 'individuality', title: REVIEW_CRITERIA[3].title, value: values.individuality },
    { key: 'atmosphere', title: 'Атмосфера / вайб', value: values.atmosphere },
  ];

  return (
    <div
      className={`average-score-badge average-score-badge--${size} ${className}`.trim()}
      role="group"
      tabIndex={0}
      aria-label={`Средняя оценка ${values.final}`}
    >
      <span className="average-score-main">{values.final}</span>
      <div className="average-score-panel">
        <div className="average-score-panel-title">
          Средняя оценка{values.count ? ` по ${values.count} рец.` : ''}
        </div>
        <div className="average-score-formula" aria-hidden>
          <span>(</span>
          <strong>
            {fmtWhole(values.rhymes)}+{fmtWhole(values.structure)}+{fmtWhole(values.implementation)}+
            {fmtWhole(values.individuality)}
          </strong>
          <span>) ×</span>
          <strong className="average-score-vibe">вайб</strong>
          <span>≈</span>
          <strong>{values.final}</strong>
        </div>
        <ul className="average-score-list">
          {criteria.map((item) => (
            <li key={item.key}>
              <span>{item.title}</span>
              <strong>{fmtWhole(item.value)}</strong>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default AverageScoreBadge;
