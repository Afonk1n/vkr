import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

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
    const userId = localStorage.getItem('userId');
    if (userId) {
      config.headers['X-User-ID'] = userId;
      console.log('API request with X-User-ID:', userId, 'to', config.url);
    } else {
      console.warn('API request without X-User-ID header to', config.url);
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
    if (error.response?.status === 401) {
      // Unauthorized - clear auth and redirect to login
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
  create: (data) => api.post('/albums', data),
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
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
};

// Search API
export const searchAPI = {
  search: (query) => api.get('/search', { params: { q: query } }),
};

// Tracks API
export const tracksAPI = {
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

