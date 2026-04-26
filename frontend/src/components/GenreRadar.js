import React from 'react';

/** Внутренний размер сетки; вокруг — отступ под длинные подписи жанров */
const CHART = 280;
const VIEW_PAD = 58;
const SIZE = CHART + VIEW_PAD * 2;
const CENTER = VIEW_PAD + CHART / 2;
const MAX_RADIUS = 86;
const LEVELS = 4;
const LABEL_OFFSET = 38;

function polarToXY(angleDeg, radius) {
  const rad = ((angleDeg - 90) * Math.PI) / 180;
  return {
    x: CENTER + radius * Math.cos(rad),
    y: CENTER + radius * Math.sin(rad),
  };
}

function pointsToPath(points) {
  return points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' ') + ' Z';
}

const GenreRadar = ({ genreStats }) => {
  if (!genreStats || genreStats.length < 3) return null;

  const data = genreStats.slice(0, 8);
  const n = data.length;
  const maxCount = Math.max(...data.map((g) => g.count), 1);
  const angleStep = 360 / n;

  const gridPaths = Array.from({ length: LEVELS }, (_, lvl) => {
    const r = (MAX_RADIUS * (lvl + 1)) / LEVELS;
    const pts = data.map((_, i) => polarToXY(i * angleStep, r));
    return pointsToPath(pts);
  });

  const axes = data.map((_, i) => {
    const end = polarToXY(i * angleStep, MAX_RADIUS);
    return { x2: end.x, y2: end.y };
  });

  const dataPoints = data.map((g, i) => {
    const r = Math.max((g.count / maxCount) * MAX_RADIUS, 4);
    return polarToXY(i * angleStep, r);
  });

  const labels = data.map((g, i) => {
    const pos = polarToXY(i * angleStep, MAX_RADIUS + LABEL_OFFSET);
    const angle = i * angleStep;
    let anchor = 'middle';
    if (angle > 25 && angle < 155) anchor = 'start';
    else if (angle > 205 && angle < 335) anchor = 'end';
    return { ...pos, text: g.name, count: g.count, anchor };
  });

  return (
    <div className="genre-radar-wrap">
      <svg
        className="genre-radar-svg"
        width="100%"
        height="auto"
        viewBox={`0 0 ${SIZE} ${SIZE}`}
        style={{ maxWidth: 400 }}
      >
        {gridPaths.map((d, i) => (
          <path key={i} d={d} fill="none" stroke="var(--border-color)" strokeWidth="1" />
        ))}
        {axes.map((a, i) => (
          <line key={i} x1={CENTER} y1={CENTER} x2={a.x2} y2={a.y2} stroke="var(--border-color)" strokeWidth="1" />
        ))}
        <path
          d={pointsToPath(dataPoints)}
          fill="var(--accent-color, #268bd2)"
          fillOpacity="0.22"
          stroke="var(--accent-color, #268bd2)"
          strokeWidth="2"
        />
        {dataPoints.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r="3.5" fill="var(--accent-color, #268bd2)" />
        ))}
        {labels.map((l, i) => (
          <text
            key={i}
            x={l.x}
            y={l.y}
            textAnchor={l.anchor}
            dominantBaseline="middle"
            fontSize="11"
            fill="var(--text-muted, #888)"
          >
            {l.text}
          </text>
        ))}
      </svg>
    </div>
  );
};

export default GenreRadar;
