import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { FilterProvider } from './context/FilterContext';
import { ThemeProvider } from './context/ThemeContext';
import Header from './components/Header';
import FeedPage from './pages/FeedPage';
import TopsPage from './pages/TopsPage';
import AlbumDetailPage from './pages/AlbumDetailPage';
import TrackDetailPage from './pages/TrackDetailPage';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import ProfilePage from './pages/ProfilePage';
import UserProfilePage from './pages/UserProfilePage';
import ArtistPage from './pages/ArtistPage';
import SearchPage from './pages/SearchPage';
import AdminPanel from './pages/AdminPanel';
import GamificationPage from './pages/GamificationPage';
import NotFoundPage from './pages/NotFoundPage';
import ProtectedRoute from './components/ProtectedRoute';
import './components/PageTransition.css';
import './App.css';

// Контент внутри роутера: ключ по pathname перезапускает мягкую fade-анимацию
// при каждом переходе между страницами.
function AppShell() {
  const location = useLocation();
  return (
    <div className="App">
      <Header />
      <main className="main-content">
        <div className="page-fade" key={location.pathname}>
          <Routes location={location}>
            <Route path="/" element={<Navigate to="/feed" replace />} />
            <Route path="/feed" element={<FeedPage />} />
            <Route path="/tops" element={<TopsPage />} />
            <Route path="/albums/:id" element={<AlbumDetailPage />} />
            <Route path="/tracks/:id" element={<TrackDetailPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/signup" element={<SignupPage />} />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <ProfilePage />
                </ProtectedRoute>
              }
            />
            <Route path="/users/:id" element={<UserProfilePage />} />
            <Route path="/artists/:name" element={<ArtistPage />} />
            <Route path="/search" element={<SearchPage />} />
            <Route
              path="/admin"
              element={
                <ProtectedRoute adminOnly>
                  <AdminPanel />
                </ProtectedRoute>
              }
            />
            <Route path="/gamification" element={<GamificationPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </div>
      </main>
    </div>
  );
}

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <FilterProvider>
          <Router>
            <AppShell />
          </Router>
        </FilterProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
