import React, { useState, useEffect } from 'react';
import { usersAPI } from '../services/api';
import { getImageUrl } from '../utils/imageUtils';
import './ProfileEditForm.css';

const SOCIAL_FIELDS = [
  { key: 'vk', label: 'VK', placeholder: 'https://vk.com/username' },
  { key: 'telegram', label: 'Telegram', placeholder: '@username' },
  { key: 'max', label: 'MAX', placeholder: '@username' },
];

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
  const [formError, setFormError] = useState('');

  useEffect(() => {
    if (!user) return;

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
  }, [user]);

  const handleChange = (event) => {
    const { name, value } = event.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSocialLinkChange = (platform, value) => {
    setFormData((prev) => ({
      ...prev,
      social_links: { ...prev.social_links, [platform]: value },
    }));
  };

  const handleAvatarChange = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      setFormError('Размер файла не должен превышать 5MB');
      return;
    }

    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      setFormError('Разрешенные форматы: JPG, PNG, WEBP');
      return;
    }
    setFormError('');

    const reader = new FileReader();
    reader.onloadend = () => setAvatarPreview(reader.result);
    reader.readAsDataURL(file);

    setUploadingAvatar(true);
    try {
      const response = await usersAPI.uploadAvatar(user.id, file);
      const updatedUserData = response.data;
      setFormData((prev) => ({ ...prev, avatar_path: updatedUserData.avatar_path }));
      setAvatarPreview(getImageUrl(updatedUserData.avatar_path));

      if (updateUser) {
        const fullUserResponse = await usersAPI.getById(user.id);
        updateUser(fullUserResponse.data);
      }
    } catch (err) {
      setFormError(`Ошибка при загрузке аватара: ${err.response?.data?.message || err.message}`);
      setAvatarPreview(user.avatar_path ? getImageUrl(user.avatar_path) : null);
    } finally {
      setUploadingAvatar(false);
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setSaving(true);
    setFormError('');
    try {
      const dataToSend = { ...formData };
      if (!dataToSend.password) delete dataToSend.password;
      await onSave(dataToSend);
    } catch (err) {
      setFormError(err.message || 'Не удалось сохранить профиль');
    } finally {
      setSaving(false);
    }
  };

  const initials = user?.username?.charAt(0).toUpperCase() || 'U';

  return (
    <div className="pef-wrapper">
      <div className="pef-header">
        <div>
          <span className="pef-kicker">Профиль</span>
          <h2 className="pef-title">Редактирование</h2>
        </div>
        <button type="button" className="pef-close" onClick={onCancel} aria-label="Закрыть">
          ×
        </button>
      </div>

      <form className="pef-form" onSubmit={handleSubmit}>
        <section className="pef-hero-card">
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

          <div className="pef-hero-copy">
            <strong>{formData.username || user?.username || 'Пользователь'}</strong>
            <span>Фото загружается сразу после выбора, остальные изменения сохраняются кнопкой ниже.</span>
            <input
              type="file"
              id="pef-avatar-input"
              accept="image/jpeg,image/jpg,image/png,image/webp"
              onChange={handleAvatarChange}
              disabled={uploadingAvatar}
              style={{ display: 'none' }}
            />
            <label
              htmlFor="pef-avatar-input"
              className={`pef-btn pef-btn-outline${uploadingAvatar ? ' pef-btn--disabled' : ''}`}
            >
              {uploadingAvatar ? 'Загрузка...' : 'Обновить фото'}
            </label>
          </div>
        </section>

        <section className="pef-card">
          <div className="pef-section-head">
            <strong>Основное</strong>
            <span>Имя, почта и короткое описание профиля</span>
          </div>
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
              rows={4}
              maxLength={220}
              placeholder="Пара слов о вкусе, любимых жанрах или подходе к рецензиям..."
            />
            <span className="pef-counter">{formData.bio.length}/220</span>
          </div>
        </section>

        <section className="pef-card">
          <div className="pef-section-head">
            <strong>Социальные сети</strong>
            <span>Покажутся в профиле компактными иконками под ником</span>
          </div>
          <div className="pef-grid-3">
            {SOCIAL_FIELDS.map(({ key, label, placeholder }) => (
              <div className="pef-field" key={key}>
                <label className="pef-label" htmlFor={`pef-${key}`}>{label}</label>
                <input
                  id={`pef-${key}`}
                  className="pef-input"
                  type="text"
                  value={formData.social_links[key] || ''}
                  onChange={(event) => handleSocialLinkChange(key, event.target.value)}
                  placeholder={placeholder}
                />
              </div>
            ))}
          </div>
        </section>

        <section className="pef-card pef-card--quiet">
          <div className="pef-section-head">
            <strong>Пароль</strong>
            <span>Заполняйте только если хотите сменить пароль</span>
          </div>
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
        </section>

        {formError && <div className="pef-message pef-message--error">{formError}</div>}

        <div className="pef-actions">
          <button type="button" className="pef-btn pef-btn-outline" onClick={onCancel}>
            Отмена
          </button>
          <button type="submit" className="pef-btn pef-btn-primary" disabled={saving || uploadingAvatar}>
            {saving ? 'Сохранение...' : 'Сохранить изменения'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default ProfileEditForm;
