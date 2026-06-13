import React, { useEffect, useRef, useState } from 'react';
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
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const navTrackRef = useRef(null);
  const themeTrackRef = useRef(null);
  const authTrackRef = useRef(null);

  const { dims: navThumbDims } = useSlidingThumb(navTrackRef, [
    location.pathname,
    isAuthenticated,
    user?.is_admin,
  ]);
  const { dims: themeThumbDims } = useSlidingThumb(themeTrackRef, [theme]);
  const { dims: authThumbDims } = useSlidingThumb(authTrackRef, [location.pathname, isAuthenticated]);

  const handleLogout = () => {
    logout();
    setMobileMenuOpen(false);
    navigate('/feed');
  };

  useEffect(() => {
    setMobileMenuOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!mobileMenuOpen) return undefined;
    const handleKeyDown = (event) => {
      if (event.key === 'Escape') setMobileMenuOpen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [mobileMenuOpen]);

  return (
    <header className="header">
      <div className="container">
        <div className={`header-content ${mobileMenuOpen ? 'header-content--menu-open' : ''}`.trim()}>
          <Link to="/feed" className="logo">
            <img src="/logo.png" alt="Мьюзик Рейтинг" className="logo-image" />
            <h1>Мьюзик Рейтинг</h1>
          </Link>
          <button
            type="button"
            className="mobile-menu-toggle"
            onClick={() => setMobileMenuOpen((open) => !open)}
            aria-expanded={mobileMenuOpen}
            aria-controls="header-mobile-panel"
            aria-label={mobileMenuOpen ? 'Закрыть меню' : 'Открыть меню'}
          >
            <span />
            <span />
            <span />
          </button>
          <SearchBar />
          <div className="header-actions" id="header-mobile-panel">
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
            <nav
              ref={!isAuthenticated ? authTrackRef : undefined}
              className={`nav-auth ${
                !isAuthenticated ? 'header-segmented nav-auth--segmented seg-sliding-track' : ''
              }`.trim()}
            >
              {isAuthenticated && (
                <>
                  <button type="button" onClick={handleLogout} className="btn-logout">
                    Выйти
                  </button>
                </>
              )}
              {!isAuthenticated && (
                <>
                  <SegmentedSlidingThumb dims={authThumbDims} />
                  <Link
                    to="/login"
                    className={`header-segment nav-auth-segment ${
                      location.pathname === '/login' ? 'header-segment--active segment-thumb-source' : ''
                    }`.trim()}
                  >
                    Вход
                  </Link>
                  <Link
                    to="/signup"
                    className={`header-segment nav-auth-segment ${
                      location.pathname === '/signup' ? 'header-segment--active segment-thumb-source' : ''
                    }`.trim()}
                  >
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
