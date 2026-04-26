import React from 'react';
import Badge from './Badge';
import './Badge.css';

const BadgeList = ({ badges, profileContext = 'self' }) => {
  if (!badges || badges.length === 0) {
    return null;
  }

  return (
    <div className="badge-list">
      {badges.map((badge, index) => (
        <Badge key={index} badge={badge} badgeId={index} profileContext={profileContext} />
      ))}
    </div>
  );
};

export default BadgeList;

