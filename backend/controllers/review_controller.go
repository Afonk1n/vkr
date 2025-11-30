package controllers

import (
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
	AlbumID              uint   `json:"album_id" binding:"required"`
	Text                 string `json:"text"`
	RatingRhymes         int    `json:"rating_rhymes" binding:"required,min=1,max=10"`
	RatingStructure      int    `json:"rating_structure" binding:"required,min=1,max=10"`
	RatingImplementation int    `json:"rating_implementation" binding:"required,min=1,max=10"`
	RatingIndividuality  int    `json:"rating_individuality" binding:"required,min=1,max=10"`
	AtmosphereRating     int    `json:"atmosphere_rating" binding:"required,min=1,max=10"` // 1-10, will be converted to multiplier
}

// UpdateReviewRequest represents review update request
type UpdateReviewRequest struct {
	Text                 string `json:"text"`
	RatingRhymes         int    `json:"rating_rhymes" binding:"min=1,max=10"`
	RatingStructure      int    `json:"rating_structure" binding:"min=1,max=10"`
	RatingImplementation int    `json:"rating_implementation" binding:"min=1,max=10"`
	RatingIndividuality  int    `json:"rating_individuality" binding:"min=1,max=10"`
	AtmosphereRating     int    `json:"atmosphere_rating" binding:"min=1,max=10"` // 1-10, will be converted to multiplier
}

// GetReviews retrieves list of reviews with filters
func (rc *ReviewController) GetReviews(c *gin.Context) {
	var reviews []models.Review
	query := rc.DB.Preload("User").Preload("Album").Preload("Album.Genre")

	// Filter by album
	if albumID := c.Query("album_id"); albumID != "" {
		query = query.Where("album_id = ?", albumID)
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
		"reviews": reviews,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}

// GetReview retrieves review by ID
func (rc *ReviewController) GetReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review

	if err := rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").First(&review, id).Error; err != nil {
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
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
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
		Text:                 req.Text,
		RatingRhymes:         req.RatingRhymes,
		RatingStructure:      req.RatingStructure,
		RatingImplementation: req.RatingImplementation,
		RatingIndividuality:  req.RatingIndividuality,
		AtmosphereMultiplier: atmosphereMultiplier,
	}

	if err := utils.ValidateReview(&review); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check if album exists
	var album models.Album
	if err := rc.DB.First(&album, req.AlbumID).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Album not found",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Calculate final score
	review.CalculateFinalScore()

	// Set status: if text is provided, status is pending (needs moderation)
	// Otherwise, status is approved
	if review.Text != "" {
		review.Status = models.ReviewStatusPending
	} else {
		review.Status = models.ReviewStatusApproved
	}

	if err := rc.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create review",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update album average rating if review is approved
	if review.Status == models.ReviewStatusApproved {
		albumController := &AlbumController{DB: rc.DB}
		if err := albumController.CalculateAverageRating(req.AlbumID); err != nil {
			// Log error but don't fail the request
			c.JSON(http.StatusCreated, review)
			return
		}
	}

	rc.DB.Preload("User").Preload("Album").Preload("Album.Genre").First(&review, review.ID)
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

	// Update fields
	if req.Text != "" {
		review.Text = req.Text
		// If text is added/changed, status should be pending
		review.Status = models.ReviewStatusPending
	}
	if req.RatingRhymes != 0 {
		review.RatingRhymes = req.RatingRhymes
	}
	if req.RatingStructure != 0 {
		review.RatingStructure = req.RatingStructure
	}
	if req.RatingImplementation != 0 {
		review.RatingImplementation = req.RatingImplementation
	}
	if req.RatingIndividuality != 0 {
		review.RatingIndividuality = req.RatingIndividuality
	}
	if req.AtmosphereRating != 0 {
		// Convert atmosphere rating to multiplier
		review.AtmosphereMultiplier = convertAtmosphereToMultiplier(req.AtmosphereRating)
	}

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

	// Update album average rating
	albumController := &AlbumController{DB: rc.DB}
	if err := albumController.CalculateAverageRating(review.AlbumID); err != nil {
		// Log error but don't fail the request
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

	// Update album average rating
	albumController := &AlbumController{DB: rc.DB}
	if err := albumController.CalculateAverageRating(albumID); err != nil {
		// Log error but don't fail the request
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

	// Update album average rating
	albumController := &AlbumController{DB: rc.DB}
	if err := albumController.CalculateAverageRating(review.AlbumID); err != nil {
		// Log error but don't fail the request
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

