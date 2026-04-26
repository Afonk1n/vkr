/**
 * Единое описание критериев рецензии (подсказки + формула).
 * Числа по полям review: rating_* и atmosphere_multiplier.
 */

export const REVIEW_CRITERIA = [
  {
    key: 'rhymes',
    short: 'РО',
    title: 'Рифмы / образы',
    hint:
      'Текст с учетом жанра: для легкой поп/электроники допускаются простые конструкции; для текстоцентричных жанров - сложная рифмовка и смысл.',
  },
  {
    key: 'structure',
    short: 'СТ',
    title: 'Структура / ритмика',
    hint:
      'Ритм и драматургия частей, контрасты; целостность трека/альбома, концепция и порядок композиций.',
  },
  {
    key: 'implementation',
    short: 'РС',
    title: 'Реализация стиля',
    hint:
      'Исполнение (вокал, речитатив, мелодия) и продакшн: сведение, звук инструментала.',
  },
  {
    key: 'individuality',
    short: 'ИХ',
    title: 'Индивидуальность / харизма',
    hint:
      'Узнаваемость тембра и манеры, умение передать эмоции и увлечь слушателя за собой.',
  },
  {
    key: 'atmosphere',
    short: 'АТ',
    title: 'Атмосфера / вайб',
    hint:
      'Субъективно: насколько удачно передана атмосфера и палитра эмоций.',
  },
];

export function getReviewScoreValues(review) {
  const rh = Number(review?.rating_rhymes) || 0;
  const st = Number(review?.rating_structure) || 0;
  const impl = Number(review?.rating_implementation) || 0;
  const ind = Number(review?.rating_individuality) || 0;
  const multRaw = review?.atmosphere_multiplier;
  const mult =
    multRaw != null && multRaw !== '' && !Number.isNaN(Number(multRaw))
      ? Number(multRaw)
      : 1;
  const base = rh + st + impl + ind;
  return { rh, st, impl, ind, mult, base };
}
