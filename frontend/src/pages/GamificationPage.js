import React, { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { usersAPI } from '../services/api';
import './GamificationPage.css';

const LEVELS = [
  { name: 'Бронзовый', points: 0, tone: 'bronze', text: 'стартовый уровень для новых авторов' },
  { name: 'Серебряный', points: 2500, tone: 'silver', text: 'профиль уже заметен по активности' },
  { name: 'Золотой', points: 8000, tone: 'gold', text: 'стабильный автор с реакцией сообщества' },
  { name: 'Изумрудный', points: 18000, tone: 'emerald', text: 'сильный участник с большим вкладом' },
  { name: 'Платиновый', points: 36000, tone: 'platinum', text: 'верхняя планка текущей системы' },
];

const XP_RULES = [
  ['Рецензия', '+320', 'Учитываются опубликованные одобренные рецензии.'],
  ['Лайк к вашей рецензии', '+55', 'Показывает, что текст оказался полезен другим.'],
  ['Поставленный лайк', '+12', 'Небольшой бонус за участие в жизни ленты.'],
  ['Авторский лайк', '+240', 'Редкая сильная реакция от автора или отмеченного аккаунта.'],
];

const ACHIEVEMENTS = [
  { name: 'Начинающий критик', icon: 'I', criteria: 'Опубликовать первую одобренную рецензию.' },
  { name: 'Опытный критик', icon: 'II', criteria: 'Опубликовать от 6 до 20 одобренных рецензий.' },
  { name: 'Мастер рецензий', icon: 'III', criteria: 'Опубликовать от 21 до 50 одобренных рецензий.' },
  { name: 'Легенда критики', icon: 'IV', criteria: 'Опубликовать 51 и более одобренных рецензий.' },
  { name: 'Универсал', icon: 'U', criteria: 'Писать о музыке так, чтобы в рецензиях встретилось 5 разных жанров.' },
  { name: 'Хип-хоп критик', icon: 'HH', criteria: 'Собрать не менее 5 одобренных рецензий в жанре хип-хоп.' },
  { name: 'Рок-ценитель', icon: 'R', criteria: 'Собрать не менее 5 одобренных рецензий в жанре рок.' },
  { name: 'Поп-эксперт', icon: 'P', criteria: 'Собрать не менее 5 одобренных рецензий в жанре поп.' },
  { name: 'Электронный знаток', icon: 'E', criteria: 'Собрать не менее 5 одобренных рецензий в электронной музыке.' },
];

const GamificationPage = () => {
  const { user } = useAuth();
  const [profileUser, setProfileUser] = useState(user);

  useEffect(() => {
    if (!user?.id) return;
    usersAPI.getById(user.id)
      .then((response) => setProfileUser(response.data))
      .catch(() => setProfileUser(user));
  }, [user]);

  const earned = useMemo(() => new Set((profileUser?.badges || []).map((badge) => badge.name)), [profileUser]);

  return (
    <div className="container">
      <div className="gamification-page">
        <header className="gamification-hero">
          <div>
            <span className="gamification-kicker">Глоссарий</span>
            <h1>Уровни и достижения</h1>
            <p>Здесь собраны правила геймификации: как растет профиль, какие уровни есть и какие достижения можно открыть.</p>
          </div>
          <Link to="/profile" className="gamification-back-link">Вернуться в профиль</Link>
        </header>

        <section className="gamification-section">
          <div className="gamification-section-head">
            <h2>Как начисляется опыт</h2>
            <p>Средняя оценка не влияет на уровень: важны рецензии, реакция сообщества и активность пользователя.</p>
          </div>
          <div className="xp-grid">
            {XP_RULES.map(([title, points, text]) => (
              <article className="xp-card" key={title}>
                <strong>{points}</strong>
                <span>{title}</span>
                <p>{text}</p>
              </article>
            ))}
          </div>
        </section>

        <section className="gamification-section">
          <div className="gamification-section-head">
            <h2>Линейка уровней</h2>
            <p>Уровень показывает общий вклад пользователя за все время.</p>
          </div>
          <div className="level-book">
            {LEVELS.map((level) => (
              <article className="level-book-row" key={level.name}>
                <div className={`level-book-gem level-book-gem--${level.tone}`}><span>{level.name.charAt(0)}</span></div>
                <div>
                  <strong>{level.name} уровень</strong>
                  <span>{level.points.toLocaleString('ru-RU')} баллов</span>
                  <p>{level.text}</p>
                </div>
              </article>
            ))}
          </div>
        </section>

        <section className="gamification-section">
          <div className="gamification-section-head">
            <h2>Книга достижений</h2>
            <p>Полученные достижения подсвечены. Остальные остаются закрытыми, но по ним видно, что нужно сделать.</p>
          </div>
          <div className="achievement-book">
            {ACHIEVEMENTS.map((achievement) => {
              const unlocked = earned.has(achievement.name);
              return (
                <article className={`achievement-page ${unlocked ? 'achievement-page--unlocked' : ''}`} key={achievement.name}>
                  <div className="achievement-seal">{achievement.icon}</div>
                  <strong>{achievement.name}</strong>
                  <span>{unlocked ? 'Получено' : 'Не открыто'}</span>
                  <p>{achievement.criteria}</p>
                </article>
              );
            })}
          </div>
        </section>
      </div>
    </div>
  );
};

export default GamificationPage;
