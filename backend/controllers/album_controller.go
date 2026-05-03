package controllers

import (
	"fmt"
	"log"
	"music-review-site/backend/middleware"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlbumController struct {
	DB *gorm.DB
}

// CreateAlbumRequest represents album creation request
type CreateAlbumRequest struct {
	Title          string `json:"title" binding:"required"`
	Artist         string `json:"artist" binding:"required"`
	GenreID        uint   `json:"genre_id" binding:"required"`
	CoverImagePath string `json:"cover_image_path"`
	Description    string `json:"description"`
	ReleaseDate    string `json:"release_date"`
}

// UpdateAlbumRequest represents album update request
type UpdateAlbumRequest struct {
	Title          string `json:"title"`
	Artist         string `json:"artist"`
	GenreID        uint   `json:"genre_id"`
	CoverImagePath string `json:"cover_image_path"`
	Description    string `json:"description"`
	ReleaseDate    string `json:"release_date"`
}

func parseAlbumReleaseDate(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func albumCoverUploadDir() string {
	if value := strings.TrimSpace(os.Getenv("COVER_UPLOAD_DIR")); value != "" {
		return value
	}
	if _, err := os.Stat("/frontend/public/preview"); err == nil {
		return "/frontend/public/preview/uploads"
	}
	return filepath.Clean("../frontend/public/preview/uploads")
}

// GetAlbums retrieves list of albums with filters
func (ac *AlbumController) GetAlbums(c *gin.Context) {
	var albums []models.Album
	query := ac.DB.Model(&models.Album{}).Preload("Genre").Preload("Likes")

	// Filter by genre
	if genreID := c.Query("genre_id"); genreID != "" {
		query = query.Where("genre_id = ?", genreID)
	}

	// Search by title or artist
	if search := c.Query("search"); search != "" {
		query = query.Where("title ILIKE ? OR artist ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Sort
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// Handle special sorting cases
	if sortBy == "release_date" {
		// For release_date, handle NULL values
		if sortOrder == "desc" {
			query = query.Order("release_date DESC NULLS LAST, created_at DESC")
		} else {
			query = query.Order("release_date ASC NULLS LAST, created_at ASC")
		}
	} else if sortBy == "title" || sortBy == "artist" {
		// For text fields, use case-insensitive sorting
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		// For numeric fields (average_rating, created_at, etc.)
		query = query.Order(sortBy + " " + sortOrder)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// Count total with same filters (before pagination)
	var total int64
	countQuery := ac.DB.Model(&models.Album{})
	if genreID := c.Query("genre_id"); genreID != "" {
		countQuery = countQuery.Where("genre_id = ?", genreID)
	}
	if search := c.Query("search"); search != "" {
		countQuery = countQuery.Where("title ILIKE ? OR artist ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	countQuery.Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch albums",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"albums":    albums,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAlbumsByArtist retrieves all albums by artist name
func (ac *AlbumController) GetAlbumsByArtist(c *gin.Context) {
	artistName := c.Param("name")
	// URL decode the artist name
	decodedName, err := url.QueryUnescape(artistName)
	if err != nil {
		decodedName = artistName
	}

	var albums []models.Album
	query := ac.DB.Model(&models.Album{}).Preload("Genre").Preload("Likes").Where("artist = ?", decodedName)

	// Sort by release_date if available, otherwise by created_at
	query = query.Order("release_date DESC NULLS LAST, created_at DESC")

	if err := query.Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch albums",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"albums": albums,
		"artist": decodedName,
		"total":  len(albums),
	})
}

// GetAlbum retrieves album by ID
func (ac *AlbumController) GetAlbum(c *gin.Context) {
	id := c.Param("id")
	var album models.Album

	if err := ac.DB.Preload("Genre").Preload("Tracks").Preload("Likes").First(&album, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Album not found",
			Code:    http.StatusNotFound,
		})
		return
	}
	if err := ac.AttachAverageScoreBreakdown(&album); err != nil {
		log.Printf("Warning: failed to attach average score breakdown for album %d: %v", album.ID, err)
	}

	c.JSON(http.StatusOK, album)
}

// CreateAlbum creates a new album
func (ac *AlbumController) CreateAlbum(c *gin.Context) {
	var req CreateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check if genre exists
	var genre models.Genre
	if err := ac.DB.First(&genre, req.GenreID).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Genre not found",
			Code:    http.StatusBadRequest,
		})
		return
	}

	album := models.Album{
		Title:          req.Title,
		Artist:         req.Artist,
		GenreID:        req.GenreID,
		CoverImagePath: req.CoverImagePath,
		Description:    req.Description,
		AverageRating:  0,
	}

	releaseDate, err := parseAlbumReleaseDate(req.ReleaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid release_date format, expected YYYY-MM-DD",
			Code:    http.StatusBadRequest,
		})
		return
	}
	album.ReleaseDate = releaseDate

	if err := ac.DB.Create(&album).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create album",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ac.DB.Preload("Genre").First(&album, album.ID)
	c.JSON(http.StatusCreated, album)
}

// UpdateAlbum updates an album
func (ac *AlbumController) UpdateAlbum(c *gin.Context) {
	id := c.Param("id")
	var album models.Album

	if err := ac.DB.First(&album, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Album not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	var req UpdateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Update fields
	if req.Title != "" {
		album.Title = req.Title
	}
	if req.Artist != "" {
		album.Artist = req.Artist
	}
	if req.GenreID != 0 {
		// Check if genre exists
		var genre models.Genre
		if err := ac.DB.First(&genre, req.GenreID).Error; err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Bad Request",
				Message: "Genre not found",
				Code:    http.StatusBadRequest,
			})
			return
		}
		album.GenreID = req.GenreID
	}
	if req.CoverImagePath != "" {
		album.CoverImagePath = req.CoverImagePath
	}
	if req.Description != "" {
		album.Description = req.Description
	}
	if req.ReleaseDate != "" {
		releaseDate, err := parseAlbumReleaseDate(req.ReleaseDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid release_date format, expected YYYY-MM-DD",
				Code:    http.StatusBadRequest,
			})
			return
		}
		album.ReleaseDate = releaseDate
	}

	if err := ac.DB.Save(&album).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update album",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ac.DB.Preload("Genre").First(&album, album.ID)
	c.JSON(http.StatusOK, album)
}

// UploadCover uploads an album cover into the public preview storage.
func (ac *AlbumController) UploadCover(c *gin.Context) {
	file, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Cover file is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if file.Size > 8*1024*1024 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Cover file is too large, max size is 8 MB",
			Code:    http.StatusBadRequest,
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Only JPG, PNG and WEBP covers are supported",
			Code:    http.StatusBadRequest,
		})
		return
	}

	uploadDir := albumCoverUploadDir()
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to prepare cover storage",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, base)
	base = strings.Trim(base, "-_")
	if base == "" {
		base = "cover"
	}

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), base, ext)
	destination := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to upload cover",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"cover_image_path": "/preview/uploads/" + filename,
	})
}

// DeleteAlbum deletes an album
func (ac *AlbumController) DeleteAlbum(c *gin.Context) {
	id := c.Param("id")
	var album models.Album

	if err := ac.DB.First(&album, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Album not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	if err := ac.DB.Delete(&album).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete album",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Album deleted successfully",
	})
}

// CalculateAverageRating calculates and updates average rating for an album
func (ac *AlbumController) CalculateAverageRating(albumID uint) error {
	var reviews []models.Review
	if err := ac.DB.Where("album_id = ? AND status = ?", albumID, models.ReviewStatusApproved).Find(&reviews).Error; err != nil {
		return err
	}

	if len(reviews) == 0 {
		return ac.DB.Model(&models.Album{}).Where("id = ?", albumID).Update("average_rating", 0).Error
	}

	var totalScore float64
	for _, review := range reviews {
		totalScore += review.FinalScore
	}

	averageRating := totalScore / float64(len(reviews))
	// Round to nearest integer
	roundedAverage := float64(int(averageRating + 0.5))
	return ac.DB.Model(&models.Album{}).Where("id = ?", albumID).Update("average_rating", roundedAverage).Error
}

// AttachAverageScoreBreakdown adds transient average criterion values to an album response.
func (ac *AlbumController) AttachAverageScoreBreakdown(album *models.Album) error {
	var avg struct {
		Count          int64
		Rhymes         float64
		Structure      float64
		Implementation float64
		Individuality  float64
		AtmosphereMult float64
		FinalScore     float64
	}

	if err := ac.DB.Model(&models.Review{}).
		Select(`
			COUNT(*) AS count,
			COALESCE(AVG(rating_rhymes), 0) AS rhymes,
			COALESCE(AVG(rating_structure), 0) AS structure,
			COALESCE(AVG(rating_implementation), 0) AS implementation,
			COALESCE(AVG(rating_individuality), 0) AS individuality,
			COALESCE(AVG(atmosphere_multiplier), 0) AS atmosphere_mult,
			COALESCE(AVG(final_score), 0) AS final_score
		`).
		Where("album_id = ? AND status = ?", album.ID, models.ReviewStatusApproved).
		Scan(&avg).Error; err != nil {
		return err
	}

	if avg.Count == 0 {
		return nil
	}

	album.ApprovedReviewsCount = avg.Count
	album.AverageRating = float64(int(avg.FinalScore + 0.5))
	album.AverageRatingRhymes = avg.Rhymes
	album.AverageRatingStructure = avg.Structure
	album.AverageRatingImplementation = avg.Implementation
	album.AverageRatingIndividuality = avg.Individuality
	album.AverageAtmosphereRating = 1 + (avg.AtmosphereMult-1.0)/(0.6072/9.0)
	return nil
}

// LikeAlbum adds a like to an album
func (ac *AlbumController) LikeAlbum(c *gin.Context) {
	albumID := c.Param("id")
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Check if album exists
	var album models.Album
	if err := ac.DB.First(&album, albumID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Album not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check if like already exists
	var existingLike models.AlbumLike
	if err := ac.DB.Where("user_id = ? AND album_id = ?", userID, albumID).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already liked", "liked": true})
		return
	}

	// Create like
	like := models.AlbumLike{
		UserID:  userID,
		AlbumID: album.ID,
	}

	if err := ac.DB.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to like album",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Album liked", "liked": true})
}

// UnlikeAlbum removes a like from an album
func (ac *AlbumController) UnlikeAlbum(c *gin.Context) {
	albumID := c.Param("id")
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Check if album exists
	var album models.Album
	if err := ac.DB.First(&album, albumID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Album not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Delete like
	if err := ac.DB.Where("user_id = ? AND album_id = ?", userID, albumID).Delete(&models.AlbumLike{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to unlike album",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Album unliked", "liked": false})
}
