import React, { useMemo, useState } from 'react';
import BadgeList from './BadgeList';
import { getImageUrl } from '../utils/imageUtils';
import './ProfileDashboard.css';

const LEVELS = [
  { name: 'Бронзовый уровень', min: 0, tone: 'bronze' },
  { name: 'Серебряный уровень', min: 2500, tone: 'silver' },
  { name: 'Золотой уровень', min: 8000, tone: 'gold' },
  { name: 'Изумрудный уровень', min: 18000, tone: 'emerald' },
  { name: 'Платиновый уровень', min: 36000, tone: 'platinum' },
];

const tabs = [
  { id: 'preferences', label: 'Предпочтения' },
  { id: 'reviews', label: 'Рецензии и оценки' },
  { id: 'liked', label: 'Понравилось' },
  { id: 'achievements', label: 'Достижения' },
];

const parseSocialLinks = (value) => {
  if (!value) return {};
  if (typeof value === 'object') return value;

  try {
    return JSON.parse(value);
  } catch (error) {
    return {};
  }
};

const normalizeImage = (path) => {
  const imageUrl = getImageUrl(path);
  return imageUrl || null;
};

const getAlbumFromReview = (review) => review?.album || review?.track?.album || null;

const uniqueByName = (items) => {
  const seen = new Set();
  return items.filter((item) => {
    if (!item?.title) return false;
    const key = item.title.toLowerCase();
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
};

const buildPreferenceSections = (profileUser, reviews) => {
  const favoriteAlbums = profileUser?.favorite_albums || [];
  const reviewAlbums = reviews.map(getAlbumFromReview).filter(Boolean);
  const albums = uniqueByName([...favoriteAlbums, ...reviewAlbums].map((album) => ({
    title: album.title,
    subtitle: album.artist?.name || album.artist_name || 'Артист',
    image: normalizeImage(album.cover_image_path),
  }))).slice(0, 3);

  const artists = uniqueByName([
    ...favoriteAlbums,
    ...reviewAlbums,
  ].map((album) => ({
    title: album.artist?.name || album.artist_name,
    subtitle: 'Артист',
    image: normalizeImage(album.artist?.avatar_path || album.cover_image_path),
  }))).slice(0, 3);

  const tracks = uniqueByName(reviews
    .filter((review) => review.track)
    .map((review) => ({
      title: review.track.title,
      subtitle: review.track.album?.artist?.name || review.track.album?.artist_name || 'Трек',
      image: normalizeImage(review.track.album?.cover_image_path),
    }))).slice(0, 3);

  return [
    { title: 'Артисты', items: artists },
    { title: 'Альбомы', items: albums },
    { title: 'Треки', items: tracks },
  ];
};

const getLevelInfo = (profileUser, reviews) => {
  const stats = profileUser?.stats || {};
  const reviewsCount = Number(stats.total_reviews ?? reviews.length ?? 0);
  const receivedLikes = Number(stats.total_likes_received ?? 0);
  const likedGiven = Number(stats.total_likes_given ?? 0);
  const authorLikes = Number(stats.author_likes_received ?? 0);
  const avgScore = Number(stats.avg_score ?? 0);
  const points = Math.round(
    reviewsCount * 320
    + receivedLikes * 55
    + likedGiven * 12
    + authorLikes * 240
    + avgScore * 8
  );

  const current = [...LEVELS].reverse().find((level) => points >= level.min) || LEVELS[0];
  const currentIndex = LEVELS.findIndex((level) => level.name === current.name);
  const next = LEVELS[currentIndex + 1] || null;
  const progress = next
    ? Math.min(100, Math.max(4, ((points - current.min) / (next.min - current.min)) * 100))
    : 100;

  return { current, next, points, progress, reviewsCount, receivedLikes, likedGiven, authorLikes };
};

const PreferenceItem = ({ item }) => (
  <div className="profile-preference-item">
    <div className="profile-preference-cover">
      {item.image ? (
        <img src={item.image} alt={item.title} />
      ) : (
        <span>{item.title.charAt(0).toUpperCase()}</span>
      )}
    </div>
    <div className="profile-preference-text">
      <strong>{item.title}</strong>
      <span>{item.subtitle}</span>
    </div>
  </div>
);

const EmptyPanel = ({ children }) => (
  <div className="profile-empty-panel">{children}</div>
);

const ProfileDashboard = ({
  profileUser,
  reviews = [],
  reviewsLoading = false,
  reviewsError = '',
  likedReviews = [],
  likedReviewsLoading = false,
  isOwner = false,
  onEditProfile,
  renderReviews,
  renderLikedReviews,
  followSlot,
  emptyReviewsText = 'Пока нет рецензий',
}) => {
  const [activeTab, setActiveTab] = useState('preferences');
  const socialLinks = useMemo(() => parseSocialLinks(profileUser?.social_links), [profileUser]);
  const level = useMemo(() => getLevelInfo(profileUser, reviews), [profileUser, reviews]);
  const preferenceSections = useMemo(() => buildPreferenceSections(profileUser, reviews), [profileUser, reviews]);
  const stats = profileUser?.stats || {};

  const avgScore = stats.avg_score || '—';
  const ratingsWithoutReview = stats.ratings_without_review ?? 0;

  return (
    <div className="profile-dashboard">
      <aside className="profile-sidebar">
        <section className="profile-card profile-identity-card">
          <div className="profile-avatar-wrap">
            {profileUser?.avatar_path && normalizeImage(profileUser.avatar_path) ? (
              <img
                src={normalizeImage(profileUser.avatar_path)}
                alt={profileUser.username}
                className="profile-avatar"
              />
            ) : (
              <div className="profile-avatar-placeholder">
                {profileUser?.username?.charAt(0).toUpperCase() || 'U'}
              </div>
            )}
          </div>
          <h1 className="profile-username">
            {profileUser?.username || 'Пользователь'}
            {profileUser?.is_verified_artist && (
              <span className="verified-badge" title="Верифицированный артист">✓</span>
            )}
          </h1>
          {profileUser?.email && isOwner && <p className="profile-email">{profileUser.email}</p>}
          {profileUser?.created_at && (
            <p className="profile-joined">
              Дата регистрации: {new Date(profileUser.created_at).toLocaleDateString('ru-RU')}
            </p>
          )}
          {profileUser?.bio && <p className="profile-bio-text">{profileUser.bio}</p>}
          {profileUser?.is_admin && <span className="admin-badge">Администратор</span>}
          {followSlot}
          {(socialLinks.vk || socialLinks.telegram || socialLinks.max) && (
            <div className="profile-social-links">
              {socialLinks.vk && <a href={socialLinks.vk} target="_blank" rel="noopener noreferrer" className="social-link">VK</a>}
              {socialLinks.telegram && (
                <a href={`https://t.me/${socialLinks.telegram.replace('@', '')}`} target="_blank" rel="noopener noreferrer" className="social-link">Telegram</a>
              )}
              {socialLinks.max && (
                <a href={socialLinks.max.startsWith('http') ? socialLinks.max : `https://max.ru/${socialLinks.max.replace('@', '')}`} target="_blank" rel="noopener noreferrer" className="social-link">MAX</a>
              )}
            </div>
          )}
          {isOwner && onEditProfile && (
            <button className="btn-edit-profile" type="button" onClick={onEditProfile}>
              Редактировать профиль
            </button>
          )}
        </section>

        <section className="profile-card profile-level-card">
          <h2>Уровень профиля</h2>
          <div className="profile-level-content">
            <div className={`profile-level-gem profile-level-gem--${level.current.tone}`}>
              <span>{level.current.name.charAt(0)}</span>
            </div>
            <div className="profile-level-copy">
              <strong>{level.current.name}</strong>
              <span>{level.points.toLocaleString('ru-RU')} баллов сообщества</span>
            </div>
          </div>
          <div className="profile-level-progress">
            <span style={{ width: `${level.progress}%` }} />
          </div>
          <p className="profile-level-next">
            {level.next ? `До ${level.next.name}: ${(level.next.min - level.points).toLocaleString('ru-RU')} баллов` : 'Максимальный уровень'}
          </p>
        </section>

        <section className="profile-card profile-stat-panel">
          <h2>Статистика</h2>
          <div className="profile-stat-list">
            <div><span>Рецензий</span><strong>{level.reviewsCount}</strong></div>
            <div><span>Оценок без рецензии</span><strong>{ratingsWithoutReview}</strong></div>
            <div><span>Получено лайков</span><strong>{level.receivedLikes}</strong></div>
            <div><span>Поставлено лайков</span><strong>{level.likedGiven}</strong></div>
            <div><span>Авторских лайков</span><strong>{level.authorLikes}</strong></div>
            <div><span>Средняя оценка</span><strong>{avgScore}</strong></div>
            {stats.top_genre && <div><span>Любимый жанр</span><strong>{stats.top_genre}</strong></div>}
          </div>
        </section>
      </aside>

      <section className="profile-main-panel">
        <div className="profile-showcase">
          <div>
            <span className="profile-showcase-kicker">Музыкальный профиль</span>
            <strong>{profileUser?.username || 'Пользователь'}</strong>
            <p>Любимые релизы, оценки и достижения в одном месте.</p>
          </div>
          <div className="profile-showcase-picks" aria-hidden="true">
            {preferenceSections.flatMap((section) => section.items).slice(0, 5).map((item) => (
              <div className="profile-showcase-pick" key={`${item.title}-${item.subtitle}`}>
                {item.image ? <img src={item.image} alt="" /> : <span>{item.title.charAt(0).toUpperCase()}</span>}
              </div>
            ))}
          </div>
        </div>

        <div className="profile-tabs" role="tablist" aria-label="Разделы профиля">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              className={`profile-tab ${activeTab === tab.id ? 'profile-tab--active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {activeTab === 'preferences' && (
          <div className="profile-preferences-grid">
            {preferenceSections.map((section) => (
              <section className="profile-preference-section" key={section.title}>
                <h2>{section.title}</h2>
                {section.items.length > 0 ? (
                  <div className="profile-preference-list">
                    {section.items.map((item) => <PreferenceItem item={item} key={`${section.title}-${item.title}`} />)}
                  </div>
                ) : (
                  <EmptyPanel>Нет данных для этого блока</EmptyPanel>
                )}
              </section>
            ))}
          </div>
        )}

        {activeTab === 'reviews' && (
          <section className="profile-reviews">
            <h2>{isOwner ? 'Мои рецензии' : 'Рецензии пользователя'} ({reviews.length})</h2>
            {reviewsError && <div className="error-message">{reviewsError}</div>}
            {reviewsLoading ? (
              <div className="loading">Загрузка...</div>
            ) : reviews.length === 0 ? (
              <div className="empty-state">{emptyReviewsText}</div>
            ) : (
              renderReviews?.()
            )}
          </section>
        )}

        {activeTab === 'liked' && (
          <section className="profile-panel-card">
            <h2>Понравилось ({likedReviews.length})</h2>
            {likedReviewsLoading ? (
              <div className="loading">Загрузка...</div>
            ) : likedReviews.length > 0 ? (
              renderLikedReviews?.()
            ) : (
              <EmptyPanel>Пока нет понравившихся рецензий</EmptyPanel>
            )}
          </section>
        )}

        {activeTab === 'achievements' && (
          <section className="profile-panel-card">
            <h2>Достижения</h2>
            {profileUser?.badges && profileUser.badges.length > 0 ? (
              <BadgeList badges={profileUser.badges} profileContext={isOwner ? 'self' : 'other'} />
            ) : (
              <EmptyPanel>Достижения пока не открыты</EmptyPanel>
            )}
          </section>
        )}
      </section>
    </div>
  );
};

export default ProfileDashboard;
