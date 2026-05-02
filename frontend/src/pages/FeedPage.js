import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
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
      const username = review.user?.username || 'Автор';
      acc.set(username, (acc.get(username) || 0) + 1);
      return acc;
    }, new Map());
    const activeAuthors = Array.from(authorMap.entries())
      .sort((a, b) => b[1] - a[1])
      .slice(0, 3);
    const topReviews = [...feedReviews]
      .sort((a, b) => Number(b.final_score || 0) - Number(a.final_score || 0))
      .slice(0, 3);

    return {
      count: feedReviews.length,
      averageScore: scores.length
        ? Math.round(scores.reduce((sum, score) => sum + score, 0) / scores.length)
        : null,
      activeAuthors,
      topReviews,
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

                {feedInsights.activeAuthors.length > 0 && (
                  <section className="feed-side-card">
                    <h3>Активные авторы</h3>
                    <div className="feed-side-list">
                      {feedInsights.activeAuthors.map(([name, count]) => (
                        <div className="feed-side-row" key={name}>
                          <span>{name}</span>
                          <strong>{count}</strong>
                        </div>
                      ))}
                    </div>
                  </section>
                )}

                {feedInsights.topReviews.length > 0 && (
                  <section className="feed-side-card">
                    <h3>Высокие оценки</h3>
                    <div className="feed-side-list">
                      {feedInsights.topReviews.map((review) => (
                        <div className="feed-side-release" key={review.id}>
                          <span>{review.album?.title || review.track?.title || 'Релиз'}</span>
                          <strong>{Math.round(Number(review.final_score || 0)) || '—'}</strong>
                        </div>
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
