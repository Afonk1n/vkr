import React, { useEffect, useMemo, useState } from 'react';
import { REVIEW_CRITERIA } from '../utils/ratingMeta';
import './ReleasePassport.css';

// Цвет по шкале 0–90 через токены проекта, чтобы графики не выбивались из темы.
const scoreColor = (score) => {
  const value = Number(score) || 0;
  if (value >= 81) return 'var(--success-color)';
  if (value >= 66) return 'var(--accent-color, var(--primary-color))';
  if (value >= 51) return 'var(--warning-color)';
  return 'var(--error-color)';
};

const avg = (items, getter) => {
  if (!items?.length) return 0;
  return items.reduce((sum, item) => sum + (Number(getter(item)) || 0), 0) / items.length;
};

const pluralizeReview = (count) => {
  const n = Math.abs(Number(count) || 0);
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod10 === 1 && mod100 !== 11) return `${count} рецензии`;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return `${count} рецензиям`;
  return `${count} рецензиям`;
};

const joinTitles = (items) => {
  if (items.length <= 2) return items.join(' и ');
  return `${items.slice(0, -1).join(', ')} и ${items[items.length - 1]}`;
};

// Те же правила, что и в AverageScoreBadge: сперва агрегаты с бэка, иначе из рецензий.
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

const valuesFromReviews = (approved) => {
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

// Плавный count-up при открытии паспорта.
const useCountUp = (target, active, duration = 900) => {
  const [value, setValue] = useState(0);
  useEffect(() => {
    if (!active) {
      setValue(0);
      return undefined;
    }
    let raf;
    const start = performance.now();
    const tick = (now) => {
      const p = Math.min(1, (now - start) / duration);
      const eased = 1 - Math.pow(1 - p, 3);
      setValue(Math.round(target * eased));
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target, active, duration]);
  return value;
};

const RADAR_ORDER = ['rhymes', 'structure', 'implementation', 'individuality', 'atmosphere'];
const CRITERIA_BY_KEY = REVIEW_CRITERIA.reduce((acc, c) => ({ ...acc, [c.key]: c }), {});

const SCORE_BUCKETS = [
  { min: 0, max: 35, label: '0–35' },
  { min: 36, max: 50, label: '36–50' },
  { min: 51, max: 65, label: '51–65' },
  { min: 66, max: 80, label: '66–80' },
  { min: 81, max: 90, label: '81–90' },
];

// --- Радар «ДНК звучания» (SVG) ---
const Radar = ({ values }) => {
  const size = 240;
  const cx = size / 2;
  const cy = size / 2;
  const maxR = 84;
  const point = (index, r) => {
    const angle = (index * (360 / RADAR_ORDER.length) - 90) * (Math.PI / 180);
    return [cx + r * Math.cos(angle), cy + r * Math.sin(angle)];
  };
  const ringPoints = (factor) =>
    RADAR_ORDER.map((_, i) => point(i, maxR * factor).join(',')).join(' ');
  const dataPoints = RADAR_ORDER.map((key, i) => {
    const v = Math.max(0, Math.min(10, Number(values[key]) || 0));
    return point(i, (v / 10) * maxR).join(',');
  }).join(' ');

  return (
    <svg viewBox={`0 0 ${size} ${size}`} className="passport-radar" role="img" aria-label="Профиль звучания по критериям">
      {[0.25, 0.5, 0.75, 1].map((f) => (
        <polygon key={f} className="passport-radar-grid" points={ringPoints(f)} />
      ))}
      {RADAR_ORDER.map((key, i) => {
        const [x, y] = point(i, maxR);
        return <line key={key} className="passport-radar-axis" x1={cx} y1={cy} x2={x} y2={y} />;
      })}
      <polygon className="passport-radar-shape" points={dataPoints} />
      {RADAR_ORDER.map((key, i) => {
        const [x, y] = point(i, maxR + 16);
        const anchor = x < cx - 4 ? 'end' : x > cx + 4 ? 'start' : 'middle';
        return (
          <text key={key} className="passport-radar-label" x={x} y={y} textAnchor={anchor} dominantBaseline="middle">
            {CRITERIA_BY_KEY[key].short}
          </text>
        );
      })}
    </svg>
  );
};

const ReleasePassport = ({ source, reviews = [], title, type = 'album' }) => {
  const [open, setOpen] = useState(false);

  const approved = useMemo(
    () => (reviews || []).filter((r) => !r.status || r.status === 'approved'),
    [reviews]
  );

  const values = useMemo(
    () => valuesFromSource(source) || valuesFromReviews(approved),
    [source, approved]
  );

  // Распределение финальных оценок по корзинам + «согласие критиков».
  const distribution = useMemo(() => {
    const scores = approved
      .map((r) => Math.round(Number(r.final_score) || 0))
      .filter((s) => s > 0);
    const buckets = SCORE_BUCKETS.map((b) => ({
      ...b,
      count: scores.filter((s) => s >= b.min && s <= b.max).length,
    }));
    const maxCount = buckets.reduce((m, b) => Math.max(m, b.count), 0);
    let spread = null;
    if (scores.length >= 2) {
      const mean = scores.reduce((a, b) => a + b, 0) / scores.length;
      const variance = scores.reduce((a, b) => a + (b - mean) ** 2, 0) / scores.length;
      const std = Math.sqrt(variance);
      const level = std < 6 ? 'high' : std < 12 ? 'mid' : 'low';
      spread = {
        level,
        label: level === 'high' ? 'Единодушие' : level === 'mid' ? 'В целом сходятся' : 'Спорный релиз',
      };
    }
    return { buckets, maxCount, total: scores.length, spread };
  }, [approved]);

  // Авто-вердикт: сильные и слабые стороны по округлённым критериям, чтобы честно обработать равные оценки.
  const verdict = useMemo(() => {
    if (!values) return null;
    const list = RADAR_ORDER.map((key) => ({
      key,
      title: CRITERIA_BY_KEY[key].title,
      value: Math.round(Number(values[key]) || 0),
    }));
    const maxValue = Math.max(...list.map((item) => item.value));
    const minValue = Math.min(...list.map((item) => item.value));
    if (maxValue === minValue) {
      return { even: true };
    }
    return {
      even: false,
      strong: list.filter((item) => item.value === maxValue).map((item) => item.title),
      weak: list.filter((item) => item.value === minValue).map((item) => item.title),
    };
  }, [values]);

  const animatedScore = useCountUp(values?.final || 0, open);

  useEffect(() => {
    if (!open) return undefined;
    const onKey = (e) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open]);

  // Без оценок паспорт не имеет смысла — кнопку не показываем.
  if (!values?.final) return null;

  const noun = type === 'track' ? 'трека' : 'релиза';

  return (
    <>
      <button
        type="button"
        className="passport-trigger"
        onClick={() => setOpen(true)}
        aria-label={`Открыть паспорт ${noun}`}
        title={`Паспорт ${noun}`}
      >
        <svg className="passport-trigger-icon" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
          <path d="M4 19V5" />
          <path d="M4 19h16" />
          <path d="M8 16v-4" />
          <path d="M12 16V8" />
          <path d="M16 16v-7" />
          <path d="M20 16v-2" />
        </svg>
      </button>

      {open && (
        <div className="passport-overlay" onClick={() => setOpen(false)} role="dialog" aria-modal="true" aria-label={`Паспорт ${noun}`}>
          <div className="passport-card" onClick={(e) => e.stopPropagation()}>
            <button type="button" className="passport-close" onClick={() => setOpen(false)} aria-label="Закрыть">×</button>

            <header className="passport-head">
              <span className="passport-kicker">Паспорт {noun}</span>
              <h3 className="passport-title">{title}</h3>
            </header>

            <div className="passport-hero">
              <div className="passport-score" style={{ '--score-color': scoreColor(values.final) }}>
                <span className="passport-score-value">{animatedScore}</span>
                <span className="passport-score-max">/ 90</span>
              </div>
              <div className="passport-hero-meta">
                <div className="passport-hero-line">Средний балл по {pluralizeReview(values.count || distribution.total)}</div>
                {distribution.spread && (
                  <span className={`passport-spread passport-spread--${distribution.spread.level}`}>
                    {distribution.spread.label}
                  </span>
                )}
              </div>
            </div>

            <div className="passport-body">
              <section className="passport-section">
                <h4 className="passport-section-title">ДНК звучания</h4>
                <Radar values={values} />
              </section>

              <section className="passport-section">
                <h4 className="passport-section-title">Разброс мнений</h4>
                {distribution.total >= 1 ? (
                  <div className="passport-histogram">
                    {distribution.buckets.map((b) => {
                      const h = distribution.maxCount ? (b.count / distribution.maxCount) * 100 : 0;
                      const mid = (b.min + b.max) / 2;
                      return (
                        <div className="passport-bar-col" key={b.label}>
                          <div className="passport-bar-track">
                            <div
                              className="passport-bar-fill"
                              style={{ height: `${h}%`, background: scoreColor(mid) }}
                            >
                              {b.count > 0 && <span className="passport-bar-count">{b.count}</span>}
                            </div>
                          </div>
                          <span className="passport-bar-label">{b.label}</span>
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <p className="passport-empty">Пока мало оценок для распределения.</p>
                )}
              </section>

              <section className="passport-section passport-section--scores">
                <h4 className="passport-section-title">Оценки</h4>
                <ul className="passport-legend">
                  {RADAR_ORDER.map((key) => (
                    <li key={key}>
                      <span className="passport-legend-tag">{CRITERIA_BY_KEY[key].short}</span>
                      <span className="passport-legend-name">{CRITERIA_BY_KEY[key].title}</span>
                      <strong>{Math.round(Number(values[key]) || 0)}</strong>
                    </li>
                  ))}
                </ul>
              </section>

              {verdict && (
                <section className="passport-section passport-section--verdict">
                  <h4 className="passport-section-title">Итог по критериям</h4>
                  <div className="passport-verdict">
                    {verdict.even ? (
                      <p>Ровный по всем параметрам — без явных провалов и пиков.</p>
                    ) : (
                      <>
                        <p>
                          <span>Лидируют</span>
                          <strong className="passport-verdict-strong">{joinTitles(verdict.strong)}</strong>
                        </p>
                        <p>
                          <span>Ниже остальных</span>
                          <strong className="passport-verdict-weak">{joinTitles(verdict.weak)}</strong>
                        </p>
                      </>
                    )}
                  </div>
                </section>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default ReleasePassport;
