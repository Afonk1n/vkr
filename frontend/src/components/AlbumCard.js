import React from 'react';
import { Link } from 'react-router-dom';
import './AlbumCard.css';

const AlbumCard = ({ album }) => {
  return (
    <Link to={`/albums/${album.id}`} className="album-card">
      <div className="album-cover">
        {album.cover_image_path ? (
          <img src={album.cover_image_path} alt={album.title} />
        ) : (
          <div className="album-cover-placeholder">🎵</div>
        )}
      </div>
      <div className="album-info">
        <h3 className="album-title">{album.title}</h3>
        <p className="album-artist">{album.artist}</p>
        {album.genre && (
          <span className="album-genre">{album.genre.name}</span>
        )}
        {album.average_rating > 0 && (
          <div className="album-rating">
            ⭐ {Math.round(album.average_rating)}
          </div>
        )}
      </div>
    </Link>
  );
};

export default AlbumCard;

