import { useCallback, useEffect, useLayoutEffect, useState } from 'react';

/**
 * Плавный «пилюльный» индикатор под активным сегментом.
 * Внутри контейнера ровно один потомок с классом .segment-thumb-source
 */
export function useSlidingThumb(trackRef, deps) {
  const [dims, setDims] = useState(null);

  const measure = useCallback(() => {
    const track = trackRef.current;
    if (!track) return;
    const active = track.querySelector('.segment-thumb-source');
    if (!active) {
      setDims(null);
      return;
    }
    const tr = track.getBoundingClientRect();
    const ar = active.getBoundingClientRect();
    setDims({
      x: ar.left - tr.left,
      y: ar.top - tr.top,
      w: ar.width,
      h: ar.height,
    });
  }, [trackRef]);

  const depKey = Array.isArray(deps) ? deps.join('\u0001') : String(deps);

  useLayoutEffect(() => {
    measure();
  }, [measure, depKey]);

  useEffect(() => {
    const onResize = () => measure();
    window.addEventListener('resize', onResize);
    const track = trackRef.current;
    const ro = track ? new ResizeObserver(onResize) : null;
    if (track && ro) {
      ro.observe(track);
      track.querySelectorAll('a, button').forEach((el) => ro.observe(el));
    }
    return () => {
      window.removeEventListener('resize', onResize);
      if (ro) ro.disconnect();
    };
  }, [measure, trackRef, depKey]);

  return { dims, measure };
}
