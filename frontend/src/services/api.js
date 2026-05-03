import axios from 'axios';

const isApiConfigured = Boolean(process.env.REACT_APP_API_URL);
const API_BASE_URL = (process.env.REACT_APP_API_URL || '/api').replace(/\/$/, '');

// Create axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config) => {
    const sessionToken = localStorage.getItem('sessionToken');
    if (sessionToken) {
      config.headers.Authorization = `Bearer ${sessionToken}`;
    }

    const userId = localStorage.getItem('userId');
    if (!sessionToken && userId) {
      config.headers['X-User-ID'] = userId;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // ВРЕМЕННО: отключаем редирект на login для демо без backend
    if (process.env.REACT_APP_USE_MOCK === 'true' || !isApiConfigured) {
      console.warn('API error (ignored in mock mode):', error);
      return Promise.reject(error);
    }
    if (error.response?.status === 401) {
      // Unauthorized - clear auth and redirect to login
      localStorage.removeItem('sessionToken');
      localStorage.removeItem('userId');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  getMe: () => api.get('/auth/me'),
};

// Albums API
export const albumsAPI = {
  getAll: (params) => api.get('/albums', { params }),
  getById: (id) => api.get(`/albums/${id}`),
  getByArtist: (artistName) => api.get(`/albums/artist/${encodeURIComponent(artistName)}`),
  create: (data) => api.post('/albums', data),
  uploadCover: (file) => {
    const formData = new FormData();
    formData.append('cover', file);
    return api.post('/albums/cover', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
  update: (id, data) => api.put(`/albums/${id}`, data),
  delete: (id) => api.delete(`/albums/${id}`),
  like: (id) => api.post(`/albums/${id}/like`),
  unlike: (id) => api.delete(`/albums/${id}/like`),
};

// Reviews API
export const reviewsAPI = {
  getAll: (params) => api.get('/reviews', { params }),
  getById: (id) => api.get(`/reviews/${id}`),
  create: (data) => api.post('/reviews', data),
  update: (id, data) => api.put(`/reviews/${id}`, data),
  delete: (id) => api.delete(`/reviews/${id}`),
  approve: (id) => api.post(`/reviews/${id}/approve`),
  reject: (id) => api.post(`/reviews/${id}/reject`),
};

// Genres API
export const genresAPI = {
  getAll: () => api.get('/genres'),
  getById: (id) => api.get(`/genres/${id}`),
  create: (data) => api.post('/genres', data),
  update: (id, data) => api.put(`/genres/${id}`, data),
  delete: (id) => api.delete(`/genres/${id}`),
};

// Users API
export const usersAPI = {
  getById: (id) => api.get(`/users/${id}`),
  getUserReviews: (id, params) => api.get(`/users/${id}/reviews`, { params }),
  getLikedReviews: (id, params) => api.get(`/users/${id}/liked-reviews`, { params }),
  follow: (id) => api.post(`/users/${id}/follow`),
  unfollow: (id) => api.delete(`/users/${id}/follow`),
  update: (id, data) => api.put(`/users/${id}`, data),
  setFavorites: (id, favorites) => api.put(`/users/${id}/favorites`, Array.isArray(favorites) ? { album_ids: favorites } : favorites),
  uploadAvatar: (id, file) => {
    const formData = new FormData();
    formData.append('avatar', file);
    return api.post(`/users/${id}/avatar`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
  delete: (id) => api.delete(`/users/${id}`),
};

// Search API
export const searchAPI = {
  search: (query) => api.get('/search', { params: { q: query } }),
};

// Tracks API
export const tracksAPI = {
  getAll: (params) => {
    // Handle array params for genre_ids
    const config = { params: {} };
    if (params) {
      Object.keys(params).forEach(key => {
        if (key === 'genre_ids' && Array.isArray(params[key])) {
          // Convert array to query string format: genre_ids[]=1&genre_ids[]=2
          config.params[key + '[]'] = params[key];
        } else {
          config.params[key] = params[key];
        }
      });
    }
    return api.get('/tracks', config);
  },
  getPopular: (params) => api.get('/tracks/popular', { params }),
  getById: (id) => api.get(`/tracks/${id}`),
  getByAlbum: (albumId) => api.get(`/albums/${albumId}/tracks`),
  create: (data) => api.post('/tracks', data),
  update: (id, data) => api.put(`/tracks/${id}`, data),
  delete: (id) => api.delete(`/tracks/${id}`),
  like: (id) => api.post(`/tracks/${id}/like`),
  unlike: (id) => api.delete(`/tracks/${id}/like`),
};

// Reviews API - add like methods
reviewsAPI.like = (id) => api.post(`/reviews/${id}/like`);
reviewsAPI.unlike = (id) => api.delete(`/reviews/${id}/like`);
reviewsAPI.getPopular = (params) => api.get('/reviews/popular', { params });

export default api;
