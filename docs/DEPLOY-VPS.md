# Runbook: деплой на VPS (Docker Compose + образы из GHCR)

Краткая инструкция для развёртывания приложения на сервере с Docker, когда backend и frontend собираются в CI и публикуются в **GitHub Container Registry** (`ghcr.io`).

## Что публикует CI

При **push** в ветки `main` или `master` workflow `.github/workflows/ci.yml` собирает и отправляет образы:

| Образ | Теги |
|-------|------|
| `ghcr.io/<owner>/<repo>/backend` | `<git-sha>`, `latest` |
| `ghcr.io/<owner>/<repo>/frontend` | `<git-sha>`, `latest` |

`<owner>` и `<repo>` — в нижнем регистре (как в URL репозитория на GitHub).

Пакеты по умолчанию наследуют видимость репозитория. Для приватного репозитория образы тоже приватные: на сервере нужна аутентификация в GHCR (см. ниже).

## Требования на сервере

- Docker Engine и плагин **Compose V2** (`docker compose`).
- Открытый порт для HTTP (по умолчанию `80`, см. переменную `FRONTEND_PUBLISH` в `compose.deploy.yml`).
- Файл `.env` с секретами и URL образов (не коммитить в git).

## 1. Клонирование и конфигурация

```bash
git clone <url-репозитория> && cd <каталог-проекта>
```

Создайте `.env` в корне проекта (пример):

```env
# Образы из GHCR (подставьте свой owner/repo в нижнем регистре)
BACKEND_IMAGE=ghcr.io/your-org/your-repo/backend:latest
FRONTEND_IMAGE=ghcr.io/your-org/your-repo/frontend:latest

DB_USER=postgres
DB_PASSWORD=<сильный-пароль>
DB_NAME=music_review_db

APP_ENV=prod
GIN_MODE=release
# Публичный URL фронта для CORS, например https://music.example.com
CORS_ALLOW_ORIGINS=https://music.example.com
MIGRATIONS_MODE=manual
SEED_ENABLED=false
DB_CREATE_ENABLED=false

# Опционально: порт хоста для nginx фронта (по умолчанию 80)
# FRONTEND_PUBLISH=8080
```

Для воспроизводимого деплоя вместо `latest` укажите конкретный тег — SHA коммита из вкладки **Packages** на GitHub или из лога успешного CI.

## 2. Вход в GHCR (если образы приватные)

Создайте **Personal Access Token (classic)** с правом `read:packages` (и при необходимости `repo`, если пакет привязан к приватному репо).

```bash
echo "<TOKEN>" | docker login ghcr.io -u <GITHUB_USERNAME> --password-stdin
```

Проверка: `docker pull` одного из образов без ошибки `denied`.

## 3. Запуск стека

```bash
docker compose -f compose.deploy.yml --env-file .env pull
docker compose -f compose.deploy.yml --env-file .env up -d
```

Статус: `docker compose -f compose.deploy.yml ps` и логи `docker compose -f compose.deploy.yml logs -f --tail=100`.

Проверки:

- Backend: `GET http://<хост>:<порт-backend-если-проброшен>/healthz` (в типичной схеме только frontend снаружи; healthcheck внутри сети compose уже есть).
- Frontend: открыть в браузере корень сайта.

## 4. Миграции БД

В `MIGRATIONS_MODE=manual` приложение **не** применяет SQL-миграции само. Нужно один раз (и после каждого обновления схемы) выполнить `golang-migrate` с тем же DSN, что и у backend.

Пример DSN:

```text
postgres://USER:PASSWORD@127.0.0.1:5432/DBNAME?sslmode=disable
```

Если PostgreSQL только внутри compose, с хоста порт можно пробросить временно или запускать migrate из одноразового контейнера в той же Docker-сети, что и `db`.

Команды (как в корневом `README.md`):

```bash
export DB_DSN="postgres://..."
migrate -path ./backend/migrations -database "$DB_DSN" up
```

Установите [golang-migrate/migrate](https://github.com/golang-migrate/migrate) локально или используйте официальный Docker-образ `migrate/migrate`.

## 5. Обновление версии

1. Задеплойте код в `main` и дождитесь зелёного CI.
2. На сервере обновите `.env`: `BACKEND_IMAGE` / `FRONTEND_IMAGE` на новый SHA или `latest`.
3. Выполните:

```bash
docker compose -f compose.deploy.yml --env-file .env pull
docker compose -f compose.deploy.yml --env-file .env up -d
```

4. При необходимости примените миграции (`migrate ... up`).

## 6. Откат (rollback)

Укажите в `.env` предыдущие теги образов (старый SHA из истории Actions или **Packages**), затем:

```bash
docker compose -f compose.deploy.yml --env-file .env pull
docker compose -f compose.deploy.yml --env-file .env up -d
```

Если откат затрагивает схему БД, согласуйте с планом миграций (`migrate down` — только осознанно, с бэкапом).

## 7. Бэкап данных

Том `pgdata` хранит данные PostgreSQL. Регулярно делайте дампы:

```bash
docker compose -f compose.deploy.yml exec db pg_dump -U "$DB_USER" "$DB_NAME" > backup.sql
```

---

При локальной сборке образов на сервере используйте `compose.prod.yml` (секция `build:`). Файл `compose.deploy.yml` предназначен для сценария «только pull готовых образов».
