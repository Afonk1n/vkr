package middleware

import (
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware checks if user is authenticated
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveAuthenticatedUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Session token is required",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
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
		userID, ok := resolveAuthenticatedUserID(c)
		if ok {
			var user models.User
			if err := db.First(&user, userID).Error; err == nil {
				c.Set("user", user)
				c.Set("user_id", user.ID)
			}
		}
		c.Next()
	}
}

func resolveAuthenticatedUserID(c *gin.Context) (uint, bool) {
	if token := bearerToken(c.GetHeader("Authorization")); token != "" {
		if userID, err := utils.ValidateSessionToken(token); err == nil {
			return userID, true
		}
	}

	if !allowUserIDHeaderFallback() {
		return 0, false
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		return 0, false
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(userID), true
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func allowUserIDHeaderFallback() bool {
	val := strings.TrimSpace(os.Getenv("AUTH_ALLOW_USER_ID_HEADER"))
	if val != "" {
		switch strings.ToLower(val) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) != "prod"
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
