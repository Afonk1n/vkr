package controllers

import (
	"encoding/json"
	"fmt"
	"music-review-site/backend/middleware"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

// GetUser retrieves user by ID
func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := uc.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	user.Password = ""
	
	// Calculate badges
	badges := uc.CalculateUserBadges(user.ID)
	userResponse := gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"avatar_path":  user.AvatarPath,
		"bio":          user.Bio,
		"social_links": user.SocialLinks,
		"is_admin":     user.IsAdmin,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
		"badges":       badges,
	}
	
	c.JSON(http.StatusOK, userResponse)
}

// GetUserReviews retrieves reviews by user ID
func (uc *UserController) GetUserReviews(c *gin.Context) {
	id := c.Param("id")
	var reviews []models.Review

	query := uc.DB.Preload("User").Preload("Album").Preload("Album.Genre").Preload("Track").Preload("Track.Album").Preload("Likes").Where("user_id = ?", id)

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
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

// UpdateUser updates user profile
func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := uc.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
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

	// Check if user is updating their own profile or is admin
	userModel, _ := middleware.GetUserFromContext(c)
	if user.ID != userID && !userModel.IsAdmin {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{
			Error:   "Forbidden",
			Message: "You don't have permission to update this user",
			Code:    http.StatusForbidden,
		})
		return
	}

	var req struct {
		Username    string            `json:"username"`
		Email       string            `json:"email"`
		AvatarPath  string            `json:"avatar_path"`
		Bio         string            `json:"bio"`
		SocialLinks map[string]string `json:"social_links"` // {"vk": "", "telegram": "", "instagram": ""}
		Password    string            `json:"password"`     // For password change
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Update username if provided
	if req.Username != "" {
		if err := utils.ValidateUsername(req.Username); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Validation Error",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		user.Username = req.Username
	}

	// Update email if provided
	if req.Email != "" {
		if !utils.ValidateEmail(req.Email) {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Validation Error",
				Message: "Invalid email format",
				Code:    http.StatusBadRequest,
			})
			return
		}
		user.Email = req.Email
	}

	// Update avatar path if provided
	if req.AvatarPath != "" {
		user.AvatarPath = req.AvatarPath
	}

	// Update bio if provided
	user.Bio = req.Bio

	// Update social links if provided
	if req.SocialLinks != nil {
		socialLinksJSON, err := json.Marshal(req.SocialLinks)
		if err == nil {
			user.SocialLinks = string(socialLinksJSON)
		}
	}

	// Update password if provided
	if req.Password != "" {
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Error:   "Validation Error",
				Message: "Password must be at least 6 characters",
				Code:    http.StatusBadRequest,
			})
			return
		}
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to hash password",
				Code:    http.StatusInternalServerError,
			})
			return
		}
		user.Password = hashedPassword
	}

	if err := uc.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update user",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	user.Password = ""
	
	// Calculate badges
	badges := uc.CalculateUserBadges(user.ID)
	userResponse := gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"avatar_path":  user.AvatarPath,
		"bio":          user.Bio,
		"social_links": user.SocialLinks,
		"is_admin":     user.IsAdmin,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
		"badges":       badges,
	}
	
	c.JSON(http.StatusOK, userResponse)
}

// DeleteUser deletes a user
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := uc.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
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

	// Check if user is deleting their own profile or is admin
	userModel, _ := middleware.GetUserFromContext(c)
	if user.ID != userID && !userModel.IsAdmin {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{
			Error:   "Forbidden",
			Message: "You don't have permission to delete this user",
			Code:    http.StatusForbidden,
		})
		return
	}

	if err := uc.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete user",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// Badge represents a user badge/achievement
type Badge struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Priority    int    `json:"priority"`
}

// CalculateUserBadges calculates badges for a user based on their reviews
func (uc *UserController) CalculateUserBadges(userID uint) []Badge {
	var reviews []models.Review
	// Get all approved reviews with genre information
	if err := uc.DB.Preload("Album").Preload("Album.Genre").Preload("Track").Preload("Track.Genres").
		Where("user_id = ? AND status = ?", userID, models.ReviewStatusApproved).
		Find(&reviews).Error; err != nil {
		return []Badge{}
	}

	if len(reviews) == 0 {
		return []Badge{}
	}

	// Count reviews by genre
	genreCounts := make(map[string]int)
	totalReviews := len(reviews)
	uniqueGenres := make(map[string]bool)

	for _, review := range reviews {
		var genres []string
		
		// Get genres from album or track
		if review.AlbumID != nil && review.Album != nil && review.Album.Genre.ID > 0 {
			genres = append(genres, review.Album.Genre.Name)
			uniqueGenres[review.Album.Genre.Name] = true
		}
		if review.TrackID != nil && review.Track != nil {
			for _, genre := range review.Track.Genres {
				if genre.ID > 0 {
					genres = append(genres, genre.Name)
					uniqueGenres[genre.Name] = true
				}
			}
		}

		// Count each genre (if review has multiple genres, count each)
		for _, genreName := range genres {
			genreCounts[genreName]++
		}
	}

	var badges []Badge

	// Badges by total count
	if totalReviews >= 51 {
		badges = append(badges, Badge{
			Name:        "Легенда критики",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Icon:        "👑",
			Priority:    1,
		})
	} else if totalReviews >= 21 {
		badges = append(badges, Badge{
			Name:        "Мастер рецензий",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Icon:        "⭐",
			Priority:    2,
		})
	} else if totalReviews >= 6 {
		badges = append(badges, Badge{
			Name:        "Опытный критик",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Icon:        "📝",
			Priority:    3,
		})
	} else if totalReviews >= 1 {
		badges = append(badges, Badge{
			Name:        "Начинающий критик",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Icon:        "🌱",
			Priority:    4,
		})
	}

	// Badges by genre (5+ reviews in a genre)
	genreIcons := map[string]string{
		"Джаз":         "🎷",
		"Поп":          "🎤",
		"Рок":          "🎸",
		"Электронная":  "🎹",
		"Хип-хоп":      "🥁",
		"Классическая": "🎻",
	}

	genreNames := map[string]string{
		"Джаз":         "Джазовый критик",
		"Поп":          "Поп-эксперт",
		"Рок":          "Рок-ценитель",
		"Электронная":  "Электронный знаток",
		"Хип-хоп":      "Хип-хоп критик",
		"Классическая": "Классический знаток",
	}

	for genreName, count := range genreCounts {
		if count >= 5 {
			icon := genreIcons[genreName]
			if icon == "" {
				icon = "🎵"
			}
			badgeName := genreNames[genreName]
			if badgeName == "" {
				badgeName = genreName + " критик"
			}
			badges = append(badges, Badge{
				Name:        badgeName,
				Description: fmt.Sprintf("%d рецензий на %s", count, genreName),
				Icon:        icon,
				Priority:    2, // Genre badges have higher priority than count badges
			})
		}
	}

	// Badge for diversity (5+ different genres)
	if len(uniqueGenres) >= 5 {
		badges = append(badges, Badge{
			Name:        "Универсал",
			Description: fmt.Sprintf("Рецензии на %d разных жанров", len(uniqueGenres)),
			Icon:        "🌈",
			Priority:    3,
		})
	}

	// Badge for specialization (80%+ reviews in one genre)
	if totalReviews > 0 {
		for genreName, count := range genreCounts {
			percentage := float64(count) / float64(totalReviews) * 100
			if percentage >= 80 {
				icon := genreIcons[genreName]
				if icon == "" {
					icon = "🎯"
				}
				badgeName := genreNames[genreName]
				if badgeName == "" {
					badgeName = genreName + " специалист"
				}
				badges = append(badges, Badge{
					Name:        badgeName + " (Специалист)",
					Description: fmt.Sprintf("%.0f%% рецензий на %s", percentage, genreName),
					Icon:        icon,
					Priority:    1, // Specialization has highest priority
				})
				break // Only one specialization badge
			}
		}
	}

	// Sort badges by priority (lower number = higher priority)
	for i := 0; i < len(badges)-1; i++ {
		for j := i + 1; j < len(badges); j++ {
			if badges[i].Priority > badges[j].Priority {
				badges[i], badges[j] = badges[j], badges[i]
			}
		}
	}

	return badges
}

// UploadAvatar handles avatar file upload
func (uc *UserController) UploadAvatar(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := uc.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
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

	// Check if user is updating their own profile or is admin
	userModel, _ := middleware.GetUserFromContext(c)
	if user.ID != userID && !userModel.IsAdmin {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{
			Error:   "Forbidden",
			Message: "You don't have permission to update this user",
			Code:    http.StatusForbidden,
		})
		return
	}

	// Get file from form
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "No file provided",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "File size exceeds 5MB limit",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".webp"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid file format. Allowed: jpg, jpeg, png, webp",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create avatars directory if it doesn't exist
	avatarsDir := "../frontend/public/avatars"
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create avatars directory",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("user_%d_%d%s", user.ID, timestamp, ext)
	filePath := filepath.Join(avatarsDir, filename)

	// Delete old avatar if exists
	if user.AvatarPath != "" && strings.HasPrefix(user.AvatarPath, "/avatars/") {
		oldFilePath := filepath.Join(avatarsDir, filepath.Base(user.AvatarPath))
		if _, err := os.Stat(oldFilePath); err == nil {
			os.Remove(oldFilePath)
		}
	}

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to save file",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update user avatar path
	user.AvatarPath = "/avatars/" + filename
	if err := uc.DB.Save(&user).Error; err != nil {
		// Try to delete uploaded file if DB update fails
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update user avatar",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}
