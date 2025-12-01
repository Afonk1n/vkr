import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { FilterProvider } from './context/FilterContext';
import { ThemeProvider } from './context/ThemeContext';
import Header from './components/Header';
import HomePage from './pages/HomePage';
import AlbumDetailPage from './pages/AlbumDetailPage';
import TrackDetailPage from './pages/TrackDetailPage';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import ProfilePage from './pages/ProfilePage';
import UserProfilePage from './pages/UserProfilePage';
import ArtistPage from './pages/ArtistPage';
import SearchPage from './pages/SearchPage';
import AdminPanel from './pages/AdminPanel';
import './App.css';

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <FilterProvider>
          <Router>
            <div className="App">
              <Header />
              <main className="main-content">
                <Routes>
                  <Route path="/" element={<HomePage />} />
                  <Route path="/albums/:id" element={<AlbumDetailPage />} />
                  <Route path="/tracks/:id" element={<TrackDetailPage />} />
                  <Route path="/login" element={<LoginPage />} />
                  <Route path="/signup" element={<SignupPage />} />
                  <Route path="/profile" element={<ProfilePage />} />
                  <Route path="/users/:id" element={<UserProfilePage />} />
                  <Route path="/artists/:name" element={<ArtistPage />} />
                  <Route path="/search" element={<SearchPage />} />
                  <Route path="/admin" element={<AdminPanel />} />
                </Routes>
              </main>
            </div>
          </Router>
        </FilterProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;

