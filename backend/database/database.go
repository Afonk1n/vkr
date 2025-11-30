package database

import (
	"fmt"
	"log"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ensureDatabaseExists checks if database exists and creates it if not
func ensureDatabaseExists() error {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return fmt.Errorf("DB_NAME environment variable is not set")
	}

	// Connect to PostgreSQL server (not to specific database)
	adminDSN := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s sslmode=%s dbname=postgres",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	// Check if database exists
	var count int64
	result := adminDB.Raw(
		"SELECT COUNT(*) FROM pg_database WHERE datname = $1",
		dbName,
	).Scan(&count)

	if result.Error != nil {
		sqlDB, _ := adminDB.DB()
		sqlDB.Close()
		return fmt.Errorf("failed to check database existence: %w", result.Error)
	}

	// Create database if it doesn't exist
	if count == 0 {
		log.Printf("Database '%s' does not exist, creating...", dbName)

		// Terminate existing connections to the database (if any)
		adminDB.Exec(fmt.Sprintf(
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()",
			dbName,
		))

		// Create database (use quote_ident for safe identifier quoting)
		createSQL := fmt.Sprintf(`CREATE DATABASE %s`, fmt.Sprintf(`"%s"`, dbName))
		if err := adminDB.Exec(createSQL).Error; err != nil {
			sqlDB, _ := adminDB.DB()
			sqlDB.Close()
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	} else {
		log.Printf("Database '%s' already exists", dbName)
	}

	// Close admin connection
	sqlDB, _ := adminDB.DB()
	sqlDB.Close()

	return nil
}

// InitDB initializes database connection and runs migrations
func InitDB() (*gorm.DB, error) {
	// Ensure database exists
	if err := ensureDatabaseExists(); err != nil {
		return nil, fmt.Errorf("database setup failed: %w", err)
	}

	// Build DSN from environment variables
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	// Open database connection
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Run migrations
	if err := runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed initial data
	if err := seedData(); err != nil {
		log.Printf("Warning: failed to seed data: %v", err)
	}

	// Update cover images for existing albums (even if seed was skipped)
	if err := updateAlbumCoverImages(); err != nil {
		log.Printf("Warning: failed to update album cover images: %v", err)
	}

	// Seed tracks (separate check, can be added even if albums exist)
	if err := seedTracks(); err != nil {
		log.Printf("Warning: failed to seed tracks: %v", err)
	}

	// Seed reviews (separate check, can be added even if users exist)
	if err := seedReviews(); err != nil {
		log.Printf("Warning: failed to seed reviews: %v", err)
	}

	// Seed track likes (for testing)
	if err := seedTrackLikes(); err != nil {
		log.Printf("Warning: failed to seed track likes: %v", err)
	}

	return DB, nil
}

// runMigrations runs database migrations
func runMigrations() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&models.User{},
		&models.Genre{},
		&models.Album{},
		&models.Track{},
		&models.TrackGenre{},
		&models.Review{},
		&models.ReviewLike{},
		&models.TrackLike{},
		&models.AlbumLike{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// seedData seeds initial data into database
func seedData() error {
	log.Println("Seeding initial data...")

	// Check if data already exists
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Data already exists, skipping seed")
		return nil
	}

	// Seed genres
	genres := []models.Genre{
		{Name: "Рок", Description: "Рок-музыка"},
		{Name: "Поп", Description: "Поп-музыка"},
		{Name: "Хип-хоп", Description: "Хип-хоп и рэп"},
		{Name: "Электронная", Description: "Электронная музыка"},
		{Name: "Джаз", Description: "Джаз"},
		{Name: "Классическая", Description: "Классическая музыка"},
	}

	var genreMap = make(map[string]models.Genre)
	for _, genre := range genres {
		if err := DB.FirstOrCreate(&genre, models.Genre{Name: genre.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed genre %s: %w", genre.Name, err)
		}
		genreMap[genre.Name] = genre
	}

	// Seed admin user
	adminPassword, _ := utils.HashPassword("admin123")
	admin := models.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: adminPassword,
		IsAdmin:  true,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Seed test user
	testPassword, _ := utils.HashPassword("test123")
	testUser := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: testPassword,
		IsAdmin:  false,
	}
	if err := DB.Create(&testUser).Error; err != nil {
		return fmt.Errorf("failed to seed test user: %w", err)
	}

	// Seed additional test users for more likes
	testUsers := []models.User{
		{Username: "musiclover1", Email: "music1@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover2", Email: "music2@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover3", Email: "music3@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover4", Email: "music4@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover5", Email: "music5@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover6", Email: "music6@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover7", Email: "music7@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover8", Email: "music8@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover9", Email: "music9@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover10", Email: "music10@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover11", Email: "music11@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover12", Email: "music12@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover13", Email: "music13@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover14", Email: "music14@example.com", Password: testPassword, IsAdmin: false},
		{Username: "musiclover15", Email: "music15@example.com", Password: testPassword, IsAdmin: false},
	}

	var allTestUsers []models.User
	allTestUsers = append(allTestUsers, admin, testUser)
	for _, user := range testUsers {
		var existingUser models.User
		if err := DB.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
			if err := DB.Create(&user).Error; err != nil {
				log.Printf("Warning: failed to create test user %s: %v", user.Username, err)
			} else {
				allTestUsers = append(allTestUsers, user)
			}
		} else {
			allTestUsers = append(allTestUsers, existingUser)
		}
	}

	// Seed albums
	albums := []models.Album{
		{
			Title:          "Abbey Road",
			Artist:         "The Beatles",
			GenreID:        genreMap["Рок"].ID,
			CoverImagePath: "/preview/1.jpg",
			Description:    "Легендарный альбом The Beatles, один из величайших в истории рок-музыки",
			AverageRating:  0,
		},
		{
			Title:          "Thriller",
			Artist:         "Michael Jackson",
			GenreID:        genreMap["Поп"].ID,
			CoverImagePath: "/preview/4.jpg",
			Description:    "Самый продаваемый альбом всех времен",
			AverageRating:  0,
		},
		{
			Title:          "The Chronic",
			Artist:         "Dr. Dre",
			GenreID:        genreMap["Хип-хоп"].ID,
			CoverImagePath: "/preview/7.jpg",
			Description:    "Классический альбом хип-хопа 90-х",
			AverageRating:  0,
		},
		{
			Title:          "Random Access Memories",
			Artist:         "Daft Punk",
			GenreID:        genreMap["Электронная"].ID,
			CoverImagePath: "/preview/8.jpg",
			Description:    "Электронный альбом с элементами диско и фанка",
			AverageRating:  0,
		},
		{
			Title:          "Kind of Blue",
			Artist:         "Miles Davis",
			GenreID:        genreMap["Джаз"].ID,
			CoverImagePath: "/preview/9.jpg",
			Description:    "Величайший джазовый альбом всех времен",
			AverageRating:  0,
		},
	}

	// Seed albums - create or update with cover images
	albumMap := map[string]string{
		"Abbey Road":             "/preview/1.jpg",
		"Thriller":               "/preview/4.jpg",
		"The Chronic":            "/preview/7.jpg",
		"Random Access Memories": "/preview/8.jpg",
		"Kind of Blue":           "/preview/9.jpg",
	}

	for _, album := range albums {
		var existingAlbum models.Album
		if err := DB.Where("title = ? AND artist = ?", album.Title, album.Artist).First(&existingAlbum).Error; err == nil {
			// Album exists, update cover_image_path if it's empty
			if existingAlbum.CoverImagePath == "" && albumMap[album.Title] != "" {
				existingAlbum.CoverImagePath = albumMap[album.Title]
				if err := DB.Save(&existingAlbum).Error; err != nil {
					log.Printf("Warning: failed to update cover_image_path for album %s: %v", album.Title, err)
				} else {
					log.Printf("Updated cover_image_path for album: %s", album.Title)
				}
			}
		} else {
			// Album doesn't exist, create it
			if err := DB.Create(&album).Error; err != nil {
				return fmt.Errorf("failed to seed album %s: %w", album.Title, err)
			}
		}
	}

	// Seed some album likes for testing
	// Use already created users
	// Admin likes first 3 albums
	for i := 0; i < 3 && i < len(albums); i++ {
		like := models.AlbumLike{
			UserID:  admin.ID,
			AlbumID: albums[i].ID,
		}
		DB.Create(&like)
	}
	// testUser likes albums 1-4
	for i := 0; i < 4 && i < len(albums); i++ {
		like := models.AlbumLike{
			UserID:  testUser.ID,
			AlbumID: albums[i].ID,
		}
		DB.Create(&like)
	}

	log.Println("Initial data seeded successfully")
	return nil
}

// seedTracks seeds tracks with multiple genres into database
func seedTracks() error {
	log.Println("Seeding tracks...")

	// Check if tracks already exist
	var trackCount int64
	DB.Model(&models.Track{}).Count(&trackCount)
	if trackCount > 0 {
		log.Println("Tracks already exist, skipping seed")
		return nil
	}

	// Get albums
	var albums []models.Album
	if err := DB.Find(&albums).Error; err != nil || len(albums) == 0 {
		log.Println("No albums found, skipping track seed")
		return nil
	}

	// Get genres
	var genres []models.Genre
	if err := DB.Find(&genres).Error; err != nil || len(genres) == 0 {
		log.Println("No genres found, skipping track seed")
		return nil
	}

	// Create genre map for easy lookup
	genreMap := make(map[string]models.Genre)
	for _, genre := range genres {
		genreMap[genre.Name] = genre
	}

	// Create tracks for albums with multiple genres
	tracks := []struct {
		AlbumTitle     string
		Title          string
		Duration       int
		TrackNum       int
		GenreNames     []string // Multiple genres per track
		CoverImagePath string   // Optional cover image path
	}{
		// Abbey Road tracks
		{"Abbey Road", "Come Together", 259, 1, []string{"Рок", "Поп"}, "/preview/1.jpg"},
		{"Abbey Road", "Something", 182, 2, []string{"Рок", "Поп"}, "/preview/2.jpg"},
		{"Abbey Road", "Maxwell's Silver Hammer", 207, 3, []string{"Рок"}, ""},
		{"Abbey Road", "Oh! Darling", 207, 4, []string{"Рок", "Поп"}, ""},
		{"Abbey Road", "Octopus's Garden", 171, 5, []string{"Рок"}, ""},
		{"Abbey Road", "I Want You (She's So Heavy)", 468, 6, []string{"Рок"}, ""},
		{"Abbey Road", "Here Comes the Sun", 185, 7, []string{"Рок", "Поп"}, "/preview/3.jpg"},
		{"Abbey Road", "Because", 164, 8, []string{"Рок"}, ""},
		{"Abbey Road", "You Never Give Me Your Money", 243, 9, []string{"Рок", "Поп"}, ""},
		{"Abbey Road", "Sun King", 147, 10, []string{"Рок"}, ""},
		{"Abbey Road", "Mean Mr. Mustard", 66, 11, []string{"Рок"}, ""},
		{"Abbey Road", "Polythene Pam", 72, 12, []string{"Рок"}, ""},
		{"Abbey Road", "She Came In Through the Bathroom Window", 118, 13, []string{"Рок", "Поп"}, ""},
		{"Abbey Road", "Golden Slumbers", 91, 14, []string{"Рок", "Поп"}, ""},
		{"Abbey Road", "Carry That Weight", 96, 15, []string{"Рок"}, ""},
		{"Abbey Road", "The End", 122, 16, []string{"Рок"}, ""},
		{"Abbey Road", "Her Majesty", 23, 17, []string{"Рок", "Поп"}, ""},

		// Thriller tracks
		{"Thriller", "Wanna Be Startin' Somethin'", 363, 1, []string{"Поп", "Электронная"}, ""},
		{"Thriller", "Baby Be Mine", 260, 2, []string{"Поп"}, ""},
		{"Thriller", "The Girl Is Mine", 222, 3, []string{"Поп"}, ""},
		{"Thriller", "Thriller", 357, 4, []string{"Поп", "Электронная"}, "/preview/4.jpg"},
		{"Thriller", "Beat It", 258, 5, []string{"Поп", "Рок"}, "/preview/5.jpg"},
		{"Thriller", "Billie Jean", 294, 6, []string{"Поп", "Электронная"}, "/preview/6.jpg"},
		{"Thriller", "Human Nature", 207, 7, []string{"Поп"}, ""},
		{"Thriller", "P.Y.T. (Pretty Young Thing)", 238, 8, []string{"Поп", "Электронная"}, ""},
		{"Thriller", "The Lady in My Life", 300, 9, []string{"Поп"}, ""},

		// The Chronic tracks
		{"The Chronic", "The Chronic (Intro)", 68, 1, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Fuck Wit Dre Day (And Everybody's Celebratin')", 269, 2, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Let Me Ride", 280, 3, []string{"Хип-хоп", "Электронная"}, ""},
		{"The Chronic", "The Day the Niggaz Took Over", 241, 4, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Nuthin' But a 'G' Thang", 214, 5, []string{"Хип-хоп"}, "/preview/7.jpg"},
		{"The Chronic", "Deeez Nuuuts", 300, 6, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Lil' Ghetto Boy", 300, 7, []string{"Хип-хоп"}, ""},
		{"The Chronic", "A Nigga Witta Gun", 207, 8, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Rat-Tat-Tat-Tat", 220, 9, []string{"Хип-хоп"}, ""},
		{"The Chronic", "The $20 Sack Pyramid", 60, 10, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Lyrical Gangbang", 252, 11, []string{"Хип-хоп"}, ""},
		{"The Chronic", "High Powered", 180, 12, []string{"Хип-хоп"}, ""},
		{"The Chronic", "The Doctor's Office", 60, 13, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Stranded on Death Row", 300, 14, []string{"Хип-хоп"}, ""},
		{"The Chronic", "The Roach (The Chronic Outro)", 120, 15, []string{"Хип-хоп"}, ""},
		{"The Chronic", "Bitches Ain't Shit", 288, 16, []string{"Хип-хоп"}, ""},

		// Random Access Memories tracks
		{"Random Access Memories", "Give Life Back to Music", 274, 1, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "The Game of Love", 321, 2, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "Giorgio by Moroder", 544, 3, []string{"Электронная"}, ""},
		{"Random Access Memories", "Within", 237, 4, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "Instant Crush", 337, 5, []string{"Электронная", "Рок"}, ""},
		{"Random Access Memories", "Lose Yourself to Dance", 353, 6, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "Touch", 498, 7, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "Get Lucky", 248, 8, []string{"Электронная", "Поп"}, "/preview/8.jpg"},
		{"Random Access Memories", "Beyond", 290, 9, []string{"Электронная"}, ""},
		{"Random Access Memories", "Motherboard", 340, 10, []string{"Электронная"}, ""},
		{"Random Access Memories", "Fragments of Time", 279, 11, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "Doin' It Right", 251, 12, []string{"Электронная", "Поп"}, ""},
		{"Random Access Memories", "Contact", 403, 13, []string{"Электронная"}, ""},

		// Kind of Blue tracks
		{"Kind of Blue", "So What", 564, 1, []string{"Джаз"}, ""},
		{"Kind of Blue", "Freddie Freeloader", 589, 2, []string{"Джаз"}, ""},
		{"Kind of Blue", "Blue in Green", 338, 3, []string{"Джаз"}, ""},
		{"Kind of Blue", "All Blues", 693, 4, []string{"Джаз"}, ""},
		{"Kind of Blue", "Flamenco Sketches", 566, 5, []string{"Джаз"}, ""},
	}

	// Create tracks and assign genres
	for _, trackData := range tracks {
		// Find album
		var album models.Album
		for _, a := range albums {
			if a.Title == trackData.AlbumTitle {
				album = a
				break
			}
		}
		if album.ID == 0 {
			continue // Skip if album not found
		}

		// Create track
		track := models.Track{
			AlbumID:        album.ID,
			Title:          trackData.Title,
			Duration:       &trackData.Duration,
			TrackNumber:    &trackData.TrackNum,
			CoverImagePath: trackData.CoverImagePath,
		}

		if err := DB.Create(&track).Error; err != nil {
			log.Printf("Warning: failed to create track %s: %v", trackData.Title, err)
			continue
		}

		// Assign multiple genres
		var trackGenres []models.Genre
		for _, genreName := range trackData.GenreNames {
			if genre, exists := genreMap[genreName]; exists {
				trackGenres = append(trackGenres, genre)
			}
		}

		if len(trackGenres) > 0 {
			if err := DB.Model(&track).Association("Genres").Append(trackGenres); err != nil {
				log.Printf("Warning: failed to assign genres to track %s: %v", trackData.Title, err)
			}
		}
	}

	log.Println("Tracks seeded successfully")
	return nil
}

// seedTrackLikes seeds track likes for testing
func seedTrackLikes() error {
	log.Println("Seeding track likes...")

	// Check if track likes already exist
	var likeCount int64
	DB.Model(&models.TrackLike{}).Count(&likeCount)
	if likeCount > 0 {
		log.Println("Track likes already exist, skipping seed")
		return nil
	}

	// Get users
	var admin, testUser models.User
	if err := DB.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		log.Println("Admin user not found, skipping track likes seed")
		return nil
	}
	if err := DB.Where("email = ?", "test@example.com").First(&testUser).Error; err != nil {
		log.Println("Test user not found, skipping track likes seed")
		return nil
	}

	// Get some tracks (first 10 tracks)
	var tracks []models.Track
	if err := DB.Limit(10).Find(&tracks).Error; err != nil || len(tracks) == 0 {
		log.Println("No tracks found, skipping track likes seed")
		return nil
	}

	// Create likes: admin likes first 5 tracks, testUser likes tracks 3-8
	for i, track := range tracks {
		if i < 5 {
			// Admin likes first 5 tracks
			like := models.TrackLike{
				UserID:  admin.ID,
				TrackID: track.ID,
			}
			DB.Create(&like)
		}
		if i >= 2 && i < 8 {
			// testUser likes tracks 3-8
			like := models.TrackLike{
				UserID:  testUser.ID,
				TrackID: track.ID,
			}
			DB.Create(&like)
		}
	}

	log.Println("Track likes seeded successfully")
	return nil
}

// seedReviews seeds test reviews into database
func seedReviews() error {
	log.Println("Seeding test reviews...")

	// Check if reviews already exist
	var reviewCount int64
	DB.Model(&models.Review{}).Count(&reviewCount)
	if reviewCount > 0 {
		log.Println("Reviews already exist, skipping seed")
		return nil
	}

	// Get users
	var admin, testUser models.User
	if err := DB.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		log.Println("Admin user not found, skipping review seed")
		return nil
	}
	if err := DB.Where("email = ?", "test@example.com").First(&testUser).Error; err != nil {
		log.Println("Test user not found, skipping review seed")
		return nil
	}

	// Get albums
	var albums []models.Album
	if err := DB.Find(&albums).Error; err != nil || len(albums) == 0 {
		log.Println("No albums found, skipping review seed")
		return nil
	}

	// Helper function to convert atmosphere rating (1-10) to multiplier (1.0000-1.6072)
	convertAtmosphereToMultiplier := func(rating int) float64 {
		step := 0.6072 / 9.0
		return 1.0000 + float64(rating-1)*step
	}

	// Create test reviews (using atmosphere ratings 1-10, converted to multiplier)
	reviews := []models.Review{
		{
			UserID:               testUser.ID,
			AlbumID:              &albums[0].ID, // Abbey Road
			Text:                 "Невероятный альбом! The Beatles на пике формы. Каждая композиция - это произведение искусства. Особенно выделяется медилейн с 'Come Together' и 'Something'. Звукорежиссура безупречна, аранжировки гениальны.",
			RatingRhymes:         9,
			RatingStructure:      10,
			RatingImplementation: 10,
			RatingIndividuality:  9,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(8), // 8/10 -> ~1.4737
			Status:               models.ReviewStatusApproved,
			ModeratedBy:          &admin.ID,
		},
		{
			UserID:               testUser.ID,
			AlbumID:              &albums[1].ID, // Thriller
			Text:                 "Классика поп-музыки! Michael Jackson создал настоящий шедевр. 'Billie Jean' и 'Beat It' - это вечные хиты. Танцевальные ритмы, запоминающиеся мелодии и неповторимый вокал.",
			RatingRhymes:         8,
			RatingStructure:      9,
			RatingImplementation: 10,
			RatingIndividuality:  10,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(7), // 7/10 -> ~1.4062
			Status:               models.ReviewStatusApproved,
			ModeratedBy:          &admin.ID,
		},
		{
			UserID:               admin.ID,
			AlbumID:              &albums[2].ID, // The Chronic
			Text:                 "Dr. Dre задал новый стандарт для хип-хопа. Битмейкинг на высшем уровне, семплы подобраны идеально. 'Nuthin' But a G Thang' - это гимн эпохи. Альбом звучит свежо даже спустя десятилетия.",
			RatingRhymes:         9,
			RatingStructure:      9,
			RatingImplementation: 10,
			RatingIndividuality:  9,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(9), // 9/10 -> ~1.5405
			Status:               models.ReviewStatusApproved,
			ModeratedBy:          &admin.ID,
		},
		{
			UserID:               testUser.ID,
			AlbumID:              &albums[3].ID, // Random Access Memories
			Text:                 "Daft Punk создали что-то особенное. Смешение электроники, диско и фанка работает идеально. 'Get Lucky' - это летний хит, который не надоедает. Альбом звучит как путешествие во времени.",
			RatingRhymes:         7,
			RatingStructure:      9,
			RatingImplementation: 10,
			RatingIndividuality:  8,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(10), // 10/10 -> 1.6072 (max)
			Status:               models.ReviewStatusApproved,
			ModeratedBy:          &admin.ID,
		},
		{
			UserID:               admin.ID,
			AlbumID:              &albums[4].ID, // Kind of Blue
			Text:                 "Абсолютный шедевр джаза. Miles Davis на этом альбоме создал что-то трансцендентное. Modal jazz в его лучшем проявлении. Каждая нота на своем месте. 'So What' - это икона жанра.",
			RatingRhymes:         10,
			RatingStructure:      10,
			RatingImplementation: 10,
			RatingIndividuality:  10,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(10), // 10/10 -> 1.6072 (max, gives 90 points)
			Status:               models.ReviewStatusApproved,
			ModeratedBy:          &admin.ID,
		},
		{
			UserID:               testUser.ID,
			AlbumID:              &albums[0].ID, // Abbey Road (вторая рецензия)
			Text:                 "Хочу добавить еще несколько слов об этом альбоме. 'Here Comes the Sun' - одна из самых красивых песен в истории музыки. George Harrison показал свой талант композитора.",
			RatingRhymes:         8,
			RatingStructure:      9,
			RatingImplementation: 9,
			RatingIndividuality:  8,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(6), // 6/10 -> ~1.3388
			Status:               models.ReviewStatusPending,
		},
		{
			UserID:               testUser.ID,
			AlbumID:              &albums[1].ID, // Thriller (только оценка, без текста)
			Text:                 "",
			RatingRhymes:         7,
			RatingStructure:      8,
			RatingImplementation: 9,
			RatingIndividuality:  9,
			AtmosphereMultiplier: convertAtmosphereToMultiplier(6), // 6/10 -> ~1.3388
			Status:               models.ReviewStatusApproved,
			ModeratedBy:          &admin.ID,
		},
	}

	// Calculate final scores and create reviews
	for i := range reviews {
		reviews[i].CalculateFinalScore()
		if err := DB.Create(&reviews[i]).Error; err != nil {
			return fmt.Errorf("failed to seed review %d: %w", i+1, err)
		}
	}

	// Get some tracks for track reviews
	var tracks []models.Track
	var allReviews []models.Review // Collect all reviews for likes seeding
	if err := DB.Limit(5).Find(&tracks).Error; err == nil && len(tracks) > 0 {
		// Add some track reviews
		trackReviews := []models.Review{
			{
				UserID:               testUser.ID,
				TrackID:              &tracks[0].ID, // First track
				Text:                 "Отличный трек! Один из лучших на альбоме. Мелодия запоминающаяся, аранжировка на высоте.",
				RatingRhymes:         8,
				RatingStructure:      9,
				RatingImplementation: 9,
				RatingIndividuality:  8,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			{
				UserID:               admin.ID,
				TrackID:              &tracks[1].ID, // Second track
				Text:                 "Классический хит! Эта композиция стала символом эпохи. Вокал неповторим.",
				RatingRhymes:         9,
				RatingStructure:      10,
				RatingImplementation: 10,
				RatingIndividuality:  9,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			{
				UserID:               testUser.ID,
				TrackID:              &tracks[2].ID, // Third track
				Text:                 "",
				RatingRhymes:         7,
				RatingStructure:      8,
				RatingImplementation: 8,
				RatingIndividuality:  7,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
		}

		for i := range trackReviews {
			trackReviews[i].CalculateFinalScore()
			if err := DB.Create(&trackReviews[i]).Error; err != nil {
				log.Printf("Warning: failed to create track review %d: %v", i+1, err)
			}
		}
	}

	// Reload all reviews from DB to get correct IDs
	if err := DB.Where("status = ?", models.ReviewStatusApproved).Find(&allReviews).Error; err != nil {
		log.Printf("Warning: failed to reload reviews for likes: %v", err)
		allReviews = reviews // Fallback to original reviews
	}

	// Update album average ratings
	for _, album := range albums {
		var reviews []models.Review
		if err := DB.Where("album_id = ? AND status = ?", album.ID, models.ReviewStatusApproved).Find(&reviews).Error; err == nil && len(reviews) > 0 {
			var totalScore float64
			for _, review := range reviews {
				totalScore += review.FinalScore
			}
			averageRating := totalScore / float64(len(reviews))
			// Round to nearest integer
			roundedAverage := float64(int(averageRating + 0.5))
			DB.Model(&album).Update("average_rating", roundedAverage)
		}
	}

	// Get all test users for likes
	var allTestUsers []models.User
	if err := DB.Find(&allTestUsers).Error; err != nil {
		log.Printf("Warning: failed to fetch users for review likes: %v", err)
		allTestUsers = []models.User{admin, testUser} // Fallback to basic users
	}

	// Seed review likes for testing - create 5-20 likes per review
	var reviewLikes []models.ReviewLike
	for i, review := range allReviews {
		if review.Status == models.ReviewStatusApproved && review.ID > 0 {
			// Generate random number of likes between 5 and 20
			numLikes := 5 + (i % 16) // Will give 5-20 likes
			if numLikes > len(allTestUsers) {
				numLikes = len(allTestUsers)
			}

			// Create likes from different users
			for j := 0; j < numLikes; j++ {
				userIndex := j % len(allTestUsers)
				// Check if like already exists
				var existingLike models.ReviewLike
				if err := DB.Where("user_id = ? AND review_id = ?", allTestUsers[userIndex].ID, review.ID).First(&existingLike).Error; err != nil {
					like := models.ReviewLike{
						UserID:   allTestUsers[userIndex].ID,
						ReviewID: review.ID,
					}
					reviewLikes = append(reviewLikes, like)
				}
			}
		}
	}

	for _, like := range reviewLikes {
		if err := DB.Create(&like).Error; err != nil {
			log.Printf("Warning: failed to create review like: %v", err)
		}
	}

	log.Println("Test reviews seeded successfully")
	return nil
}

// updateAlbumCoverImages updates cover_image_path for existing albums
func updateAlbumCoverImages() error {
	albumMap := map[string]string{
		"Abbey Road":             "/preview/1.jpg",
		"Thriller":               "/preview/4.jpg",
		"The Chronic":            "/preview/7.jpg",
		"Random Access Memories": "/preview/8.jpg",
		"Kind of Blue":           "/preview/9.jpg",
	}

	for title, coverPath := range albumMap {
		var album models.Album
		if err := DB.Where("title = ?", title).First(&album).Error; err == nil {
			if album.CoverImagePath == "" {
				album.CoverImagePath = coverPath
				if err := DB.Save(&album).Error; err != nil {
					log.Printf("Warning: failed to update cover_image_path for album %s: %v", title, err)
				} else {
					log.Printf("Updated cover_image_path for album: %s -> %s", title, coverPath)
				}
			}
		}
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
