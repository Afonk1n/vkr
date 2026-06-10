import React from 'react';
import { Link } from 'react-router-dom';

const NotFoundPage = () => (
  <div className="container">
    <div className="empty-state empty-state--soft" style={{ textAlign: 'center', padding: '64px 16px' }}>
      <h1 style={{ fontSize: '3rem', margin: 0 }}>404</h1>
      <p>Такой страницы нет — возможно, ссылка устарела.</p>
      <Link to="/feed" className="btn-edit">На главную</Link>
    </div>
  </div>
);

export default NotFoundPage;
