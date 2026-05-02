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
    social_links: { vk: '', telegram: '', max: '' },
    password: '',
  });
  const [avatarPreview, setAvatarPreview] = useState(null);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (user) {
      setFormData({
        username: user.username || '',
        email: user.email || '',
        bio: user.bio || '',
        avatar_path: user.avatar_path || '',
        social_links: user.social_links
          ? (typeof user.social_links === 'string' ? JSON.parse(user.social_links) : user.social_links)
          : { vk: '', telegram: '', max: '' },
        password: '',
      });
      setAvatarPreview(user.avatar_path ? getImageUrl(user.avatar_path) : null);
    }
  }, [user]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSocialLinkChange = (platform, value) => {
    setFormData(prev => ({
      ...prev,
      social_links: { ...prev.social_links, [platform]: value },
    }));
  };

  const handleAvatarChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      alert('Размер файла не должен превышать 5MB');
      return;
    }
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      alert('Разрешенные форматы: JPG, PNG, WEBP');
      return;
    }

    // Show preview immediately
    const reader = new FileReader();
    reader.onloadend = () => setAvatarPreview(reader.result);
    reader.readAsDataURL(file);

    // Auto-upload right away
    setUploadingAvatar(true);
    try {
      const response = await usersAPI.uploadAvatar(user.id, file);
      const updatedUserData = response.data;
      setFormData(prev => ({ ...prev, avatar_path: updatedUserData.avatar_path }));
      setAvatarPreview(getImageUrl(updatedUserData.avatar_path));
      if (updateUser) {
        const fullUserResponse = await usersAPI.getById(user.id);
        updateUser(fullUserResponse.data);
      }
    } catch (err) {
      alert('Ошибка при загрузке аватара: ' + (err.response?.data?.message || err.message));
      setAvatarPreview(user.avatar_path ? getImageUrl(user.avatar_path) : null);
    } finally {
      setUploadingAvatar(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const dataToSend = { ...formData };
      if (!dataToSend.password) delete dataToSend.password;
      await onSave(dataToSend);
    } finally {
      setSaving(false);
    }
  };

  const initials = user?.username?.charAt(0).toUpperCase() || 'U';

  return (
    <div className="pef-wrapper">
      <div className="pef-header">
        <h2 className="pef-title">Редактирование профиля</h2>
      </div>

      <form className="pef-form" onSubmit={handleSubmit}>
        {/* Avatar */}
        <div className="pef-section">
          <span className="pef-section-label">Аватар</span>
          <div className="pef-avatar-row">
            <div className="pef-avatar-wrap">
              {uploadingAvatar && (
                <div className="pef-avatar-overlay">
                  <span>...</span>
                </div>
              )}
              {avatarPreview ? (
                <img src={avatarPreview} alt="avatar" className="pef-avatar-img" />
              ) : (
                <div className="pef-avatar-placeholder">{initials}</div>
              )}
            </div>
            <div className="pef-avatar-info">
              <input
                type="file"
                id="pef-avatar-input"
                accept="image/jpeg,image/jpg,image/png,image/webp"
                onChange={handleAvatarChange}
                style={{ display: 'none' }}
              />
              <label
                htmlFor="pef-avatar-input"
                className={`pef-btn pef-btn-outline${uploadingAvatar ? ' pef-btn--disabled' : ''}`}
              >
                {uploadingAvatar ? 'Загрузка...' : 'Выбрать фото'}
              </label>
              <span className="pef-hint">JPG, PNG, WEBP · до 5 МБ · загружается автоматически</span>
            </div>
          </div>
        </div>

        {/* Basic info */}
        <div className="pef-grid-2">
          <div className="pef-field">
            <label className="pef-label" htmlFor="pef-username">Имя пользователя</label>
            <input
              id="pef-username"
              className="pef-input"
              type="text"
              name="username"
              value={formData.username}
              onChange={handleChange}
              required
            />
          </div>
          <div className="pef-field">
            <label className="pef-label" htmlFor="pef-email">Email</label>
            <input
              id="pef-email"
              className="pef-input"
              type="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              required
            />
          </div>
        </div>

        <div className="pef-field">
          <label className="pef-label" htmlFor="pef-bio">Биография</label>
          <textarea
            id="pef-bio"
            className="pef-input pef-textarea"
            name="bio"
            value={formData.bio}
            onChange={handleChange}
            rows={3}
            placeholder="Расскажите о себе..."
          />
        </div>

        {/* Social links */}
        <div className="pef-section">
          <span className="pef-section-label">Социальные сети</span>
          <div className="pef-grid-3">
            {[
              { key: 'vk', label: 'VK', placeholder: 'https://vk.com/username' },
              { key: 'telegram', label: 'Telegram', placeholder: '@username' },
              { key: 'max', label: 'MAX', placeholder: '@username' },
            ].map(({ key, label, placeholder }) => (
              <div className="pef-field" key={key}>
                <label className="pef-label">{label}</label>
                <input
                  className="pef-input"
                  type="text"
                  value={formData.social_links[key] || ''}
                  onChange={(e) => handleSocialLinkChange(key, e.target.value)}
                  placeholder={placeholder}
                />
              </div>
            ))}
          </div>
        </div>

        {/* Password */}
        <div className="pef-field">
          <label className="pef-label" htmlFor="pef-password">Новый пароль</label>
          <input
            id="pef-password"
            className="pef-input"
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            placeholder="Оставьте пустым, чтобы не менять"
          />
        </div>

        <div className="pef-actions">
          <button type="submit" className="pef-btn pef-btn-primary" disabled={saving || uploadingAvatar}>
            {saving ? 'Сохранение...' : 'Сохранить изменения'}
          </button>
          <button type="button" className="pef-btn pef-btn-outline" onClick={onCancel}>
            Отмена
          </button>
        </div>
      </form>
    </div>
  );
};

export default ProfileEditForm;
