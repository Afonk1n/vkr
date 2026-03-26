# Roadmap: production-like без оверинжиниринга

Цель: привести учебный проект к виду “реально работающего продукта”, не уходя в микросервисы/Kubernetes раньше времени.

## Принципы

- Конфигурация только через env, **секреты не в git**.
- Минимум инфраструктуры, максимум воспроизводимости.
- “Prod-like” практики включаем поэтапно, чтобы не блокировать разработку функционала.

---

## Этап 1 — продуктовый локальный контур (для спокойной доработки фич)

**Что внедряем**

- Единый конфиг через env: убрать хардкод (например, CORS `localhost:3000`, API URL) → управлять через переменные окружения.
- Dev vs prod-like режимы:
  - `APP_ENV=dev|prod`
  - `SEED_ENABLED=true|false`
  - `DB_CREATE_ENABLED=true|false`
  - (временный вариант) `MIGRATIONS_MODE=auto|manual`
- Минимальная prod-готовность backend:
  - `GET /healthz`
  - graceful shutdown
  - разумные таймауты HTTP

### Решения Этапа 1 (фиксируем до внедрения)

#### Контракт окружения (env)

**Backend**

- **`APP_ENV`**: `dev` (по умолчанию) | `prod`
- **`PORT`**: порт HTTP сервера (дефолт `8080`)
- **`GIN_MODE`**: `debug` | `release`
- **`CORS_ALLOW_ORIGINS`**: список origins через запятую  
  Пример dev: `http://localhost:3000`
- **`DB_HOST`**, **`DB_PORT`**, **`DB_USER`**, **`DB_PASSWORD`**, **`DB_NAME`**, **`DB_SSLMODE`**
- **`SEED_ENABLED`**: `true|false`  
  - dev: `true`
  - prod-like: `false`
- **`DB_CREATE_ENABLED`**: `true|false` (создание БД приложением)  
  - dev: `true` (удобство локальной разработки)
  - prod-like: `false`
- **`MIGRATIONS_MODE`** (временный): `auto|manual`  
  - dev: `auto` (текущий AutoMigrate)
  - prod-like: `manual` (подготовка к Этапу 2 с SQL-миграциями)

**Frontend**

- **`REACT_APP_API_URL`**: base URL для API  
  Пример dev: `http://localhost:8080/api`
- **`REACT_APP_USE_MOCK`**: `true|false` (если используется)

#### Dev vs prod-like (правила)

- **Dev**:
  - разрешены `SEED_ENABLED=true`, `DB_CREATE_ENABLED=true`, `MIGRATIONS_MODE=auto`
- **Prod-like**:
  - запрещены seed/автосоздание БД/автомиграции в рантайме (`SEED_ENABLED=false`, `DB_CREATE_ENABLED=false`, `MIGRATIONS_MODE=manual`)
  - CORS, порты, режим Gin — только через env

#### Definition of Done (Этап 1)

- Нет хардкода `localhost` для CORS/API URL (всё через env).
- Есть `/healthz` (и он не требует auth).
- Backend корректно завершает работу (graceful shutdown).
- Dev/prod-like режимы влияют на seed/создание БД/миграции согласно правилам выше.

**Роли/агенты**

- `software-architect`: целевая картина, что делаем/не делаем, границы.
- `backend-architect` / `senior-developer`: внедрить флаги, healthz, shutdown в Go.
- `frontend-developer`: конфиг фронта (API URL), убрать лишние режимы/хардкоды.
- `security-engineer`: политика секретов + зафиксировать, что `X-User-ID` (если остаётся) — **dev-only**.

### Прогресс (зафиксировано)

**Backend**

- Внедрено чтение CORS origins из env `CORS_ALLOW_ORIGINS` (дефолт `http://localhost:3000`).
- Добавлен `GET /healthz` (без auth); оставлен `GET /health` для совместимости.
- Добавлены graceful shutdown и HTTP таймауты (через `http.Server`).
- В `database.InitDB()` добавлены флаги dev vs prod-like:
  - `APP_ENV` (дефолт `dev`)
  - `SEED_ENABLED` (дефолт `true` в dev / `false` в prod)
  - `DB_CREATE_ENABLED` (дефолт `true` в dev / `false` в prod)
  - `MIGRATIONS_MODE` (дефолт `auto` в dev / `manual` в prod) — управляет `AutoMigrate`
- Добавлены шаблоны env:
  - `backend/.env.example`
  - `frontend/.env.example`

**Проверка**

- `curl -s http://localhost:8080/healthz` → `{"status":"ok"}`

**Известные нюансы (не блокируют Этап 1)**

- В seed-логах возможны сообщения про дубликаты лайков (`duplicated key not allowed`) — часть seed-лайков не создаётся из-за уникальных индексов.
- Gin предупреждает про повторно подключенные middleware Logger/Recovery (можно почистить позже, когда будем делать “release-ready” конфиг).

### Дальше (следующий агент)

- Этапы 1–2 (локальный контур и миграции) реализованы в базовом объёме.
- **Следующий осмысленный шаг** — перейти к Этапу 4 (CI), т.к. Docker/Compose-контур из Этапа 3 уже присутствует.
- Обратиться к роли **`devops-automator`** для настройки минимального CI:
  - GitHub Actions (или аналог) для lint/test/build backend;
  - build frontend;
  - сборка Docker-образов (без обязательного деплоя).

---

## Этап 2 — контроль схемы БД (как в компании)

**Что внедряем**

- Переход с `AutoMigrate` на **SQL-миграции** (goose или golang-migrate).
- Seed оставить только для dev, в prod-like отключить.

**Роли/агенты**

- `software-architect`: ADR “миграции vs automigrate”.
- `backend-architect`: внедрение миграций и корректный lifecycle старта.
- `security-engineer`: политика запуска миграций (отдельной командой/джобой), запрет “самовольных” изменений схемы в prod.

### Прогресс (зафиксировано)

- Принято ADR о переходе с AutoMigrate на SQL-миграции (golang-migrate) и политике их запуска.
- Реализован каталог `backend/migrations/` с начальными миграциями:
  - `0001_init_schema` — базовая схема (users, genres, albums, tracks, track_genres, reviews, review_likes, track_likes, album_likes).
  - `0002_fix_reviews_nullable` — делает `album_id` и `track_id` в `reviews` nullable.
- Обновлены `README.md` и `Documentation.md`, описывающие:
  - использование golang-migrate для применения миграций;
  - dev/prod-like правила (`MIGRATIONS_MODE=auto` только для dev, `manual` для prod-like);
  - что seed включён только в dev (через `SEED_ENABLED=true`).

---

## Этап 3 — контур сборки и запуска (Docker/Compose)

**Что внедряем**

- Dockerfile для backend и frontend (multi-stage, минимальные образы где уместно).
- `docker compose` для dev и отдельный `compose.prod.yml` для демонстрации.
- Healthchecks в Compose.

**Роли/агенты**

- `devops-automator`: Docker/Compose, env contract, run commands.
- `security-engineer`: hardening контейнеров, безопасные дефолты.

### Прогресс (зафиксировано)

- Добавлены Docker-артефакты:
  - `backend/Dockerfile` (prod build + dev target)
  - `frontend/Dockerfile.dev` (CRA dev server)
  - `frontend/Dockerfile` + `frontend/nginx.conf` (prod build + nginx + proxy `/api`)
  - `docker-compose.yml` (dev: db + backend + frontend + healthchecks)
  - `compose.prod.yml` (prod-like demo: nginx + `/api` proxy + prod defaults)
- Примечание для Windows/WSL: для нормальной скорости и file-watcher лучше держать рабочую копию в файловой системе WSL (ext4), а не на `/mnt/d/...`.

---

## Этап 4 — CI (минимум, но реально)

**Что внедряем**

- GitHub Actions: lint/test/build backend + build frontend + сборка Docker images (без деплоя или с опциональным).

**Роли/агенты**

- `devops-automator`: pipeline.
- `git-workflow-master`: стиль веток/PR, если будешь оформлять процесс.

### Прогресс (зафиксировано)

- Добавлен workflow `.github/workflows/ci.yml`:
  - Job `backend`: `go vet`, `go test ./...`, `go build ./...` в каталоге `backend`.
  - Job `frontend`: `npm install` и `npm run build` в каталоге `frontend` (с `REACT_APP_API_URL=http://localhost:8080/api`).
  - Job `docker`: сборка образов из `backend/Dockerfile` и `frontend/Dockerfile`; при **push** в `main`/`master` — **push в GHCR** (`ghcr.io/<owner>/<repo>/backend|frontend`, теги `<sha>` и `latest`). На PR — только сборка без публикации.
- Для деплоя на VPS: `compose.deploy.yml` (образы из реестра) и runbook `docs/DEPLOY-VPS.md`.
- CI запускается на `push` и `pull_request` в ветки `main`/`master` и служит базовой проверкой, что проект собирается до мерджа.

---

## Этап 5 — демо-деплой на VPS (без Kubernetes)

**Что внедряем**

- 1 VPS + Docker Compose “prod-like”.
- Деплой: `pull images` + `restart` (через CI или вручную по runbook).
- Домены/TLS — опционально, если нужно для презентации.

**Роли/агенты**

- `devops-automator`: деплой-процесс + runbook.
- `security-engineer`: SSH keys, firewall, секреты.

---

## Этап 6 — enterprise IaC: Terraform + (опционально) Ansible

Подключать после того, как приложение и Compose стабилизированы.

**Что внедряем**

- Terraform: создать VPS + firewall (+ IP/DNS по необходимости).
- Ansible: bootstrap сервера (docker, users, ufw) + deploy compose.

**Роли/агенты**

- `devops-automator`: Terraform/Ansible скелет и структура `infra/`.
- `security-engineer`: секреты (Ansible Vault/CI secrets), least privilege.

---

## Сознательно не делаем пока

- Kubernetes
- микросервисы/очереди
- тяжёлая observability-платформа

Причина: высокий риск “съесть время”, низкая отдача для текущего масштаба и целей ВКР.

