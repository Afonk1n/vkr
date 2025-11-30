import React, { useState, useEffect } from 'react';
import { genresAPI } from '../services/api';
import './FilterModal.css';

const FilterModal = ({ isOpen, onClose, onApplyFilters, currentFilters }) => {
  const [genres, setGenres] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState({
    genre_id: currentFilters?.genre_id || null,
    sort_by: currentFilters?.sort_by || 'created_at',
    sort_order: currentFilters?.sort_order || 'desc',
  });

  useEffect(() => {
    if (isOpen) {
      setFilters({
        genre_id: currentFilters?.genre_id || null,
        sort_by: currentFilters?.sort_by || 'created_at',
        sort_order: currentFilters?.sort_order || 'desc',
      });
    }
  }, [isOpen, currentFilters]);

  useEffect(() => {
    const fetchGenres = async () => {
      try {
        const response = await genresAPI.getAll();
        setGenres(response.data);
      } catch (error) {
        console.error('Error fetching genres:', error);
      } finally {
        setLoading(false);
      }
    };

    if (isOpen) {
      fetchGenres();
    }
  }, [isOpen]);

  const handleGenreChange = (e) => {
    setFilters({ ...filters, genre_id: e.target.value || null });
  };

  const handleSortChange = (e) => {
    const [sortBy, sortOrder] = e.target.value.split('_');
    setFilters({ ...filters, sort_by: sortBy, sort_order: sortOrder });
  };

  const handleApply = () => {
    onApplyFilters(filters);
    onClose();
  };

  const handleClear = () => {
    const clearedFilters = {
      genre_id: null,
      sort_by: 'created_at',
      sort_order: 'desc',
    };
    setFilters(clearedFilters);
    onApplyFilters(clearedFilters);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="filter-modal-overlay" onClick={onClose}>
      <div className="filter-modal" onClick={(e) => e.stopPropagation()}>
        <div className="filter-modal-header">
          <h2>Фильтры</h2>
          <button className="filter-modal-close" onClick={onClose}>
            ×
          </button>
        </div>
        <div className="filter-modal-content">
          <div className="filter-group">
            <label htmlFor="modal-genre">Жанр</label>
            <select
              id="modal-genre"
              value={filters.genre_id || ''}
              onChange={handleGenreChange}
              disabled={loading}
            >
              <option value="">Все жанры</option>
              {genres.map((genre) => (
                <option key={genre.id} value={genre.id}>
                  {genre.name}
                </option>
              ))}
            </select>
          </div>
          <div className="filter-group">
            <label htmlFor="modal-sort">Сортировка</label>
            <select
              id="modal-sort"
              value={`${filters.sort_by}_${filters.sort_order}`}
              onChange={handleSortChange}
            >
              <option value="created_at_desc">Новые сначала</option>
              <option value="created_at_asc">Старые сначала</option>
              <option value="average_rating_desc">По рейтингу (высокий)</option>
              <option value="average_rating_asc">По рейтингу (низкий)</option>
              <option value="title_asc">По названию (А-Я)</option>
              <option value="title_desc">По названию (Я-А)</option>
            </select>
          </div>
        </div>
        <div className="filter-modal-footer">
          <button onClick={handleClear} className="btn-clear">
            Сбросить
          </button>
          <button onClick={handleApply} className="btn-apply">
            Применить
          </button>
        </div>
      </div>
    </div>
  );
};

export default FilterModal;

