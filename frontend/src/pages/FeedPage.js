import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { Link } from 'react-router-dom';
import ReviewCardSmall from '../components/ReviewCardSmall';
import { SegmentedSlidingThumb } from '../components/SegmentedSlidingThumb';
import { reviewsAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { useSlidingThumb } from '../hooks/useSlidingThumb';
import './HomePage.css';

const FEED_REVIEWS = 20;

const FeedPage = () => {
  const { isAuthenticated } = useAuth();
  const feedSegRef = useRef(null);
  const [feedSource, setFeedSource] = useState('all');
  const { dims: feedThumbDims } = useSlidingThumb(feedSegRef, [feedSource, isAuthenticated]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [feedReviews, setFeedReviews] = useState([]);
  const feedInsights = useMemo(() => {
    const scores = feedReviews
      .map((review) => Number(review.final_score))
      .filter((score) => Number.isFinite(score) && score > 0);
    const authorMap = feedReviews.reduce((acc, review) => {
      const userId = review.user?.id || review.user_id;
      const username = review.user?.username || 'Автор';
      const key = userId || username;
      const current = acc.get(key) || { id: userId, username, count: 0 };
      acc.set(key, { ...current, count: current.count + 1 });
      return acc;
    }, new Map());
    const activeAuthors = Array.from(authorMap.values())
      .sort((a, b) => b.count - a.count)
      .slice(0, 3);
    const topReviews = [...feedReviews]
      .sort((a, b) => Number(b.final_score || 0) - Number(a.final_score || 0))
      .slice(0, 3)
      .map((review) => ({
        id: review.id,
        title: review.album?.title || review.track?.title || 'Релиз',
        score: Math.round(Number(review.final_score || 0)) || null,
        href: review.album?.id || review.album_id
          ? `/albums/${review.album?.id || review.album_id}`
          : `/tracks/${review.track?.id || review.track_id}`,
      }));
    const hotDiscussions = Array.from(feedReviews
      .filter((review) => review.album?.id || review.album_id)
      .reduce((acc, review) => {
        const id = review.album?.id || review.album_id;
        const current = acc.get(id) || {
          id,
          title: review.album?.title || 'Альбом',
          count: 0,
        };
        acc.set(id, { ...current, count: current.count + 1 });
        return acc;
      }, new Map())
      .values())
      .sort((a, b) => b.count - a.count)
      .slice(0, 3);

    return {
      count: feedReviews.length,
      averageScore: scores.length
        ? Math.round(scores.reduce((sum, score) => sum + score, 0) / scores.length)
        : null,
      activeAuthors,
      topReviews,
      hotDiscussions,
    };
  }, [feedReviews]);

  const loadFeed = useCallback(async () => {
    setLoading(true);
    setError('');
    const feedParams = {
      page_size: FEED_REVIEWS,
      page: 1,
      sort_by: 'created_at',
      sort_order: 'desc',
    };
    if (isAuthenticated && feedSource === 'following') {
      feedParams.following = true;
    }
    try {
      const feedRes = await reviewsAPI.getAll(feedParams);
      const rawFeed = feedRes.data?.reviews ?? [];
      const list = Array.isArray(rawFeed) ? rawFeed : [];
      setFeedReviews(list.filter((r) => r.album_id));
    } catch (e) {
      console.error(e);
      if (e.response?.status === 401) {
        setError('Войдите, чтобы видеть ленту подписок.');
      } else {
        setError('Не удалось загрузить ленту.');
      }
    } finally {
      setLoading(false);
    }
  }, [feedSource, isAuthenticated]);

  useEffect(() => {
    if (!isAuthenticated) setFeedSource('all');
  }, [isAuthenticated]);

  useEffect(() => {
    loadFeed();
  }, [loadFeed]);

  return (
    <div className="container">
      <header className="feed-page-intro">
        <p className="feed-page-lead">
          Свежие одобренные рецензии. «Подписки» — только авторы, на которых вы подписаны.
        </p>
      </header>

      {error && <div className="error-message">{error}</div>}

      {loading ? (
        <div className="home-skeleton-wrap" aria-busy="true">
          <div className="home-skeleton home-skeleton--wide" />
          <div className="home-skeleton home-skeleton--wide" />
        </div>
      ) : (
        <section className="home-section home-section--feed">
          <div className="feed-toolbar">
            <h2 className="section-title section-title--inline">Рецензии</h2>
            <div
              ref={feedSegRef}
              className="feed-segments seg-sliding-track"
              role="group"
              aria-label="Источник ленты"
            >
              <SegmentedSlidingThumb dims={feedThumbDims} />
              <button
                type="button"
                className={`feed-segment ${feedSource === 'all' ? 'feed-segment--active segment-thumb-source' : ''}`}
                onClick={() => setFeedSource('all')}
              >
                Все
              </button>
              <button
                type="button"
                className={`feed-segment ${
                  feedSource === 'following' ? 'feed-segment--active segment-thumb-source' : ''
                }`}
                onClick={() => isAuthenticated && setFeedSource('following')}
                disabled={!isAuthenticated}
                title={!isAuthenticated ? 'Войдите, чтобы видеть подписки' : undefined}
              >
                Подписки
              </button>
            </div>
          </div>

          {feedReviews.length === 0 ? (
            <div className="empty-state empty-state--soft">
              {feedSource === 'following'
                ? 'Пока пусто: подпишитесь на авторов в их профилях — их рецензии появятся здесь.'
                : 'Пока нет рецензий в ленте'}
            </div>
          ) : (
            <div className="feed-layout">
              <div className="reviews-feed-list">
                {feedReviews.map((review) => (
                  <ReviewCardSmall key={review.id} review={review} onUpdate={loadFeed} />
                ))}
              </div>
              <aside className="feed-sidebar" aria-label="Сводка ленты">
                <section className="feed-side-card feed-side-card--summary">
                  <span className="feed-side-kicker">Сейчас в ленте</span>
                  <div className="feed-side-stats">
                    <div>
                      <strong>{feedInsights.count}</strong>
                      <span>рецензий</span>
                    </div>
                    <div>
                      <strong>{feedInsights.averageScore || '—'}</strong>
                      <span>средняя</span>
                    </div>
                  </div>
                </section>

                <section className="feed-side-card">
                  <h3>Быстрые переходы</h3>
                  <div className="feed-quick-links">
                    <Link to="/tops">Топы релизов</Link>
                    <Link to="/search">Поиск музыки</Link>
                    {isAuthenticated && <button type="button" onClick={() => setFeedSource('following')}>Мои подписки</button>}
                  </div>
                </section>

                {feedInsights.activeAuthors.length > 0 && (
                  <section className="feed-side-card">
                    <h3>Активные авторы</h3>
                    <div className="feed-side-list">
                      {feedInsights.activeAuthors.map((author) => (
                        <Link className="feed-side-row" to={author.id ? `/users/${author.id}` : '/feed'} key={author.id || author.username}>
                          <span>{author.username}</span>
                          <strong>{author.count}</strong>
                        </Link>
                      ))}
                    </div>
                  </section>
                )}

                {feedInsights.hotDiscussions.length > 0 && (
                  <section className="feed-side-card">
                    <h3>Обсуждают чаще</h3>
                    <div className="feed-side-list">
                      {feedInsights.hotDiscussions.map((album) => (
                        <Link className="feed-side-release" to={`/albums/${album.id}`} key={album.id}>
                          <span>{album.title}</span>
                          <strong>{album.count}</strong>
                        </Link>
                      ))}
                    </div>
                  </section>
                )}

                {feedInsights.topReviews.length > 0 && (
                  <section className="feed-side-card">
                    <h3>Высокие оценки</h3>
                    <div className="feed-side-list">
                      {feedInsights.topReviews.map((review) => (
                        <Link className="feed-side-release" to={review.href} key={review.id}>
                          <span>{review.title}</span>
                          <strong>{review.score || '—'}</strong>
                        </Link>
                      ))}
                    </div>
                  </section>
                )}
              </aside>
            </div>
          )}
        </section>
      )}
    </div>
  );
};

export default FeedPage;
