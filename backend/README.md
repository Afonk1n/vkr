# Backend - Music Review Site

Backend API для сайта рецензирования музыки, написанный на Go с использованием Gin framework и PostgreSQL.

## Требования

- Go 1.21 или выше
- PostgreSQL 12 или выше

## Установка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

3. Настройте переменные окружения в `.env`:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=music_review_db
DB_SSLMODE=disable

PORT=8080
GIN_MODE=debug

JWT_SECRET=your-secret-key-change-this-in-production
```

4. Создайте базу данных PostgreSQL:
```sql
CREATE DATABASE music_review_db;
```

5. Запустите сервер:
```bash
go run main.go
```

Сервер запустится на порту 8080 (или на порту, указанном в переменной окружения PORT).

## Структура проекта

```
backend/
├── main.go                 # Точка входа приложения
├── routes/                 # Маршруты API
│   └── routes.go
├── controllers/            # Контроллеры для обработки запросов
│   ├── auth_controller.go
│   ├── album_controller.go
│   ├── review_controller.go
│   ├── genre_controller.go
│   └── user_controller.go
├── models/                 # Модели данных
│   ├── user.go
│   ├── album.go
│   ├── genre.go
│   ├── review.go
│   ├── track.go
│   └── review_like.go
├── middleware/             # Middleware для аутентификации и авторизации
│   └── auth.go
├── database/               # Настройка базы данных и миграции
│   └── database.go
└── utils/                  # Вспомогательные функции
    ├── validator.go
    ├── password.go
    └── errors.go
```

## API Endpoints

### Аутентификация
- `POST /api/auth/register` - Регистрация пользователя
- `POST /api/auth/login` - Вход в систему
- `GET /api/auth/me` - Получение информации о текущем пользователе

### Жанры
- `GET /api/genres` - Получение списка жанров
- `GET /api/genres/:id` - Получение жанра по ID
- `POST /api/genres` - Создание жанра (только админ)
- `PUT /api/genres/:id` - Обновление жанра (только админ)
- `DELETE /api/genres/:id` - Удаление жанра (только админ)

### Альбомы
- `GET /api/albums` - Получение списка альбомов (с фильтрацией и пагинацией)
- `GET /api/albums/:id` - Получение альбома по ID
- `POST /api/albums` - Создание альбома (авторизованный пользователь)
- `PUT /api/albums/:id` - Обновление альбома (только админ)
- `DELETE /api/albums/:id` - Удаление альбома (только админ)

### Рецензии
- `GET /api/reviews` - Получение списка рецензий (с фильтрацией)
- `GET /api/reviews/:id` - Получение рецензии по ID
- `POST /api/reviews` - Создание рецензии (авторизованный пользователь)
- `PUT /api/reviews/:id` - Обновление рецензии (автор или админ)
- `DELETE /api/reviews/:id` - Удаление рецензии (автор или админ)
- `POST /api/reviews/:id/approve` - Одобрение рецензии (только админ)
- `POST /api/reviews/:id/reject` - Отклонение рецензии (только админ)

### Пользователи
- `GET /api/users/:id` - Получение пользователя по ID
- `GET /api/users/:id/reviews` - Получение рецензий пользователя
- `PUT /api/users/:id` - Обновление профиля пользователя
- `DELETE /api/users/:id` - Удаление пользователя

## Система оценки

Рецензии оцениваются по следующей формуле:
- **Базовые критерии** (1-10 баллов каждый):
  - Рифмы/Образы
  - Структура/Ритмика
  - Реализация стиля
  - Индивидуальность/Харизма
- **Множитель** Атмосфера/Вайб (1.0000-1.6072)
- **Итоговый балл** = (Рифмы + Структура + Реализация + Индивидуальность) × 1.4 × Атмосфера/Вайб

## Тестовые пользователи

После первого запуска создаются тестовые пользователи:
- **Администратор:**
  - Email: `admin@example.com`
  - Пароль: `admin123`
- **Тестовый пользователь:**
  - Email: `test@example.com`
  - Пароль: `test123`

## Авторизация

Для доступа к защищенным эндпоинтам необходимо передавать заголовок `X-User-ID` с ID пользователя. В production рекомендуется использовать JWT токены.

## База данных

База данных автоматически создается при первом запуске приложения с помощью миграций GORM. Также автоматически заполняются начальные данные (жанры, тестовые пользователи, альбомы).

## Разработка

Для запуска в режиме разработки:
```bash
GIN_MODE=debug go run main.go
```

Для сборки:
```bash
go build -o music-review-backend main.go
```

