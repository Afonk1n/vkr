import React, { useRef } from 'react';
import { Link, NavLink, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import SearchBar from './SearchBar';
import { useSlidingThumb } from '../hooks/useSlidingThumb';
import { SegmentedSlidingThumb } from './SegmentedSlidingThumb';
import './Header.css';

const navSegmentClass = ({ isActive }) =>
  `header-segment${isActive ? ' header-segment--active segment-thumb-source' : ''}`;

const Header = () => {
  const { user, logout, isAuthenticated } = useAuth();
  const { theme, setTheme } = useTheme();
  const navigate = useNavigate();
  const location = useLocation();
  const navTrackRef = useRef(null);
  const themeTrackRef = useRef(null);

  const { dims: navThumbDims } = useSlidingThumb(navTrackRef, [
    location.pathname,
    isAuthenticated,
    user?.is_admin,
  ]);
  const { dims: themeThumbDims } = useSlidingThumb(themeTrackRef, [theme]);

  const handleLogout = () => {
    logout();
    navigate('/feed');
  };

  return (
    <header className="header">
      <div className="container">
        <div className="header-content">
          <Link to="/feed" className="logo">
            <img src="/logo.png" alt="Мьюзик Рейтинг" className="logo-image" />
            <h1>Мьюзик Рейтинг</h1>
          </Link>
          <SearchBar />
          <div className="header-actions">
            <div className="header-sliders">
              <div
                ref={navTrackRef}
                className="header-segmented seg-sliding-track"
                role="tablist"
                aria-label="Разделы сайта"
              >
                <SegmentedSlidingThumb dims={navThumbDims} />
                <NavLink to="/feed" className={navSegmentClass} end>
                  Лента
                </NavLink>
                <NavLink to="/tops" className={navSegmentClass}>
                  Топы
                </NavLink>
                {isAuthenticated ? (
                  <NavLink to="/profile" className={navSegmentClass}>
                    Профиль
                  </NavLink>
                ) : (
                  <Link to="/login" className="header-segment">
                    Профиль
                  </Link>
                )}
                {user?.is_admin && (
                  <NavLink to="/admin" className={navSegmentClass}>
                    Админка
                  </NavLink>
                )}
              </div>
              <div
                ref={themeTrackRef}
                className="header-segmented header-segmented--theme seg-sliding-track"
                role="group"
                aria-label="Тема оформления"
              >
                <SegmentedSlidingThumb dims={themeThumbDims} />
                <button
                  type="button"
                  className={`header-segment ${theme === 'dark' ? 'header-segment--active segment-thumb-source' : ''}`}
                  onClick={() => setTheme('dark')}
                  aria-pressed={theme === 'dark'}
                >
                  Тёмная
                </button>
                <button
                  type="button"
                  className={`header-segment ${theme === 'light' ? 'header-segment--active segment-thumb-source' : ''}`}
                  onClick={() => setTheme('light')}
                  aria-pressed={theme === 'light'}
                >
                  Светлая
                </button>
              </div>
            </div>
            <nav className="nav-auth">
              {isAuthenticated && (
                <>
                  <button type="button" onClick={handleLogout} className="btn-logout">
                    Выйти
                  </button>
                </>
              )}
              {!isAuthenticated && (
                <>
                  <Link to="/login" className="nav-text-link">
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
      </div>
    </header>
  );
};

export default Header;
