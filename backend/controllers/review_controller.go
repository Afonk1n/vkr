package controllers

import (
	"fmt"
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

// convertAtmosphereToMultiplier converts atmosphere rating (1-10) to multiplier (1.0000-1.6072)
// Formula: multiplier = 1.0000 + (rating - 1) * 0.0674666...
// This ensures max score of 90 when all ratings are 10
func convertAtmosphereToMultiplier(rating int) float64 {
	step := 0.6072 / 9.0
	return 1.0000 + float64(rating-1)*step
}

type ReviewController struct {
	DB *gorm.DB
}

// CreateReviewRequest represents review creation request
type CreateReviewRequest struct {
	AlbumID              *uint  `json:"album_id"` // Optional - either album_id or track_id must be provided
	TrackID              *uint  `json:"track_id"` // Optional - either album_id or track_id must be provided
	Text                 string `json:"text"`
	RatingRhymes         int    `json:"rating_rhymes" binding:"required,min=1,max=10"`
	RatingStructure      int    `json:"rating_structure" binding:"required,min=1,max=10"`
	RatingImplementation int    `json:"rating_implementation" binding:"required,min=1,max=10"`
	RatingIndividuality  int    `json:"rating_individuality" binding:"required,min=1,max=10"`
	AtmosphereRating     int    `json:"atmosphere_rating" binding:"required,min=1,max=10"` // 1-10, will be converted to multiplier
}

// UpdateReviewRequest represents review update request
type UpdateReviewRequest struct {
	Text                 *string `json:"text"` // Pointer to detect if field was provided
	RatingRhymes         int     `json:"rating_rhymes" binding:"min=1,max=10"`
	RatingStructure      int     `json:"rating_structure" binding:"min=1,max=10"`
	RatingImplementation int     `json:"rating_implementation" binding:"min=1,max=10"`
	RatingIndividuality  int     `json:"rating_individuality" binding:"min=1,max=10"`
	AtmosphereRating     int     `json:"atmosphere_rating" binding:"min=1,max=10"` // 1-10, will be converted to multiplier
}

// GetReviews retrieves list of reviews with filters
func (rc *ReviewController) GetReviews(c *gin.Context) {
	var reviews []models.Review
	query := rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").Preload("Track").Preload("Track.Album").Preload("Likes")

	// Filter by album
	if albumID := c.Query("album_id"); albumID != "" {
		query = query.Where("album_id = ?", albumID)
	}

	// Filter by track
	if trackID := c.Query("track_id"); trackID != "" {
		query = query.Where("track_id = ?", trackID)
	}

	// Filter by user
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	} else {
		// By default, show only approved reviews
		query = query.Where("status = ?", models.ReviewStatusApproved)
	}

	// Sort
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	query = query.Order(sortBy + " " + sortOrder)

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&models.Review{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch reviews",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":   reviews,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetReview retrieves review by ID
func (rc *ReviewController) GetReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review

	if err := rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").Preload("Track").Preload("Track.Album").Preload("Track.Genres").Preload("Likes").First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// CreateReview creates a new review
func (rc *ReviewController) CreateReview(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		log.Printf("CreateReview: user not authenticated (no X-User-ID header)")
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Необходимо войти в систему для создания рецензии",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	log.Printf("CreateReview: user %d is creating a review", userID)

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON in CreateReview: %v", err)
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: fmt.Sprintf("Invalid request data: %v", err.Error()),
			Code:    http.StatusBadRequest,
		})
		return
	}

	log.Printf("CreateReview request: AlbumID=%v, TrackID=%v, Ratings=%d/%d/%d/%d, Atmosphere=%d",
		req.AlbumID, req.TrackID, req.RatingRhymes, req.RatingStructure, req.RatingImplementation, req.RatingIndividuality, req.AtmosphereRating)

	// Validate that either album_id or track_id is provided
	if req.AlbumID == nil && req.TrackID == nil {
		log.Printf("CreateReview: neither album_id nor track_id provided")
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Необходимо указать album_id или track_id",
			Code:    http.StatusBadRequest,
		})
		return
	}
	if req.AlbumID != nil && req.TrackID != nil {
		log.Printf("CreateReview: both album_id and track_id provided")
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Можно указать только album_id или track_id, но не оба одновременно",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Convert atmosphere rating (1-10) to multiplier (1.0000-1.6072)
	atmosphereMultiplier := convertAtmosphereToMultiplier(req.AtmosphereRating)

	// Validate review data
	review := models.Review{
		UserID:               userID,
		AlbumID:              req.AlbumID,
		TrackID:              req.TrackID,
		Text:                 req.Text,
		RatingRhymes:         req.RatingRhymes,
		RatingStructure:      req.RatingStructure,
		RatingImplementation: req.RatingImplementation,
		RatingIndividuality:  req.RatingIndividuality,
		AtmosphereMultiplier: atmosphereMultiplier,
	}

	log.Printf("Review before validation: UserID=%d, AlbumID=%v, TrackID=%v, Ratings=%d/%d/%d/%d, AtmosphereMultiplier=%f",
		review.UserID, review.AlbumID, review.TrackID, review.RatingRhymes, review.RatingStructure,
		review.RatingImplementation, review.RatingIndividuality, review.AtmosphereMultiplier)

	if err := utils.ValidateReview(&review); err != nil {
		log.Printf("Validation error in CreateReview: %v", err)
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Validation Error",
			Message: fmt.Sprintf("Ошибка валидации: %v", err.Error()),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check if album or track exists
	if req.AlbumID != nil {
		var album models.Album
		if err := rc.DB.First(&album, *req.AlbumID).Error; err != nil {
			log.Printf("Album %d not found: %v", *req.AlbumID, err)
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Bad Request",
				Message: fmt.Sprintf("Альбом с ID %d не найден", *req.AlbumID),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Check if user already has a review for this album
		var existingReview models.Review
		if err := rc.DB.Where("user_id = ? AND album_id = ? AND deleted_at IS NULL", userID, *req.AlbumID).First(&existingReview).Error; err == nil {
			log.Printf("User %d already has a review for album %d", userID, *req.AlbumID)
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Error:   "Conflict",
				Message: "У вас уже есть рецензия для этого альбома. Пожалуйста, отредактируйте существующую рецензию.",
				Code:    http.StatusConflict,
			})
			return
		}
	} else if req.TrackID != nil {
		var track models.Track
		if err := rc.DB.First(&track, *req.TrackID).Error; err != nil {
			log.Printf("Track %d not found: %v", *req.TrackID, err)
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Bad Request",
				Message: fmt.Sprintf("Трек с ID %d не найден", *req.TrackID),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Check if user already has a review for this track
		var existingReview models.Review
		if err := rc.DB.Where("user_id = ? AND track_id = ? AND deleted_at IS NULL", userID, *req.TrackID).First(&existingReview).Error; err == nil {
			log.Printf("User %d already has a review for track %d", userID, *req.TrackID)
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Error:   "Conflict",
				Message: "У вас уже есть рецензия для этого трека. Пожалуйста, отредактируйте существующую рецензию.",
				Code:    http.StatusConflict,
			})
			return
		}
	}

	// Calculate final score
	review.CalculateFinalScore()

	// Все новые рецензии проходят модерацию (для удобства тестирования)
	review.Status = models.ReviewStatusPending

	if err := rc.DB.Create(&review).Error; err != nil {
		// Log detailed error for debugging
		log.Printf("Error creating review: %v", err)
		log.Printf("Review data: UserID=%d, AlbumID=%v, TrackID=%v, Text=%s",
			review.UserID, review.AlbumID, review.TrackID, review.Text)

		// Provide more detailed error message
		errorMessage := "Failed to create review"
		if err.Error() != "" {
			errorMessage = fmt.Sprintf("Failed to create review: %v", err)
		}

		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: errorMessage,
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update album average rating if review is approved and is for an album
	if review.Status == models.ReviewStatusApproved && review.AlbumID != nil {
		albumController := &AlbumController{DB: rc.DB}
		if err := albumController.CalculateAverageRating(*review.AlbumID); err != nil {
			// Log error but don't fail the request
		}
	}

	// Update track average rating if review is approved and is for a track
	if review.Status == models.ReviewStatusApproved && review.TrackID != nil {
		trackController := &TrackController{DB: rc.DB}
		if err := trackController.CalculateAverageRating(*review.TrackID); err != nil {
			// Log error but don't fail the request
		}
	}

	// Preload relationships
	query := rc.DB.Preload("User").Preload("Likes")
	if review.AlbumID != nil {
		query = query.Preload("Album").Preload("Album.Genre")
	}
	if review.TrackID != nil {
		query = query.Preload("Track").Preload("Track.Album").Preload("Track.Genres")
	}
	query.First(&review, review.ID)
	c.JSON(http.StatusCreated, review)
}

// UpdateReview updates a review
func (rc *ReviewController) UpdateReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review

	if err := rc.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	user, _ := middleware.GetUserFromContext(c)
	// Check if user is owner or admin
	if review.UserID != userID && !user.IsAdmin {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{
			Error:   "Forbidden",
			Message: "You don't have permission to update this review",
			Code:    http.StatusForbidden,
		})
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Сохраняем исходные значения для проверки изменений
	originalText := review.Text
	textChanged := false

	// Обновляем текст только если поле было передано в запросе
	if req.Text != nil {
		newText := *req.Text
		if newText != originalText {
			textChanged = true
			review.Text = newText
		}
	}

	// Update ratings
	if req.RatingRhymes != 0 && req.RatingRhymes != review.RatingRhymes {
		review.RatingRhymes = req.RatingRhymes
	}
	if req.RatingStructure != 0 && req.RatingStructure != review.RatingStructure {
		review.RatingStructure = req.RatingStructure
	}
	if req.RatingImplementation != 0 && req.RatingImplementation != review.RatingImplementation {
		review.RatingImplementation = req.RatingImplementation
	}
	if req.RatingIndividuality != 0 && req.RatingIndividuality != review.RatingIndividuality {
		review.RatingIndividuality = req.RatingIndividuality
	}
	if req.AtmosphereRating != 0 {
		newMultiplier := convertAtmosphereToMultiplier(req.AtmosphereRating)
		if newMultiplier != review.AtmosphereMultiplier {
			review.AtmosphereMultiplier = newMultiplier
		}
	}

	// Логика изменения статуса для обычных пользователей:
	// - Если изменился текст → на модерацию
	// - Если изменились только оценки → статус не меняется (остаётся approved)
	// - Админ может редактировать без изменения статуса
	if !user.IsAdmin {
		if textChanged {
			// Если текст изменился, отправляем на модерацию
			review.Status = models.ReviewStatusPending
		}
		// Если изменились только оценки, статус остаётся как был (approved или pending)
	}
	// Админы могут редактировать без изменения статуса

	// Validate updated review
	if err := utils.ValidateReview(&review); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Recalculate final score
	review.CalculateFinalScore()

	if err := rc.DB.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update album average rating if review is for an album
	if review.AlbumID != nil {
		albumController := &AlbumController{DB: rc.DB}
		if err := albumController.CalculateAverageRating(*review.AlbumID); err != nil {
			// Log error but don't fail the request
		}
	}

	rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").First(&review, review.ID)
	c.JSON(http.StatusOK, review)
}

// DeleteReview deletes a review
func (rc *ReviewController) DeleteReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review

	if err := rc.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	user, _ := middleware.GetUserFromContext(c)
	// Check if user is owner or admin
	if review.UserID != userID && !user.IsAdmin {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{
			Error:   "Forbidden",
			Message: "You don't have permission to delete this review",
			Code:    http.StatusForbidden,
		})
		return
	}

	albumID := review.AlbumID
	if err := rc.DB.Delete(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update album average rating if review was for an album
	if albumID != nil {
		albumController := &AlbumController{DB: rc.DB}
		if err := albumController.CalculateAverageRating(*albumID); err != nil {
			// Log error but don't fail the request
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Review deleted successfully",
	})
}

// ApproveReview approves a review (admin only)
func (rc *ReviewController) ApproveReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review

	if err := rc.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	review.Status = models.ReviewStatusApproved
	review.ModeratedBy = &userID
	now := time.Now()
	review.ModeratedAt = &now

	if err := rc.DB.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to approve review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update album average rating if review is for an album
	if review.AlbumID != nil {
		albumController := &AlbumController{DB: rc.DB}
		if err := albumController.CalculateAverageRating(*review.AlbumID); err != nil {
			// Log error but don't fail the request
		}
	}

	rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").First(&review, review.ID)
	c.JSON(http.StatusOK, review)
}

// RejectReview rejects a review (admin only)
func (rc *ReviewController) RejectReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review

	if err := rc.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	review.Status = models.ReviewStatusRejected
	review.ModeratedBy = &userID
	now := time.Now()
	review.ModeratedAt = &now

	if err := rc.DB.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to reject review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").First(&review, review.ID)
	c.JSON(http.StatusOK, review)
}

// LikeReview adds a like to a review
func (rc *ReviewController) LikeReview(c *gin.Context) {
	reviewID := c.Param("id")
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Check if review exists
	var review models.Review
	if err := rc.DB.First(&review, reviewID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check if like already exists
	var existingLike models.ReviewLike
	if err := rc.DB.Where("user_id = ? AND review_id = ?", userID, reviewID).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already liked", "liked": true})
		return
	}

	// Create like
	like := models.ReviewLike{
		UserID:   userID,
		ReviewID: review.ID,
	}

	if err := rc.DB.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to like review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review liked", "liked": true})
}

// UnlikeReview removes a like from a review
func (rc *ReviewController) UnlikeReview(c *gin.Context) {
	reviewID := c.Param("id")
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Check if review exists
	var review models.Review
	if err := rc.DB.First(&review, reviewID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Review not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Delete like
	if err := rc.DB.Where("user_id = ? AND review_id = ?", userID, reviewID).Delete(&models.ReviewLike{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to unlike review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review unliked", "liked": false})
}

// GetPopularReviews retrieves most liked reviews from last 24 hours
func (rc *ReviewController) GetPopularReviews(c *gin.Context) {
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	// Get reviews from last 24 hours
	last24Hours := time.Now().Add(-24 * time.Hour)

	var reviews []models.Review
	// Get all approved reviews from last 24 hours with likes count, prioritizing reviews with albums
	query := rc.DB.Model(&models.Review{}).
		Preload("User").
		Preload("Album").
		Preload("Album.Genre").
		Preload("Track").
		Preload("Track.Album").
		Preload("Track.Genres").
		Preload("Likes").
		Where("status = ? AND created_at >= ?", models.ReviewStatusApproved, last24Hours).
		Where("album_id IS NOT NULL"). // Только рецензии с альбомами
		Order("created_at DESC").
		Limit(limit * 2) // Get more to sort by likes

	if err := query.Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch popular reviews",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Sort by likes count
	for i := 0; i < len(reviews); i++ {
		for j := i + 1; j < len(reviews); j++ {
			if len(reviews[i].Likes) < len(reviews[j].Likes) {
				reviews[i], reviews[j] = reviews[j], reviews[i]
			}
		}
	}

	// Limit results
	if len(reviews) > limit {
		reviews = reviews[:limit]
	}

	c.JSON(http.StatusOK, reviews)
}
