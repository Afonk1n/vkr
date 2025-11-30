import React, { useState, useEffect } from 'react';
import { genresAPI } from '../services/api';
import './Filters.css';

const Filters = ({ onFilterChange, filters }) => {
  const [genres, setGenres] = useState([]);
  const [loading, setLoading] = useState(true);

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

    fetchGenres();
  }, []);

  const handleGenreChange = (e) => {
    onFilterChange({ ...filters, genre_id: e.target.value || null });
  };

  const handleSearchChange = (e) => {
    onFilterChange({ ...filters, search: e.target.value });
  };

  const handleSortChange = (e) => {
    const [sortBy, sortOrder] = e.target.value.split('_');
    onFilterChange({ ...filters, sort_by: sortBy, sort_order: sortOrder });
  };

  const clearFilters = () => {
    onFilterChange({
      genre_id: null,
      search: '',
      sort_by: 'created_at',
      sort_order: 'desc',
    });
  };

  return (
    <div className="filters">
      <div className="filters-row">
        <div className="filter-group">
          <label htmlFor="search">Поиск</label>
          <input
            type="text"
            id="search"
            placeholder="Название или исполнитель..."
            value={filters.search || ''}
            onChange={handleSearchChange}
          />
        </div>
        <div className="filter-group">
          <label htmlFor="genre">Жанр</label>
          <select
            id="genre"
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
          <label htmlFor="sort">Сортировка</label>
          <select
            id="sort"
            value={`${filters.sort_by || 'created_at'}_${filters.sort_order || 'desc'}`}
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
        <button onClick={clearFilters} className="btn-clear">
          Сбросить
        </button>
      </div>
    </div>
  );
};

export default Filters;

