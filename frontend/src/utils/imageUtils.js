/**
 * Утилита для работы с изображениями
 * Формирует правильный URL для изображений из public папки
 */

/**
 * Получает URL изображения обложки
 * @param {string} imagePath - Путь к изображению из базы данных (например, "/preview/1.jpg")
 * @returns {string} - Полный URL для использования в img src
 */
export const getImageUrl = (imagePath) => {
  if (!imagePath) return null;
  
  // В Create React App файлы из public доступны по корневому пути
  // Путь должен начинаться с / и указывать на файл в public папке
  // Например: /preview/1.jpg -> http://localhost:3000/preview/1.jpg
  
  // Если путь уже начинается с /, используем его напрямую
  if (imagePath.startsWith('/')) {
    return imagePath;
  }
  
  // Если путь относительный, добавляем /
  return `/${imagePath}`;
};

/**
 * Проверяет, является ли строка валидным путем к изображению
 * @param {string} path - Путь для проверки
 * @returns {boolean}
 */
export const isValidImagePath = (path) => {
  if (!path) return false;
  const imageExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg'];
  return imageExtensions.some(ext => path.toLowerCase().endsWith(ext));
};
