import React, { useState, useEffect } from 'react';
import AlbumCard from '../components/AlbumCard';
import TrackCard from '../components/TrackCard';
import ReviewCardSmall from '../components/ReviewCardSmall';
import { mockAlbums, mockTracks, mockReviews } from '../data/mockData';
import './HomePage.css';

// ВРЕМЕННЫЙ КОД ДЛЯ ДЕМОНСТРАЦИИ БЕЗ BACKEND
// ВСЕ API ВЫЗОВЫ УДАЛЕНЫ - ТОЛЬКО МОКОВЫЕ ДАННЫЕ

const HomePage = () => {
  const [popularTracks, setPopularTracks] = useState([]);
  const [latestAlbums, setLatestAlbums] = useState([]);
  const [popularReviews, setPopularReviews] = useState([]);
  const [loading, setLoading] = useState(true);

  // Загрузка моковых данных
  useEffect(() => {
    console.log('Loading mock data...', { 
      mockAlbumsCount: mockAlbums?.length || 0, 
      mockTracksCount: mockTracks?.length || 0, 
      mockReviewsCount: mockReviews?.length || 0 
    });
    
    if (!mockAlbums || mockAlbums.length === 0) {
      console.error('Mock albums are empty or undefined!');
      setLoading(false);
      return;
    }
    
    // Загружаем данные сразу
    const latest = mockAlbums.slice(0, 5);
    const popular = mockTracks.slice(0, 8);
    const reviews = mockReviews.slice(0, 6);
    
    console.log('Setting mock data:', { 
      latestCount: latest.length, 
      popularCount: popular.length, 
      reviewsCount: reviews.length 
    });
    
    setLatestAlbums(latest);
    setPopularTracks(popular);
    setPopularReviews(reviews);
    setLoading(false);
    
    console.log('Mock data loaded successfully');
  }, []);

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="container">
      <h1 className="page-title">Актуальное</h1>
      
      {/* Latest Albums Section */}
      {latestAlbums && latestAlbums.length > 0 ? (
        <section className="home-section">
          <h2 className="section-title">Последние релизы</h2>
          <div className="albums-grid">
            {latestAlbums.map((album) => (
              <AlbumCard key={album.id} album={album} />
            ))}
          </div>
        </section>
      )}

      {/* Popular Tracks Section */}
      {popularTracks && popularTracks.length > 0 && (
        <section className="home-section">
          <h2 className="section-title">Топ треков за сутки</h2>
          <div className="tracks-list">
            {popularTracks.map((track) => (
              <TrackCard key={track.id} track={track} />
            ))}
          </div>
        </section>
      )}

      {/* Popular Reviews Section */}
      {popularReviews && popularReviews.length > 0 && (
        (() => {
          const validReviews = popularReviews
            .filter(review => review && review.album_id && review.album)
            .slice(0, 6);
          
          if (validReviews.length === 0) {
            return null;
          }

          return (
            <section className="home-section">
              <h2 className="section-title">Топ рецензий за сутки</h2>
              <div className="reviews-grid-popular">
                {validReviews.map((review) => (
                  <ReviewCardSmall key={`review-${review.id}`} review={review} />
                ))}
              </div>
            </section>
          );
        })()
      )}
    </div>
  );
};

export default HomePage;

