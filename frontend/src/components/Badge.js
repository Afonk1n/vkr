import React, { useState, useEffect, useRef, useCallback } from 'react';
import { createPortal } from 'react-dom';
import './Badge.css';

/** Закрытие сразу; зазор до поповера перекрыт прозрачным ::before в Badge.css */
const HIDE_MS = 0;

/** Если бэкенд без поля criteria (старый ответ) */
const CRITERIA_FALLBACK = {
  'Легенда критики': '51 и более одобренных рецензий на сайте.',
  'Мастер рецензий': 'От 21 до 50 одобренных рецензий.',
  'Опытный критик': 'От 6 до 20 одобренных рецензий.',
  'Начинающий критик': 'Первая одобренная рецензия.',
  Универсал: 'Не менее 5 разных жанров в одобренных рецензиях.',
};

function criteriaText(badge, profileContext) {
  if (badge.criteria && String(badge.criteria).trim()) return badge.criteria;
  if (CRITERIA_FALLBACK[badge.name]) return CRITERIA_FALLBACK[badge.name];
  return profileContext === 'other'
    ? 'Звание выдаётся автоматически по одобренным рецензиям в этом профиле.'
    : 'Звание выдаётся автоматически по вашим одобренным рецензиям.';
}

function factHeading(profileContext) {
  return profileContext === 'other' ? 'По данным в профиле:' : 'Сейчас у вас:';
}

const Badge = ({ badge, badgeId = 0, profileContext = 'self' }) => {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState({ left: 0, top: 0 });
  const [touchMode, setTouchMode] = useState(false);
  const wrapRef = useRef(null);
  const hideTimerRef = useRef(null);

  const clearHideTimer = useCallback(() => {
    if (hideTimerRef.current) {
      clearTimeout(hideTimerRef.current);
      hideTimerRef.current = null;
    }
  }, []);

  const close = useCallback(() => {
    clearHideTimer();
    setOpen(false);
  }, [clearHideTimer]);

  const scheduleHide = useCallback(() => {
    clearHideTimer();
    hideTimerRef.current = window.setTimeout(() => {
      hideTimerRef.current = null;
      setOpen(false);
    }, HIDE_MS);
  }, [clearHideTimer]);

  useEffect(() => {
    const mq = window.matchMedia('(hover: none)');
    const sync = () => setTouchMode(mq.matches);
    sync();
    mq.addEventListener('change', sync);
    return () => mq.removeEventListener('change', sync);
  }, []);

  useEffect(() => {
    if (!open) return undefined;
    const onKey = (e) => {
      if (e.key === 'Escape') close();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, close]);

  useEffect(() => () => clearHideTimer(), [clearHideTimer]);

  if (!badge) return null;

  const updatePosition = () => {
    const el = wrapRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const w = 280;
    const margin = 10;
    let left = r.left + r.width / 2 - w / 2;
    left = Math.max(margin, Math.min(left, window.innerWidth - w - margin));
    const top = Math.min(r.bottom + 8, window.innerHeight - 200);
    setPos({ left, top: Math.max(margin, top) });
  };

  const showPopover = () => {
    clearHideTimer();
    updatePosition();
    setOpen(true);
  };

  const onBadgeClick = (e) => {
    e.stopPropagation();
    if (!touchMode) return;
    setOpen((v) => {
      if (v) {
        clearHideTimer();
        return false;
      }
      updatePosition();
      return true;
    });
  };

  const tooltipId = `badge-tip-${badgeId}`;
  const heading = factHeading(profileContext);

  const popover =
    open &&
    createPortal(
      <>
        {touchMode && (
          <button type="button" className="badge-popover-backdrop badge-popover-backdrop--light" aria-label="Закрыть" onClick={close} />
        )}
        <div
          className="badge-popover"
          style={{ left: pos.left, top: pos.top, width: 280 }}
          role="tooltip"
          id={tooltipId}
          onMouseEnter={clearHideTimer}
          onMouseLeave={scheduleHide}
          onClick={(e) => e.stopPropagation()}
        >
          <div className="badge-popover-header">
            <span className="badge-popover-icon" aria-hidden>
              {badge.icon}
            </span>
            <h4 className="badge-popover-title">{badge.name}</h4>
            {touchMode && (
              <button type="button" className="badge-popover-close" onClick={close} aria-label="Закрыть">
                ×
              </button>
            )}
          </div>
          <p className="badge-popover-criteria">{criteriaText(badge, profileContext)}</p>
          {badge.description && (
            <p className="badge-popover-fact">
              <strong>{heading}</strong> {badge.description}
            </p>
          )}
        </div>
      </>,
      document.body
    );

  return (
    <>
      <button
        type="button"
        ref={wrapRef}
        className="badge"
        data-priority={badge.priority}
        onMouseEnter={touchMode ? undefined : showPopover}
        onMouseLeave={touchMode ? undefined : scheduleHide}
        onClick={onBadgeClick}
        aria-expanded={open}
        aria-haspopup="true"
        aria-describedby={open ? tooltipId : undefined}
      >
        <span className="badge-icon">{badge.icon}</span>
        <span className="badge-name">{badge.name}</span>
      </button>
      {popover}
    </>
  );
};

export default Badge;
