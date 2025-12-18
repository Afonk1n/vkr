# Временные изменения для демонстрации без backend

## ⚠️ ВНИМАНИЕ: Эти изменения временные для отчёта!

После завершения отчёта нужно **откатить все изменения**, связанные с моковыми данными.

## Что было изменено:

1. **`frontend/src/data/mockData.js`** - НОВЫЙ ФАЙЛ с моковыми данными
   - Удалить этот файл после отчёта

2. **`frontend/src/pages/HomePage.js`** - добавлен флаг `USE_MOCK_DATA = true`
   - Вернуть оригинальную логику загрузки данных

3. **`frontend/src/components/AlbumCard.js`** - добавлена проверка для мокового режима
   - Убрать проверку `process.env.REACT_APP_USE_MOCK`

4. **`frontend/src/components/TrackCard.js`** - добавлена проверка для мокового режима
   - Убрать проверку `process.env.REACT_APP_USE_MOCK`

5. **`frontend/src/components/ReviewCardSmall.js`** - добавлена проверка для мокового режима
   - Убрать проверку `process.env.REACT_APP_USE_MOCK`

6. **`frontend/src/services/api.js`** - отключен редирект на login при ошибках
   - Вернуть оригинальную обработку ошибок

## Как откатить изменения:

### Вариант 1: Через Git (рекомендуется)

```bash
# Посмотреть изменённые файлы
git status

# Откатить конкретные файлы
git checkout HEAD -- frontend/src/pages/HomePage.js
git checkout HEAD -- frontend/src/components/AlbumCard.js
git checkout HEAD -- frontend/src/components/TrackCard.js
git checkout HEAD -- frontend/src/components/ReviewCardSmall.js
git checkout HEAD -- frontend/src/services/api.js

# Удалить новый файл с моками
rm frontend/src/data/mockData.js
```

### Вариант 2: Вручную

1. Удалить файл `frontend/src/data/mockData.js`
2. В `HomePage.js` убрать:
   - `import { mockAlbums, mockTracks, mockReviews } from '../data/mockData';`
   - `const USE_MOCK_DATA = true;`
   - Функцию `loadMockData()`
   - Все проверки `if (USE_MOCK_DATA)`
3. В компонентах (AlbumCard, TrackCard, ReviewCardSmall) убрать проверки мокового режима
4. В `api.js` вернуть оригинальную обработку ошибок

## Что работает в демо-режиме:

✅ Главная страница отображает:
- 5 последних альбомов
- 8 популярных треков
- 6 популярных рецензий

✅ Лайки работают локально (без сохранения на сервере)

❌ Не работает:
- Переход на страницы альбомов/треков (будут ошибки 404)
- Поиск (показывает пустые результаты)
- Авторизация
- Все остальные страницы

## Для отчёта:

Главная страница должна отображаться корректно на Vercel с моковыми данными.

