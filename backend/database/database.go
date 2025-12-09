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

	// Seed album likes (for testing)
	if err := seedAlbumLikes(); err != nil {
		log.Printf("ERROR: failed to seed album likes: %v", err)
	} else {
		log.Println("✓ Album likes seeding completed successfully")
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

	// Check if genres already exist in sufficient quantity (15 жанров)
	var existingGenreCount int64
	DB.Model(&models.Genre{}).Count(&existingGenreCount)
	if existingGenreCount >= 15 {
		log.Printf("Genres already exist (%d genres), skipping genre seed to avoid duplicates", existingGenreCount)
		// Still need to reload genres for album creation
	} else {
		// Seed genres - 15 самых популярных жанров в РФ за последние годы
		genresToCreate := []models.Genre{
			{Name: "Поп", Description: "Поп-музыка"},
			{Name: "Рэп", Description: "Рэп"},
			{Name: "Хип-хоп", Description: "Хип-хоп"},
			{Name: "Рок", Description: "Рок-музыка"},
			{Name: "Электронная", Description: "Электронная музыка"},
			{Name: "Поп-рок", Description: "Поп-рок"},
			{Name: "Инди-поп", Description: "Инди-поп"},
			{Name: "Альтернативный рок", Description: "Альтернативный рок"},
			{Name: "R&B", Description: "R&B"},
			{Name: "Соул", Description: "Соул"},
			{Name: "Трэп", Description: "Трэп"},
			{Name: "Дрилл", Description: "Дрилл"},
			{Name: "Фолк", Description: "Фолк"},
			{Name: "Шансон", Description: "Шансон"},
			{Name: "Метал", Description: "Метал"},
		}

		// Create genres if they don't exist (use FirstOrCreate to avoid duplicates)
		createdGenres := 0
		existingGenres := 0
		for _, genre := range genresToCreate {
			var existingGenre models.Genre
			result := DB.Where("name = ?", genre.Name).FirstOrCreate(&existingGenre, genre)
			if result.Error != nil {
				log.Printf("ERROR: Failed to create/find genre %s: %v", genre.Name, result.Error)
				return fmt.Errorf("failed to seed genre %s: %w", genre.Name, result.Error)
			}
			if result.RowsAffected > 0 {
				createdGenres++
				log.Printf("✓ Created genre: %s (ID: %d)", existingGenre.Name, existingGenre.ID)
			} else {
				existingGenres++
				log.Printf("  Genre already exists: %s (ID: %d)", existingGenre.Name, existingGenre.ID)
			}
		}
		log.Printf("Genres: %d created, %d already existed", createdGenres, existingGenres)
	}

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

	// Check if albums already exist in sufficient quantity
	var existingAlbumCount int64
	DB.Model(&models.Album{}).Count(&existingAlbumCount)
	if existingAlbumCount >= 12 {
		log.Printf("Albums already exist (%d albums), skipping album seed to avoid duplicates", existingAlbumCount)
		// Still need to reload albums for likes
	} else {
		// Seed albums - verify genre IDs before using them
		rockGenre, rockExists := genreMap["Рок"]
		popGenre, popExists := genreMap["Поп"]
		hiphopGenre, hiphopExists := genreMap["Хип-хоп"]
		electronicGenre, electronicExists := genreMap["Электронная"]

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

	// Helper function to create time pointer
	createDate := func(year int, month time.Month, day int) *time.Time {
		t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		return &t
	}

	albums := []models.Album{
		// Баста (Basta / Ноггано) - Хип-хоп
		{Title: "Баста 1", Artist: "Баста", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/basta1.jpg", Description: "Первый студийный альбом Басты", ReleaseDate: createDate(2006, 1, 1), AverageRating: 0},
		{Title: "Баста 2", Artist: "Баста", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/basta2.jpg", Description: "Второй студийный альбом Басты", ReleaseDate: createDate(2007, 1, 1), AverageRating: 0},
		{Title: "Ноггано", Artist: "Баста", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/noggano.jpg", Description: "Альбом под псевдонимом Ноггано", ReleaseDate: createDate(2008, 1, 1), AverageRating: 0},
		{Title: "Баста 3", Artist: "Баста", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/basta3.jpg", Description: "Третий студийный альбом Басты", ReleaseDate: createDate(2010, 1, 1), AverageRating: 0},
		
		// Скриптонит (Scriptonite) - Хип-хоп
		{Title: "Дом с нормальными явлениями", Artist: "Скриптонит", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/domsnormyavleniyami.jpg", Description: "Дебютный альбом Скриптонита", ReleaseDate: createDate(2015, 1, 1), AverageRating: 0},
		{Title: "Праздник на улице 36", Artist: "Скриптонит", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/prazdnikulica36.jpg", Description: "Второй альбом Скриптонита", ReleaseDate: createDate(2017, 1, 1), AverageRating: 0},
		{Title: "2004", Artist: "Скриптонит", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/2004.jpg", Description: "Третий альбом Скриптонита", ReleaseDate: createDate(2018, 1, 1), AverageRating: 0},
		{Title: "Уроборос: улочка и аллея", Artist: "Скриптонит & 104", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/uroboros.jpg", Description: "Альбом Скриптонита & 104", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
		
		// ANNA ASTI - Поп
		{Title: "Феникс", Artist: "ANNA ASTI", GenreID: popGenre.ID, CoverImagePath: "/preview/fenix.png", Description: "Дебютный альбом ANNA ASTI", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
		{Title: "Царица", Artist: "ANNA ASTI", GenreID: popGenre.ID, CoverImagePath: "/preview/carica.png", Description: "Второй альбом ANNA ASTI", ReleaseDate: createDate(2023, 1, 1), AverageRating: 0},
		
		// Zivert - Поп
		{Title: "Vinyl #1", Artist: "Zivert", GenreID: popGenre.ID, CoverImagePath: "/preview/venil1.jpg", Description: "Дебютный альбом Zivert", ReleaseDate: createDate(2018, 1, 1), AverageRating: 0},
		{Title: "Vinyl #2", Artist: "Zivert", GenreID: popGenre.ID, CoverImagePath: "/preview/venil2.jpg", Description: "Второй альбом Zivert", ReleaseDate: createDate(2019, 1, 1), AverageRating: 0},
		{Title: "Сияй", Artist: "Zivert", GenreID: popGenre.ID, CoverImagePath: "/preview/siyai.jpg", Description: "Третий альбом Zivert", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
		
		// IOWA - Поп
		{Title: "Import", Artist: "IOWA", GenreID: popGenre.ID, CoverImagePath: "/preview/import.jpg", Description: "Первый альбом IOWA", ReleaseDate: createDate(2012, 1, 1), AverageRating: 0},
		{Title: "Export", Artist: "IOWA", GenreID: popGenre.ID, CoverImagePath: "/preview/export.jpg", Description: "Второй альбом IOWA", ReleaseDate: createDate(2015, 1, 1), AverageRating: 0},
		{Title: "Французский альбом", Artist: "IOWA", GenreID: popGenre.ID, CoverImagePath: "/preview/french.jpg", Description: "Третий альбом IOWA", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
		
		// Клава Кока (Klava Koka) - Поп
		{Title: "Неприлично о личном", Artist: "Клава Кока", GenreID: popGenre.ID, CoverImagePath: "/preview/neprelichnoolicnom.jpg", Description: "Дебютный альбом Клавы Коки", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
		{Title: "Красное вино", Artist: "Клава Кока", GenreID: popGenre.ID, CoverImagePath: "/preview/krasnoevino.jpg", Description: "Второй альбом Клавы Коки", ReleaseDate: createDate(2024, 1, 1), AverageRating: 0},
		
		// ЛСП (LSP) - Хип-хоп
		{Title: "Magic City", Artist: "ЛСП", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/magiccity.jpg", Description: "Первый альбом ЛСП", ReleaseDate: createDate(2015, 1, 1), AverageRating: 0},
		{Title: "Tragic City", Artist: "ЛСП", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/tragiccity.jpg", Description: "Второй альбом ЛСП", ReleaseDate: createDate(2017, 1, 1), AverageRating: 0},
		{Title: "SAD SOUNDS", Artist: "ЛСП", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/sadsounds.png", Description: "Третий альбом ЛСП", ReleaseDate: createDate(2020, 1, 1), AverageRating: 0},
		
		// The Hatters - Рок/Инди
		{Title: "Безумие", Artist: "The Hatters", GenreID: rockGenre.ID, CoverImagePath: "/preview/bezumie.jpg", Description: "Первый альбом The Hatters", ReleaseDate: createDate(2016, 1, 1), AverageRating: 0},
		{Title: "Третий", Artist: "The Hatters", GenreID: rockGenre.ID, CoverImagePath: "/preview/tretiy.jpg", Description: "Третий альбом The Hatters", ReleaseDate: createDate(2018, 1, 1), AverageRating: 0},
		{Title: "Четвёртый", Artist: "The Hatters", GenreID: rockGenre.ID, CoverImagePath: "/preview/chetvertiy.jpg", Description: "Четвёртый альбом The Hatters", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
		
		// Miyagi (Miyagi & Эндшпиль / Miyagi & Andy Panda) - Хип-хоп
		{Title: "Hajime 1", Artist: "Miyagi & Эндшпиль", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/hajime1.jpg", Description: "Первый альбом Miyagi & Эндшпиль", ReleaseDate: createDate(2016, 1, 1), AverageRating: 0},
		{Title: "Buster Keaton", Artist: "Miyagi & Andy Panda", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/BusterKeaton.jpg", Description: "Альбом Miyagi & Andy Panda", ReleaseDate: createDate(2018, 1, 1), AverageRating: 0},
		{Title: "Yamakasi", Artist: "Miyagi & Andy Panda", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/Yamakasi.jpg", Description: "Альбом Miyagi & Andy Panda", ReleaseDate: createDate(2020, 1, 1), AverageRating: 0},
		{Title: "Million Dollars: Happiness", Artist: "Miyagi & Andy Panda", GenreID: hiphopGenre.ID, CoverImagePath: "/preview/MillionDollars.jpg", Description: "Альбом Miyagi & Andy Panda", ReleaseDate: createDate(2021, 1, 1), AverageRating: 0},
	}

	// Seed albums - create or update with cover images
	albumMap := map[string]string{
		"Баста 1": "/preview/basta1.jpg",
		"Баста 2": "/preview/basta2.jpg",
		"Ноггано": "/preview/noggano.jpg",
		"Баста 3": "/preview/basta3.jpg",
		"Дом с нормальными явлениями": "/preview/domsnormyavleniyami.jpg",
		"Праздник на улице 36": "/preview/prazdnikulica36.jpg",
		"2004": "/preview/2004.jpg",
		"Уроборос: улочка и аллея": "/preview/uroboros.jpg",
		"Феникс": "/preview/fenix.png",
		"Царица": "/preview/carica.png",
		"Vinyl #1": "/preview/venil1.jpg",
		"Vinyl #2": "/preview/venil2.jpg",
		"Сияй": "/preview/siyai.jpg",
		"Import": "/preview/import.jpg",
		"Export": "/preview/export.jpg",
		"Французский альбом": "/preview/french.jpg",
		"Неприлично о личном": "/preview/neprelichnoolicnom.jpg",
		"Красное вино": "/preview/krasnoevino.jpg",
		"Magic City": "/preview/magiccity.jpg",
		"Tragic City": "/preview/tragiccity.jpg",
		"SAD SOUNDS": "/preview/sadsounds.png",
		"Безумие": "/preview/bezumie.jpg",
		"Третий": "/preview/tretiy.jpg",
		"Четвёртый": "/preview/chetvertiy.jpg",
		"Hajime 1": "/preview/hajime1.jpg",
		"Buster Keaton": "/preview/BusterKeaton.jpg",
		"Yamakasi": "/preview/Yamakasi.jpg",
		"Million Dollars: Happiness": "/preview/MillionDollars.jpg",
	}

	createdAlbums := 0
	existingAlbums := 0
	skippedAlbums := 0
	for _, album := range albums {
		// Verify genre ID is valid before creating
		if album.GenreID == 0 {
			log.Printf("ERROR: Album %s has invalid GenreID (0), skipping", album.Title)
			skippedAlbums++
			continue
		}
		
		var existingAlbum models.Album
		result := DB.Where("title = ? AND artist = ?", album.Title, album.Artist).FirstOrCreate(&existingAlbum, album)
		if result.Error != nil {
			log.Printf("ERROR: Failed to create/find album %s: %v", album.Title, result.Error)
			skippedAlbums++
			continue
		}
		
		if result.RowsAffected > 0 {
			// Album was created
			createdAlbums++
			log.Printf("✓ Created album: %s by %s (ID: %d, GenreID: %d)", album.Title, album.Artist, existingAlbum.ID, existingAlbum.GenreID)
		} else {
			// Album already exists, update cover_image_path if it's empty
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
		}
	}
		log.Printf("Albums seeding complete: %d created, %d already existed, %d skipped", createdAlbums, existingAlbums, skippedAlbums)
	}

	// Reload albums from DB to get correct IDs
	var allAlbums []models.Album
	if err := DB.Find(&allAlbums).Error; err != nil {
		log.Printf("Warning: failed to reload albums: %v", err)
		allAlbums = []models.Album{} // Fallback to empty slice
	} else {
		log.Printf("Reloaded %d albums from database", len(allAlbums))
	}

	// Album likes are now seeded in seedAlbumLikes() function

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

	// Check if tracks already exist in sufficient quantity
	var existingTrackCount int64
	DB.Model(&models.Track{}).Count(&existingTrackCount)
	if existingTrackCount >= 50 {
		log.Printf("Tracks already exist (%d tracks), skipping track seed to avoid duplicates", existingTrackCount)
		return nil
	}

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
		// Баста - Баста 1 (2006)
		{"Баста 1", "Мой друг", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Наше лето (feat. Гуф)", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Свобода", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Ростов", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Водяной", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Так плачем было (feat. Лигалайз)", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Без тебя", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Мама", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Город дорог", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 1", "Реквием", 242, 10, []string{"Хип-хоп", "Рэп"}, ""},

		// Баста - Баста 2 (2007)
		{"Баста 2", "Intro", 60, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Моя игра", 240, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Осень", 267, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Выпускной (Медлячок)", 251, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Город", 234, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Самурай", 228, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Дождь", 245, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Life", 239, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Снится сон", 223, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 2", "Outro", 50, 10, []string{"Хип-хоп", "Рэп"}, ""},

		// Баста - Ноггано (2008)
		{"Ноггано", "Куба", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Вечный жид", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Родина", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Выпускной", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Водяной", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Ствол", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Рим", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Мама", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Ноггано", "Медлячок (Remix)", 256, 9, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},
		{"Ноггано", "Осень (Remix)", 242, 10, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},

		// Баста - Баста 3 (2010)
		{"Баста 3", "Сансара", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Чёрное солнце", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Выпускной (Баста 3)", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Где я", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Свобода или смерть", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Дым", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Война", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Любовь и страх", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Баста 3", "Мой рок-н-ролл (feat. Смоки Мо)", 256, 9, []string{"Хип-хоп", "Рок", "Поп-рок"}, ""},
		{"Баста 3", "Outro", 50, 10, []string{"Хип-хоп", "Рэп"}, ""},

		// Скриптонит - Дом с нормальными явлениями (2015)
		{"Дом с нормальными явлениями", "Вне игры", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "RBG", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "Мы любим...", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "Экзистенциальная холка", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "Люби меня", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "Право на выбор", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "ПТВ", 239, 7, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},
		{"Дом с нормальными явлениями", "Гастроль", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "Феномен", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "MDM", 242, 10, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},
		{"Дом с нормальными явлениями", "Тем, кто с нами", 250, 11, []string{"Хип-хоп", "Рэп"}, ""},
		{"Дом с нормальными явлениями", "Статистика", 235, 12, []string{"Хип-хоп", "Рэп"}, ""},

		// Скриптонит - Праздник на улице 36 (2017)
		{"Праздник на улице 36", "Время тяжёлое", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Праздник на улице 36", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Стиль", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Личный рай", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Пуля-дура", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Смок", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Слишком сильная любовь", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Кино", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Зеро", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Моя", 242, 10, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "По полной", 250, 11, []string{"Хип-хоп", "Рэп"}, ""},
		{"Праздник на улице 36", "Ливень (Bonus Track)", 235, 12, []string{"Хип-хоп", "Рэп"}, ""},

		// Скриптонит - 2004 (2018)
		{"2004", "2004", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Герой", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Барбисайз", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Нас не видят", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Фурия", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Улица", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Ангел", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Блок", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Физрук", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Твой первый диск", 242, 10, []string{"Хип-хоп", "Рэп"}, ""},
		{"2004", "Неважно", 250, 11, []string{"Хип-хоп", "Рэп"}, ""},

		// Скриптонит & 104 - Уроборос: улочка и аллея (2021)
		{"Уроборос: улочка и аллея", "Улочка", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Аллея", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Девочка с картинки", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Мама, я танцую", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Микрофон", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "До рассвета", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Бассейн", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Кепка", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Давным-давно", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Один", 242, 10, []string{"Хип-хоп", "Рэп"}, ""},
		{"Уроборос: улочка и аллея", "Так и должно быть", 250, 11, []string{"Хип-хоп", "Рэп"}, ""},

		// ANNA ASTI - Феникс (2021)
		{"Феникс", "По барам", 240, 1, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Феникс", 267, 2, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Царица", 251, 3, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Берега", 234, 4, []string{"Поп", "Инди-поп"}, ""},
		{"Феникс", "Гармония", 228, 5, []string{"Поп", "Инди-поп"}, ""},
		{"Феникс", "Дикая", 245, 6, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Я не боюсь", 239, 7, []string{"Поп", "Инди-поп"}, ""},
		{"Феникс", "Крылья", 223, 8, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Монро", 256, 9, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Психиатр", 242, 10, []string{"Поп", "Инди-поп"}, ""},
		{"Феникс", "Стелс", 250, 11, []string{"Поп", "Поп-рок"}, ""},
		{"Феникс", "Три дня", 235, 12, []string{"Поп", "Инди-поп"}, ""},

		// ANNA ASTI - Царица (2023)
		{"Царица", "Интерлюдия: По барам", 60, 1, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Феникс", 267, 2, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Гармония", 251, 3, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Голая", 234, 4, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Берега", 228, 5, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Интерлюдия: Три дня", 60, 6, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Поцелуи", 245, 7, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Дикая", 239, 8, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Стелс", 223, 9, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Интерлюдия: Царица", 60, 10, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Монро", 256, 11, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Почему?", 242, 12, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Интерлюдия: Крылья", 60, 13, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Трафик", 250, 14, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Нас двое", 235, 15, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Царица", 240, 16, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Без тебя", 267, 17, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Априори", 251, 18, []string{"Поп", "Поп-рок"}, ""},
		{"Царица", "Интерлюдия: Психиатр", 60, 19, []string{"Поп", "Инди-поп"}, ""},
		{"Царица", "Я не боюсь", 234, 20, []string{"Поп", "Инди-поп"}, ""},

		// Zivert - Vinyl #1 (2018)
		{"Vinyl #1", "Life", 201, 1, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #1", "Beverly Hills", 192, 2, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #1", "Fly", 197, 3, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #1", "Зелёные волны", 205, 4, []string{"Поп"}, ""},
		{"Vinyl #1", "Ещё хочу", 198, 5, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #1", "Credo", 200, 6, []string{"Поп"}, ""},
		{"Vinyl #1", "Поребрик", 195, 7, []string{"Поп"}, ""},
		{"Vinyl #1", "В метро", 203, 8, []string{"Поп"}, ""},
		{"Vinyl #1", "Паруса", 189, 9, []string{"Поп"}, ""},

		// Zivert - Vinyl #2 (2019)
		{"Vinyl #2", "Credo", 200, 1, []string{"Поп"}, ""},
		{"Vinyl #2", "Паруса", 189, 2, []string{"Поп"}, ""},
		{"Vinyl #2", "Ещё хочу", 198, 3, []string{"Поп", "Электронная"}, ""},
		{"Vinyl #2", "Чак", 195, 4, []string{"Поп"}, ""},
		{"Vinyl #2", "Рокки", 203, 5, []string{"Поп"}, ""},
		{"Vinyl #2", "Анестезия", 197, 6, []string{"Поп"}, ""},
		{"Vinyl #2", "Натуре мама", 201, 7, []string{"Поп"}, ""},
		{"Vinyl #2", "Бродвей", 205, 8, []string{"Поп"}, ""},
		{"Vinyl #2", "ЯТЛ (feat. M'Dee)", 192, 9, []string{"Поп"}, ""},

		// Zivert - Сияй (2021)
		{"Сияй", "Сияй", 201, 1, []string{"Поп", "Электронная"}, ""},
		{"Сияй", "Никаких больше вечеринок", 200, 2, []string{"Поп"}, ""},
		{"Сияй", "Лайки", 195, 3, []string{"Поп"}, ""},
		{"Сияй", "Good Bye", 198, 4, []string{"Поп", "Электронная"}, ""},
		{"Сияй", "Добрая сказка", 203, 5, []string{"Поп"}, ""},
		{"Сияй", "Мотылёк", 197, 6, []string{"Поп"}, ""},
		{"Сияй", "Крошка", 189, 7, []string{"Поп"}, ""},
		{"Сияй", "Forever Young", 205, 8, []string{"Поп"}, ""},
		{"Сияй", "Бесконечно", 192, 9, []string{"Поп"}, ""},
		{"Сияй", "Новая", 201, 10, []string{"Поп"}, ""},

		// IOWA - Import (2012)
		{"Import", "Улыбайся", 240, 1, []string{"Поп", "Электронная"}, ""},
		{"Import", "Маршрутка", 267, 2, []string{"Поп", "Электронная"}, ""},
		{"Import", "Бьёт бит", 251, 3, []string{"Поп", "Электронная"}, ""},
		{"Import", "Ищу тебя", 234, 4, []string{"Поп", "Инди-поп"}, ""},
		{"Import", "130", 228, 5, []string{"Поп", "Электронная"}, ""},
		{"Import", "Безответно", 245, 6, []string{"Поп", "Инди-поп"}, ""},
		{"Import", "Без тебя", 239, 7, []string{"Поп", "Инди-поп"}, ""},
		{"Import", "Облако", 223, 8, []string{"Поп", "Электронная"}, ""},
		{"Import", "Три слова", 256, 9, []string{"Поп", "Инди-поп"}, ""},

		// IOWA - Export (2015)
		{"Export", "Тает", 240, 1, []string{"Поп", "Электронная"}, ""},
		{"Export", "Простая песня", 267, 2, []string{"Поп", "Инди-поп"}, ""},
		{"Export", "Бьёт бит", 251, 3, []string{"Поп", "Электронная"}, ""},
		{"Export", "Улыбайся", 234, 4, []string{"Поп", "Электронная"}, ""},
		{"Export", "Ищи меня", 228, 5, []string{"Поп", "Инди-поп"}, ""},
		{"Export", "Безответно", 245, 6, []string{"Поп", "Инди-поп"}, ""},
		{"Export", "130", 239, 7, []string{"Поп", "Электронная"}, ""},
		{"Export", "Такси", 223, 8, []string{"Поп", "Электронная"}, ""},
		{"Export", "Несчастный случай", 256, 9, []string{"Поп", "Инди-поп"}, ""},
		{"Export", "Маршрутка", 242, 10, []string{"Поп", "Электронная"}, ""},
		{"Export", "Без тебя", 250, 11, []string{"Поп", "Инди-поп"}, ""},

		// IOWA - Французский альбом (2021)
		{"Французский альбом", "Видели ночь", 240, 1, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Последний раз", 267, 2, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Любовь, которой больше нет", 251, 3, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Один", 234, 4, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Прелюдия", 60, 5, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Она вернётся", 228, 6, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Посмотри в глаза", 245, 7, []string{"Поп", "Инди-поп"}, ""},
		{"Французский альбом", "Ты мне снишься", 239, 8, []string{"Поп", "Инди-поп"}, ""},

		// Клава Кока - Неприлично о личном (2021)
		{"Неприлично о личном", "Начнем сначала", 240, 1, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Мне так хорошо", 267, 2, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Помада", 251, 3, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Нас уночит", 234, 4, []string{"Поп", "Инди-поп"}, ""},
		{"Неприлично о личном", "Крошка моя", 228, 5, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Неприлично о личном", 245, 6, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Химия", 239, 7, []string{"Поп", "Инди-поп"}, ""},
		{"Неприлично о личном", "Малыш", 223, 8, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Треки", 256, 9, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Будто первая любовь", 242, 10, []string{"Поп", "Инди-поп"}, ""},
		{"Неприлично о личном", "Косы", 250, 11, []string{"Поп", "Поп-рок"}, ""},
		{"Неприлично о личном", "Пропади", 235, 12, []string{"Поп", "Инди-поп"}, ""},

		// Клава Кока - Красное вино (2024)
		{"Красное вино", "Красное вино", 240, 1, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "Дикая", 267, 2, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "Молодость", 251, 3, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "Отпусти", 234, 4, []string{"Поп", "Инди-поп"}, ""},
		{"Красное вино", "Хочешь, я к тебе приеду?", 228, 5, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "Не в себе", 245, 6, []string{"Поп", "Инди-поп"}, ""},
		{"Красное вино", "Танцуй красиво", 239, 7, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "Я и ты", 223, 8, []string{"Поп", "Инди-поп"}, ""},
		{"Красное вино", "Мандарины", 256, 9, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "Слухи", 242, 10, []string{"Поп", "Поп-рок"}, ""},
		{"Красное вино", "С Новым годом, малыш", 250, 11, []string{"Поп", "Инди-поп"}, ""},
		{"Красное вино", "Родная", 235, 12, []string{"Поп", "Поп-рок"}, ""},

		// ЛСП - Magic City (2015)
		{"Magic City", "Intro", 60, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Канкан", 240, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Body Talk", 267, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Номера", 251, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Айдище", 234, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Назад", 228, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Танцевать", 245, 7, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},
		{"Magic City", "Маленький принц", 239, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Крыши", 223, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Мечтатели", 256, 10, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Чайлдфри", 242, 11, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Тройник", 250, 12, []string{"Хип-хоп", "Рэп"}, ""},
		{"Magic City", "Неваляшка", 235, 13, []string{"Хип-хоп", "Рэп"}, ""},

		// ЛСП - Tragic City (2017)
		{"Tragic City", "Intro (Выпускной)", 60, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Крыши", 223, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Номера", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Тройник", 250, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Чайлдфри", 242, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Неваляшка", 235, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Маленький принц", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Танцевать", 245, 8, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},
		{"Tragic City", "Айдище", 234, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Мечтатели", 256, 10, []string{"Хип-хоп", "Рэп"}, ""},
		{"Tragic City", "Outro (Путь домой)", 50, 11, []string{"Хип-хоп", "Рэп"}, ""},

		// ЛСП - SAD SOUNDS (2020)
		{"SAD SOUNDS", "Intro", 60, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Плак-Плак", 240, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Without You (feat. МОТ)", 267, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Монетка", 251, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Привет", 234, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Хлоп-Хлоп", 228, 6, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},
		{"SAD SOUNDS", "Ау", 245, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Киса", 239, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"SAD SOUNDS", "Outro", 50, 9, []string{"Хип-хоп", "Рэп"}, ""},

		// The Hatters - Безумие (2016)
		{"Безумие", "Янтарь", 240, 1, []string{"Рок", "Поп-рок"}, ""},
		{"Безумие", "Солнце Монако", 267, 2, []string{"Рок", "Поп-рок"}, ""},
		{"Безумие", "Безумие", 251, 3, []string{"Рок", "Альтернативный рок"}, ""},
		{"Безумие", "Болен тобой", 234, 4, []string{"Рок", "Поп-рок"}, ""},
		{"Безумие", "Клоун", 228, 5, []string{"Рок", "Альтернативный рок"}, ""},
		{"Безумие", "Розовое вино (feat. Jah Khalib)", 245, 6, []string{"Рок", "Поп-рок"}, ""},
		{"Безумие", "Тает дым", 239, 7, []string{"Рок", "Поп-рок"}, ""},
		{"Безумие", "Косатка", 223, 8, []string{"Рок", "Альтернативный рок"}, ""},
		{"Безумие", "Наше лето", 256, 9, []string{"Рок", "Поп-рок"}, ""},
		{"Безумие", "Санрайз", 242, 10, []string{"Рок", "Поп-рок"}, ""},

		// The Hatters - Третий (2018)
		{"Третий", "Какая разница", 240, 1, []string{"Рок", "Поп-рок"}, ""},
		{"Третий", "Маршрут", 267, 2, []string{"Рок", "Поп-рок"}, ""},
		{"Третий", "Русский ковчег", 251, 3, []string{"Рок", "Альтернативный рок"}, ""},
		{"Третий", "Невеста", 234, 4, []string{"Рок", "Поп-рок"}, ""},
		{"Третий", "Солнце Монако", 228, 5, []string{"Рок", "Поп-рок"}, ""},
		{"Третий", "Яд", 245, 6, []string{"Рок", "Альтернативный рок"}, ""},
		{"Третий", "Безумие", 239, 7, []string{"Рок", "Альтернативный рок"}, ""},
		{"Третий", "Санрайз", 223, 8, []string{"Рок", "Поп-рок"}, ""},
		{"Третий", "Болен тобой", 256, 9, []string{"Рок", "Поп-рок"}, ""},
		{"Третий", "Скажи", 242, 10, []string{"Рок", "Поп-рок"}, ""},

		// The Hatters - Четвёртый (2021)
		{"Четвёртый", "Старлетка", 240, 1, []string{"Рок", "Поп-рок"}, ""},
		{"Четвёртый", "Всё решено", 267, 2, []string{"Рок", "Поп-рок"}, ""},
		{"Четвёртый", "Я твоя", 251, 3, []string{"Рок", "Поп-рок"}, ""},
		{"Четвёртый", "Пациент", 234, 4, []string{"Рок", "Альтернативный рок"}, ""},
		{"Четвёртый", "Пляж", 228, 5, []string{"Рок", "Поп-рок"}, ""},
		{"Четвёртый", "Песня 404", 245, 6, []string{"Рок", "Альтернативный рок"}, ""},
		{"Четвёртый", "Мир сошёл с ума", 239, 7, []string{"Рок", "Альтернативный рок"}, ""},
		{"Четвёртый", "Марта", 223, 8, []string{"Рок", "Поп-рок"}, ""},
		{"Четвёртый", "Рок-н-ролл", 256, 9, []string{"Рок", "Поп-рок"}, ""},
		{"Четвёртый", "Амстердам", 242, 10, []string{"Рок", "Поп-рок"}, ""},

		// Miyagi & Эндшпиль - Hajime 1 (2016)
		{"Hajime 1", "Hajime", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "Captain", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "Умка", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "Angel", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "Ламбада (feat. Рем Дигга)", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "Fire Man", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "People", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "Momento", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Hajime 1", "I Got Love (feat. Эндшпиль)", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},

		// Miyagi & Andy Panda - Buster Keaton (2018)
		{"Buster Keaton", "Kosandra", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Там ревели горы", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Ударь", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Minor", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Привет", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Забеги", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Тепло", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Buster Keaton", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "По волнам", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Buster Keaton", "Found Love", 242, 10, []string{"Хип-хоп", "Рэп"}, ""},

		// Miyagi & Andy Panda - Yamakasi (2020)
		{"Yamakasi", "Yamakasi", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Марал", 267, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Ты меня не узнал", 251, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Патрон", 234, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Сюда", 228, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "I Got Love", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Мой друг", 239, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Медлячок", 223, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Колизей", 256, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Yamakasi", "Там ревели горы (Remix)", 242, 10, []string{"Хип-хоп", "Рэп", "Электронная"}, ""},

		// Miyagi & Andy Panda - Million Dollars: Happiness (2021)
		{"Million Dollars: Happiness", "Million Dollars", 240, 1, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Тепло", 239, 2, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "По волнам", 256, 3, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Привет", 228, 4, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Ударь", 251, 5, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Забеги", 245, 6, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Kosandra", 240, 7, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Там ревели горы", 267, 8, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Minor", 234, 9, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Buster Keaton", 223, 10, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Found Love", 242, 11, []string{"Хип-хоп", "Рэп"}, ""},
		{"Million Dollars: Happiness", "Сontent", 250, 12, []string{"Хип-хоп", "Рэп"}, ""},
	}

	// Create tracks and assign genres
	createdTracks := 0
	existingTracks := 0
	skippedTracks := 0
	trackGenreAssignments := 0
	trackGenreErrors := 0

	for _, trackData := range tracks {
		// Find album by title and artist (if needed)
		var album models.Album
		if err := DB.Where("title = ?", trackData.AlbumTitle).First(&album).Error; err != nil {
			log.Printf("  WARNING: Album '%s' not found, skipping track '%s'", trackData.AlbumTitle, trackData.Title)
			skippedTracks++
			continue // Skip if album not found
		}

		// Check if track already exists, if not create it (use FirstOrCreate to avoid duplicates)
		var track models.Track
		trackToCreate := models.Track{
			AlbumID:        album.ID,
			Title:          trackData.Title,
			Duration:       &trackData.Duration,
			TrackNumber:    &trackData.TrackNum,
			CoverImagePath: trackData.CoverImagePath,
		}
		
		result := DB.Where("album_id = ? AND title = ?", album.ID, trackData.Title).FirstOrCreate(&track, trackToCreate)
		if result.Error != nil {
			log.Printf("ERROR: Failed to create/find track %s: %v", trackData.Title, result.Error)
			skippedTracks++
			continue
		}
		
		if result.RowsAffected > 0 {
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
			// Check current genres for this track to avoid unnecessary updates
			var currentGenres []models.Genre
			DB.Model(&track).Association("Genres").Find(&currentGenres)
			
			// Check if genres need to be updated (compare by ID)
			needsUpdate := false
			if len(currentGenres) != len(trackGenres) {
				needsUpdate = true
			} else {
				currentGenreIDs := make(map[uint]bool)
				for _, g := range currentGenres {
					currentGenreIDs[g.ID] = true
				}
				for _, g := range trackGenres {
					if !currentGenreIDs[g.ID] {
						needsUpdate = true
						break
					}
				}
			}
			
			if needsUpdate {
				// Use Replace to update genres (only if needed)
				if err := DB.Model(&track).Association("Genres").Replace(trackGenres); err != nil {
					log.Printf("ERROR: Failed to assign genres to track %s: %v", trackData.Title, err)
					trackGenreErrors++
				} else {
					trackGenreAssignments++
				}
			} else {
				// Genres already match, skip update
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

	// Get current time for setting created_at - distribute over last 7 days
	// 30% within last 24 hours, 70% over last week
	now := time.Now()
	hoursAgo := 0

	// Seed likes for ALL tracks (not just a few per artist)
	trackLikes := []models.TrackLike{}

	// Process all tracks
	for trackIndex, track := range tracks {
		// Generate random number of likes between 5 and 30 for testing "Актуальное"
		// Use combination of trackIndex and track.ID to create variation
		// Using modulo 26 to get range 0-25, then add 5 to get 5-30
		numLikes := 5 + ((trackIndex*7 + int(track.ID)) % 26) // Will give 5-30 likes with variation
		if numLikes > 30 {
			numLikes = 30
		}
		if numLikes < 5 {
			numLikes = 5
		}

		// Calculate how many likes should be in last 24 hours (30%)
		likesInLast24Hours := int(float64(numLikes) * 0.3)

		// Create likes from different users
		// Distribute users cyclically to ensure variety
		startIndex := trackIndex % len(allTestUsers)
		likesCreated := 0
		
		for j := 0; j < numLikes && likesCreated < numLikes; j++ {
			userIndex := (startIndex + j) % len(allTestUsers)
			// Check if like already exists
			var existingLike models.TrackLike
			if err := DB.Where("user_id = ? AND track_id = ?", allTestUsers[userIndex].ID, track.ID).First(&existingLike).Error; err != nil {
				// Create new like
				like := models.TrackLike{
					UserID:  allTestUsers[userIndex].ID,
					TrackID: track.ID,
				}
				// Set created_at: 30% within last 24 hours, 70% over last 7 days
				if likesCreated < likesInLast24Hours {
					// Within last 24 hours
					like.CreatedAt = now.Add(-time.Duration(hoursAgo%24) * time.Hour)
				} else {
					// Over last 7 days (24-168 hours)
					hoursOffset := 24 + (hoursAgo % 144) // 24-168 hours
					like.CreatedAt = now.Add(-time.Duration(hoursOffset) * time.Hour)
				}
				trackLikes = append(trackLikes, like)
				hoursAgo++
				likesCreated++
			} else {
				// Update existing like's created_at
				if likesCreated < likesInLast24Hours {
					existingLike.CreatedAt = now.Add(-time.Duration(hoursAgo%24) * time.Hour)
				} else {
					hoursOffset := 24 + (hoursAgo % 144)
					existingLike.CreatedAt = now.Add(-time.Duration(hoursOffset) * time.Hour)
				}
				if err := DB.Save(&existingLike).Error; err != nil {
					log.Printf("Warning: failed to update track like created_at: %v", err)
				}
				hoursAgo++
				likesCreated++
			}
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

// seedAlbumLikes seeds album likes for testing
func seedAlbumLikes() error {
	log.Println("Seeding album likes...")

	// Get all test users
	var allTestUsers []models.User
	if err := DB.Find(&allTestUsers).Error; err != nil {
		log.Printf("ERROR: Failed to query users: %v", err)
		return fmt.Errorf("failed to query users: %w", err)
	}
	if len(allTestUsers) == 0 {
		log.Printf("WARNING: No users found in database, skipping album likes seed")
		return nil
	}
	log.Printf("Found %d users for album likes", len(allTestUsers))

	// Get all albums
	var albums []models.Album
	if err := DB.Find(&albums).Error; err != nil {
		log.Printf("ERROR: Failed to query albums: %v", err)
		return fmt.Errorf("failed to query albums: %w", err)
	}
	if len(albums) == 0 {
		log.Printf("WARNING: No albums found in database, skipping album likes seed")
		return nil
	}
	log.Printf("Found %d albums for album likes", len(albums))

	// Get current time for setting created_at - distribute over last 7 days
	// 30% within last 24 hours, 70% over last week
	now := time.Now()
	hoursAgo := 0

	// Seed likes for ALL albums
	albumLikes := []models.AlbumLike{}

	// Process all albums
	for albumIndex, album := range albums {
		// Generate random number of likes between 5 and 30 for testing "Актуальное"
		// Use combination of albumIndex and album.ID to create variation
		// Using modulo 26 to get range 0-25, then add 5 to get 5-30
		numLikes := 5 + ((albumIndex*11 + int(album.ID)) % 26) // Will give 5-30 likes with variation
		if numLikes > 30 {
			numLikes = 30
		}
		if numLikes < 5 {
			numLikes = 5
		}

		// Calculate how many likes should be in last 24 hours (30%)
		likesInLast24Hours := int(float64(numLikes) * 0.3)

		// Create likes from different users
		// Distribute users cyclically to ensure variety
		startIndex := albumIndex % len(allTestUsers)
		likesCreated := 0

		for j := 0; j < numLikes && likesCreated < numLikes; j++ {
			userIndex := (startIndex + j) % len(allTestUsers)
			// Check if like already exists
			var existingLike models.AlbumLike
			if err := DB.Where("user_id = ? AND album_id = ?", allTestUsers[userIndex].ID, album.ID).First(&existingLike).Error; err != nil {
				// Create new like
				like := models.AlbumLike{
					UserID:  allTestUsers[userIndex].ID,
					AlbumID: album.ID,
				}
				// Set created_at: 30% within last 24 hours, 70% over last 7 days
				if likesCreated < likesInLast24Hours {
					// Within last 24 hours
					like.CreatedAt = now.Add(-time.Duration(hoursAgo%24) * time.Hour)
				} else {
					// Over last 7 days (24-168 hours)
					hoursOffset := 24 + (hoursAgo % 144) // 24-168 hours
					like.CreatedAt = now.Add(-time.Duration(hoursOffset) * time.Hour)
				}
				albumLikes = append(albumLikes, like)
				hoursAgo++
				likesCreated++
			} else {
				// Update existing like's created_at
				if likesCreated < likesInLast24Hours {
					existingLike.CreatedAt = now.Add(-time.Duration(hoursAgo%24) * time.Hour)
				} else {
					hoursOffset := 24 + (hoursAgo % 144)
					existingLike.CreatedAt = now.Add(-time.Duration(hoursOffset) * time.Hour)
				}
				if err := DB.Save(&existingLike).Error; err != nil {
					log.Printf("Warning: failed to update album like created_at: %v", err)
				}
				hoursAgo++
				likesCreated++
			}
		}
	}

	// Create all new likes in batch
	createdLikes := 0
	failedLikes := 0
	for _, like := range albumLikes {
		if err := DB.Create(&like).Error; err != nil {
			log.Printf("ERROR: Failed to create album like (UserID: %d, AlbumID: %d): %v", like.UserID, like.AlbumID, err)
			failedLikes++
		} else {
			createdLikes++
		}
	}

	log.Printf("Album likes seeding complete: %d created, %d failed", createdLikes, failedLikes)
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
		
		// Find albums by title for reviews
		var basta1, basta2, noggano, basta3 models.Album
		var domNorm, prazdnik36, album2004, uroboros models.Album
		var fenix, carica models.Album
		var vinyl1, vinyl2, siyai models.Album
		var importAlbum, exportAlbum, frenchAlbum models.Album
		var neprilichno, krasnoeVino models.Album
		var magicCity, tragicCity, sadSounds models.Album
		var bezumie, tretiy, chetvertiy models.Album
		var hajime1, busterKeaton, yamakasi, millionDollars models.Album
		
		DB.Where("title = ? AND artist = ?", "Баста 1", "Баста").First(&basta1)
		DB.Where("title = ? AND artist = ?", "Баста 2", "Баста").First(&basta2)
		DB.Where("title = ? AND artist = ?", "Ноггано", "Баста").First(&noggano)
		DB.Where("title = ? AND artist = ?", "Баста 3", "Баста").First(&basta3)
		DB.Where("title = ? AND artist = ?", "Дом с нормальными явлениями", "Скриптонит").First(&domNorm)
		DB.Where("title = ? AND artist = ?", "Праздник на улице 36", "Скриптонит").First(&prazdnik36)
		DB.Where("title = ? AND artist = ?", "2004", "Скриптонит").First(&album2004)
		DB.Where("title = ? AND artist = ?", "Уроборос: улочка и аллея", "Скриптонит & 104").First(&uroboros)
		DB.Where("title = ? AND artist = ?", "Феникс", "ANNA ASTI").First(&fenix)
		DB.Where("title = ? AND artist = ?", "Царица", "ANNA ASTI").First(&carica)
		DB.Where("title = ? AND artist = ?", "Vinyl #1", "Zivert").First(&vinyl1)
		DB.Where("title = ? AND artist = ?", "Vinyl #2", "Zivert").First(&vinyl2)
		DB.Where("title = ? AND artist = ?", "Сияй", "Zivert").First(&siyai)
		DB.Where("title = ? AND artist = ?", "Import", "IOWA").First(&importAlbum)
		DB.Where("title = ? AND artist = ?", "Export", "IOWA").First(&exportAlbum)
		DB.Where("title = ? AND artist = ?", "Французский альбом", "IOWA").First(&frenchAlbum)
		DB.Where("title = ? AND artist = ?", "Неприлично о личном", "Клава Кока").First(&neprilichno)
		DB.Where("title = ? AND artist = ?", "Красное вино", "Клава Кока").First(&krasnoeVino)
		DB.Where("title = ? AND artist = ?", "Magic City", "ЛСП").First(&magicCity)
		DB.Where("title = ? AND artist = ?", "Tragic City", "ЛСП").First(&tragicCity)
		DB.Where("title = ? AND artist = ?", "SAD SOUNDS", "ЛСП").First(&sadSounds)
		DB.Where("title = ? AND artist = ?", "Безумие", "The Hatters").First(&bezumie)
		DB.Where("title = ? AND artist = ?", "Третий", "The Hatters").First(&tretiy)
		DB.Where("title = ? AND artist = ?", "Четвёртый", "The Hatters").First(&chetvertiy)
		DB.Where("title = ? AND artist = ?", "Hajime 1", "Miyagi & Эндшпиль").First(&hajime1)
		DB.Where("title = ? AND artist = ?", "Buster Keaton", "Miyagi & Andy Panda").First(&busterKeaton)
		DB.Where("title = ? AND artist = ?", "Yamakasi", "Miyagi & Andy Panda").First(&yamakasi)
		DB.Where("title = ? AND artist = ?", "Million Dollars: Happiness", "Miyagi & Andy Panda").First(&millionDollars)
		
		// Create test reviews (using atmosphere ratings 1-10, converted to multiplier)
		reviews := []models.Review{
			// Баста - Баста 1 (Хип-хоп)
			{
				UserID:               testUser.ID,
				AlbumID:              &basta1.ID,
				Text:                 "Первый альбом Басты - это классика русского хип-хопа, которая не теряет актуальности. Рифмы сложные, многослойные, с игрой слов - каждый куплет продуман до мелочей. Особенно выделяются треки 'Мой друг' и 'Наше лето' с Гуфом - здесь чувствуется настоящая химия между артистами. Структура треков выстроена идеально: биты качают, куплеты не провисают, припевы цепляют. Продакшн для своего времени на высшем уровне - семплы подобраны идеально, басы мощные, но не перегружают. Подача Басты узнаваема с первых секунд - уверенная, мощная, с правильной интонацией. Альбом создает атмосферу начала 2000-х, ностальгии и одновременно свежести, что делает его вечным.",
				RatingRhymes:         9,
				RatingStructure:      9,
				RatingImplementation: 9,
				RatingIndividuality:  9,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// Баста - Баста 2 (Хип-хоп)
			{
				UserID:               admin.ID,
				AlbumID:              &basta2.ID,
				Text:                 "Второй альбом Басты показывает эволюцию артиста - здесь больше экспериментов, но основа остается узнаваемой. Тексты стали глубже, образы ярче - особенно в треках 'Осень' и 'Выпускной (Медлячок)'. Структура альбома продумана: от интро до аутро всё выстроено логично, каждый трек на своем месте. Битмейкинг улучшился - семплы более разнообразные, аранжировки интереснее. Подача Басты стала более уверенной и зрелой. Альбом создает атмосферу роста, поиска и одновременно уверенности в себе.",
				RatingRhymes:         9,
				RatingStructure:      9,
				RatingImplementation: 10,
				RatingIndividuality:  9,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// Скриптонит - Дом с нормальными явлениями (Хип-хоп)
			{
				UserID:               testUser.ID,
				AlbumID:              &domNorm.ID,
				Text:                 "Дебютный альбом Скриптонита - это настоящий прорыв в русском хип-хопе. Тексты наполнены глубокими образами и метафорами, которые работают на нескольких уровнях. Особенно выделяются 'Вне игры' и 'MDM' - здесь чувствуется уникальный стиль артиста. Структура треков нестандартная, но работает идеально - переходы плавные, динамика выдержана. Продакшн качественный - биты качают, аранжировки интересные, не перегружены. Подача Скриптонита узнаваема - характерный голос, манера чтения, стиль. Альбом создает атмосферу казахстанского хип-хопа, которая привнесла свежесть в жанр.",
				RatingRhymes:         10,
				RatingStructure:      9,
				RatingImplementation: 10,
				RatingIndividuality:  10,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(10),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// ANNA ASTI - Феникс (Поп)
			{
				UserID:               admin.ID,
				AlbumID:              &fenix.ID,
				Text:                 "Дебютный альбом ANNA ASTI - это качественный поп с душой. Тексты простые, но искренние - они говорят о том, что близко каждому. Особенно выделяются треки 'Феникс' и 'Царица' - здесь чувствуется характер артиста. Структура песен классическая для поп-музыки, но работает идеально: запоминающиеся припевы, динамичные куплеты. Продакшн на высоте - каждый элемент на своем месте, синтезаторы звучат современно. Вокал ANNA ASTI узнаваем - мощный, эмоциональный, с характерной манерой подачи. Альбом создает позитивную, вдохновляющую атмосферу, которая поднимает настроение.",
				RatingRhymes:         8,
				RatingStructure:      9,
				RatingImplementation: 10,
				RatingIndividuality:  10,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// Zivert - Vinyl #1 (Поп)
			{
				UserID:               testUser.ID,
				AlbumID:              &vinyl1.ID,
				Text:                 "Дебютный альбом Zivert стал символом эпохи в русской поп-музыке. Тексты простые, но искренние - они говорят о том, что близко каждому, без излишней пафосности. Особенно выделяются треки 'Life' и 'Credo' - здесь чувствуется философия артиста. Структура песен классическая для поп-музыки, но работает идеально: запоминающиеся припевы, динамичные куплеты, бит качает без перебора. Продакшн на высоте - каждый элемент на своем месте, синтезаторы звучат современно, но не навязчиво. Вокал Zivert узнаваем с первых нот - легкий, воздушный, с характерной манерой подачи. Альбом создает позитивную, танцевальную атмосферу, которая поднимает настроение и не надоедает даже после многократного прослушивания.",
				RatingRhymes:         8,
				RatingStructure:      9,
				RatingImplementation: 10,
				RatingIndividuality:  10,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// IOWA - Import (Поп)
			{
				UserID:               admin.ID,
				AlbumID:              &importAlbum.ID,
				Text:                 "Первый альбом IOWA - это качественный поп с элементами электроники. Тексты простые, но цепляющие - особенно 'Улыбайся' и 'Маршрутка' стали хитами. Структура песен стандартная, но работает - припевы запоминаются, куплеты развивают тему. Продакшн качественный - электронные элементы звучат современно, аранжировки не перегружены. Вокал узнаваем - легкий, воздушный, с характерной манерой. Альбом создает позитивную атмосферу, которая поднимает настроение.",
				RatingRhymes:         7,
				RatingStructure:      8,
				RatingImplementation: 9,
				RatingIndividuality:  9,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// ЛСП - Magic City (Хип-хоп)
			{
				UserID:               testUser.ID,
				AlbumID:              &magicCity.ID,
				Text:                 "Первый альбом ЛСП - это уникальный взгляд на русский хип-хоп. Тексты наполнены образами и метафорами, которые работают на эмоциональном уровне. Особенно выделяются треки 'Крыши' и 'Номера' - здесь чувствуется стиль артиста. Структура треков интересная - переходы плавные, динамика выдержана. Продакшн качественный - биты качают, аранжировки интересные. Подача ЛСП узнаваема - характерный голос, манера чтения. Альбом создает атмосферу магического города, которая цепляет и не отпускает.",
				RatingRhymes:         9,
				RatingStructure:      9,
				RatingImplementation: 9,
				RatingIndividuality:  10,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// The Hatters - Безумие (Рок)
			{
				UserID:               admin.ID,
				AlbumID:              &bezumie.ID,
				Text:                 "Первый альбом The Hatters - это качественный рок с элементами инди. Тексты наполнены образами и метафорами - особенно выделяется 'Солнце Монако'. Структура композиций продумана - переходы плавные, динамика выдержана. Продакшн качественный - инструменты звучат объемно, аранжировки не перегружены. Вокал узнаваем - эмоциональный, с характерной манерой подачи. Альбом создает атмосферу безумия и свободы, которая цепляет и не отпускает.",
				RatingRhymes:         8,
				RatingStructure:      9,
				RatingImplementation: 9,
				RatingIndividuality:  9,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
				Status:               models.ReviewStatusApproved,
				ModeratedBy:          &admin.ID,
			},
			// Miyagi - Hajime 1 (Хип-хоп)
			{
				UserID:               testUser.ID,
				AlbumID:              &hajime1.ID,
				Text:                 "Первый альбом Miyagi & Эндшпиль - это уникальный взгляд на русский хип-хоп. Тексты наполнены образами и метафорами, которые работают на эмоциональном уровне. Особенно выделяются треки 'Hajime' и 'I Got Love' - здесь чувствуется стиль дуэта. Структура треков интересная - переходы плавные, динамика выдержана. Продакшн качественный - биты качают, аранжировки интересные, с элементами восточной музыки. Подача Miyagi узнаваема - характерный голос, манера чтения. Альбом создает атмосферу начала пути, которая цепляет и не отпускает.",
				RatingRhymes:         9,
				RatingStructure:      9,
				RatingImplementation: 10,
				RatingIndividuality:  10,
				AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
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

		// Get some tracks for track reviews (from new albums)
		var track1, track2, track3, track4, track5 models.Track
		DB.Where("title = ?", "Мой друг").First(&track1) // Баста 1
		DB.Where("title = ?", "Вне игры").First(&track2) // Скриптонит
		DB.Where("title = ?", "Феникс").First(&track3) // ANNA ASTI
		DB.Where("title = ?", "Life").First(&track4) // Zivert
		DB.Where("title = ?", "Улыбайся").First(&track5) // IOWA
		
		if track1.ID > 0 || track2.ID > 0 || track3.ID > 0 || track4.ID > 0 || track5.ID > 0 {
			// Add some track reviews
			trackReviews := []models.Review{}
			
			if track1.ID > 0 {
				trackReviews = append(trackReviews, models.Review{
					UserID:               testUser.ID,
					TrackID:              &track1.ID,
					Text:                 "Классический трек Басты, который открывает альбом. Текст наполнен образами дружбы и верности, рифмы сложные, многослойные. Структура композиции выстроена идеально - куплеты плавно переходят в запоминающийся припев. Продакшн качественный, бит качает, но не перегружает. Подача Басты узнаваема - уверенная, мощная. Трек создает атмосферу дружбы и братства, которая цепляет с первых секунд.",
					RatingRhymes:         9,
					RatingStructure:      9,
					RatingImplementation: 9,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				})
			}
			
			if track2.ID > 0 {
				trackReviews = append(trackReviews, models.Review{
					UserID:               admin.ID,
					TrackID:              &track2.ID,
					Text:                 "Открывающий трек альбома Скриптонита - это заявление о выходе из игры. Текст наполнен глубокими образами и метафорами, которые работают на эмоциональном уровне. Структура трека интересная - переходы плавные, динамика выдержана. Продакшн качественный - бит качает, аранжировки интересные. Подача Скриптонита узнаваема - характерный голос, манера чтения. Трек создает атмосферу начала пути, которая цепляет и не отпускает.",
					RatingRhymes:         10,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				})
			}
			
			if track3.ID > 0 {
				trackReviews = append(trackReviews, models.Review{
					UserID:               testUser.ID,
					TrackID:              &track3.ID,
					Text:                 "Титульный трек альбома ANNA ASTI - это мощная композиция о возрождении. Текст наполнен образами феникса и возрождения, которые работают на эмоциональном уровне. Структура композиции выстроена идеально - куплеты плавно переходят в запоминающийся припев. Продакшн качественный - каждый элемент на своем месте. Вокал ANNA ASTI узнаваем - мощный, эмоциональный. Трек создает вдохновляющую атмосферу, которая поднимает настроение.",
					RatingRhymes:         8,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				})
			}
			
			if track4.ID > 0 {
				trackReviews = append(trackReviews, models.Review{
					UserID:               admin.ID,
					TrackID:              &track4.ID,
					Text:                 "Открывающий трек альбома Zivert - это гимн жизни. Текст простой, но искренний - он говорит о том, что близко каждому. Структура композиции классическая для поп-музыки, но работает идеально. Продакшн на высоте - синтезаторы звучат современно. Вокал Zivert узнаваем - легкий, воздушный. Трек создает позитивную атмосферу, которая поднимает настроение.",
					RatingRhymes:         8,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				})
			}
			
			if track5.ID > 0 {
				trackReviews = append(trackReviews, models.Review{
					UserID:               testUser.ID,
					TrackID:              &track5.ID,
					Text:                 "Хитовый трек IOWA - это качественный поп с элементами электроники. Текст простой, но цепляющий - особенно запоминается припев. Структура композиции стандартная, но работает - припев запоминается с первого прослушивания. Продакшн качественный - электронные элементы звучат современно. Вокал узнаваем - легкий, воздушный. Трек создает позитивную атмосферу, которая поднимает настроение.",
					RatingRhymes:         7,
					RatingStructure:      8,
					RatingImplementation: 9,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				})
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
		if len(allTestUsersForReviews) >= 3 {
			additionalReviews := []models.Review{
				// Баста - Ноггано (Хип-хоп)
				{
					UserID:               allTestUsersForReviews[2].ID,
					AlbumID:              &noggano.ID,
					Text:                 "Альбом Ноггано - это продолжение эволюции Басты. Рифмы сложные, многослойные, с игрой слов - особенно выделяются треки 'Куба' и 'Вечный жид'. Структура треков выстроена идеально: бит меняется в нужных местах, куплеты не провисают, припевы цепляют. Битмейкинг на высшем уровне - семплы подобраны идеально, басы качают, но не перегружают. Подача Басты узнаваема - уверенная, мощная, с правильной интонацией. Альбом создает атмосферу городской жизни, борьбы и надежды, которая резонирует с аудиторией.",
					RatingRhymes:         9,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(9),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				// Скриптонит - Праздник на улице 36 (Хип-хоп)
				{
					UserID:               allTestUsersForReviews[3].ID,
					AlbumID:              &prazdnik36.ID,
					Text:                 "Второй альбом Скриптонита показывает рост артиста. Тексты наполнены образами и метафорами, которые работают на эмоциональном уровне. Особенно выделяются треки 'Праздник на улице 36' и 'Смок' - здесь чувствуется уникальный стиль артиста. Структура треков интересная - переходы плавные, динамика выдержана. Продакшн качественный - биты качают, аранжировки интересные. Подача Скриптонита узнаваема - характерный голос, манера чтения. Альбом создает атмосферу праздника и одновременно глубины, которая цепляет и не отпускает.",
					RatingRhymes:         9,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				// ANNA ASTI - Царица (Поп)
				{
					UserID:               allTestUsersForReviews[4].ID,
					AlbumID:              &carica.ID,
					Text:                 "Второй альбом ANNA ASTI - это продолжение качественного попа с душой. Тексты простые, но искренние - особенно выделяется титульный трек 'Царица'. Структура песен классическая для поп-музыки, но работает идеально. Продакшн на высоте - каждый элемент на своем месте. Вокал ANNA ASTI узнаваем - мощный, эмоциональный. Альбом создает позитивную, вдохновляющую атмосферу, которая поднимает настроение.",
					RatingRhymes:         8,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				// Zivert - Vinyl #2 (Поп)
				{
					UserID:               allTestUsersForReviews[5].ID,
					AlbumID:              &vinyl2.ID,
					Text:                 "Второй альбом Zivert продолжает традиции первого. Тексты простые, но искренние - они говорят о том, что близко каждому. Структура песен классическая для поп-музыки, но работает идеально. Продакшн на высоте - синтезаторы звучат современно. Вокал Zivert узнаваем - легкий, воздушный. Альбом создает позитивную, танцевальную атмосферу, которая поднимает настроение.",
					RatingRhymes:         8,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  10,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				// IOWA - Export (Поп)
				{
					UserID:               allTestUsersForReviews[6].ID,
					AlbumID:              &exportAlbum.ID,
					Text:                 "Второй альбом IOWA - это качественный поп с элементами электроники. Тексты простые, но цепляющие - особенно 'Тает' и 'Простая песня'. Структура песен стандартная, но работает - припевы запоминаются. Продакшн качественный - электронные элементы звучат современно. Вокал узнаваем - легкий, воздушный. Альбом создает позитивную атмосферу, которая поднимает настроение.",
					RatingRhymes:         7,
					RatingStructure:      8,
					RatingImplementation: 9,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(7),
					Status:               models.ReviewStatusApproved,
					ModeratedBy:          &admin.ID,
				},
				// Клава Кока - Неприлично о личном (Поп)
				{
					UserID:               allTestUsersForReviews[7].ID,
					AlbumID:              &neprilichno.ID,
					Text:                 "Дебютный альбом Клавы Коки - это качественный поп с личными историями. Тексты простые, но искренние - они говорят о личном, без излишней пафосности. Структура песен классическая для поп-музыки, но работает идеально. Продакшн на высоте - каждый элемент на своем месте. Вокал Клавы Коки узнаваем - легкий, эмоциональный. Альбом создает позитивную атмосферу, которая поднимает настроение.",
					RatingRhymes:         8,
					RatingStructure:      9,
					RatingImplementation: 10,
					RatingIndividuality:  9,
					AtmosphereMultiplier: convertAtmosphereToMultiplier(8),
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

	// Seed review likes for testing - create 5-30 likes per review for testing "Актуальное"
	// Distribute: 30% within last 24 hours, 70% over last 7 days
	nowForLikes := time.Now()
	likeHoursAgo := 0

	var reviewLikes []models.ReviewLike
	for i, review := range allReviews {
		if review.Status == models.ReviewStatusApproved && review.ID > 0 {
			// Generate random number of likes between 5 and 30
			// Use combination of review index and review.ID to create variation
			numLikes := 5 + ((i*13 + int(review.ID)) % 26) // Will give 5-30 likes with variation
			if numLikes > 30 {
				numLikes = 30
			}
			if numLikes < 5 {
				numLikes = 5
			}

			// Calculate how many likes should be in last 24 hours (30%)
			likesInLast24Hours := int(float64(numLikes) * 0.3)

			// Create likes from different users, используем циклическое распределение для разнообразия
			startIndex := i % len(allTestUsers) // Начинаем с разных пользователей для каждой рецензии
			likesCreated := 0
			for j := 0; j < numLikes && likesCreated < numLikes; j++ {
				userIndex := (startIndex + j) % len(allTestUsers)
				// Проверяем что пользователь не автор рецензии
				if allTestUsers[userIndex].ID == review.UserID {
					continue
				}
				// Check if like already exists
				var existingLike models.ReviewLike
				if err := DB.Where("user_id = ? AND review_id = ?", allTestUsers[userIndex].ID, review.ID).First(&existingLike).Error; err != nil {
					// Create new like
					like := models.ReviewLike{
						UserID:   allTestUsers[userIndex].ID,
						ReviewID: review.ID,
					}
					// Set created_at: 30% within last 24 hours, 70% over last 7 days
					if likesCreated < likesInLast24Hours {
						// Within last 24 hours
						like.CreatedAt = nowForLikes.Add(-time.Duration(likeHoursAgo%24) * time.Hour)
					} else {
						// Over last 7 days (24-168 hours)
						hoursOffset := 24 + (likeHoursAgo % 144) // 24-168 hours
						like.CreatedAt = nowForLikes.Add(-time.Duration(hoursOffset) * time.Hour)
					}
					reviewLikes = append(reviewLikes, like)
					likeHoursAgo++
					likesCreated++
				} else {
					// Update existing like's created_at
					if likesCreated < likesInLast24Hours {
						existingLike.CreatedAt = nowForLikes.Add(-time.Duration(likeHoursAgo%24) * time.Hour)
					} else {
						hoursOffset := 24 + (likeHoursAgo % 144)
						existingLike.CreatedAt = nowForLikes.Add(-time.Duration(hoursOffset) * time.Hour)
					}
					if err := DB.Save(&existingLike).Error; err != nil {
						log.Printf("Warning: failed to update review like created_at: %v", err)
					}
					likeHoursAgo++
					likesCreated++
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
		"Жить в твоей голове":    "/preview/1.jpg",
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
