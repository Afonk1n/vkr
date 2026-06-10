import React, { useEffect, useState } from 'react';
import { albumsAPI, tracksAPI } from '../services/api';
import AlbumCard from './AlbumCard';
import TrackCard from './TrackCard';
import './SimilarReleases.css';

const MAX_ITEMS = 6;

// Блок «похожих» релизов того же жанра. Чисто на существующих эндпоинтах:
// для альбома — альбомы того же жанра, для трека — треки того же жанра,
// исключаем текущий и сортируем по среднему баллу.
const SimilarReleases = ({ type, genreId, genreIds, excludeId, title }) => {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const genre = genreId || (Array.isArray(genreIds) && genreIds[0]);
    if (!genre) {
      setItems([]);
      setLoading(false);
      return undefined;
    }

    let ignore = false;
    setLoading(true);

    const request =
      type === 'track'
        ? tracksAPI.getAll({ genre_ids: [genre], sort_by: 'average_rating', sort_order: 'desc', page_size: 12 })
        : albumsAPI.getAll({ genre_id: genre, sort_by: 'average_rating', sort_order: 'desc', page_size: 12 });

    request
      .then((res) => {
        if (ignore) return;
        const raw = type === 'track'
          ? (res.data?.tracks ?? res.data ?? [])
          : (res.data?.albums ?? res.data ?? []);
        const list = (Array.isArray(raw) ? raw : [])
          .filter((it) => it && it.id !== excludeId)
          .slice(0, MAX_ITEMS);
        setItems(list);
      })
      .catch(() => {
        if (!ignore) setItems([]);
      })
      .finally(() => {
        if (!ignore) setLoading(false);
      });

    return () => { ignore = true; };
  }, [type, genreId, genreIds, excludeId]);

  // Пока грузится или нечего показать — не занимаем место пустотой.
  if (loading || items.length === 0) return null;

  return (
    <section className="similar-releases">
      <h2 className="section-title">{title || (type === 'track' ? 'Похожие треки' : 'Похожие релизы')}</h2>
      <div className={type === 'track' ? 'tracks-list' : 'albums-grid'}>
        {items.map((item) =>
          type === 'track'
            ? <TrackCard key={item.id} track={item} />
            : <AlbumCard key={item.id} album={item} />
        )}
      </div>
    </section>
  );
};

export default SimilarReleases;
