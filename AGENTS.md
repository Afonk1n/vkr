# AGENTS.md

Этот файл — короткая навигация для ИИ-агентов и новых разработчиков. Подробности по продукту лежат в [Documentation.md](Documentation.md). Здесь только то, что нужно, чтобы быстро сориентироваться, безопасно править код и не сломать привычные правила проекта.

## Краткое описание

**Mustreview** — учебный веб-сервис музыкальных рецензий. React 18 + Go (Gin) + PostgreSQL, всё запускается через Docker Compose. Сделан как ВКР, защита в июне 2026.

## Карта репозитория

```text
vkr/
  backend/                  Go API, Gin + GORM
    controllers/            HTTP-обработчики (по сущностям)
    database/               InitDB, AutoMigrate, сидер
    middleware/             AuthMiddleware, OptionalAuthMiddleware, AdminMiddleware
    migrations/             ручные SQL-миграции 0001..0004
    models/                 GORM-модели
    routes/routes.go        вся таблица маршрутов в одном файле
    utils/                  токены сессий, хеш паролей
    main.go                 точка входа, CORS, graceful shutdown
    Dockerfile              multi-stage (dev/prod)
  frontend/                 React 18 (CRA, react-scripts 5)
    src/pages/              страницы (FeedPage, AlbumDetailPage, ProfilePage, AdminPanel, ...)
    src/components/         UI-блоки (Header, ReviewCard, ProfileDashboard, ...)
    src/context/AuthContext подписанный сессионный токен в localStorage
    src/services/           axios-клиент
    public/preview/         обложки альбомов и треков
    public/avatars/         аватары пользователей
    Dockerfile              prod (nginx раздаёт build)
    Dockerfile.dev          dev (react-scripts start с polling)
    nginx.conf              прокси /api -> backend:8080 в prod
  docker-compose.yml        локальный dev: db + backend(dev) + frontend(dev), hot reload
  compose.prod.yml          prod-like локально: build backend + nginx-frontend
  compose.deploy.yml        VPS-вариант: образы из GHCR вместо build
  .github/workflows/ci.yml  go vet/test/build, frontend build, compose smoke, push в GHCR
  README.md                 короткая инструкция запуска
  Documentation.md          техническая документация (модели, API, UX-решения)
  Пояснительная записка/    материалы ВКР: .docx, главы, скрипты сборки
```

## Главные команды

| Что | Команда |
| --- | --- |
| Поднять dev | `docker compose up --build` |
| Поднять prod-like локально | `docker compose -f compose.prod.yml up --build` |
| Сборка/тесты backend | `cd backend; go vet ./...; go test ./...; go build ./...` |
| Сборка frontend | `cd frontend; npm install; npm run build` |
| Проверить compose-файлы | `docker compose -f <файл> config` (нужен `BACKEND_IMAGE`/`FRONTEND_IMAGE` для `compose.deploy.yml`) |
| Health | `GET http://localhost:8080/healthz`, `GET http://localhost/` |

В `Documentation.md` есть пример с `GOCACHE` под Windows — это нужно, потому что у Go по умолчанию кэш в `%LOCALAPPDATA%`, и на ограниченных проектных дисках это иногда падает.

## Архитектура коротко

- **Авторизация**: подписанный bearer-токен (`utils/session.go`), TTL берётся из `SESSION_TTL_HOURS`. Для dev оставлен fallback `X-User-ID`, в prod отключён через `AUTH_ALLOW_USER_ID_HEADER=false`.
- **Роли**: `is_admin` на пользователе. Админка модерации — `/api/reviews/:id/approve|reject`, `AdminMiddleware`.
- **БД**: PostgreSQL, GORM + ручные миграции в `backend/migrations`. `MIGRATIONS_MODE=auto|manual`, `DB_CREATE_ENABLED` создаёт БД, `SEED_ENABLED` запускает идемпотентный сидер.
- **Сидер**: в [`backend/database/database.go`](backend/database/database.go), создаёт `admin@example.com`/`admin123` и `test@example.com`/`test123`, демо-альбомы, треки, рецензии (approved и pending), лайки. Не дублирует уже существующие сущности.
- **Маршруты**: единая регистрация в [`backend/routes/routes.go`](backend/routes/routes.go) — туда же добавлять новые. Конкретные маршруты (`/:id/tracks`, `/popular`) объявлены ДО `/:id`, чтобы Gin не съел их как параметр.
- **Frontend**: страницы в `src/pages`, API-клиент один (`src/services/api.js`). Темизация через CSS-переменные, светлая + тёмная.

## Конвенции и фишки

- Системные `alert`/`confirm` НЕ используются — все подтверждения внутри UI.
- Оценка: 4 параметра (1–10) + атмосфера (множитель), итог приводится примерно к 1–90. Формула спрятана от пользователя; см. `Documentation.md` §8.
- Уровень профиля считается по активности и реакции сообщества, НЕ по среднему баллу. Это сознательно: чтобы поощрять активность, а не «накрутку оценок».
- Любимое (артисты/альбомы/треки) — максимум 3 в каждой категории; если у юзера ничего не выбрано, профиль автоматически собирает блоки из его рецензий.
- Лайки — оптимистичные на фронте, без ожидания ответа API.
- Демо-обложки лежат в `frontend/public/preview`, загружаемые — в `frontend/public/preview/uploads` (volume `cover_uploads`).
- Аватары — `frontend/public/avatars`, в dev смонтированы как bind-mount, чтобы не терялись при пересборке.

## Переменные окружения

| Переменная | Где | Дефолт | Комментарий |
| --- | --- | --- | --- |
| `APP_ENV` | backend | `prod` | используется только как тэг |
| `PORT` | backend | `8080` | порт Gin |
| `GIN_MODE` | backend | `release` | в dev = `debug` |
| `CORS_ALLOW_ORIGINS` | backend | `http://localhost:3000` | запятая-список доменов |
| `DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE` | backend | `db/5432/postgres/postgres/music_review_db/disable` | подключение к PG |
| `DB_CREATE_ENABLED` | backend | `false` | создать БД, если её нет |
| `MIGRATIONS_MODE` | backend | `manual` | `auto` запускает AutoMigrate |
| `SEED_ENABLED` | backend | `false` | накатить демо-данные |
| `SESSION_SECRET` | backend | `change-me-in-prod` | **обязательно поменять в prod** |
| `SESSION_TTL_HOURS` | backend | `168` | срок жизни токена (часы) |
| `AUTH_ALLOW_USER_ID_HEADER` | backend | `false` | dev-fallback `X-User-ID` |
| `REACT_APP_API_URL` | frontend (build-time) | `http://localhost:8080/api` | используется в проде |
| `BACKEND_IMAGE` / `FRONTEND_IMAGE` | compose.deploy | — | образы из GHCR |
| `FRONTEND_PUBLISH` | compose.deploy | `80` | внешний порт nginx |

## Правила для ИИ-агентов

1. **Рабочий язык — русский.** Коммит-сообщения короткие, без многострочного описания и без `Co-Authored-By`.
2. **Перед `git push`** — обязательно спросить, в какую ветку. В `main` без явного разрешения не пушить.
3. **Деструктивные git-операции** (`reset --hard`, `push --force`, удаление веток) — только с явного разрешения.
4. Если меняешь маршрут или модель — обнови соответствующий раздел в [Documentation.md](Documentation.md). README/AGENTS правь только когда правда изменилась структура.
5. Новые миграции — формат `NNNN_name.up.sql` / `NNNN_name.down.sql`, инкремент номера от последнего. AutoMigrate в дев-режиме их не подхватит, нужны изменения в `database/database.go`.
6. Не плодить отдельных файлов под мелкие правки — `routes.go` намеренно один.
7. Не использовать `alert`/`confirm` во фронте.
8. После любых правок API: проверь `go vet ./... && go build ./...` и `npm run build`. Если меняешь compose — `docker compose -f <файл> config`.
9. Sense check для модерации: рецензия может быть `pending|approved|rejected`, в ленту попадают только `approved`. Лайки на `pending` тоже возможны, но не светятся публично.

## Где что искать

| Хочу… | Файл/папка |
| --- | --- |
| добавить эндпоинт | [`backend/routes/routes.go`](backend/routes/routes.go) + новый метод контроллера в `backend/controllers/` |
| поменять модель данных | `backend/models/` + новая миграция в `backend/migrations/` |
| поменять сид-данные | [`backend/database/database.go`](backend/database/database.go) |
| добавить страницу | `frontend/src/pages/` + регистрация роута в `frontend/src/App.js` |
| общий axios-клиент | `frontend/src/services/api.js` |
| глобальные стили / темы | `frontend/src/index.css`, переменные в каждом *.css |
| поменять CI | [`.github/workflows/ci.yml`](.github/workflows/ci.yml) |
| гайд по деплою | [`docs/DEPLOY-VPS.md`](docs/DEPLOY-VPS.md) |
| ВКР-документация | `Пояснительная записка/` |

## Известные ограничения

- Юнит-тестов по факту нет (`go test ./...` проходит, но `*_test.go` отсутствуют) — учебный проект.
- Frontend без TypeScript, без linter-pipeline в CI (только `npm run build`).
- Авторские реакции на рецензии — задел в данных есть, в UI пока заглушка.
- Мобильная вёрстка — приоритет ниже десктопной (см. `Documentation.md` §10).
