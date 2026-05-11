import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { albumsAPI, genresAPI, reviewsAPI, tracksAPI } from '../services/api';
import ReviewCard from '../components/ReviewCard';
import { SegmentedSlidingThumb } from '../components/SegmentedSlidingThumb';
import { useSlidingThumb } from '../hooks/useSlidingThumb';
import { getImageUrl } from '../utils/imageUtils';
import './AdminPanel.css';

const randomDuration = () => Math.floor(Math.random() * (240 - 60 + 1)) + 60;

const formatDuration = (seconds) => {
  const total = Number.isFinite(seconds) ? seconds : randomDuration();
  const minutes = Math.floor(total / 60);
  const rest = String(total % 60).padStart(2, '0');
  return `${minutes}:${rest}`;
};

const parseDuration = (value) => {
  const trimmed = value.trim();
  const match = trimmed.match(/^(\d{1,2}):([0-5]\d)$/);
  if (match) {
    const minutes = Number(match[1]);
    const seconds = Number(match[2]);
    const total = minutes * 60 + seconds;
    return total >= 30 ? total : null;
  }
  const numeric = Number(trimmed);
  if (Number.isFinite(numeric) && numeric > 0) {
    return Math.round(numeric);
  }
  return null;
};

const createDraftTrack = (index) => {
  const duration = randomDuration();
  return {
    draftId: `${Date.now()}-${index}-${Math.random().toString(16).slice(2)}`,
    title: '',
    duration,
    durationText: formatDuration(duration),
  };
};

const initialReleaseForm = {
  title: '',
  artist: '',
  genreId: '',
  releaseDate: '',
  description: '',
  coverFile: null,
  coverPreview: '',
};

const statusLabels = {
  pending: 'На проверке',
  approved: 'Одобрено',
  rejected: 'Отклонено',
};

const AdminPanel = () => {
  const { isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();
  const [activePage, setActivePage] = useState('moderation');
  const [pendingReviews, setPendingReviews] = useState([]);
  const [status, setStatus] = useState('pending');
  const [stats, setStats] = useState({ pending: 0, approved: 0, rejected: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [genres, setGenres] = useState([]);
  const [releaseForm, setReleaseForm] = useState(initialReleaseForm);
  const [trackDrafts, setTrackDrafts] = useState([createDraftTrack(1), createDraftTrack(2), createDraftTrack(3)]);
  const [releaseSaving, setReleaseSaving] = useState(false);
  const [releaseError, setReleaseError] = useState('');
  const [releaseMessage, setReleaseMessage] = useState(null);

  const adminPageRef = useRef(null);
  const { dims: adminPageThumbDims } = useSlidingThumb(adminPageRef, [activePage]);

  const statusRef = useRef(null);
  const { dims: statusThumbDims } = useSlidingThumb(statusRef, [status]);

  const filledTracks = useMemo(
    () => trackDrafts.filter((track) => track.title.trim()),
    [trackDrafts]
  );

  const fetchReviews = useCallback(async (nextStatus = status) => {
    setLoading(true);
    setError('');
    try {
      const [current, pending, approved, rejected] = await Promise.all([
        reviewsAPI.getAll({ status: nextStatus, page_size: 30 }),
        reviewsAPI.getAll({ status: 'pending', page_size: 1 }),
        reviewsAPI.getAll({ status: 'approved', page_size: 1 }),
        reviewsAPI.getAll({ status: 'rejected', page_size: 1 }),
      ]);
      setPendingReviews(current.data.reviews || []);
      setStats({
        pending: pending.data.total || 0,
        approved: approved.data.total || 0,
        rejected: rejected.data.total || 0,
      });
    } catch (err) {
      setError('Ошибка загрузки рецензий');
      console.error('Error fetching reviews:', err);
    } finally {
      setLoading(false);
    }
  }, [status]);

  const fetchGenres = useCallback(async () => {
    try {
      const response = await genresAPI.getAll();
      const nextGenres = response.data || [];
      setGenres(nextGenres);
      setReleaseForm((form) => ({
        ...form,
        genreId: form.genreId || String(nextGenres[0]?.id || ''),
      }));
    } catch (err) {
      setReleaseError('Не удалось загрузить жанры');
      console.error('Error fetching genres:', err);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated || !isAdmin) {
      navigate('/feed');
      return;
    }
    fetchReviews(status);
    fetchGenres();
  }, [isAuthenticated, isAdmin, navigate, status, fetchReviews, fetchGenres]);

  useEffect(() => {
    return () => {
      if (releaseForm.coverPreview?.startsWith('blob:')) {
        URL.revokeObjectURL(releaseForm.coverPreview);
      }
    };
  }, [releaseForm.coverPreview]);

  const handleApprove = async (reviewId) => {
    try {
      await reviewsAPI.approve(reviewId);
      fetchReviews(status);
    } catch (err) {
      setError('Ошибка при одобрении рецензии');
      console.error('Error approving review:', err);
    }
  };

  const handleReject = async (reviewId) => {
    try {
      await reviewsAPI.reject(reviewId);
      fetchReviews(status);
    } catch (err) {
      setError('Ошибка при отклонении рецензии');
      console.error('Error rejecting review:', err);
    }
  };

  const updateReleaseField = (field, value) => {
    setReleaseMessage(null);
    setReleaseError('');
    setReleaseForm((form) => ({ ...form, [field]: value }));
  };

  const handleCoverChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (releaseForm.coverPreview?.startsWith('blob:')) {
      URL.revokeObjectURL(releaseForm.coverPreview);
    }
    updateReleaseField('coverFile', file);
    setReleaseForm((form) => ({
      ...form,
      coverFile: file,
      coverPreview: URL.createObjectURL(file),
    }));
  };

  const updateTrackDraft = (draftId, patch) => {
    setReleaseMessage(null);
    setReleaseError('');
    setTrackDrafts((tracks) => tracks.map((track) => (
      track.draftId === draftId ? { ...track, ...patch } : track
    )));
  };

  const randomizeTrackDuration = (draftId) => {
    const duration = randomDuration();
    updateTrackDraft(draftId, {
      duration,
      durationText: formatDuration(duration),
    });
  };

  const addTrackDraft = () => {
    setTrackDrafts((tracks) => [...tracks, createDraftTrack(tracks.length + 1)]);
  };

  const removeTrackDraft = (draftId) => {
    setTrackDrafts((tracks) => (tracks.length > 1 ? tracks.filter((track) => track.draftId !== draftId) : tracks));
  };

  const resetReleaseForm = () => {
    if (releaseForm.coverPreview?.startsWith('blob:')) {
      URL.revokeObjectURL(releaseForm.coverPreview);
    }
    setReleaseForm({
      ...initialReleaseForm,
      genreId: String(genres[0]?.id || ''),
    });
    setTrackDrafts([createDraftTrack(1), createDraftTrack(2), createDraftTrack(3)]);
  };

  const handleCreateRelease = async (event) => {
    event.preventDefault();
    setReleaseSaving(true);
    setReleaseError('');
    setReleaseMessage(null);

    try {
      if (!releaseForm.title.trim() || !releaseForm.artist.trim() || !releaseForm.genreId) {
        throw new Error('Заполни название альбома, артиста и жанр');
      }
      if (filledTracks.length === 0) {
        throw new Error('Добавь хотя бы один трек');
      }

      const normalizedTracks = filledTracks.map((track, index) => {
        const duration = parseDuration(track.durationText);
        if (!duration) {
          throw new Error(`Проверь длительность у трека ${index + 1}`);
        }
        return { ...track, duration };
      });

      let coverImagePath = '';
      if (releaseForm.coverFile) {
        const coverResponse = await albumsAPI.uploadCover(releaseForm.coverFile);
        coverImagePath = coverResponse.data.cover_image_path || '';
      }

      const albumResponse = await albumsAPI.create({
        title: releaseForm.title.trim(),
        artist: releaseForm.artist.trim(),
        genre_id: Number(releaseForm.genreId),
        description: releaseForm.description.trim(),
        release_date: releaseForm.releaseDate || undefined,
        cover_image_path: coverImagePath,
      });

      const album = albumResponse.data;
      await Promise.all(normalizedTracks.map((track, index) => tracksAPI.create({
        album_id: album.id,
        title: track.title.trim(),
        duration: track.duration,
        track_number: index + 1,
        genre_ids: [Number(releaseForm.genreId)],
      })));

      setReleaseMessage({
        title: 'Релиз создан',
        text: `${album.title} теперь есть в каталоге, добавлено треков: ${normalizedTracks.length}.`,
        albumId: album.id,
      });
      resetReleaseForm();
    } catch (err) {
      setReleaseError(err.message || 'Не удалось создать релиз');
      console.error('Error creating release:', err);
    } finally {
      setReleaseSaving(false);
    }
  };

  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return (
    <div className="container">
      <div className="admin-panel">
        <div className="admin-header">
          <div>
            <span className="admin-kicker">Панель управления</span>
            <h1>Панель модерации</h1>
            <p>Модерация рецензий и быстрый конструктор музыкального каталога для демо-данных.</p>
          </div>
          <div
            ref={adminPageRef}
            className="admin-page-switch seg-sliding-track"
            role="tablist"
            aria-label="Раздел админки"
          >
            <SegmentedSlidingThumb dims={adminPageThumbDims} />
            <button
              type="button"
              role="tab"
              aria-selected={activePage === 'moderation'}
              className={`admin-page-segment ${activePage === 'moderation' ? 'admin-page-segment--active segment-thumb-source' : ''}`}
              onClick={() => setActivePage('moderation')}
            >
              Модерация
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={activePage === 'release'}
              className={`admin-page-segment ${activePage === 'release' ? 'admin-page-segment--active segment-thumb-source' : ''}`}
              onClick={() => setActivePage('release')}
            >
              Новый релиз
            </button>
          </div>
        </div>

        {activePage === 'moderation' ? (
          <>
            <div className="admin-moderation-top">
              <div
                ref={statusRef}
                className="admin-status-switch seg-sliding-track"
                role="tablist"
                aria-label="Статус рецензий"
              >
                <SegmentedSlidingThumb dims={statusThumbDims} />
                {Object.entries(statusLabels).map(([key, label]) => (
                  <button
                    key={key}
                    type="button"
                    role="tab"
                    aria-selected={status === key}
                    className={`admin-status-segment ${status === key ? 'admin-status-segment--active segment-thumb-source' : ''}`}
                    onClick={() => setStatus(key)}
                  >
                    <span>{label}</span>
                    <strong>{stats[key]}</strong>
                  </button>
                ))}
              </div>
            </div>

            {error && <div className="error-message">{error}</div>}

            {loading ? (
              <div className="loading">Загрузка...</div>
            ) : pendingReviews.length === 0 ? (
              <div className="empty-state admin-empty-state">
                {status === 'pending' ? 'Нет рецензий на модерации' : 'В этом статусе пока пусто'}
              </div>
            ) : (
              <div className="reviews-list">
                {pendingReviews.map((review) => (
                  <ReviewCard
                    key={review.id}
                    review={review}
                    hideLike={true}
                    moderationActions={
                      <div className="moderation-actions">
                        <button
                          onClick={() => handleApprove(review.id)}
                          className="btn-approve"
                          title="Одобрить"
                        >
                          Одобрить
                        </button>
                        <button
                          onClick={() => handleReject(review.id)}
                          className="btn-reject"
                          title="Отклонить"
                        >
                          Отклонить
                        </button>
                      </div>
                    }
                  />
                ))}
              </div>
            )}
          </>
        ) : (
          <form className="admin-release-panel" onSubmit={handleCreateRelease}>
            <div className="admin-release-heading">
              <div>
                <span className="admin-kicker">Каталог</span>
                <h2>Добавить альбом и треки</h2>
                <p>Длительности можно оставить сгенерированными или поправить вручную в формате 3:24.</p>
              </div>
              <button type="submit" className="admin-save-release" disabled={releaseSaving}>
                {releaseSaving ? 'Сохранение...' : 'Сохранить релиз'}
              </button>
            </div>

            {releaseError && <div className="error-message">{releaseError}</div>}
            {releaseMessage && (
              <div className="admin-success-message">
                <strong>{releaseMessage.title}</strong>
                <span>{releaseMessage.text}</span>
                <Link to={`/albums/${releaseMessage.albumId}`}>Открыть альбом</Link>
              </div>
            )}

            <div className="admin-release-layout">
              <label className="admin-cover-uploader">
                <span className="admin-cover-preview">
                  {releaseForm.coverPreview ? (
                    <img
                      src={releaseForm.coverPreview.startsWith('blob:') ? releaseForm.coverPreview : getImageUrl(releaseForm.coverPreview)}
                      alt="Предпросмотр обложки"
                    />
                  ) : (
                    <span>Обложка</span>
                  )}
                </span>
                <input type="file" accept="image/png,image/jpeg,image/webp" onChange={handleCoverChange} />
                <strong>Загрузить обложку</strong>
                <small>JPG, PNG или WEBP до 8 МБ</small>
              </label>

              <div className="admin-form-grid">
                <label className="admin-field">
                  <span>Название альбома</span>
                  <input
                    value={releaseForm.title}
                    onChange={(event) => updateReleaseField('title', event.target.value)}
                    placeholder="Например, Новый альбом"
                  />
                </label>
                <label className="admin-field">
                  <span>Артист</span>
                  <input
                    value={releaseForm.artist}
                    onChange={(event) => updateReleaseField('artist', event.target.value)}
                    placeholder="Имя артиста"
                  />
                </label>
                <label className="admin-field">
                  <span>Жанр</span>
                  <select
                    value={releaseForm.genreId}
                    onChange={(event) => updateReleaseField('genreId', event.target.value)}
                  >
                    {genres.map((genre) => (
                      <option key={genre.id} value={genre.id}>{genre.name}</option>
                    ))}
                  </select>
                </label>
                <label className="admin-field">
                  <span>Дата релиза</span>
                  <input
                    type="date"
                    value={releaseForm.releaseDate}
                    onChange={(event) => updateReleaseField('releaseDate', event.target.value)}
                  />
                </label>
                <label className="admin-field admin-field--wide">
                  <span>Описание</span>
                  <textarea
                    value={releaseForm.description}
                    onChange={(event) => updateReleaseField('description', event.target.value)}
                    rows={4}
                    placeholder="Коротко о релизе для страницы альбома"
                  />
                </label>
              </div>
            </div>

            <section className="admin-track-builder" aria-label="Треки релиза">
              <div className="admin-track-builder-head">
                <div>
                  <h3>Треклист</h3>
                  <p>{filledTracks.length} треков будет создано вместе с альбомом.</p>
                </div>
                <button type="button" className="admin-soft-button" onClick={addTrackDraft}>
                  Добавить трек
                </button>
              </div>

              <div className="admin-track-list">
                {trackDrafts.map((track, index) => (
                  <div className="admin-track-row" key={track.draftId}>
                    <span className="admin-track-number">{index + 1}</span>
                    <input
                      value={track.title}
                      onChange={(event) => updateTrackDraft(track.draftId, { title: event.target.value })}
                      placeholder="Название трека"
                    />
                    <input
                      className="admin-track-duration"
                      value={track.durationText}
                      onChange={(event) => updateTrackDraft(track.draftId, { durationText: event.target.value })}
                      placeholder="3:24"
                    />
                    <button
                      type="button"
                      className="admin-icon-button"
                      onClick={() => randomizeTrackDuration(track.draftId)}
                      title="Сгенерировать длительность"
                    >
                      ↻
                    </button>
                    <button
                      type="button"
                      className="admin-icon-button"
                      onClick={() => removeTrackDraft(track.draftId)}
                      title="Убрать строку"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            </section>
          </form>
        )}
      </div>
    </div>
  );
};

export default AdminPanel;
