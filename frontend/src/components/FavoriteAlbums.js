import React, { useState, useEffect, useCallback, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from 'react-router-dom';
import { albumsAPI, usersAPI } from '../services/api';
import { getImageUrl } from '../utils/imageUtils';
import './FavoriteAlbums.css';

const SLOTS = 3;
const MODAL_W = 300;
const MODAL_H = 400;

const normalizeAlbums = (albums) => (Array.isArray(albums) ? albums : []);

function clampModalPosition(clientX, clientY) {
  const margin = 12;
  let left = clientX - MODAL_W / 2;
  let top = clientY + 14;
  left = Math.max(margin, Math.min(left, window.innerWidth - MODAL_W - margin));
  top = Math.max(margin, Math.min(top, window.innerHeight - MODAL_H - margin));
  return { left, top };
}

const FavoriteAlbums = ({ albums, isOwner = false, userId, onUpdate }) => {
  const navigate = useNavigate();
  const [editing, setEditing] = useState(false);
  const [modalPos, setModalPos] = useState({ left: 0, top: 0 });
  const [search, setSearch] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [saveError, setSaveError] = useState('');
  const [current, setCurrent] = useState(() => normalizeAlbums(albums));
  const searchInputRef = useRef(null);

  useEffect(() => {
    if (!editing) {
      setCurrent(normalizeAlbums(albums));
    }
  }, [albums, editing]);

  const handleCancel = useCallback(() => {
    setCurrent(normalizeAlbums(albums));
    setSearch('');
    setSearchResults([]);
    setSaveError('');
    setEditing(false);
  }, [albums]);

  useEffect(() => {
    if (!editing) return undefined;
    const onKey = (e) => {
      if (e.key === 'Escape') handleCancel();
    };
    window.addEventListener('keydown', onKey);
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', onKey);
      document.body.style.overflow = '';
    };
  }, [editing, handleCancel]);

  useEffect(() => {
    if (editing && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [editing]);

  const openEditor = (clientX, clientY) => {
    setModalPos(clampModalPosition(clientX, clientY));
    setSaveError('');
    setSearch('');
    setSearchResults([]);
    setEditing(true);
  };

  const handleSearch = async (q) => {
    setSearch(q);
    if (q.trim().length < 2) {
      setSearchResults([]);
      return;
    }
    try {
      const res = await albumsAPI.getAll({ search: q, page_size: 8 });
      setSearchResults(res.data?.albums || res.data || []);
    } catch {
      setSearchResults([]);
    }
  };

  const handleAdd = (album) => {
    if (current.find((a) => a.id === album.id)) return;
    if (current.length >= SLOTS) return;
    setCurrent([...current, album]);
    setSearch('');
    setSearchResults([]);
  };

  const handleRemove = (id) => setCurrent(current.filter((a) => a.id !== id));

  const handleSave = async () => {
    setSaveError('');
    try {
      const { data } = await usersAPI.setFavorites(userId, current.map((a) => a.id));
      const saved = data?.favorite_albums;
      onUpdate && onUpdate(Array.isArray(saved) ? saved : current);
      setEditing(false);
    } catch (e) {
      console.error(e);
      setSaveError('Не удалось сохранить. Проверьте сеть и попробуйте снова.');
    }
  };

  const pointFromElement = (el) => {
    const r = el.getBoundingClientRect();
    return { clientX: r.left + r.width / 2, clientY: r.bottom + 4 };
  };

  const modal =
    editing &&
    createPortal(
      <>
        <button
          type="button"
          className="fav-modal-backdrop"
          aria-label="Закрыть"
          onClick={handleCancel}
        />
        <div
          className="fav-modal"
          style={{ left: modalPos.left, top: modalPos.top, width: MODAL_W }}
          role="dialog"
          aria-modal="true"
          aria-labelledby="fav-modal-title"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="fav-modal-header">
            <h4 id="fav-modal-title">Любимые альбомы</h4>
            <button type="button" className="fav-modal-close" onClick={handleCancel} aria-label="Закрыть">
              ×
            </button>
          </div>
          <p className="fav-modal-hint">До трёх альбомов. Поиск по каталогу.</p>
          <input
            ref={searchInputRef}
            className="fav-search-input fav-search-input--modal"
            type="text"
            placeholder="Начните вводить название…"
            value={search}
            onChange={(e) => handleSearch(e.target.value)}
          />
          {searchResults.length > 0 && (
            <div className="fav-search-results fav-search-results--modal">
              {searchResults.map((a) => (
                <div key={a.id} className="fav-search-item" onClick={() => handleAdd(a)}>
                  <img
                    src={getImageUrl(a.cover_image_path)}
                    alt={a.title}
                    className="fav-search-thumb"
                    onError={(e) => {
                      e.target.style.display = 'none';
                    }}
                  />
                  <div>
                    <div className="fav-search-title">{a.title}</div>
                    <div className="fav-search-artist">{a.artist}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
          <div className="fav-modal-slots-preview">
            {Array.from({ length: SLOTS }).map((_, i) => {
              const album = current[i];
              return album ? (
                <div key={album.id} className="fav-modal-slot">
                  <img src={getImageUrl(album.cover_image_path)} alt="" className="fav-modal-slot-cover" />
                  <button type="button" className="fav-modal-slot-remove" onClick={() => handleRemove(album.id)}>
                    ×
                  </button>
                </div>
              ) : (
                <div key={i} className="fav-modal-slot fav-modal-slot--empty" />
              );
            })}
          </div>
          {saveError && (
            <p className="fav-save-error fav-save-error--modal" role="alert">
              {saveError}
            </p>
          )}
          <div className="fav-actions fav-actions--modal">
            <button type="button" className="btn-save-favorites" onClick={handleSave}>
              Сохранить
            </button>
            <button type="button" className="btn-cancel-favorites" onClick={handleCancel}>
              Отмена
            </button>
          </div>
        </div>
      </>,
      document.body
    );

  return (
    <div className="favorite-albums favorite-albums-card">
      <div className="favorite-albums-header">
        <h3>Любимые альбомы</h3>
        {isOwner && !editing && (
          <button
            type="button"
            className="btn-edit-favorites"
            onClick={(e) => openEditor(e.clientX, e.clientY)}
          >
            {current.length === 0 ? 'Добавить' : 'Изменить'}
          </button>
        )}
      </div>
      <p className="favorite-albums-sub">
        {isOwner
          ? 'До трёх альбомов отображаются у всех, кто открывает ваш профиль.'
          : 'Подборка из трёх любимых релизов.'}
      </p>

      <div className="favorite-albums-slots">
        {Array.from({ length: SLOTS }).map((_, i) => {
          const album = current[i];
          return album ? (
            <div
              key={album.id}
              className="fav-slot filled"
              onClick={() => !editing && navigate(`/albums/${album.id}`)}
            >
              <img
                src={getImageUrl(album.cover_image_path)}
                alt={album.title}
                className="fav-cover"
                onError={(e) => {
                  e.target.style.display = 'none';
                }}
              />
              <div className="fav-info">
                <span className="fav-title">{album.title}</span>
                <span className="fav-artist">{album.artist}</span>
              </div>
            </div>
          ) : (
            <div
              key={i}
              className={`fav-slot empty ${isOwner ? 'fav-slot--clickable' : ''}`}
              role={isOwner ? 'button' : undefined}
              onClick={
                isOwner
                  ? (e) => {
                      const p = pointFromElement(e.currentTarget);
                      openEditor(p.clientX, p.clientY);
                    }
                  : undefined
              }
              tabIndex={isOwner ? 0 : undefined}
              aria-label={isOwner ? 'Добавить любимый альбом' : undefined}
              onKeyDown={
                isOwner
                  ? (e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        const p = pointFromElement(e.currentTarget);
                        openEditor(p.clientX, p.clientY);
                      }
                    }
                  : undefined
              }
            >
              <span className="fav-placeholder">
                <span className="fav-placeholder-hint">{isOwner ? 'добавить' : 'пусто'}</span>
              </span>
            </div>
          );
        })}
      </div>

      {modal}
    </div>
  );
};

export default FavoriteAlbums;
