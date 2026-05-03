# Mustreview

Учебный веб-сервис для музыкальных рецензий: оценки альбомов и треков, лента, лайки, подписки, профили пользователей, уровни, достижения и админская модерация.

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

Сидер также создает демо-авторов, верифицированные аккаунты артистов, одобренные и ожидающие модерации рецензии, лайки альбомов, треков и рецензий.

## Стек

React 18, Go + Gin, GORM, PostgreSQL, Docker Compose.

## DevOps

В репозитории есть GitHub Actions pipeline: Go vet/test/build, frontend build, production compose smoke-test и сборка/push Docker-образов в GHCR.

Подробная техническая документация: [Documentation.md](Documentation.md).
