import React from 'react';
import './Skeleton.css';

// Базовый «мерцающий» блок. Размеры задаются через пропсы.
export const Skeleton = ({ width = '100%', height = 16, radius, className = '', style = {} }) => (
  <span
    className={`skeleton ${className}`.trim()}
    style={{ width, height, ...(radius != null ? { borderRadius: radius } : {}), ...style }}
    aria-hidden="true"
  />
);

// Скелет страницы альбома/трека: обложка + информация + строки.
export const DetailSkeleton = () => (
  <div className="container" aria-busy="true">
    <div className="skeleton-detail">
      <Skeleton className="skeleton-detail-cover" />
      <div className="skeleton-detail-info">
        <Skeleton width="60%" height={34} />
        <Skeleton width="40%" height={20} />
        <Skeleton width="30%" height={20} />
        <Skeleton width={120} height={44} radius="var(--radius-lg)" style={{ marginTop: '0.5rem' }} />
      </div>
    </div>
    <div className="skeleton-lines">
      {[90, 80, 85, 70].map((w, i) => (
        <Skeleton key={i} width={`${w}%`} height={56} radius="var(--radius-lg)" />
      ))}
    </div>
  </div>
);

// Скелет сетки карточек (каталог/поиск).
export const CardGridSkeleton = ({ count = 8 }) => (
  <div className="skeleton-grid" aria-busy="true">
    {Array.from({ length: count }).map((_, i) => (
      <Skeleton key={i} className="skeleton-card" />
    ))}
  </div>
);

// Скелет вертикального списка (рецензии, модерация).
export const ListSkeleton = ({ count = 4 }) => (
  <div className="skeleton-list" aria-busy="true">
    {Array.from({ length: count }).map((_, i) => (
      <Skeleton key={i} className="skeleton-row" />
    ))}
  </div>
);

export default Skeleton;
