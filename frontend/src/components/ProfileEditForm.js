import React, { useState, useEffect } from 'react';
import './ProfileEditForm.css';

const ProfileEditForm = ({ user, onSave, onCancel }) => {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    bio: '',
    avatar_path: '',
    social_links: {
      vk: '',
      telegram: '',
      instagram: '',
    },
    password: '',
  });

  useEffect(() => {
    if (user) {
      setFormData({
        username: user.username || '',
        email: user.email || '',
        bio: user.bio || '',
        avatar_path: user.avatar_path || '',
        social_links: user.social_links ? (typeof user.social_links === 'string' ? JSON.parse(user.social_links) : user.social_links) : { vk: '', telegram: '', instagram: '' },
        password: '',
      });
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
        <label>Аватар (URL)</label>
        <input
          type="url"
          name="avatar_path"
          value={formData.avatar_path}
          onChange={handleChange}
          placeholder="https://example.com/avatar.jpg"
        />
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
              type="url"
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
            <label>Instagram</label>
            <input
              type="url"
              value={formData.social_links.instagram || ''}
              onChange={(e) => handleSocialLinkChange('instagram', e.target.value)}
              placeholder="https://instagram.com/username"
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

