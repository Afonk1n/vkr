package middleware

import (
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware checks if user is authenticated
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from header or session
		// For simplicity, we'll use a simple header-based auth
		// In production, use JWT tokens
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User ID is required",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid user ID",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Store user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Next()
	}
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't require authentication
func OptionalAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr != "" {
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err == nil {
				var user models.User
				if err := db.First(&user, uint(userID)).Error; err == nil {
					c.Set("user", user)
					c.Set("user_id", user.ID)
				}
			}
		}
		c.Next()
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		userModel, ok := user.(models.User)
		if !ok || !userModel.IsAdmin {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Error:   "Forbidden",
				Message: "Admin access required",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserFromContext gets user from context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userModel, ok := user.(models.User)
	if !ok {
		return nil, false
	}

	return &userModel, true
}

// GetUserIDFromContext gets user ID from context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return userIDUint, true
}
