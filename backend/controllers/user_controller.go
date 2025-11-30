package controllers

import (
	"encoding/json"
	"music-review-site/backend/middleware"
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"strconv"

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
	c.JSON(http.StatusOK, user)
}

// GetUserReviews retrieves reviews by user ID
func (uc *UserController) GetUserReviews(c *gin.Context) {
	id := c.Param("id")
	var reviews []models.Review

	query := uc.DB.Preload("Album").Preload("Album.Genre").Where("user_id = ?", id)

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
		"reviews": reviews,
		"total":   total,
		"page":    page,
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
	c.JSON(http.StatusOK, user)
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

