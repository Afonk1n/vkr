package database

import (
	"fmt"
	"log"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"os"
	"time"

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

	// Check database state before seeding
	log.Println("=== Database state BEFORE seeding ===")
	logDatabaseState()

	// Seed initial data
	log.Println("=== Starting data seeding ===")
	if err := seedData(); err != nil {
		log.Printf("ERROR: failed to seed data: %v", err)
	} else {
		log.Println("✓ Data seeding completed successfully")
	}

	// Update cover images for existing albums (even if seed was skipped)
	if err := updateAlbumCoverImages(); err != nil {
		log.Printf("Warning: failed to update album cover images: %v", err)
	}

	// Seed tracks (separate check, can be added even if albums exist)
	if err := seedTracks(); err != nil {
		log.Printf("ERROR: failed to seed tracks: %v", err)
	} else {
		log.Println("✓ Tracks seeding completed successfully")
	}

	// Seed reviews (separate check, can be added even if users exist)
	if err := seedReviews(); err != nil {
		log.Printf("ERROR: failed to seed reviews: %v", err)
	} else {
		log.Println("✓ Reviews seeding completed successfully")
	}

	// Seed track likes (for testing)
	if err := seedTrackLikes(); err != nil {
		log.Printf("ERROR: failed to seed track likes: %v", err)
	} else {
		log.Println("✓ Track likes seeding completed successfully")
	}
	log.Println("=== Data seeding finished ===")

	// Check database state after seeding
	log.Println("=== Database state AFTER seeding ===")
	logDatabaseState()

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

	// Fix reviews table constraints - album_id and track_id should be nullable
	// This fixes the issue where GORM might have created NOT NULL constraints
	if err := fixReviewsTableConstraints(); err != nil {
		log.Printf("Warning: failed to fix reviews table constraints: %v", err)
		// Don't fail migration, just log warning
	}

	log.Println("Migrations completed successfully")
	return nil
}

// fixReviewsTableConstraints fixes the constraints on reviews table
// to ensure album_id and track_id are nullable
func fixReviewsTableConstraints() error {
	// Check if table exists
	var exists bool
	if err := DB.Raw(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'reviews')",
	).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check if reviews table exists: %w", err)
	}

	if !exists {
		log.Println("Reviews table does not exist, skipping constraint fix")
		return nil
	}

	// Check current constraints
	var albumIDNullable bool
	var trackIDNullable bool

	if err := DB.Raw(`
		SELECT 
			is_nullable = 'YES' as album_id_nullable
		FROM information_schema.columns 
		WHERE table_name = 'reviews' AND column_name = 'album_id'
	`).Scan(&albumIDNullable).Error; err != nil {
		return fmt.Errorf("failed to check album_id constraint: %w", err)
	}

	if err := DB.Raw(`
		SELECT 
			is_nullable = 'YES' as track_id_nullable
		FROM information_schema.columns 
		WHERE table_name = 'reviews' AND column_name = 'track_id'
	`).Scan(&trackIDNullable).Error; err != nil {
		return fmt.Errorf("failed to check track_id constraint: %w", err)
	}

	// Fix album_id if needed
	if !albumIDNullable {
		log.Println("Fixing album_id constraint in reviews table (making it nullable)...")
		if err := DB.Exec("ALTER TABLE reviews ALTER COLUMN album_id DROP NOT NULL").Error; err != nil {
			return fmt.Errorf("failed to alter album_id column: %w", err)
		}
		log.Println("album_id constraint fixed")
	}

	// Fix track_id if needed
	if !trackIDNullable {
		log.Println("Fixing track_id constraint in reviews table (making it nullable)...")
		if err := DB.Exec("ALTER TABLE reviews ALTER COLUMN track_id DROP NOT NULL").Error; err != nil {
			return fmt.Errorf("failed to alter track_id column: %w", err)
		}
		log.Println("track_id constraint fixed")
	}

	return nil
}

// seedData seeds initial data into database
func seedData() error {
	log.Println("Seeding initial data...")

	// Seed genres - reload from DB after creation to get correct IDs
	genresToCreate := []models.Genre{
		{Name: "Рок", Description: "Рок-музыка"},
		{Name: "Поп", Description: "Поп-музыка"},
		{Name: "Хип-хоп", Description: "Хип-хоп и рэп"},
		{Name: "Электронная", Description: "Электронная музыка"},
		{Name: "Джаз", Description: "Джаз"},
		{Name: "Классическая", Description: "Классическая музыка"},
	}

	// Create genres if they don't exist
	createdGenres := 0
	existingGenres := 0
	for _, genre := range genresToCreate {
		var existingGenre models.Genre
		if err := DB.Where("name = ?", genre.Name).First(&existingGenre).Error; err != nil {
			// Genre doesn't exist, create it
			if err := DB.Create(&genre).Error; err != nil {
				log.Printf("ERROR: Failed to create genre %s: %v", genre.Name, err)
				return fmt.Errorf("failed to seed genre %s: %w", genre.Name, err)
			}
			createdGenres++
			log.Printf("✓ Created genre: %s (ID: %d)", genre.Name, genre.ID)
		} else {
			existingGenres++
			log.Printf("  Genre already exists: %s (ID: %d)", genre.Name, existingGenre.ID)
		}
	}
	log.Printf("Genres: %d created, %d already existed", createdGenres, existingGenres)

	// Reload all genres from DB to get correct IDs
	var allGenres []models.Genre
	if err := DB.Find(&allGenres).Error; err != nil {
		return fmt.Errorf("failed to reload genres: %w", err)
	}

	// Create genre map with correct IDs
	genreMap := make(map[string]models.Genre)
	for _, genre := range allGenres {
		genreMap[genre.Name] = genre
		if genre.ID == 0 {
			return fmt.Errorf("genre %s has invalid ID (0)", genre.Name)
		}
	}

	if len(genreMap) == 0 {
		return fmt.Errorf("no genres found in database after seeding")
	}
	log.Printf("Loaded %d genres from database", len(genreMap))

	// Seed admin user
	adminPassword, _ := utils.HashPassword("admin123")
	var admin models.User
	if err := DB.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		// User doesn't exist, create it
		admin = models.User{
			Username:    "admin",
			Email:       "admin@example.com",
			Password:    adminPassword,
			SocialLinks: "{}", // Valid JSON for jsonb field
			IsAdmin:     true,
		}
		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("ERROR: Failed to create admin user: %v", err)
			return fmt.Errorf("failed to seed admin user: %w", err)
		}
		log.Printf("✓ Created admin user (ID: %d, Email: %s)", admin.ID, admin.Email)
	} else {
		log.Printf("  Admin user already exists (ID: %d, Email: %s)", admin.ID, admin.Email)
	}

	// Seed test user
	testPassword, _ := utils.HashPassword("test123")
	var testUser models.User
	if err := DB.Where("email = ?", "test@example.com").First(&testUser).Error; err != nil {
		// User doesn't exist, create it
		testUser = models.User{
			Username:    "testuser",
			Email:       "test@example.com",
			Password:    testPassword,
			SocialLinks: "{}", // Valid JSON for jsonb field
			IsAdmin:     false,
		}
		if err := DB.Create(&testUser).Error; err != nil {
			log.Printf("ERROR: Failed to create test user: %v", err)
			return fmt.Errorf("failed to seed test user: %w", err)
		}
		log.Printf("✓ Created test user (ID: %d, Email: %s)", testUser.ID, testUser.Email)
	} else {
		log.Printf("  Test user already exists (ID: %d, Email: %s)", testUser.ID, testUser.Email)
	}

	// Seed additional test users for more likes
	emptySocialLinks := "{}" // Valid JSON for jsonb field
	testUsers := []models.User{
		{Username: "musiclover1", Email: "music1@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover2", Email: "music2@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover3", Email: "music3@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover4", Email: "music4@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover5", Email: "music5@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover6", Email: "music6@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover7", Email: "music7@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover8", Email: "music8@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover9", Email: "music9@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover10", Email: "music10@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover11", Email: "music11@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover12", Email: "music12@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover13", Email: "music13@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover14", Email: "music14@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
		{Username: "musiclover15", Email: "music15@example.com", Password: testPassword, SocialLinks: emptySocialLinks, IsAdmin: false},
	}

	var allTestUsers []models.User
	allTestUsers = append(allTestUsers, admin, testUser)
	createdTestUsers := 0
	existingTestUsers := 0
	for _, user := range testUsers {
		var existingUser models.User
		if err := DB.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
			if err := DB.Create(&user).Error; err != nil {
				log.Printf("ERROR: Failed to create test user %s: %v", user.Username, err)
			} else {
				createdTestUsers++
				allTestUsers = append(allTestUsers, user)
				log.Printf("  Created test user: %s (ID: %d)", user.Username, user.ID)
			}
		} else {
			existingTestUsers++
			allTestUsers = append(allTestUsers, existingUser)
		}
	}
	log.Printf("Test users: %d created, %d already existed (total: %d)", createdTestUsers, existingTestUsers, len(allTestUsers))

	// Seed albums - verify genre IDs before using them
	rockGenre, rockExists := genreMap["Рок"]
	popGenre, popExists := genreMap["Поп"]
	hiphopGenre, hiphopExists := genreMap["Хип-хоп"]
	electronicGenre, electronicExists := genreMap["Электронная"]
	jazzGenre, jazzExists := genreMap["Джаз"]

	if !rockExists || rockGenre.ID == 0 {
		return fmt.Errorf("Рок genre not found or has invalid ID")
	}
	if !popExists || popGenre.ID == 0 {
		return fmt.Errorf("Поп genre not found or has invalid ID")
	}
	if !hiphopExists || hiphopGenre.ID == 0 {
		return fmt.Errorf("Хип-хоп genre not found or has invalid ID")
	}
	if !electronicExists || electronicGenre.ID == 0 {
		return fmt.Errorf("Электронная genre not found or has invalid ID")
	}
	if !jazzExists || jazzGenre.ID == 0 {
		return fmt.Errorf("Джаз genre not found or has invalid ID")
	}

	albums := []models.Album{
		{
			Title:          "Жить в твоей голове",
			Artist:         "Земфира",
			GenreID:        rockGenre.ID,
			CoverImagePath: "/preview/1.jpg",
			Description:    "Седьмой студийный альбом Земфиры, выпущенный в 2021 году",
			AverageRating:  0,
		},
		{
			Title:          "Vinyl #1",
			Artist:         "Zivert",
			GenreID:        popGenre.ID,
			CoverImagePath: "/preview/4.jpg",
			Description:    "Дебютный студийный альбом Zivert, один из самых успешных поп-альбомов 2019 года",
			AverageRating:  0,
		},
		{
			Title:          "Горгород",
			Artist:         "Oxxxymiron",
			GenreID:        hiphopGenre.ID,
			CoverImagePath: "/preview/7.jpg",
			Description:    "Концептуальный альбом Oxxxymiron, выпущенный в 2015 году",
			AverageRating:  0,
		},
		{
			Title:          "До свидания",
			Artist:         "IC3PEAK",
			GenreID:        electronicGenre.ID,
			CoverImagePath: "/preview/8.jpg",
			Description:    "Электронный альбом дуэта IC3PEAK с элементами индастриала и трип-хопа",
			AverageRating:  0,
		},
		{
			Title:          "Раскраски для взрослых",
			Artist:         "Монеточка",
			GenreID:        popGenre.ID,
			CoverImagePath: "/preview/9.jpg",
			Description:    "Второй студийный альбом Монеточки, выпущенный в 2018 году",
			AverageRating:  0,
		},
	}

	// Seed albums - create or update with cover images
	albumMap := map[string]string{
		"Жить в твоей голове":     "/preview/1.jpg",
		"Vinyl #1":                "/preview/4.jpg",
		"Горгород":                "/preview/7.jpg",
		"До свидания":             "/preview/8.jpg",
		"Раскраски для взрослых":   "/preview/9.jpg",
	}

	createdAlbums := 0
	existingAlbums := 0
	skippedAlbums := 0
	for _, album := range albums {
		var existingAlbum models.Album
		if err := DB.Where("title = ? AND artist = ?", album.Title, album.Artist).First(&existingAlbum).Error; err == nil {
			// Album exists, update cover_image_path if it's empty
			existingAlbums++
			if existingAlbum.CoverImagePath == "" && albumMap[album.Title] != "" {
				existingAlbum.CoverImagePath = albumMap[album.Title]
				if err := DB.Save(&existingAlbum).Error; err != nil {
					log.Printf("ERROR: Failed to update cover_image_path for album %s: %v", album.Title, err)
				} else {
					log.Printf("  Updated cover_image_path for album: %s (ID: %d)", album.Title, existingAlbum.ID)
				}
			} else {
				log.Printf("  Album already exists: %s by %s (ID: %d, GenreID: %d)", album.Title, album.Artist, existingAlbum.ID, existingAlbum.GenreID)
			}
		} else {
			// Album doesn't exist, create it
			// Verify genre ID is valid before creating
			if album.GenreID == 0 {
				log.Printf("ERROR: Album %s has invalid GenreID (0), skipping", album.Title)
				skippedAlbums++
				continue
			}
			if err := DB.Create(&album).Error; err != nil {
				log.Printf("ERROR: Failed to create album %s: %v", album.Title, err)
				return fmt.Errorf("failed to seed album %s: %w", album.Title, err)
			}
			createdAlbums++
			log.Printf("✓ Created album: %s by %s (ID: %d, GenreID: %d)", album.Title, album.Artist, album.ID, album.GenreID)
		}
	}
	log.Printf("Albums seeding complete: %d created, %d already existed, %d skipped", createdAlbums, existingAlbums, skippedAlbums)

	// Reload albums from DB to get correct IDs
	var allAlbums []models.Album
	if err := DB.Find(&allAlbums).Error; err != nil {
		log.Printf("Warning: failed to reload albums: %v", err)
		allAlbums = albums // Fallback to original albums
	} else {
		log.Printf("Reloaded %d albums from database", len(allAlbums))
	}

	// Seed some album likes for testing
	// Use already created users
	// Admin likes first 3 albums
	for i := 0; i < 3 && i < len(allAlbums); i++ {
		// Check if like already exists
		var existingLike models.AlbumLike
		if err := DB.Where("user_id = ? AND album_id = ?", admin.ID, allAlbums[i].ID).First(&existingLike).Error; err != nil {
			like := models.AlbumLike{
				UserID:  admin.ID,
				AlbumID: allAlbums[i].ID,
			}
			if err := DB.Create(&like).Error; err != nil {
				log.Printf("Warning: failed to create album like: %v", err)
			}
		}
	}
	// testUser likes albums 1-4
	for i := 0; i < 4 && i < len(allAlbums); i++ {
		// Check if like already exists
		var existingLike models.AlbumLike
		if err := DB.Where("user_id = ? AND album_id = ?", testUser.ID, allAlbums[i].ID).First(&existingLike).Error; err != nil {
			like := models.AlbumLike{
				UserID:  testUser.ID,
				AlbumID: allAlbums[i].ID,
			}
			if err := DB.Create(&like).Error; err != nil {
				log.Printf("Warning: failed to create album like: %v", err)
			}
		}
	}

	// Final verification - check that data was actually created
	var userCount, albumCount, genreCount int64
	DB.Model(&models.User{}).Count(&userCount)
	DB.Model(&models.Album{}).Count(&albumCount)
	DB.Model(&models.Genre{}).Count(&genreCount)

	log.Printf("Initial data seeded successfully: %d users, %d albums, %d genres", len(allTestUsers), len(allAlbums), genreCount)

	if userCount == 0 {
		return fmt.Errorf("no users created - seeding failed")
	}
	if albumCount == 0 {
		return fmt.Errorf("no albums created - seeding failed")
	}
	if genreCount == 0 {
		return fmt.Errorf("no genres created - seeding failed")
	}

	log.Printf("Verification: DB contains %d users, %d albums, %d genres", userCount, albumCount, genreCount)
	return nil
}

// seedTracks seeds tracks with multiple genres into database
func seedTracks() error {
	log.Println("Seeding tracks...")

	// Get albums
	var albums []models.Album
	if err := DB.Find(&albums).Error; err != nil {
		log.Printf("ERROR: Failed to query albums: %v", err)
		return fmt.Errorf("failed to query albums: %w", err)
	}
	if len(albums) == 0 {
		log.Printf("WARNING: No albums found in database, skipping track seed")
		return nil
	}
	log.Printf("Found %d albums, creating tracks...", len(albums))

	// Get genres
	var genres []models.Genre
	if err := DB.Find(&genres).Error; err != nil {
		log.Printf("ERROR: Failed to query genres: %v", err)
		return fmt.Errorf("failed to query genres: %w", err)
	}
	if len(genres) == 0 {
		log.Printf("WARNING: No genres found in database, skipping track seed")
		return nil
	}
	log.Printf("Found %d genres for track assignment", len(genres))

	// Create genre map for easy lookup
	genreMap := make(map[string]models.Genre)
	for _, genre := range genres {
		genreMap[genre.Name] = genre
		log.Printf("  Genre mapped: %s (ID: %d)", genre.Name, genre.ID)
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
		// Земфира - Жить в твоей голове
		{"Жить в твоей голове", "Крым", 245, 1, []string{"Рок"}, "/preview/1.jpg"},
		{"Жить в твоей голове", "Ок", 198, 2, []string{"Рок", "Поп"}, "/preview/2.jpg"},
		{"Жить в твоей голове", "Жить в твоей голове", 223, 3, []string{"Рок"}, ""},
		{"Жить в твоей голове", "Мы разбиваемся", 267, 4, []string{"Рок"}, ""},
		{"Жить в твоей голове", "Таблетки", 201, 5, []string{"Рок"}, ""},
		{"Жить в твоей голове", "Без шансов", 189, 6, []string{"Рок"}, ""},
		{"Жить в твоей голове", "Абъюза", 215, 7, []string{"Рок", "Поп"}, "/preview/3.jpg"},
		{"Жить в твоей голове", "Ракеты", 192, 8, []string{"Рок"}, ""},
		{"Жить в твоей голове", "Снег идёт", 204, 9, []string{"Рок", "Поп"}, ""},

		// Zivert - Vinyl #1
		{"Vinyl #1", "Life", 201, 1, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #1", "Credo", 200, 2, []string{"Поп"}, ""},
		{"Vinyl #1", "ЯТЛ", 195, 3, []string{"Поп"}, ""},
		{"Vinyl #1", "Ещё хочу", 198, 4, []string{"Поп", "Электронная"}, "/preview/4.jpg"},
		{"Vinyl #1", "Анечка", 203, 5, []string{"Поп", "Рок"}, "/preview/5.jpg"},
		{"Vinyl #1", "Fly", 197, 6, []string{"Поп", "Электронная"}, "/preview/6.jpg"},
		{"Vinyl #1", "Чак", 189, 7, []string{"Поп"}, ""},
		{"Vinyl #1", "Beverly Hills", 192, 8, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #1", "Океанами стали", 205, 9, []string{"Поп"}, ""},

		// Oxxxymiron - Горгород
		{"Горгород", "Где нас нет", 240, 1, []string{"Хип-хоп"}, ""},
		{"Горгород", "Город под подошвой", 267, 2, []string{"Хип-хоп"}, ""},
		{"Горгород", "Не с начала", 251, 3, []string{"Хип-хоп", "Электронная"}, ""},
		{"Горгород", "Переплетено", 234, 4, []string{"Хип-хоп"}, ""},
		{"Горгород", "Дежавю", 228, 5, []string{"Хип-хоп"}, "/preview/7.jpg"},
		{"Горгород", "Лондон против всех", 245, 6, []string{"Хип-хоп"}, ""},
		{"Горгород", "Полигон", 239, 7, []string{"Хип-хоп"}, ""},
		{"Горгород", "Колыбельная", 223, 8, []string{"Хип-хоп"}, ""},
		{"Горгород", "Где нас нет (ремикс)", 256, 9, []string{"Хип-хоп"}, ""},

		// IC3PEAK - До свидания
		{"До свидания", "Смерти больше нет", 198, 1, []string{"Электронная", "Поп"}, ""},
		{"До свидания", "Сказка", 215, 2, []string{"Электронная", "Поп"}, ""},
		{"До свидания", "Марш", 189, 3, []string{"Электронная"}, ""},
		{"До свидания", "Грустная сука", 203, 4, []string{"Электронная", "Поп"}, ""},
		{"До свидания", "Слёзы", 192, 5, []string{"Электронная", "Рок"}, ""},
		{"До свидания", "Кости", 207, 6, []string{"Электронная", "Поп"}, ""},
		{"До свидания", "До свидания", 221, 7, []string{"Электронная", "Поп"}, ""},
		{"До свидания", "Плакать", 195, 8, []string{"Электронная", "Поп"}, "/preview/8.jpg"},
		{"До свидания", "Трипута", 201, 9, []string{"Электронная"}, ""},
		{"До свидания", "Это не любовь", 189, 10, []string{"Электронная"}, ""},

		// Монеточка - Раскраски для взрослых
		{"Раскраски для взрослых", "Каждый раз", 198, 1, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "Нимфоманка", 203, 2, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "90", 195, 3, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "Последняя дискотека", 201, 4, []string{"Поп", "Электронная"}, ""},
		{"Раскраски для взрослых", "Раскраски для взрослых", 207, 5, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "Уйди, но останься", 192, 6, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "Песенка", 189, 7, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "Крошка", 205, 8, []string{"Поп"}, ""},
		{"Раскраски для взрослых", "Люди", 198, 9, []string{"Поп"}, "/preview/9.jpg"},
	}

	// Create tracks and assign genres
	createdTracks := 0
	existingTracks := 0
	skippedTracks := 0
	trackGenreAssignments := 0
	trackGenreErrors := 0

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
			log.Printf("  WARNING: Album '%s' not found, skipping track '%s'", trackData.AlbumTitle, trackData.Title)
			skippedTracks++
			continue // Skip if album not found
		}

		// Check if track already exists, if not create it
		var track models.Track
		trackExists := DB.Where("album_id = ? AND title = ?", album.ID, trackData.Title).First(&track).Error == nil
		
		if !trackExists {
			// Track doesn't exist, create it
			track = models.Track{
				AlbumID:        album.ID,
				Title:          trackData.Title,
				Duration:       &trackData.Duration,
				TrackNumber:    &trackData.TrackNum,
				CoverImagePath: trackData.CoverImagePath,
			}

			if err := DB.Create(&track).Error; err != nil {
				log.Printf("ERROR: Failed to create track %s: %v", trackData.Title, err)
				skippedTracks++
				continue
			}
			createdTracks++
			log.Printf("  ✓ Created track: %s (ID: %d, AlbumID: %d)", trackData.Title, track.ID, album.ID)
		} else {
			existingTracks++
		}

		// Assign multiple genres - use Replace to avoid duplicates
		var trackGenres []models.Genre
		for _, genreName := range trackData.GenreNames {
			if genre, exists := genreMap[genreName]; exists {
				// Check for duplicates in trackGenres
				duplicate := false
				for _, g := range trackGenres {
					if g.ID == genre.ID {
						duplicate = true
						break
					}
				}
				if !duplicate {
					trackGenres = append(trackGenres, genre)
				}
			} else {
				log.Printf("  WARNING: Genre '%s' not found for track '%s'", genreName, trackData.Title)
			}
		}

		if len(trackGenres) > 0 {
			// Use Replace instead of Append to avoid duplicates for existing tracks
			if err := DB.Model(&track).Association("Genres").Replace(trackGenres); err != nil {
				log.Printf("ERROR: Failed to assign genres to track %s: %v", trackData.Title, err)
				trackGenreErrors++
			} else {
				trackGenreAssignments++
			}
		}
	}

	log.Printf("Tracks seeding complete: %d created, %d already existed, %d skipped", createdTracks, existingTracks, skippedTracks)
	log.Printf("Track genre assignments: %d successful, %d errors", trackGenreAssignments, trackGenreErrors)
	return nil
}

// seedTrackLikes seeds track likes for testing
func seedTrackLikes() error {
	log.Println("Seeding track likes...")

	// Get all test users
	var allTestUsers []models.User
	if err := DB.Find(&allTestUsers).Error; err != nil {
		log.Printf("ERROR: Failed to query users: %v", err)
		return fmt.Errorf("failed to query users: %w", err)
	}
	if len(allTestUsers) == 0 {
		log.Printf("WARNING: No users found in database, skipping track likes seed")
		return nil
	}
	log.Printf("Found %d users for track likes", len(allTestUsers))

	// Get all tracks with their albums to distribute likes across different artists
	var tracks []models.Track
	if err := DB.Preload("Album").Find(&tracks).Error; err != nil {
		log.Printf("ERROR: Failed to query tracks: %v", err)
		return fmt.Errorf("failed to query tracks: %w", err)
	}
	if len(tracks) == 0 {
		log.Printf("WARNING: No tracks found in database, skipping track likes seed")
		return nil
	}
	log.Printf("Found %d tracks for track likes", len(tracks))

	// Group tracks by album/artist to ensure diversity
	tracksByArtist := make(map[string][]models.Track)
	for _, track := range tracks {
		if track.Album.Artist != "" {
			tracksByArtist[track.Album.Artist] = append(tracksByArtist[track.Album.Artist], track)
		}
	}

	// Get current time for setting created_at within last 24 hours
	now := time.Now()
	hoursAgo := 0

	// Seed likes: distribute across different artists, ensuring diversity
	trackLikes := []models.TrackLike{}
	trackIndex := 0

	// Process artists in order to ensure diversity
	artistNames := make([]string, 0, len(tracksByArtist))
	for artist := range tracksByArtist {
		artistNames = append(artistNames, artist)
	}

	// Select top tracks from each artist to get diverse results
	for _, artistName := range artistNames {
		artistTracks := tracksByArtist[artistName]
		// Take top 2-3 tracks from each artist to ensure diversity
		maxTracksPerArtist := 3
		if len(artistTracks) < maxTracksPerArtist {
			maxTracksPerArtist = len(artistTracks)
		}

		for i := 0; i < maxTracksPerArtist; i++ {
			track := artistTracks[i]
			// Generate number of likes between 5 and 20, prioritizing different artists
			numLikes := 10 + (trackIndex % 11) // Will give 10-20 likes
			if numLikes > len(allTestUsers) {
				numLikes = len(allTestUsers)
			}

			// Create likes from different users
			startIndex := trackIndex % len(allTestUsers)
			for j := 0; j < numLikes; j++ {
				userIndex := (startIndex + j) % len(allTestUsers)
				// Check if like already exists
				var existingLike models.TrackLike
				if err := DB.Where("user_id = ? AND track_id = ?", allTestUsers[userIndex].ID, track.ID).First(&existingLike).Error; err != nil {
					// Create new like with created_at within last 24 hours
					like := models.TrackLike{
						UserID:  allTestUsers[userIndex].ID,
						TrackID: track.ID,
					}
					// Set created_at to be within last 24 hours, distributed over time
					like.CreatedAt = now.Add(-time.Duration(hoursAgo%24) * time.Hour)
					trackLikes = append(trackLikes, like)
					hoursAgo++
				} else {
					// Update existing like's created_at to be within last 24 hours
					existingLike.CreatedAt = now.Add(-time.Duration(hoursAgo%24) * time.Hour)
					if err := DB.Save(&existingLike).Error; err != nil {
						log.Printf("Warning: failed to update track like created_at: %v", err)
					}
					hoursAgo++
				}
			}
			trackIndex++
		}
	}

	// Create all new likes in batch
	createdLikes := 0
	failedLikes := 0
	for _, like := range trackLikes {
		if err := DB.Create(&like).Error; err != nil {
			log.Printf("ERROR: Failed to create track like (UserID: %d, TrackID: %d): %v", like.UserID, like.TrackID, err)
			failedLikes++
		} else {
			createdLikes++
		}
	}

	log.Printf("Track likes seeding complete: %d created, %d failed", createdLikes, failedLikes)
	return nil
}

// seedReviews seeds test reviews into database
func seedReviews() error {
	log.Println("Seeding test reviews...")

	// Get users first (needed for both new and existing reviews)
	var admin, testUser models.User
	if err := DB.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		log.Printf("ERROR: Admin user not found: %v, skipping review seed", err)
		return nil
	}
	log.Printf("Found admin user (ID: %d)", admin.ID)

	if err := DB.Where("email = ?", "test@example.com").First(&testUser).Error; err != nil {
		log.Printf("ERROR: Test user not found: %v, skipping review seed", err)
		return nil
	}
	log.Printf("Found test user (ID: %d)", testUser.ID)

	// Get albums
	var albums []models.Album
	if err := DB.Find(&albums).Error; err != nil {
		log.Printf("ERROR: Failed to query albums: %v", err)
		return fmt.Errorf("failed to query albums: %w", err)
	}
	if len(albums) == 0 {
		log.Printf("WARNING: No albums found in database, skipping review seed")
		return nil
	}
	log.Printf("Found %d albums for reviews", len(albums))

	// Check if reviews already exist
	var reviewCount int64
	DB.Model(&models.Review{}).Count(&reviewCount)
	reviewsExist := reviewCount > 0
	log.Printf("Current review count in database: %d", reviewCount)

	// Helper function to convert atmosphere rating (1-10) to multiplier (1.0000-1.6072)
	convertAtmosphereToMultiplier := func(rating int) float64 {
		step := 0.6072 / 9.0
		return 1.0000 + float64(rating-1)*step
	}

	// Only create new reviews if they don't exist
	var allReviews []models.Review
	createdReviews := 0
	failedReviews := 0
	if !reviewsExist {
		log.Println("No reviews found, creating new reviews...")
		// Create test reviews (using atmosphere ratings 1-10, converted to multiplier)
		// Альбомы: [0] Жить в твоей голове, [1] Vinyl #1, [2] Горгород, [3] До свидания, [4] Раскраски для взрослых
		reviews := []models.Review{
			{
				UserID:               testUser.ID,
				AlbumID:              &albums[0].ID, // Жить в твоей голове
				Text:                 "Невероятный альбом! Земфира на пике формы. Каждая композиция - это произведение искусства. Особенно выделяются 'Крым' и 'Ок'. Звукорежиссура безупречна, аранжировки гениальны.",
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
				AlbumID:              &albums[1].ID, // Vinyl #1
				Text:                 "Классика поп-музыки! Zivert создала настоящий шедевр. 'Life' и 'Ещё хочу' - это вечные хиты. Танцевальные ритмы, запоминающиеся мелодии и неповторимый вокал.",
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
				AlbumID:              &albums[2].ID, // Горгород
				Text:                 "Oxxxymiron задал новый стандарт для хип-хопа. Битмейкинг на высшем уровне, семплы подобраны идеально. 'Где нас нет' - это гимн эпохи. Альбом звучит свежо даже спустя годы.",
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
				AlbumID:              &albums[3].ID, // До свидания
				Text:                 "IC3PEAK создали что-то особенное. Смешение электроники, индастриала и трип-хопа работает идеально. 'Смерти больше нет' - это мощный хит, который не надоедает. Альбом звучит как путешествие в темную сторону.",
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
				AlbumID:              &albums[4].ID, // Раскраски для взрослых
				Text:                 "Абсолютный шедевр поп-музыки. Монеточка на этом альбоме создала что-то трансцендентное. Искренность и ирония в их лучшем проявлении. Каждая песня на своем месте. 'Каждый раз' - это икона жанра.",
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
				AlbumID:              &albums[0].ID, // Жить в твоей голове (вторая рецензия)
				Text:                 "Хочу добавить еще несколько слов об этом альбоме. 'Снег идёт' - одна из самых красивых песен в истории русской музыки. Земфира показала свой талант композитора.",
				RatingRhymes:         8,
				RatingStructure:      9,
				RatingImplementation: 9,
				RatingIndividuality:  8,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(6), // 6/10 -> ~1.3388
				Status:               models.ReviewStatusPending,
			},
			{
				UserID:               testUser.ID,
				AlbumID:              &albums[1].ID, // Vinyl #1 (только оценка, без текста)
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
				log.Printf("ERROR: Failed to create review %d: %v", i+1, err)
				failedReviews++
				return fmt.Errorf("failed to seed review %d: %w", i+1, err)
			}
			createdReviews++
			log.Printf("  ✓ Created review %d (ID: %d, UserID: %d, AlbumID: %v)", i+1, reviews[i].ID, reviews[i].UserID, reviews[i].AlbumID)
		}

		// Get some tracks for track reviews
		var seedTracks []models.Track
		if err := DB.Limit(5).Find(&seedTracks).Error; err == nil && len(seedTracks) > 0 {
			// Add some track reviews
			trackReviews := []models.Review{
				{
					UserID:               testUser.ID,
					TrackID:              &seedTracks[0].ID, // First track
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
					TrackID:              &seedTracks[1].ID, // Second track
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
					TrackID:              &seedTracks[2].ID, // Third track
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
					log.Printf("ERROR: Failed to create track review %d: %v", i+1, err)
					failedReviews++
				} else {
					createdReviews++
					log.Printf("  ✓ Created track review %d (ID: %d, TrackID: %v)", i+1, trackReviews[i].ID, trackReviews[i].TrackID)
				}
			}
		}

		// Get all test users (need to reload after creation)
		var allTestUsersForReviews []models.User
		if err := DB.Find(&allTestUsersForReviews).Error; err != nil {
			log.Printf("Warning: failed to fetch users for additional reviews: %v", err)
			allTestUsersForReviews = []models.User{admin, testUser}
		}

		// Добавляем больше рецензий для разных альбомов от разных пользователей
		// Альбомы: [0] Жить в твоей голове, [1] Vinyl #1, [2] Горгород, [3] До свидания, [4] Раскраски для взрослых
		if len(allTestUsersForReviews) >= 3 && len(albums) >= 5 {
			additionalReviews := []models.Review{
				{
					UserID:               allTestUsersForReviews[2].ID, // musiclover1
					AlbumID:              &albums[0].ID,                // Жить в твоей голове
					Text:                 "Классика рока! Земфира показывает своё мастерство на этом альбоме.",
					RatingRhymes:         9,
					RatingStructure:      9,
					RatingImplementation: 9,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				{
					UserID:               allTestUsersForReviews[3].ID, // musiclover2
					AlbumID:              &albums[1].ID,                // Vinyl #1
					Text:                 "Великолепный альбом Zivert! 'Vinyl #1' - это легенда.",
					RatingRhymes:         9,
					RatingStructure:      10,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(10),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				{
					UserID:               allTestUsersForReviews[4].ID, // musiclover3
					AlbumID:              &albums[2].ID,                // Горгород
					Text:                 "Oxxxymiron создал шедевр хип-хопа. Битмейкинг вне времени.",
					RatingRhymes:         10,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				{
					UserID:               allTestUsersForReviews[5].ID, // musiclover4
					AlbumID:              &albums[3].ID,                // До свидания
					Text:                 "IC3PEAK на высоте! Электроника встречается с индастриалом.",
					RatingRhymes:         8,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  8,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(10),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				{
					UserID:               allTestUsersForReviews[6].ID, // musiclover5
					AlbumID:              &albums[4].ID,                // Раскраски для взрослых
					Text:                 "Монеточка создала поп-шедевр. Каждая композиция - произведение искусства.",
					RatingRhymes:         10,
					RatingStructure:      10,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(10),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				{
					UserID:               allTestUsersForReviews[7].ID, // musiclover6
					AlbumID:              &albums[0].ID,                // Жить в твоей голове (третья рецензия)
					Text:                 "Альбом на все времена. Земфира в лучшей форме.",
					RatingRhymes:         10,
					RatingStructure:      10,
					RatingImplementation: 10,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
			}

			// Calculate final scores and create additional reviews
			for i := range additionalReviews {
				additionalReviews[i].CalculateFinalScore()
				if err := DB.Create(&additionalReviews[i]).Error; err != nil {
					log.Printf("ERROR: Failed to create additional review %d: %v", i+1, err)
					failedReviews++
				} else {
					createdReviews++
					log.Printf("  ✓ Created additional review %d (ID: %d)", i+1, additionalReviews[i].ID)
				}
			}
		}
		log.Printf("Reviews creation complete: %d created, %d failed", createdReviews, failedReviews)
	} else {
		log.Println("Reviews already exist, skipping creation")
	} // End of if !reviewsExist

	// Reload all reviews from DB to get correct IDs (including newly created ones)
	// This is done regardless of whether reviews existed before
	if err := DB.Where("status = ?", models.ReviewStatusApproved).Find(&allReviews).Error; err != nil {
		log.Printf("Warning: failed to reload reviews for likes: %v", err)
		// If we can't load reviews, we can't proceed with likes
		if len(allReviews) == 0 {
			log.Println("No reviews found, skipping review likes seed")
			return nil
		}
	}

	// Update created_at for approved reviews to be within last 24 hours (for popular reviews to show)
	now := time.Now()
	for i := range allReviews {
		if allReviews[i].Status == models.ReviewStatusApproved && allReviews[i].ID > 0 {
			// Set created_at to be within last 24 hours, distributed over time
			hoursAgo := i % 24 // Distribute over 24 hours
			newCreatedAt := now.Add(-time.Duration(hoursAgo) * time.Hour)
			// Use Update with Where to ensure the update happens
			if err := DB.Model(&models.Review{}).Where("id = ?", allReviews[i].ID).Update("created_at", newCreatedAt).Error; err != nil {
				log.Printf("Warning: failed to update review created_at for review %d: %v", allReviews[i].ID, err)
			}
		}
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

	// Update track average ratings
	var allTracks []models.Track
	if err := DB.Find(&allTracks).Error; err == nil {
		for _, track := range allTracks {
			var trackReviews []models.Review
			if err := DB.Where("track_id = ? AND status = ?", track.ID, models.ReviewStatusApproved).Find(&trackReviews).Error; err == nil && len(trackReviews) > 0 {
				var totalScore float64
				for _, review := range trackReviews {
					totalScore += review.FinalScore
				}
				averageRating := totalScore / float64(len(trackReviews))
				// Round to nearest integer
				roundedAverage := float64(int(averageRating + 0.5))
				DB.Model(&track).Update("average_rating", roundedAverage)
			}
		}
	}

	// Get all test users for likes
	var allTestUsers []models.User
	if err := DB.Find(&allTestUsers).Error; err != nil {
		log.Printf("Warning: failed to fetch users for review likes: %v", err)
		allTestUsers = []models.User{admin, testUser} // Fallback to basic users
	}

	// Seed review likes for testing - create 5-20 likes per review, распределённые по разным пользователям
	// Update created_at for review likes to be within last 24 hours
	nowForLikes := time.Now()
	likeHoursAgo := 0

	var reviewLikes []models.ReviewLike
	for i, review := range allReviews {
		if review.Status == models.ReviewStatusApproved && review.ID > 0 && review.AlbumID != nil {
			// Generate number of likes between 5 and 20, распределяем чтобы первые рецензии имели больше лайков
			numLikes := 5 + (i % 16) // Will give 5-20 likes
			if numLikes > len(allTestUsers) {
				numLikes = len(allTestUsers)
			}

			// Create likes from different users, используем циклическое распределение для разнообразия
			startIndex := i % len(allTestUsers) // Начинаем с разных пользователей для каждой рецензии
			for j := 0; j < numLikes; j++ {
				userIndex := (startIndex + j) % len(allTestUsers)
				// Проверяем что пользователь не автор рецензии
				if allTestUsers[userIndex].ID == review.UserID {
					continue
				}
				// Check if like already exists
				var existingLike models.ReviewLike
				if err := DB.Where("user_id = ? AND review_id = ?", allTestUsers[userIndex].ID, review.ID).First(&existingLike).Error; err != nil {
					// Create new like with created_at within last 24 hours
					like := models.ReviewLike{
						UserID:   allTestUsers[userIndex].ID,
						ReviewID: review.ID,
					}
					// Set created_at to be within last 24 hours
					like.CreatedAt = nowForLikes.Add(-time.Duration(likeHoursAgo%24) * time.Hour)
					reviewLikes = append(reviewLikes, like)
					likeHoursAgo++
				} else {
					// Update existing like's created_at to be within last 24 hours
					existingLike.CreatedAt = nowForLikes.Add(-time.Duration(likeHoursAgo%24) * time.Hour)
					if err := DB.Save(&existingLike).Error; err != nil {
						log.Printf("Warning: failed to update review like created_at: %v", err)
					}
					likeHoursAgo++
				}
			}
		}
	}

	createdReviewLikes := 0
	failedReviewLikes := 0
	for _, like := range reviewLikes {
		if err := DB.Create(&like).Error; err != nil {
			log.Printf("ERROR: Failed to create review like (UserID: %d, ReviewID: %d): %v", like.UserID, like.ReviewID, err)
			failedReviewLikes++
		} else {
			createdReviewLikes++
		}
	}

	log.Printf("Review likes seeding complete: %d created, %d failed", createdReviewLikes, failedReviewLikes)
	log.Printf("Reviews seeding summary: %d reviews created, %d review likes created", createdReviews, createdReviewLikes)
	return nil
}

// updateAlbumCoverImages updates cover_image_path for existing albums
func updateAlbumCoverImages() error {
	albumMap := map[string]string{
		"Жить в твоей голове":   "/preview/1.jpg",
		"Vinyl #1":               "/preview/4.jpg",
		"Горгород":               "/preview/7.jpg",
		"До свидания":            "/preview/8.jpg",
		"Раскраски для взрослых": "/preview/9.jpg",
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

// LogDatabaseState logs the current state of all database tables
// This function can be called externally for diagnostics
func LogDatabaseState() {
	logDatabaseState()
}

// logDatabaseState logs the current state of all database tables
func logDatabaseState() {
	if DB == nil {
		log.Println("ERROR: Database connection is nil")
		return
	}

	var counts struct {
		Users       int64
		Genres      int64
		Albums      int64
		Tracks      int64
		TrackGenres int64
		Reviews     int64
		ReviewLikes int64
		TrackLikes  int64
		AlbumLikes  int64
	}

	// Count records in each table
	DB.Model(&models.User{}).Count(&counts.Users)
	DB.Model(&models.Genre{}).Count(&counts.Genres)
	DB.Model(&models.Album{}).Count(&counts.Albums)
	DB.Model(&models.Track{}).Count(&counts.Tracks)
	DB.Model(&models.TrackGenre{}).Count(&counts.TrackGenres)
	DB.Model(&models.Review{}).Count(&counts.Reviews)
	DB.Model(&models.ReviewLike{}).Count(&counts.ReviewLikes)
	DB.Model(&models.TrackLike{}).Count(&counts.TrackLikes)
	DB.Model(&models.AlbumLike{}).Count(&counts.AlbumLikes)

	log.Printf("📊 Database Statistics:")
	log.Printf("   Users:       %d", counts.Users)
	log.Printf("   Genres:      %d", counts.Genres)
	log.Printf("   Albums:      %d", counts.Albums)
	log.Printf("   Tracks:      %d", counts.Tracks)
	log.Printf("   TrackGenres: %d", counts.TrackGenres)
	log.Printf("   Reviews:     %d", counts.Reviews)
	log.Printf("   ReviewLikes: %d", counts.ReviewLikes)
	log.Printf("   TrackLikes:  %d", counts.TrackLikes)
	log.Printf("   AlbumLikes:  %d", counts.AlbumLikes)

	// Check for empty tables
	emptyTables := []string{}
	if counts.Users == 0 {
		emptyTables = append(emptyTables, "users")
	}
	if counts.Genres == 0 {
		emptyTables = append(emptyTables, "genres")
	}
	if counts.Albums == 0 {
		emptyTables = append(emptyTables, "albums")
	}
	if counts.Tracks == 0 {
		emptyTables = append(emptyTables, "tracks")
	}
	if counts.TrackGenres == 0 {
		emptyTables = append(emptyTables, "track_genres")
	}
	if counts.Reviews == 0 {
		emptyTables = append(emptyTables, "reviews")
	}
	if counts.ReviewLikes == 0 {
		emptyTables = append(emptyTables, "review_likes")
	}
	if counts.TrackLikes == 0 {
		emptyTables = append(emptyTables, "track_likes")
	}
	if counts.AlbumLikes == 0 {
		emptyTables = append(emptyTables, "album_likes")
	}

	if len(emptyTables) > 0 {
		log.Printf("⚠️  WARNING: Empty tables detected: %v", emptyTables)
	} else {
		log.Println("✓ All tables contain data")
	}

	// Additional detailed checks
	if counts.Albums > 0 {
		var albums []models.Album
		DB.Find(&albums)
		log.Printf("   Album details: %d albums found", len(albums))
		for i, album := range albums {
			if i < 5 { // Show first 5
				log.Printf("      - [%d] %s by %s (GenreID: %d)", album.ID, album.Title, album.Artist, album.GenreID)
			}
		}
	}

	if counts.Tracks > 0 {
		var tracks []models.Track
		DB.Preload("Album").Find(&tracks)
		log.Printf("   Track details: %d tracks found", len(tracks))
		for i, track := range tracks {
			if i < 5 { // Show first 5
				albumTitle := "N/A"
				if track.Album.ID > 0 {
					albumTitle = track.Album.Title
				}
				log.Printf("      - [%d] %s (AlbumID: %d, Album: %s)", track.ID, track.Title, track.AlbumID, albumTitle)
			}
		}
	}

	if counts.Genres > 0 {
		var genres []models.Genre
		DB.Find(&genres)
		log.Printf("   Genre details: %d genres found", len(genres))
		for _, genre := range genres {
			log.Printf("      - [%d] %s", genre.ID, genre.Name)
		}
	}
}
