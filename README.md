# Mustreview

Учебный веб-сервис для музыкальных рецензий: оценки альбомов и треков, лента, лайки, подписки, профили пользователей, уровни и достижения.

## Запуск

```bash
docker compose up --build
```

| Сервис | URL |
| --- | --- |
| Frontend | http://localhost:3000 |
| API | http://localhost:8080/api |

## Тестовые аккаунты

| Роль | Email | Пароль |
| --- | --- | --- |
| Админ | `admin@example.com` | `admin123` |
| Пользователь | `test@example.com` | `test123` |

## Стек

React 18, Go + Gin, GORM, PostgreSQL, Docker Compose.

Подробная техническая документация: [Documentation.md](Documentation.md).
