import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import './Header.css';

const Header = () => {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <header className="header">
      <div className="container">
        <div className="header-content">
          <Link to="/" className="logo">
            <h1>🎵 Music Review</h1>
          </Link>
          <nav className="nav">
            <Link to="/" className="nav-link">
              Альбомы
            </Link>
            {isAuthenticated && (
              <>
                <Link to="/profile" className="nav-link">
                  Профиль
                </Link>
                {user?.is_admin && (
                  <Link to="/admin" className="nav-link">
                    Админ-панель
                  </Link>
                )}
                <button onClick={handleLogout} className="btn-logout">
                  Выйти
                </button>
              </>
            )}
            {!isAuthenticated && (
              <>
                <Link to="/login" className="nav-link">
                  Вход
                </Link>
                <Link to="/signup" className="btn-primary">
                  Регистрация
                </Link>
              </>
            )}
          </nav>
        </div>
      </div>
    </header>
  );
};

export default Header;

