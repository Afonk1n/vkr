/**
 * Converts atmosphere rating (1-10) to multiplier (1.0000-1.6072)
 * Formula: multiplier = 1.0000 + (rating - 1) * 0.0674666...
 * This ensures max score of 90 when all ratings are 10
 * 
 * @param {number} atmosphereRating - Atmosphere rating (1-10)
 * @returns {number} Atmosphere multiplier (1.0000-1.6072)
 */
export const convertAtmosphereToMultiplier = (atmosphereRating) => {
  // Linear conversion: 1 -> 1.0000, 10 -> 1.6072
  // Range: 0.6072, divided by 9 steps
  const step = 0.6072 / 9;
  return 1.0000 + (atmosphereRating - 1) * step;
};

/**
 * Converts atmosphere multiplier (1.0000-1.6072) to rating (1-10)
 * 
 * @param {number} multiplier - Atmosphere multiplier (1.0000-1.6072)
 * @returns {number} Atmosphere rating (1-10)
 */
export const convertMultiplierToAtmosphere = (multiplier) => {
  // Reverse conversion
  const step = 0.6072 / 9;
  return Math.round(1 + (multiplier - 1.0000) / step);
};

/**
 * Calculates the final score based on the rating formula
 * Formula: (Рифмы+Структура+Реализация+Индивидуальность) × 1.4 × Атмосфера/Вайб
 * Result is rounded to the nearest integer
 * 
 * @param {number} ratingRhymes - Rating for rhymes/images (1-10)
 * @param {number} ratingStructure - Rating for structure/rhythm (1-10)
 * @param {number} ratingImplementation - Rating for style implementation (1-10)
 * @param {number} ratingIndividuality - Rating for individuality/charisma (1-10)
 * @param {number} atmosphereRating - Atmosphere rating (1-10)
 * @returns {number} Final score (rounded to integer)
 */
export const calculateFinalScore = (
  ratingRhymes,
  ratingStructure,
  ratingImplementation,
  ratingIndividuality,
  atmosphereRating
) => {
  const baseScore = ratingRhymes + ratingStructure + ratingImplementation + ratingIndividuality;
  const atmosphereMultiplier = convertAtmosphereToMultiplier(atmosphereRating);
  const score = baseScore * 1.4 * atmosphereMultiplier;
  return Math.round(score);
};

/**
 * Validates rating value (1-10)
 */
export const validateRating = (rating) => {
  return rating >= 1 && rating <= 10;
};

/**
 * Validates atmosphere rating (1-10)
 */
export const validateAtmosphereRating = (rating) => {
  return rating >= 1 && rating <= 10;
};

/**
 * Formats the final score as an integer
 */
export const formatScore = (score) => {
  return Math.round(score).toString();
};

