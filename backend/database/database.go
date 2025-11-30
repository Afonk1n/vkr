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

	// Seed reviews (separate check, can be added even if users exist)
	if err := seedReviews(); err != nil {
		log.Printf("Warning: failed to seed reviews: %v", err)
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
		&models.Review{},
		&models.ReviewLike{},
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

	// Seed albums
	albums := []models.Album{
		{
			Title:         "Abbey Road",
			Artist:        "The Beatles",
			GenreID:       genreMap["Рок"].ID,
			Description:   "Легендарный альбом The Beatles, один из величайших в истории рок-музыки",
			AverageRating: 0,
		},
		{
			Title:         "Thriller",
			Artist:        "Michael Jackson",
			GenreID:       genreMap["Поп"].ID,
			Description:   "Самый продаваемый альбом всех времен",
			AverageRating: 0,
		},
		{
			Title:         "The Chronic",
			Artist:        "Dr. Dre",
			GenreID:       genreMap["Хип-хоп"].ID,
			Description:   "Классический альбом хип-хопа 90-х",
			AverageRating: 0,
		},
		{
			Title:         "Random Access Memories",
			Artist:        "Daft Punk",
			GenreID:       genreMap["Электронная"].ID,
			Description:   "Электронный альбом с элементами диско и фанка",
			AverageRating: 0,
		},
		{
			Title:         "Kind of Blue",
			Artist:        "Miles Davis",
			GenreID:       genreMap["Джаз"].ID,
			Description:   "Величайший джазовый альбом всех времен",
			AverageRating: 0,
		},
	}

	for _, album := range albums {
		if err := DB.Create(&album).Error; err != nil {
			return fmt.Errorf("failed to seed album %s: %w", album.Title, err)
		}
	}

	log.Println("Initial data seeded successfully")
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
			AlbumID:              albums[0].ID, // Abbey Road
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
			AlbumID:              albums[1].ID, // Thriller
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
			AlbumID:              albums[2].ID, // The Chronic
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
			AlbumID:              albums[3].ID, // Random Access Memories
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
			AlbumID:              albums[4].ID, // Kind of Blue
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
			AlbumID:              albums[0].ID, // Abbey Road (вторая рецензия)
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
			AlbumID:              albums[1].ID, // Thriller (только оценка, без текста)
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

	log.Println("Test reviews seeded successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
