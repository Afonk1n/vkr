# Инструкция по настройке Vercel для проекта

## Проблема
Vercel не определяет фреймворк автоматически ("No framework detected"), потому что проект имеет структуру монорепо (frontend и backend в одной папке).

## Решение

### Вариант 1: Настройка через интерфейс Vercel (РЕКОМЕНДУЕТСЯ)

1. Откройте проект в Vercel Dashboard
2. Перейдите в **Settings** → **General**
3. Найдите раздел **Root Directory**
4. Установите **Root Directory**: `frontend`
5. Сохраните изменения

6. Перейдите в **Settings** → **Build & Development Settings**
7. Убедитесь, что настройки следующие:
   - **Framework Preset**: Create React App
   - **Build Command**: `npm run build` (будет выполняться в папке frontend)
   - **Output Directory**: `build`
   - **Install Command**: `npm install`
   - **Development Command**: `npm start`

8. Сохраните изменения

9. Сделайте новый deployment (пересоберите проект)

### Вариант 2: Использование текущей конфигурации

Если вы не хотите менять Root Directory, текущий `vercel.json` должен работать, но нужно убедиться что:

1. В настройках проекта **Build Command** установлен: `cd frontend && npm install && npm run build`
2. **Output Directory** установлен: `frontend/build`

### Проверка

После настройки и пересборки:
- Главная страница должна открываться без ошибки 404
- Все маршруты React Router должны работать
- Статические файлы (изображения, CSS, JS) должны загружаться

### Текущие файлы конфигурации

- `vercel.json` - конфигурация для SPA routing
- `package.json` (в корне) - прокси для сборки frontend

### Если всё ещё не работает

1. Проверьте логи сборки в Vercel Dashboard
2. Убедитесь, что папка `frontend/build` создаётся после сборки
3. Проверьте, что файл `frontend/build/index.html` существует
4. Убедитесь, что в `vercel.json` правильно указан `outputDirectory`

