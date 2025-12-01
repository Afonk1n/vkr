import React from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import SearchBar from './SearchBar';
import './Header.css';

const Header = () => {
  const { user, logout, isAuthenticated } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <header className="header">
      <div className="container">
        <div className="header-content">
          <Link to="/" className="logo">
            <h1>🎵 Мьюзик Рейтинг</h1>
          </Link>
          <SearchBar />
          <nav className="nav">
            <button
              className="theme-toggle"
              onClick={toggleTheme}
              title={theme === 'dark' ? 'Переключить на светлую тему' : 'Переключить на тёмную тему'}
            >
              {theme === 'dark' ? '☀️' : '🌙'}
            </button>
            <Link to="/" className="nav-link">
              Актуальное
            </Link>
            <Link to="/search" className="nav-link">
              Каталог
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

