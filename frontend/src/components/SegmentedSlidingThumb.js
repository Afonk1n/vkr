import React from 'react';
import './SegmentedSliding.css';

/** Плашка-индикатор под активным сегментом (абсолютное позиционирование + transform) */
export function SegmentedSlidingThumb({ dims }) {
  const ready = dims && dims.w > 0;
  return (
    <span
      aria-hidden
      className={`seg-sliding-thumb ${ready ? 'seg-sliding-thumb--ready' : ''}`.trim()}
      style={{
        width: ready ? dims.w : 0,
        height: ready ? dims.h : 0,
        transform: `translate(${ready ? dims.x : 0}px, ${ready ? dims.y : 0}px)`,
      }}
    />
  );
}
