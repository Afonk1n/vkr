import React from 'react';
import './Badge.css';

const Badge = ({ badge }) => {
  if (!badge) return null;

  return (
    <div className="badge" title={badge.description} data-priority={badge.priority}>
      <span className="badge-icon">{badge.icon}</span>
      <span className="badge-name">{badge.name}</span>
    </div>
  );
};

export default Badge;

