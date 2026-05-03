import React, { createContext, useState, useContext, useEffect } from 'react';
import { authAPI } from '../services/api';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in
    const userId = localStorage.getItem('userId');
    const savedUser = localStorage.getItem('user');

    if (userId && savedUser) {
      try {
        setUser(JSON.parse(savedUser));
      } catch (error) {
        console.error('Error parsing user data:', error);
        localStorage.removeItem('userId');
        localStorage.removeItem('user');
      }
    }
    setLoading(false);
  }, []);

  const login = async (email, password) => {
    try {
      const response = await authAPI.login({ email, password });
      const { user: userData, user_id, session_token } = response.data;

      if (session_token) {
        localStorage.setItem('sessionToken', session_token);
      }
      localStorage.setItem('userId', user_id.toString());
      localStorage.setItem('user', JSON.stringify(userData));
      setUser(userData);
      
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.message || 'Ошибка входа',
      };
    }
  };

  const register = async (username, email, password) => {
    try {
      const response = await authAPI.register({ username, email, password });
      const { user: registeredUser, session_token: registerToken, user_id: registerUserId } = response.data;

      let userData = registeredUser;
      let userId = registerUserId;
      let sessionToken = registerToken;
      if (!sessionToken || !userId) {
        const loginResponse = await authAPI.login({ email, password });
        userData = loginResponse.data.user;
        userId = loginResponse.data.user_id;
        sessionToken = loginResponse.data.session_token;
      }

      if (sessionToken) {
        localStorage.setItem('sessionToken', sessionToken);
      }
      localStorage.setItem('userId', userId.toString());
      localStorage.setItem('user', JSON.stringify(userData));
      setUser(userData);
      
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.message || 'Ошибка регистрации',
      };
    }
  };

  const logout = () => {
    localStorage.removeItem('sessionToken');
    localStorage.removeItem('userId');
    localStorage.removeItem('user');
    setUser(null);
  };

  const updateUser = (userData) => {
    // Update user data in context and localStorage without re-login
    setUser(userData);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const value = {
    user,
    login,
    register,
    logout,
    updateUser,
    loading,
    isAuthenticated: !!user,
    isAdmin: user?.is_admin || false,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

