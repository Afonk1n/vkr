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
      'Текст с учётом жанра: для лёгкой поп/электроники допускаются простые конструкции; для текстоцентричных жанров — сложная рифмовка и смысл.',
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
      'Узнаваемость тембра и манеры, умение передать эмоции и «увести» слушателя за собой.',
  },
  {
    key: 'atmosphere',
    short: 'АТ',
    title: 'Атмосфера / вайб (1–10 → множитель)',
    hint:
      'Субъективно: насколько удачно передана атмосфера и палитра эмоций. Из шкалы 1–10 считается множитель к формуле.',
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

export function formatMultiplierShort(m) {
  if (m == null || Number.isNaN(Number(m))) return '1';
  let s = (Math.round(Number(m) * 10000) / 10000).toFixed(4);
  if (s.includes('.')) {
    s = s.replace(/0+$/, '').replace(/\.$/, '');
  }
  return s || '1';
}
