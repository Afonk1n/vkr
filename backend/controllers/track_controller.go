package controllers

import (
	"log"
	"music-review-site/backend/middleware"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"sort"
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

	// Среднее считаем агрегатом на чтении (read-only), без UPDATE на каждый трек.
	for i := range tracks {
		if err := tc.AttachAverageScoreBreakdown(&tracks[i]); err != nil {
			log.Printf("Warning: failed to attach average score breakdown for track %d: %v", tracks[i].ID, err)
		}
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

	// Среднее считаем агрегатом на чтении (read-only). Треки уже загружены
	// со всеми связями основным запросом — повторная загрузка и UPDATE не нужны.
	for i := range tracks {
		if err := tc.AttachAverageScoreBreakdown(&tracks[i]); err != nil {
			log.Printf("Warning: failed to attach average score breakdown for track %d: %v", tracks[i].ID, err)
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

	// Среднее — агрегатом на чтении, без UPDATE.
	if err := tc.AttachAverageScoreBreakdown(&track); err != nil {
		log.Printf("Warning: failed to attach average score breakdown for track %d: %v", track.ID, err)
	}

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

	// Для демо берём по одному лидеру от каждого артиста. Иначе при плотном
	// каталоге один исполнитель легко занимает весь топ несколькими треками.
	type popularTrackRow struct {
		TrackID   uint
		LikeCount int64
	}
	var rankedRows []popularTrackRow
	rankingSQL := `
		WITH counts AS (
			SELECT t.id AS track_id, a.artist, COUNT(tl.id) AS like_count
			FROM tracks t
			JOIN albums a ON a.id = t.album_id AND a.deleted_at IS NULL
			LEFT JOIN track_likes tl ON tl.track_id = t.id
				AND tl.created_at >= ? AND tl.deleted_at IS NULL
			WHERE t.deleted_at IS NULL
			GROUP BY t.id, a.artist
		), ranked AS (
			SELECT track_id, like_count,
				ROW_NUMBER() OVER (PARTITION BY artist ORDER BY like_count DESC, track_id DESC) AS artist_rank
			FROM counts
		)
		SELECT track_id, like_count
		FROM ranked
		WHERE artist_rank = 1
		ORDER BY like_count DESC, track_id DESC
		LIMIT ?`
	if err := tc.DB.Raw(rankingSQL, since, limit).Scan(&rankedRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch popular tracks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	trackIDs := make([]uint, 0, len(rankedRows))
	trackOrder := make(map[uint]int, len(rankedRows))
	for index, row := range rankedRows {
		trackIDs = append(trackIDs, row.TrackID)
		trackOrder[row.TrackID] = index
	}

	var tracks []models.Track
	if len(trackIDs) > 0 {
		if err := tc.DB.Preload("Album").Preload("Album.Genre").Preload("Genres").Preload("Likes").
			Where("id IN ?", trackIDs).Find(&tracks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Internal Server Error", Message: "Failed to fetch popular tracks", Code: http.StatusInternalServerError})
			return
		}
		sort.SliceStable(tracks, func(i, j int) bool { return trackOrder[tracks[i].ID] < trackOrder[tracks[j].ID] })
	}

	// Среднее — агрегатом на чтении; дедуп жанров на уже загруженном треке.
	for i := range tracks {
		if err := tc.AttachAverageScoreBreakdown(&tracks[i]); err != nil {
			log.Printf("Warning: failed to attach average score breakdown for track %d: %v", tracks[i].ID, err)
		}
		seen := make(map[uint]bool, len(tracks[i].Genres))
		unique := tracks[i].Genres[:0]
		for _, genre := range tracks[i].Genres {
			if !seen[genre.ID] {
				seen[genre.ID] = true
				unique = append(unique, genre)
			}
		}
		tracks[i].Genres = unique
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

	// Жёсткое удаление (см. уникальный индекс ux_track_like_pair).
	if err := tc.DB.Unscoped().Where("user_id = ? AND track_id = ?", userID, trackID).Delete(&models.TrackLike{}).Error; err != nil {
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

// AttachAverageScoreBreakdown adds transient average criterion values to a track response.
func (tc *TrackController) AttachAverageScoreBreakdown(track *models.Track) error {
	var avg struct {
		Count          int64
		Rhymes         float64
		Structure      float64
		Implementation float64
		Individuality  float64
		AtmosphereMult float64
		FinalScore     float64
	}

	if err := tc.DB.Model(&models.Review{}).
		Select(`
			COUNT(*) AS count,
			COALESCE(AVG(rating_rhymes), 0) AS rhymes,
			COALESCE(AVG(rating_structure), 0) AS structure,
			COALESCE(AVG(rating_implementation), 0) AS implementation,
			COALESCE(AVG(rating_individuality), 0) AS individuality,
			COALESCE(AVG(atmosphere_multiplier), 0) AS atmosphere_mult,
			COALESCE(AVG(final_score), 0) AS final_score
		`).
		Where("track_id = ? AND status = ?", track.ID, models.ReviewStatusApproved).
		Scan(&avg).Error; err != nil {
		return err
	}

	if avg.Count == 0 {
		return nil
	}

	track.ApprovedReviewsCount = avg.Count
	track.AverageRating = float64(int(avg.FinalScore + 0.5))
	track.AverageRatingRhymes = avg.Rhymes
	track.AverageRatingStructure = avg.Structure
	track.AverageRatingImplementation = avg.Implementation
	track.AverageRatingIndividuality = avg.Individuality
	track.AverageAtmosphereRating = 1 + (avg.AtmosphereMult-1.0)/(0.6072/9.0)
	return nil
}
