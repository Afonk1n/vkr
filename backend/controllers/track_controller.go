package controllers

import (
	"log"
	"music-review-site/backend/middleware"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TrackController struct {
	DB *gorm.DB
}

// CreateTrackRequest represents track creation request
type CreateTrackRequest struct {
	AlbumID     uint   `json:"album_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Duration    *int   `json:"duration"`
	TrackNumber *int   `json:"track_number"`
	GenreIDs    []uint `json:"genre_ids"` // Array of genre IDs
}

// UpdateTrackRequest represents track update request
type UpdateTrackRequest struct {
	Title       string `json:"title"`
	Duration    *int   `json:"duration"`
	TrackNumber *int   `json:"track_number"`
	GenreIDs    []uint `json:"genre_ids"` // Array of genre IDs
}

// GetTracks retrieves tracks for an album
func (tc *TrackController) GetTracks(c *gin.Context) {
	albumID := c.Param("id")
	var tracks []models.Track

	if err := tc.DB.Preload("Likes").Preload("Genres").Where("album_id = ?", albumID).Order("track_number ASC, created_at ASC").Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch tracks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, tracks)
}

// GetAllTracks retrieves all tracks with filtering, sorting and pagination
func (tc *TrackController) GetAllTracks(c *gin.Context) {
	var tracks []models.Track
	query := tc.DB.Model(&models.Track{}).Preload("Album").Preload("Album.Genre").Preload("Genres").Preload("Likes")

	// Filter by genre_ids (array) - AND logic: track must have ALL selected genres
	if genreIDsParam := c.QueryArray("genre_ids[]"); len(genreIDsParam) > 0 {
		genreIDs := make([]uint, 0)
		for _, idStr := range genreIDsParam {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				genreIDs = append(genreIDs, uint(id))
			}
		}
		if len(genreIDs) > 0 {
			// Use subquery to find tracks that have ALL selected genres
			// For each genre, we check if track has it, then count matches
			query = query.Where(`
				(SELECT COUNT(DISTINCT genre_id) 
				 FROM track_genres 
				 WHERE track_id = tracks.id AND genre_id IN (?)
				) = ?`, genreIDs, len(genreIDs))
		}
	}

	// Search by title or artist (through album)
	if search := c.Query("search"); search != "" {
		query = query.Where("tracks.title ILIKE ? OR EXISTS (SELECT 1 FROM albums WHERE albums.id = tracks.album_id AND albums.artist ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	// Sort
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// Handle special sorting cases
	switch sortBy {
	case "release_date":
		if sortOrder == "desc" {
			query = query.Order("(SELECT release_date FROM albums WHERE albums.id = tracks.album_id) DESC NULLS LAST, tracks.created_at DESC")
		} else {
			query = query.Order("(SELECT release_date FROM albums WHERE albums.id = tracks.album_id) ASC NULLS LAST, tracks.created_at ASC")
		}
	case "title":
		if sortOrder == "desc" {
			query = query.Order("tracks.title DESC")
		} else {
			query = query.Order("tracks.title ASC")
		}
	case "average_rating":
		if sortOrder == "desc" {
			query = query.Order("tracks.average_rating DESC NULLS LAST, tracks.created_at DESC")
		} else {
			query = query.Order("tracks.average_rating ASC NULLS LAST, tracks.created_at ASC")
		}
	case "likes_count":
		// Sort by number of likes
		if sortOrder == "desc" {
			query = query.Order("(SELECT COUNT(*) FROM track_likes WHERE track_likes.track_id = tracks.id) DESC, tracks.created_at DESC")
		} else {
			query = query.Order("(SELECT COUNT(*) FROM track_likes WHERE track_likes.track_id = tracks.id) ASC, tracks.created_at ASC")
		}
	default: // created_at
		if sortOrder == "desc" {
			query = query.Order("tracks.created_at DESC")
		} else {
			query = query.Order("tracks.created_at ASC")
		}
	}

	// Count total with same filters (before pagination)
	var total int64
	countQuery := tc.DB.Model(&models.Track{})
	
	// Apply same filters to count query
	if genreIDsParam := c.QueryArray("genre_ids[]"); len(genreIDsParam) > 0 {
		genreIDs := make([]uint, 0)
		for _, idStr := range genreIDsParam {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				genreIDs = append(genreIDs, uint(id))
			}
		}
		if len(genreIDs) > 0 {
			countQuery = countQuery.Where(`
				(SELECT COUNT(DISTINCT genre_id) 
				 FROM track_genres 
				 WHERE track_id = tracks.id AND genre_id IN (?)
				) = ?`, genreIDs, len(genreIDs))
		}
	}
	if search := c.Query("search"); search != "" {
		countQuery = countQuery.Where("tracks.title ILIKE ? OR EXISTS (SELECT 1 FROM albums WHERE albums.id = tracks.album_id AND albums.artist ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}
	countQuery.Count(&total)

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch tracks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Calculate average ratings for all tracks
	for i := range tracks {
		if err := tc.CalculateAverageRating(tracks[i].ID); err != nil {
			log.Printf("Warning: failed to calculate average rating for track %d: %v", tracks[i].ID, err)
		}
		// Reload track to get updated rating
		var updatedTrack models.Track
		if err := tc.DB.Preload("Album").Preload("Album.Genre").Preload("Genres").Preload("Likes").First(&updatedTrack, tracks[i].ID).Error; err == nil {
			tracks[i] = updatedTrack
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tracks":    tracks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTrack retrieves track by ID
func (tc *TrackController) GetTrack(c *gin.Context) {
	id := c.Param("id")
	var track models.Track

	if err := tc.DB.Preload("Album").Preload("Album.Genre").Preload("Likes").Preload("Genres").First(&track, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Track not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Calculate average rating
	if err := tc.CalculateAverageRating(track.ID); err != nil {
		log.Printf("Warning: failed to calculate average rating for track %d: %v", track.ID, err)
	}
	// Reload track to get updated rating
	tc.DB.First(&track, id)

	c.JSON(http.StatusOK, track)
}

// CreateTrack creates a new track
func (tc *TrackController) CreateTrack(c *gin.Context) {
	var req CreateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check if album exists
	var album models.Album
	if err := tc.DB.First(&album, req.AlbumID).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Album not found",
			Code:    http.StatusBadRequest,
		})
		return
	}

	track := models.Track{
		AlbumID:     req.AlbumID,
		Title:       req.Title,
		Duration:    req.Duration,
		TrackNumber: req.TrackNumber,
	}

	if err := tc.DB.Create(&track).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create track",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Associate genres if provided
	if len(req.GenreIDs) > 0 {
		var genres []models.Genre
		if err := tc.DB.Where("id IN ?", req.GenreIDs).Find(&genres).Error; err == nil {
			tc.DB.Model(&track).Association("Genres").Replace(genres)
		}
	}

	tc.DB.Preload("Album").Preload("Genres").First(&track, track.ID)
	c.JSON(http.StatusCreated, track)
}

// UpdateTrack updates a track
func (tc *TrackController) UpdateTrack(c *gin.Context) {
	id := c.Param("id")
	var track models.Track

	if err := tc.DB.First(&track, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Track not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	var req UpdateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if req.Title != "" {
		track.Title = req.Title
	}
	if req.Duration != nil {
		track.Duration = req.Duration
	}
	if req.TrackNumber != nil {
		track.TrackNumber = req.TrackNumber
	}

	if err := tc.DB.Save(&track).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update track",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update genres if provided
	if req.GenreIDs != nil {
		var genres []models.Genre
		if len(req.GenreIDs) > 0 {
			if err := tc.DB.Where("id IN ?", req.GenreIDs).Find(&genres).Error; err == nil {
				tc.DB.Model(&track).Association("Genres").Replace(genres)
			}
		} else {
			// Clear all genres if empty array
			tc.DB.Model(&track).Association("Genres").Clear()
		}
	}

	tc.DB.Preload("Album").Preload("Genres").First(&track, track.ID)
	c.JSON(http.StatusOK, track)
}

// DeleteTrack deletes a track
func (tc *TrackController) DeleteTrack(c *gin.Context) {
	id := c.Param("id")
	var track models.Track

	if err := tc.DB.First(&track, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Track not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	if err := tc.DB.Delete(&track).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete track",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track deleted successfully"})
}

// GetPopularTracks retrieves most liked tracks from last 24 hours
func (tc *TrackController) GetPopularTracks(c *gin.Context) {
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}
	since := time.Now().Add(-24 * time.Hour)

	var tracks []models.Track
	// Get tracks with likes from last 24 hours, ordered by like count
	query := tc.DB.Model(&models.Track{}).
		Preload("Album").
		Preload("Album.Genre").
		Preload("Genres").
		Preload("Likes").
		Joins("LEFT JOIN track_likes ON tracks.id = track_likes.track_id AND track_likes.created_at >= ? AND track_likes.deleted_at IS NULL", since).
		Group("tracks.id").
		Order("COUNT(track_likes.id) DESC, tracks.created_at DESC").
		Limit(limit)

	if err := query.Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch popular tracks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Calculate average ratings for all tracks
	for i := range tracks {
		if err := tc.CalculateAverageRating(tracks[i].ID); err != nil {
			log.Printf("Warning: failed to calculate average rating for track %d: %v", tracks[i].ID, err)
		}
		// Reload track to get updated rating with all relationships
		var updatedTrack models.Track
		if err := tc.DB.Preload("Album").Preload("Album.Genre").Preload("Genres").Preload("Likes").First(&updatedTrack, tracks[i].ID).Error; err == nil {
			// Remove duplicate genres by ID
			genreMap := make(map[uint]models.Genre)
			for _, genre := range updatedTrack.Genres {
				if _, exists := genreMap[genre.ID]; !exists {
					genreMap[genre.ID] = genre
				}
			}
			// Rebuild genres slice without duplicates
			updatedTrack.Genres = make([]models.Genre, 0, len(genreMap))
			for _, genre := range genreMap {
				updatedTrack.Genres = append(updatedTrack.Genres, genre)
			}
			tracks[i] = updatedTrack
		}
	}

	c.JSON(http.StatusOK, tracks)
}

// LikeTrack adds a like to a track
func (tc *TrackController) LikeTrack(c *gin.Context) {
	trackID := c.Param("id")
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Check if track exists
	var track models.Track
	if err := tc.DB.First(&track, trackID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Track not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check if like already exists
	var existingLike models.TrackLike
	if err := tc.DB.Where("user_id = ? AND track_id = ?", userID, trackID).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already liked", "liked": true})
		return
	}

	// Create like
	like := models.TrackLike{
		UserID:  userID,
		TrackID: track.ID,
	}

	if err := tc.DB.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to like track",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track liked", "liked": true})
}

// UnlikeTrack removes a like from a track
func (tc *TrackController) UnlikeTrack(c *gin.Context) {
	trackID := c.Param("id")
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Check if track exists
	var track models.Track
	if err := tc.DB.First(&track, trackID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Track not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Delete like
	if err := tc.DB.Where("user_id = ? AND track_id = ?", userID, trackID).Delete(&models.TrackLike{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to unlike track",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track unliked", "liked": false})
}

// CalculateAverageRating calculates and updates average rating for a track
func (tc *TrackController) CalculateAverageRating(trackID uint) error {
	var reviews []models.Review
	if err := tc.DB.Where("track_id = ? AND status = ?", trackID, models.ReviewStatusApproved).Find(&reviews).Error; err != nil {
		return err
	}

	if len(reviews) == 0 {
		return tc.DB.Model(&models.Track{}).Where("id = ?", trackID).Update("average_rating", 0).Error
	}

	var totalScore float64
	for _, review := range reviews {
		totalScore += review.FinalScore
	}

	averageRating := totalScore / float64(len(reviews))
	// Round to nearest integer
	roundedAverage := float64(int(averageRating + 0.5))
	return tc.DB.Model(&models.Track{}).Where("id = ?", trackID).Update("average_rating", roundedAverage).Error
}
