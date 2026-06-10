# Mustreview

Учебный веб-сервис для музыкальных рецензий: оценки альбомов и треков по 5 критериям, лента, лайки, подписки, профили пользователей, уровни, достижения и админская модерация. Есть «паспорт релиза» (радар критериев + распределение оценок), блоки похожих релизов и расширенные топы.

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

Подробная техническая документация: [Documentation.md](Documentation.md). Навигация для разработчиков и ИИ-агентов: [AGENTS.md](AGENTS.md). Деплой на VPS: [docs/DEPLOY-VPS.md](docs/DEPLOY-VPS.md).
