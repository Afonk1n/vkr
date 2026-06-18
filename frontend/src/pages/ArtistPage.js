import React, { useState, useEffect, useCallback } from 'react';
import { Link, useParams } from 'react-router-dom';
import { albumsAPI } from '../services/api';
import AlbumCard from '../components/AlbumCard';
import { getImageUrl } from '../utils/imageUtils';
import './ArtistPage.css';

const VerifiedMark = () => (
  <span
    className="artist-verified"
    title="Артист зарегистрирован в «Мьюзик-рейтинг»"
    aria-label="Подтверждённый артист"
  >
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path className="artist-verified-shape" d="M12 1.8 14.4 4l3.2-.2.8 3.1 2.7 1.7-1.3 2.9 1.3 2.9-2.7 1.7-.8 3.1-3.2-.2L12 22.2 9.6 20l-3.2.2-.8-3.1-2.7-1.7 1.3-2.9-1.3-2.9 2.7-1.7.8-3.1 3.2.2L12 1.8Z" />
      <path className="artist-verified-check" d="m8.2 12.1 2.4 2.4 5.2-5.3" />
    </svg>
  </span>
);

const ArtistPage = () => {
  const { name } = useParams();
  const [artistData, setArtistData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchAlbums = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const decodedName = decodeURIComponent(name);
      const response = await albumsAPI.getByArtist(decodedName);
      setArtistData(response.data);
    } catch (err) {
      setError('Не удалось загрузить страницу артиста');
      console.error('Error fetching artist:', err);
    } finally {
      setLoading(false);
    }
  }, [name]);

  useEffect(() => {
    fetchAlbums();
  }, [fetchAlbums]);

  const albums = artistData?.albums || [];
  const artistName = artistData?.artist || decodeURIComponent(name);
  const verifiedAccount = artistData?.verified_account;
  if (loading) {
    return <div className="container"><div className="loading">Загрузка...</div></div>;
  }

  if (error || albums.length === 0) {
    return (
      <div className="container">
        <div className="error-message">{error || `Релизы артиста «${artistName}» не найдены`}</div>
      </div>
    );
  }

  const totalLikes = albums.reduce((sum, album) => sum + (album.likes?.length || 0), 0);
  const heroCover = verifiedAccount?.avatar_path || albums[0]?.cover_image_path;
  const bestAlbum = [...albums]
    .filter((album) => Number(album.average_rating) > 0)
    .sort((a, b) => Number(b.average_rating) - Number(a.average_rating))[0];

  return (
    <div className="container">
      <div className="artist-page">
        <section className="artist-hero">
          <div className="artist-hero-glow" aria-hidden />
          <div className="artist-avatar-section">
            {heroCover ? (
              <img src={getImageUrl(heroCover)} alt={artistName} className="artist-avatar" />
            ) : (
              <div className="artist-avatar-placeholder">{(artistName || 'А')[0].toUpperCase()}</div>
            )}
            <div className="artist-release-count">
              <strong>{albums.length}</strong>
              <span>{albums.length === 1 ? 'релиз' : albums.length < 5 ? 'релиза' : 'релизов'}</span>
            </div>
          </div>

          <div className="artist-info-section">
            <span className="artist-kicker">Артист в каталоге</span>
            <h1 className="artist-name">
              {artistName}
              {verifiedAccount && <VerifiedMark />}
            </h1>

            <p className="artist-description">
              {verifiedAccount?.bio || 'Каталог релизов, пользовательские оценки и обсуждения артиста в «Мьюзик-рейтинг».'}
            </p>

            <div className="artist-stats" aria-label="Статистика артиста">
              <div className="artist-stat-item"><span>Композиций</span><strong>{artistData?.total_tracks || 0}</strong></div>
              <div className="artist-stat-item"><span>Рецензий</span><strong>{artistData?.approved_reviews_count || 0}</strong></div>
              <div className="artist-stat-item"><span>Средний балл</span><strong>{Math.round(Number(artistData?.average_rating || 0)) || '—'}</strong></div>
              <div className="artist-stat-item"><span>Лайков релизов</span><strong>{totalLikes}</strong></div>
            </div>

          </div>

          {verifiedAccount && (
            <aside className="artist-account-card">
              <span className="artist-account-label"><VerifiedMark /> Официальный аккаунт</span>
              <div className="artist-account-identity">
                <span className="artist-profile-link-icon">@</span>
                <div>
                  <small>Профиль в сообществе</small>
                  <strong>{verifiedAccount.username}</strong>
                </div>
              </div>
              <p>Автор может участвовать в обсуждении и отмечать рецензии слушателей.</p>
              <Link className="artist-profile-link" to={`/users/${verifiedAccount.id}`}>
                Открыть профиль <b aria-hidden>→</b>
              </Link>
            </aside>
          )}
        </section>

        {bestAlbum && (
          <section className="artist-highlight">
            <img src={getImageUrl(bestAlbum.cover_image_path)} alt="" />
            <div className="artist-highlight-copy">
              <span>Выбор сообщества</span>
              <strong>{bestAlbum.title}</strong>
              <small>Самый высоко оценённый релиз артиста</small>
            </div>
            <div className="artist-highlight-score">
              <strong>{Math.round(Number(bestAlbum.average_rating))}</strong>
              <span>средний балл</span>
            </div>
            <Link to={`/albums/${bestAlbum.id}`}>Открыть релиз <b aria-hidden>→</b></Link>
          </section>
        )}

        <section className="artist-albums-section">
          <div className="artist-section-head">
            <div>
              <span className="artist-kicker">Дискография</span>
              <h2 className="section-title">
                Релизы
                <span className="artist-release-badge">{albums.length}</span>
              </h2>
            </div>
          </div>
          <div className="albums-grid">
            {albums.map((album) => <AlbumCard key={album.id} album={album} />)}
          </div>
        </section>
      </div>
    </div>
  );
};

export default ArtistPage;
