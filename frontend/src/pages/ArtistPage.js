import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { albumsAPI } from '../services/api';
import AlbumCard from '../components/AlbumCard';
import { getImageUrl } from '../utils/imageUtils';
import './ArtistPage.css';

const parseSocialLinks = (value) => {
  if (!value) return {};
  if (typeof value === 'object') return value;
  try {
    return JSON.parse(value);
  } catch (_) {
    return {};
  }
};

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
  const socialLinks = useMemo(
    () => parseSocialLinks(verifiedAccount?.social_links),
    [verifiedAccount?.social_links]
  );

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
          </div>

          <div className="artist-info-section">
            <div className="artist-title-row">
              <div>
                <span className="artist-kicker">Страница артиста</span>
                <h1 className="artist-name">
                  {artistName}
                  {verifiedAccount && <span className="artist-verified" title="Артист зарегистрирован в «Мьюзик-рейтинг»">✓</span>}
                </h1>
              </div>
              {verifiedAccount && (
                <Link className="artist-profile-link" to={`/users/${verifiedAccount.id}`}>
                  <span className="artist-profile-link-icon">@</span>
                  <span>
                    <small>Подтверждённый аккаунт</small>
                    <strong>{verifiedAccount.username}</strong>
                  </span>
                  <b aria-hidden>→</b>
                </Link>
              )}
            </div>

            <p className="artist-description">
              {verifiedAccount?.bio || 'Каталог релизов, пользовательские оценки и обсуждения артиста в «Мьюзик-рейтинг».'}
            </p>

            <div className="artist-stats">
              <div className="artist-stat-item"><span>Релизов</span><strong>{albums.length}</strong></div>
              <div className="artist-stat-item"><span>Композиций</span><strong>{artistData?.total_tracks || 0}</strong></div>
              <div className="artist-stat-item"><span>Рецензий</span><strong>{artistData?.approved_reviews_count || 0}</strong></div>
              <div className="artist-stat-item"><span>Средний балл</span><strong>{Math.round(Number(artistData?.average_rating || 0)) || '—'}</strong></div>
              <div className="artist-stat-item"><span>Лайков релизов</span><strong>{totalLikes}</strong></div>
            </div>

            {(socialLinks.vk || socialLinks.telegram) && (
              <div className="artist-socials">
                {socialLinks.vk && <a href={socialLinks.vk} target="_blank" rel="noreferrer">VK</a>}
                {socialLinks.telegram && <a href={socialLinks.telegram} target="_blank" rel="noreferrer">Telegram</a>}
              </div>
            )}
          </div>
        </section>

        {bestAlbum && (
          <section className="artist-highlight">
            <span>Самый высоко оценённый релиз</span>
            <Link to={`/albums/${bestAlbum.id}`}>
              <strong>{bestAlbum.title}</strong>
              <b>{Math.round(Number(bestAlbum.average_rating))} баллов</b>
            </Link>
          </section>
        )}

        <section className="artist-albums-section">
          <div className="artist-section-head">
            <div>
              <span className="artist-kicker">Дискография</span>
              <h2 className="section-title">Релизы</h2>
            </div>
            <span>{albums.length}</span>
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
