import React, { useState, useEffect } from 'react';
import { usersAPI } from '../services/api';
import { getImageUrl } from '../utils/imageUtils';
import './ProfileEditForm.css';

const ProfileEditForm = ({ user, onSave, onCancel, updateUser }) => {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    bio: '',
    avatar_path: '',
    social_links: {
      vk: '',
      telegram: '',
      max: '',
    },
    password: '',
  });
  const [avatarFile, setAvatarFile] = useState(null);
  const [avatarPreview, setAvatarPreview] = useState(null);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);

  useEffect(() => {
    if (user) {
      setFormData({
        username: user.username || '',
        email: user.email || '',
        bio: user.bio || '',
        avatar_path: user.avatar_path || '',
        social_links: user.social_links ? (typeof user.social_links === 'string' ? JSON.parse(user.social_links) : user.social_links) : { vk: '', telegram: '', max: '' },
        password: '',
      });
      setAvatarPreview(user.avatar_path ? getImageUrl(user.avatar_path) : null);
    }
  }, [user]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSocialLinkChange = (platform, value) => {
    setFormData(prev => ({
      ...prev,
      social_links: {
        ...prev.social_links,
        [platform]: value,
      },
    }));
  };

  const handleAvatarChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      // Validate file size (max 5MB)
      if (file.size > 5 * 1024 * 1024) {
        alert('Размер файла не должен превышать 5MB');
        return;
      }

      // Validate file type
      const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];
      if (!allowedTypes.includes(file.type)) {
        alert('Разрешенные форматы: JPG, PNG, WEBP');
        return;
      }

      setAvatarFile(file);
      
      // Create preview
      const reader = new FileReader();
      reader.onloadend = () => {
        setAvatarPreview(reader.result);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleAvatarUpload = async () => {
    if (!avatarFile || !user) return;

    setUploadingAvatar(true);
    try {
      const response = await usersAPI.uploadAvatar(user.id, avatarFile);
      const updatedUserData = response.data;
      
      // Update form data
      setFormData(prev => ({
        ...prev,
        avatar_path: updatedUserData.avatar_path,
      }));
      setAvatarPreview(getImageUrl(updatedUserData.avatar_path));
      setAvatarFile(null);
      
      // Update user in context to prevent logout
      if (updateUser) {
        // Fetch full user data to get badges and other fields
        const fullUserResponse = await usersAPI.getById(user.id);
        updateUser(fullUserResponse.data);
      }
      
      alert('Аватар успешно загружен');
    } catch (err) {
      alert('Ошибка при загрузке аватара: ' + (err.response?.data?.message || err.message));
      console.error('Error uploading avatar:', err);
    } finally {
      setUploadingAvatar(false);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const dataToSend = {
      ...formData,
      password: formData.password || undefined, // Only send if provided
    };
    if (!dataToSend.password) {
      delete dataToSend.password;
    }
    onSave(dataToSend);
  };

  return (
    <form className="profile-edit-form" onSubmit={handleSubmit}>
      <div className="form-group">
        <label>Имя пользователя</label>
        <input
          type="text"
          name="username"
          value={formData.username}
          onChange={handleChange}
          required
        />
      </div>

      <div className="form-group">
        <label>Email</label>
        <input
          type="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          required
        />
      </div>

      <div className="form-group">
        <label>Аватар</label>
        <div className="avatar-upload-section">
          <div className="avatar-preview">
            {avatarPreview ? (
              <img src={avatarPreview} alt="Avatar preview" className="avatar-preview-image" />
            ) : (
              <div className="avatar-preview-placeholder">
                {user?.username?.charAt(0).toUpperCase() || 'U'}
              </div>
            )}
          </div>
          <div className="avatar-upload-controls">
            <input
              type="file"
              id="avatar-upload"
              accept="image/jpeg,image/jpg,image/png,image/webp"
              onChange={handleAvatarChange}
              style={{ display: 'none' }}
            />
            <label htmlFor="avatar-upload" className="btn-upload-avatar">
              Выбрать файл
            </label>
            {avatarFile && (
              <button
                type="button"
                onClick={handleAvatarUpload}
                disabled={uploadingAvatar}
                className="btn-upload-submit"
              >
                {uploadingAvatar ? 'Загрузка...' : 'Загрузить'}
              </button>
            )}
            <div className="avatar-upload-hint">
              Максимум 5MB. Форматы: JPG, PNG, WEBP
            </div>
            <div className="avatar-url-fallback">
              <label>Или укажите путь/URL:</label>
              <input
                type="text"
                name="avatar_path"
                value={formData.avatar_path}
                onChange={handleChange}
                placeholder="/avatars/user.jpg или https://example.com/avatar.jpg"
              />
            </div>
          </div>
        </div>
      </div>

      <div className="form-group">
        <label>Биография</label>
        <textarea
          name="bio"
          value={formData.bio}
          onChange={handleChange}
          rows="4"
          placeholder="Расскажите о себе..."
        />
      </div>

      <div className="form-group">
        <label>Социальные сети</label>
        <div className="social-links">
          <div className="social-link-item">
            <label>VK</label>
            <input
              type="text"
              value={formData.social_links.vk || ''}
              onChange={(e) => handleSocialLinkChange('vk', e.target.value)}
              placeholder="https://vk.com/username"
            />
          </div>
          <div className="social-link-item">
            <label>Telegram</label>
            <input
              type="text"
              value={formData.social_links.telegram || ''}
              onChange={(e) => handleSocialLinkChange('telegram', e.target.value)}
              placeholder="@username"
            />
          </div>
          <div className="social-link-item">
            <label>MAX</label>
            <input
              type="text"
              value={formData.social_links.max || ''}
              onChange={(e) => handleSocialLinkChange('max', e.target.value)}
              placeholder="https://max.ru/username или @username"
            />
          </div>
        </div>
      </div>

      <div className="form-group">
        <label>Новый пароль (оставьте пустым, чтобы не менять)</label>
        <input
          type="password"
          name="password"
          value={formData.password}
          onChange={handleChange}
          placeholder="Минимум 6 символов"
        />
      </div>

      <div className="form-actions">
        <button type="submit" className="btn-primary">Сохранить</button>
        <button type="button" onClick={onCancel} className="btn-secondary">Отмена</button>
      </div>
    </form>
  );
};

export default ProfileEditForm;

