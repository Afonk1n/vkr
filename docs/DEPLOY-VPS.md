# Деплой Mustreview на VPS

Универсальный runbook: подходит для Timeweb Cloud, Selectel, Yandex Cloud (Compute Instance), VK Cloud, Hetzner и любого другого VPS с Ubuntu 22.04/24.04. Реальный деплой не выполнялся — это пошаговая инструкция, готовая к копипасте.

Архитектура одна и та же:
- Один VPS (минимум 2 vCPU, 2 ГБ RAM, 20 ГБ SSD).
- Docker + docker compose.
- Стек поднимается из готовых образов `ghcr.io/<owner>/<repo>/{backend,frontend}` через [`compose.deploy.yml`](../compose.deploy.yml).
- Nginx уже встроен в фронтовый образ и проксирует `/api` на backend.
- TLS — отдельным caddy/nginx-reverse-proxy или `traefik`, см. §7.

## 1. Что должно быть до старта

- Аккаунт у облачного провайдера и созданная VM (Ubuntu 22.04 LTS, public IPv4).
- SSH-доступ по ключу.
- Доменное имя (можно бесплатное `*.duckdns.org`/`*.nip.io`, для защиты ВКР этого хватит).
- Репозиторий `Afonk1n/vkr` на GitHub собирается в GHCR (см. CI: `.github/workflows/ci.yml`, job `docker`).
- Личный access token GitHub (`read:packages`) если репозиторий приватный.

## 2. Подготовка сервера

```bash
ssh ubuntu@<server-ip>

# обновление
sudo apt update && sudo apt upgrade -y

# docker по официальной инструкции
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# проверка
docker version
docker compose version
```

Открыть фаервол:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```

## 3. Логин в GHCR (только если образы приватные)

```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u <github-username> --password-stdin
```

`GHCR_TOKEN` = personal access token c правом `read:packages`. Для публичного репозитория шаг можно пропустить.

## 4. Скачать compose-файл

На сервере достаточно одного файла — собственно образа уже опубликованы.

```bash
mkdir -p ~/mustreview && cd ~/mustreview
curl -fsSL https://raw.githubusercontent.com/Afonk1n/vkr/main/compose.deploy.yml -o compose.deploy.yml
```

(Можно и просто `git clone`, но тогда на сервер уезжает весь репозиторий, что не обязательно для деплоя из реестра.)

## 5. Настроить .env

Создать `~/mustreview/.env`:

```env
# Образы (latest = свежий коммит main, конкретный SHA — для пин-релиза)
BACKEND_IMAGE=ghcr.io/afonk1n/vkr/backend:latest
FRONTEND_IMAGE=ghcr.io/afonk1n/vkr/frontend:latest

# Внешний порт публикации фронта (80 = http, можно 8080 если нужен reverse-proxy)
FRONTEND_PUBLISH=80

# Postgres
DB_USER=postgres
DB_PASSWORD=<сгенерируй: openssl rand -hex 16>
DB_NAME=music_review_db
DB_SSLMODE=disable

# Backend
APP_ENV=prod
GIN_MODE=release
CORS_ALLOW_ORIGINS=https://<твой-домен>
SEED_ENABLED=true            # первый запуск — да, потом перевести в false
DB_CREATE_ENABLED=true       # первый запуск — да
MIGRATIONS_MODE=auto
SESSION_SECRET=<сгенерируй: openssl rand -hex 32>
SESSION_TTL_HOURS=168
AUTH_ALLOW_USER_ID_HEADER=false
```

> **Важно**: `SESSION_SECRET` нельзя оставлять дефолтным `change-me-in-prod` — токены подделают.
> После первого старта `SEED_ENABLED=false` и `DB_CREATE_ENABLED=false`, чтобы не пересоздавать БД при перезапуске.

```bash
chmod 600 .env
```

## 6. Первый старт

```bash
cd ~/mustreview
docker compose -f compose.deploy.yml --env-file .env pull
docker compose -f compose.deploy.yml --env-file .env up -d
docker compose -f compose.deploy.yml ps
```

Дождаться, пока все три сервиса станут `healthy`:

```bash
watch -n 2 'docker compose -f compose.deploy.yml ps'
```

Проверки:

```bash
curl -fsSL http://127.0.0.1/                          # фронт отдаётся nginx
curl -fsSL http://127.0.0.1/api/albums | head -c 300  # API проксируется через nginx
docker compose -f compose.deploy.yml exec backend wget -qO- http://localhost:8080/healthz
```

Тестовые логины из сидера (см. README):
- `admin@example.com` / `admin123`
- `test@example.com` / `test123`

## 7. TLS и домен

Простой способ — Caddy перед nginx-фронтом.

`~/mustreview/Caddyfile`:

```caddy
mustreview.example.com {
  reverse_proxy localhost:8080
}
```

> при этом в `.env` ставь `FRONTEND_PUBLISH=8080`, чтобы 80/443 остались Caddy.

Поднимаешь Caddy:

```bash
docker run -d --name caddy --restart=always \
  --network host \
  -v ~/mustreview/Caddyfile:/etc/caddy/Caddyfile \
  -v caddy_data:/data -v caddy_config:/config \
  caddy:2
```

DNS-запись `A` → IP сервера, Caddy сам выпустит Let's Encrypt сертификат.

Альтернатива — поднять nginx-reverse-proxy + certbot, или встроенные средства облака (например, балансировщик Yandex Cloud / Selectel с TLS).

После того как HTTPS заработает, **обнови `CORS_ALLOW_ORIGINS` на `https://...`** и сделай:

```bash
docker compose -f compose.deploy.yml --env-file .env up -d
```

## 8. Обновление

CI после каждого пуша в `main` обновляет `ghcr.io/.../{backend,frontend}:latest`. Чтобы подтянуть свежее на сервере:

```bash
cd ~/mustreview
docker compose -f compose.deploy.yml --env-file .env pull
docker compose -f compose.deploy.yml --env-file .env up -d
docker image prune -f
```

Для воспроизводимости лучше пинить по SHA: в `.env` поставить
`BACKEND_IMAGE=ghcr.io/afonk1n/vkr/backend:<sha>` и тогда `up -d` гарантированно поднимет ровно тот образ.

## 9. Бэкап БД

Простейший вариант — pg_dump раз в сутки:

```bash
docker compose -f compose.deploy.yml exec -T db \
  pg_dump -U "$DB_USER" "$DB_NAME" | gzip > ~/backups/mustreview-$(date +%F).sql.gz
```

Положить в crontab + сложить в S3-совместимое хранилище провайдера.

Восстановление:

```bash
gunzip -c ~/backups/mustreview-YYYY-MM-DD.sql.gz | \
  docker compose -f compose.deploy.yml exec -T db psql -U "$DB_USER" -d "$DB_NAME"
```

## 10. Логи и траблшутинг

```bash
docker compose -f compose.deploy.yml logs -f backend
docker compose -f compose.deploy.yml logs -f frontend
docker compose -f compose.deploy.yml logs -f db
```

Типичные проблемы:

| Симптом | Причина | Решение |
| --- | --- | --- |
| backend не healthy, в логах `failed to connect to database` | `db` ещё стартует или пароль не совпал | подождать; проверить `.env`, `DB_PASSWORD` совпадает с `POSTGRES_PASSWORD` |
| фронт открывается, но `/api/...` 404 | nginx во фронт-образе ходит на `http://backend:8080` по docker-сети — backend упал | посмотреть `logs backend` |
| 401 на `/api/auth/me` после смены домена | разные `CORS_ALLOW_ORIGINS` / куки | обновить `CORS_ALLOW_ORIGINS` и перезапустить backend |
| дубли в БД после рестарта | `SEED_ENABLED` остался `true` | сидер идемпотентен, но всё равно поставить `false` для прода |
| диск растёт | старые образы | `docker image prune -f`, `docker system df` для контроля |

## 11. Чек-лист перед демо

- [ ] `curl https://<домен>/healthz` → 200, ну или `/api/albums` → JSON.
- [ ] Залогинились admin'ом и обычным юзером, лента грузится.
- [ ] Создание рецензии → видна в `/admin` как pending → approve → появилась в `/feed`.
- [ ] Загрузка аватара работает (volume `cover_uploads` смонтирован).
- [ ] `SEED_ENABLED=false`, `DB_CREATE_ENABLED=false` после первого старта.
- [ ] `SESSION_SECRET` не дефолтный.
- [ ] Бэкап БД хотя бы один сделан и проверен.

## 12. Стоимость (ориентир, июнь 2026)

| Провайдер | Конфигурация | ~цена/мес |
| --- | --- | --- |
| Timeweb Cloud | 2 vCPU / 2 ГБ / 40 ГБ NVMe | ~400 ₽ |
| Selectel | Cloud Server S 2 vCPU / 2 ГБ | ~500 ₽ |
| Yandex Cloud | Compute Instance s2.micro (2 vCPU / 2 ГБ) | ~700 ₽ + трафик |
| Hetzner | CPX11 (2 vCPU / 2 ГБ / 40 ГБ) | ~5 € |

Для демо-проекта подходит любая мелкая VPS — узким местом будет сборка образов в CI, а не runtime.
