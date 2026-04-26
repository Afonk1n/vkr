import React, { useState, useEffect, useCallback, useRef } from 'react';
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
            {!isAuthenticated && (
              <Link className="feed-login-hint" to="/login">
                Войти
              </Link>
            )}
          </div>

          {feedReviews.length === 0 ? (
            <div className="empty-state empty-state--soft">
              {feedSource === 'following'
                ? 'Пока пусто: подпишитесь на авторов в их профилях — их рецензии появятся здесь.'
                : 'Пока нет рецензий в ленте'}
            </div>
          ) : (
            <div className="reviews-feed-list">
              {feedReviews.map((review) => (
                <ReviewCardSmall key={review.id} review={review} onUpdate={loadFeed} />
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  );
};

export default FeedPage;
