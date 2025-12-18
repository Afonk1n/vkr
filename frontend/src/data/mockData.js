// Моковые данные для демонстрации без backend
// ВРЕМЕННЫЙ ФАЙЛ - для отчёта, потом откатить

export const mockAlbums = [
  {
    id: 1,
    title: "Баста 1",
    artist: "Баста",
    cover_image_path: "/preview/basta1.jpg",
    average_rating: 8.5,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }],
    genre: { id: 1, name: "Рэп" },
    release_date: "2006-01-01",
    description: "Первый сольный альбом Басты"
  },
  {
    id: 2,
    title: "Баста 2",
    artist: "Баста",
    cover_image_path: "/preview/basta2.jpg",
    average_rating: 8.7,
    likes: [{ user_id: 1 }, { user_id: 2 }],
    genre: { id: 1, name: "Рэп" },
    release_date: "2007-01-01",
    description: "Второй сольный альбом Басты"
  },
  {
    id: 3,
    title: "Баста 3",
    artist: "Баста",
    cover_image_path: "/preview/basta3.jpg",
    average_rating: 9.0,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 4 }],
    genre: { id: 1, name: "Рэп" },
    release_date: "2010-01-01",
    description: "Третий сольный альбом Басты"
  },
  {
    id: 4,
    title: "Третий",
    artist: "Ноггано",
    cover_image_path: "/preview/tretiy.jpg",
    average_rating: 7.8,
    likes: [{ user_id: 1 }],
    genre: { id: 1, name: "Рэп" },
    release_date: "2008-01-01",
    description: "Третий альбом Ноггано"
  },
  {
    id: 5,
    title: "Четвёртый",
    artist: "Ноггано",
    cover_image_path: "/preview/chetvertiy.jpg",
    average_rating: 8.2,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }],
    genre: { id: 1, name: "Рэп" },
    release_date: "2009-01-01",
    description: "Четвёртый альбом Ноггано"
  }
];

export const mockTracks = [
  {
    id: 1,
    title: "Мой друг",
    album: {
      id: 1,
      title: "Баста 1",
      artist: "Баста",
      cover_image_path: "/preview/basta1.jpg"
    },
    cover_image_path: "/preview/basta1.jpg",
    average_rating: 8.5,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 4 }, { user_id: 5 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "3:45"
  },
  {
    id: 2,
    title: "Город дорог",
    album: {
      id: 2,
      title: "Баста 2",
      artist: "Баста",
      cover_image_path: "/preview/basta2.jpg"
    },
    cover_image_path: "/preview/basta2.jpg",
    average_rating: 8.7,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 4 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "4:12"
  },
  {
    id: 3,
    title: "Сансара",
    album: {
      id: 3,
      title: "Баста 3",
      artist: "Баста",
      cover_image_path: "/preview/basta3.jpg"
    },
    cover_image_path: "/preview/basta3.jpg",
    average_rating: 9.0,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 4 }, { user_id: 5 }, { user_id: 6 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "4:30"
  },
  {
    id: 4,
    title: "Мама",
    album: {
      id: 1,
      title: "Баста 1",
      artist: "Баста",
      cover_image_path: "/preview/basta1.jpg"
    },
    cover_image_path: "/preview/basta1.jpg",
    average_rating: 8.9,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "3:20"
  },
  {
    id: 5,
    title: "Выпускной",
    album: {
      id: 2,
      title: "Баста 2",
      artist: "Баста",
      cover_image_path: "/preview/basta2.jpg"
    },
    cover_image_path: "/preview/basta2.jpg",
    average_rating: 8.3,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 4 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "4:05"
  },
  {
    id: 6,
    title: "Весна",
    album: {
      id: 3,
      title: "Баста 3",
      artist: "Баста",
      cover_image_path: "/preview/basta3.jpg"
    },
    cover_image_path: "/preview/basta3.jpg",
    average_rating: 8.6,
    likes: [{ user_id: 1 }, { user_id: 2 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "3:55"
  },
  {
    id: 7,
    title: "Лето",
    album: {
      id: 4,
      title: "Третий",
      artist: "Ноггано",
      cover_image_path: "/preview/tretiy.jpg"
    },
    cover_image_path: "/preview/tretiy.jpg",
    average_rating: 7.9,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "4:15"
  },
  {
    id: 8,
    title: "Осень",
    album: {
      id: 5,
      title: "Четвёртый",
      artist: "Ноггано",
      cover_image_path: "/preview/chetvertiy.jpg"
    },
    cover_image_path: "/preview/chetvertiy.jpg",
    average_rating: 8.1,
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 4 }],
    genres: [{ id: 1, name: "Рэп" }],
    duration: "3:50"
  }
];

export const mockReviews = [
  {
    id: 1,
    text: "Отличный альбом! Баста показал свой талант во всей красе. Каждая композиция продумана до мелочей.",
    final_score: 85.5,
    rating_rhymes: 9,
    rating_structure: 8,
    rating_implementation: 9,
    rating_individuality: 10,
    atmosphere_multiplier: 1.4,
    status: "approved",
    album_id: 1,
    album: {
      id: 1,
      title: "Баста 1",
      artist: "Баста",
      cover_image_path: "/preview/basta1.jpg"
    },
    user: {
      id: 1,
      username: "МузыкальныйКритик",
      avatar_path: null
    },
    likes: [{ user_id: 2 }, { user_id: 3 }],
    created_at: "2024-01-15T10:00:00Z"
  },
  {
    id: 2,
    text: "Второй альбом Басты - это настоящий прорыв! Звучание стало более зрелым, тексты глубже.",
    final_score: 88.2,
    rating_rhymes: 9,
    rating_structure: 9,
    rating_implementation: 9,
    rating_individuality: 9,
    atmosphere_multiplier: 1.5,
    status: "approved",
    album_id: 2,
    album: {
      id: 2,
      title: "Баста 2",
      artist: "Баста",
      cover_image_path: "/preview/basta2.jpg"
    },
    user: {
      id: 2,
      username: "Meloman",
      avatar_path: null
    },
    likes: [{ user_id: 1 }, { user_id: 3 }, { user_id: 4 }],
    created_at: "2024-01-16T14:30:00Z"
  },
  {
    id: 3,
    text: "Третий альбом - вершина творчества! Каждая песня - это отдельная история, которая затягивает.",
    final_score: 92.5,
    rating_rhymes: 10,
    rating_structure: 10,
    rating_implementation: 9,
    rating_individuality: 10,
    atmosphere_multiplier: 1.6,
    status: "approved",
    album_id: 3,
    album: {
      id: 3,
      title: "Баста 3",
      artist: "Баста",
      cover_image_path: "/preview/basta3.jpg"
    },
    user: {
      id: 3,
      username: "SoundExplorer",
      avatar_path: null
    },
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 4 }, { user_id: 5 }],
    created_at: "2024-01-17T09:15:00Z"
  },
  {
    id: 4,
    text: "Ноггано продолжает радовать! Третий альбом показывает рост артиста и его музыкальный вкус.",
    final_score: 78.5,
    rating_rhymes: 8,
    rating_structure: 7,
    rating_implementation: 8,
    rating_individuality: 8,
    atmosphere_multiplier: 1.3,
    status: "approved",
    album_id: 4,
    album: {
      id: 4,
      title: "Третий",
      artist: "Ноггано",
      cover_image_path: "/preview/tretiy.jpg"
    },
    user: {
      id: 1,
      username: "МузыкальныйКритик",
      avatar_path: null
    },
    likes: [{ user_id: 2 }],
    created_at: "2024-01-18T11:20:00Z"
  },
  {
    id: 5,
    text: "Четвёртый альбом Ноггано - это смесь экспериментов и классики. Интересное звучание!",
    final_score: 80.2,
    rating_rhymes: 8,
    rating_structure: 8,
    rating_implementation: 8,
    rating_individuality: 7,
    atmosphere_multiplier: 1.4,
    status: "approved",
    album_id: 5,
    album: {
      id: 5,
      title: "Четвёртый",
      artist: "Ноггано",
      cover_image_path: "/preview/chetvertiy.jpg"
    },
    user: {
      id: 2,
      username: "Meloman",
      avatar_path: null
    },
    likes: [{ user_id: 1 }, { user_id: 3 }],
    created_at: "2024-01-19T16:45:00Z"
  },
  {
    id: 6,
    text: "Ещё одна рецензия на Баста 1. Классика жанра, которая не теряет актуальности даже спустя годы.",
    final_score: 86.8,
    rating_rhymes: 9,
    rating_structure: 8,
    rating_implementation: 9,
    rating_individuality: 9,
    atmosphere_multiplier: 1.5,
    status: "approved",
    album_id: 1,
    album: {
      id: 1,
      title: "Баста 1",
      artist: "Баста",
      cover_image_path: "/preview/basta1.jpg"
    },
    user: {
      id: 4,
      username: "MusicLover",
      avatar_path: null
    },
    likes: [{ user_id: 1 }, { user_id: 2 }, { user_id: 3 }, { user_id: 5 }],
    created_at: "2024-01-20T13:10:00Z"
  }
];

