package controllers

import (
	"encoding/json"
	"fmt"
	"math"
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

	badges := uc.CalculateUserBadges(user.ID)
	stats := uc.CalculateUserStats(user.ID)
	genreStats := uc.CalculateGenreStats(user.ID)
	favoriteAlbums := uc.GetFavoriteAlbums(user.FavoriteAlbumIDs)
	favoriteArtists := uc.GetFavoriteArtists(user.FavoriteArtists)
	favoriteTracks := uc.GetFavoriteTracks(user.FavoriteTrackIDs)

	var followersCount, followingCount int64
	uc.DB.Model(&models.UserFollow{}).Where("following_id = ?", user.ID).Count(&followersCount)
	uc.DB.Model(&models.UserFollow{}).Where("follower_id = ?", user.ID).Count(&followingCount)

	userResponse := gin.H{
		"id":                 user.ID,
		"username":           user.Username,
		"email":              user.Email,
		"avatar_path":        user.AvatarPath,
		"bio":                user.Bio,
		"social_links":       user.SocialLinks,
		"is_admin":           user.IsAdmin,
		"is_verified_artist": user.IsVerifiedArtist,
		"favorite_album_ids": user.FavoriteAlbumIDs,
		"favorite_artists":   favoriteArtists,
		"favorite_track_ids": user.FavoriteTrackIDs,
		"created_at":         user.CreatedAt,
		"updated_at":         user.UpdatedAt,
		"badges":             badges,
		"stats":              stats,
		"genre_stats":        genreStats,
		"favorite_albums":    favoriteAlbums,
		"favorite_tracks":    favoriteTracks,
		"followers_count":    followersCount,
		"following_count":    followingCount,
	}

	isFollowing := false
	if viewerID, ok := middleware.GetUserIDFromContext(c); ok && viewerID != user.ID {
		var fc int64
		uc.DB.Model(&models.UserFollow{}).Where("follower_id = ? AND following_id = ?", viewerID, user.ID).Count(&fc)
		isFollowing = fc > 0
	}
	userResponse["is_following"] = isFollowing

	c.JSON(http.StatusOK, userResponse)
}

// FollowUser subscribes the current user to another user.
func (uc *UserController) FollowUser(c *gin.Context) {
	targetID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Bad Request", Message: "Некорректный id", Code: http.StatusBadRequest})
		return
	}
	targetID := uint(targetID64)
	followerID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Unauthorized", Message: "Нужна авторизация", Code: http.StatusUnauthorized})
		return
	}
	if targetID == followerID {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Bad Request", Message: "Нельзя подписаться на себя", Code: http.StatusBadRequest})
		return
	}
	var target models.User
	if err := uc.DB.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Not Found", Message: "Пользователь не найден", Code: http.StatusNotFound})
		return
	}
	var existing models.UserFollow
	if err := uc.DB.Where("follower_id = ? AND following_id = ?", followerID, targetID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"following": true})
		return
	}
	uf := models.UserFollow{FollowerID: followerID, FollowingID: targetID}
	if err := uc.DB.Create(&uf).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Internal Server Error", Message: "Не удалось подписаться", Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": true})
}

// UnfollowUser removes subscription to another user.
func (uc *UserController) UnfollowUser(c *gin.Context) {
	targetID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Bad Request", Message: "Некорректный id", Code: http.StatusBadRequest})
		return
	}
	targetID := uint(targetID64)
	followerID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "Unauthorized", Message: "Нужна авторизация", Code: http.StatusUnauthorized})
		return
	}
	res := uc.DB.Where("follower_id = ? AND following_id = ?", followerID, targetID).Delete(&models.UserFollow{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Internal Server Error", Message: "Не удалось отписаться", Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": false})
}

// SetFavoriteAlbums sets up to 3 favorite albums, artists and tracks for a user.
func (uc *UserController) SetFavoriteAlbums(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := uc.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Not Found", Message: "User not found", Code: http.StatusNotFound})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists || user.ID != userID {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "Forbidden", Message: "Not allowed", Code: http.StatusForbidden})
		return
	}

	var req struct {
		AlbumIDs    []uint   `json:"album_ids"`
		ArtistNames []string `json:"artist_names"`
		TrackIDs    []uint   `json:"track_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Bad Request", Message: err.Error(), Code: http.StatusBadRequest})
		return
	}
	if len(req.AlbumIDs) > 3 {
		req.AlbumIDs = req.AlbumIDs[:3]
	}
	if len(req.ArtistNames) > 3 {
		req.ArtistNames = req.ArtistNames[:3]
	}
	if len(req.TrackIDs) > 3 {
		req.TrackIDs = req.TrackIDs[:3]
	}

	idsJSON, _ := json.Marshal(req.AlbumIDs)
	user.FavoriteAlbumIDs = string(idsJSON)
	artistsJSON, _ := json.Marshal(cleanArtistNames(req.ArtistNames))
	user.FavoriteArtists = string(artistsJSON)
	trackIDsJSON, _ := json.Marshal(req.TrackIDs)
	user.FavoriteTrackIDs = string(trackIDsJSON)
	uc.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"favorite_albums":  uc.GetFavoriteAlbums(user.FavoriteAlbumIDs),
		"favorite_artists": uc.GetFavoriteArtists(user.FavoriteArtists),
		"favorite_tracks":  uc.GetFavoriteTracks(user.FavoriteTrackIDs),
	})
}

func cleanArtistNames(names []string) []string {
	cleaned := make([]string, 0, len(names))
	seen := make(map[string]bool)
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		cleaned = append(cleaned, name)
	}
	if len(cleaned) > 3 {
		return cleaned[:3]
	}
	return cleaned
}

// GetFavoriteAlbums loads album objects from a JSON IDs string
func (uc *UserController) GetFavoriteAlbums(idsJSON string) []models.Album {
	if idsJSON == "" || idsJSON == "[]" || idsJSON == "null" {
		return []models.Album{}
	}
	var ids []uint
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil || len(ids) == 0 {
		return []models.Album{}
	}
	var albums []models.Album
	uc.DB.Preload("Genre").Where("id IN ?", ids).Find(&albums)
	// Preserve order
	ordered := make([]models.Album, 0, len(ids))
	albumMap := make(map[uint]models.Album)
	for _, a := range albums {
		albumMap[a.ID] = a
	}
	for _, id := range ids {
		if a, ok := albumMap[id]; ok {
			ordered = append(ordered, a)
		}
	}
	return ordered
}

func (uc *UserController) GetFavoriteArtists(namesJSON string) []ArtistSearchResult {
	if namesJSON == "" || namesJSON == "[]" || namesJSON == "null" {
		return []ArtistSearchResult{}
	}
	var names []string
	if err := json.Unmarshal([]byte(namesJSON), &names); err != nil || len(names) == 0 {
		return []ArtistSearchResult{}
	}

	result := make([]ArtistSearchResult, 0, len(names))
	for _, name := range cleanArtistNames(names) {
		var count int64
		var firstAlbum models.Album
		uc.DB.Model(&models.Album{}).Where("artist = ?", name).Count(&count)
		uc.DB.Where("artist = ?", name).Order("created_at ASC").First(&firstAlbum)
		result = append(result, ArtistSearchResult{
			Name:           name,
			Count:          int(count),
			CoverImagePath: firstAlbum.CoverImagePath,
		})
	}
	return result
}

func (uc *UserController) GetFavoriteTracks(idsJSON string) []TrackSearchResult {
	if idsJSON == "" || idsJSON == "[]" || idsJSON == "null" {
		return []TrackSearchResult{}
	}
	var ids []uint
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil || len(ids) == 0 {
		return []TrackSearchResult{}
	}
	var tracks []models.Track
	uc.DB.Preload("Album").Where("id IN ?", ids).Find(&tracks)
	trackMap := make(map[uint]models.Track)
	for _, track := range tracks {
		trackMap[track.ID] = track
	}

	ordered := make([]TrackSearchResult, 0, len(ids))
	for _, id := range ids {
		if track, ok := trackMap[id]; ok {
			cover := track.CoverImagePath
			if cover == "" {
				cover = track.Album.CoverImagePath
			}
			ordered = append(ordered, TrackSearchResult{
				ID:             track.ID,
				Title:          track.Title,
				AlbumID:        track.AlbumID,
				AlbumTitle:     track.Album.Title,
				Artist:         track.Album.Artist,
				CoverImagePath: cover,
			})
		}
	}
	return ordered
}

type UserStats struct {
	TotalReviews         int     `json:"total_reviews"`
	RatingsWithoutReview int64   `json:"ratings_without_review"`
	AvgScore             float64 `json:"avg_score"`
	TotalLikesReceived   int64   `json:"total_likes_received"`
	TotalLikesGiven      int64   `json:"total_likes_given"`
	AuthorLikesReceived  int64   `json:"author_likes_received"`
	TopGenre             string  `json:"top_genre"`
}

// CalculateUserStats returns profile statistics for a user
func (uc *UserController) CalculateUserStats(userID uint) UserStats {
	var stats UserStats

	var reviews []models.Review
	uc.DB.Where("user_id = ? AND status = ?", userID, models.ReviewStatusApproved).Find(&reviews)
	stats.TotalReviews = len(reviews)

	if stats.TotalReviews > 0 {
		var totalScore float64
		for _, r := range reviews {
			totalScore += r.FinalScore
		}
		stats.AvgScore = math.Round(totalScore/float64(stats.TotalReviews)*10) / 10
	}

	var reviewIDs []uint
	for _, r := range reviews {
		reviewIDs = append(reviewIDs, r.ID)
	}
	if len(reviewIDs) > 0 {
		uc.DB.Model(&models.ReviewLike{}).Where("review_id IN ?", reviewIDs).Count(&stats.TotalLikesReceived)
	}
	uc.DB.Model(&models.Review{}).
		Where("user_id = ? AND status = ? AND btrim(coalesce(text, '')) = ''", userID, models.ReviewStatusApproved).
		Count(&stats.RatingsWithoutReview)
	uc.DB.Model(&models.ReviewLike{}).Where("user_id = ?", userID).Count(&stats.TotalLikesGiven)

	genreStats := uc.CalculateGenreStats(userID)
	if len(genreStats) > 0 {
		stats.TopGenre = genreStats[0].Name
	}

	return stats
}

type GenreStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// CalculateGenreStats returns review counts per genre for radar chart
func (uc *UserController) CalculateGenreStats(userID uint) []GenreStat {
	var reviews []models.Review
	uc.DB.Preload("Album").Preload("Album.Genre").Preload("Track").Preload("Track.Genres").
		Where("user_id = ? AND status = ?", userID, models.ReviewStatusApproved).
		Find(&reviews)

	counts := make(map[string]int)
	for _, r := range reviews {
		if r.AlbumID != nil && r.Album != nil && r.Album.Genre.ID > 0 {
			counts[r.Album.Genre.Name]++
		}
		if r.TrackID != nil && r.Track != nil {
			for _, g := range r.Track.Genres {
				if g.ID > 0 {
					counts[g.Name]++
				}
			}
		}
	}

	result := make([]GenreStat, 0, len(counts))
	for name, count := range counts {
		result = append(result, GenreStat{Name: name, Count: count})
	}
	// Sort descending by count
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Count > result[i].Count {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	// Return top 8 genres max
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

// GetUserLikedReviews retrieves reviews liked by a user.
func (uc *UserController) GetUserLikedReviews(c *gin.Context) {
	id := c.Param("id")
	var likes []models.ReviewLike

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := uc.DB.
		Preload("Review.User").
		Preload("Review.Album").
		Preload("Review.Album.Genre").
		Preload("Review.Track").
		Preload("Review.Track.Album").
		Preload("Review.Likes").
		Where("user_id = ?", id).
		Order("created_at desc")

	var total int64
	query.Model(&models.ReviewLike{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Find(&likes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch liked reviews",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	reviews := make([]models.Review, 0, len(likes))
	for _, like := range likes {
		if like.Review.ID != 0 {
			reviews = append(reviews, like.Review)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":   reviews,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
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

	badges := uc.CalculateUserBadges(user.ID)
	stats := uc.CalculateUserStats(user.ID)
	genreStats := uc.CalculateGenreStats(user.ID)
	favoriteAlbums := uc.GetFavoriteAlbums(user.FavoriteAlbumIDs)
	favoriteArtists := uc.GetFavoriteArtists(user.FavoriteArtists)
	favoriteTracks := uc.GetFavoriteTracks(user.FavoriteTrackIDs)

	userResponse := gin.H{
		"id":                 user.ID,
		"username":           user.Username,
		"email":              user.Email,
		"avatar_path":        user.AvatarPath,
		"bio":                user.Bio,
		"social_links":       user.SocialLinks,
		"is_admin":           user.IsAdmin,
		"is_verified_artist": user.IsVerifiedArtist,
		"favorite_album_ids": user.FavoriteAlbumIDs,
		"favorite_artists":   favoriteArtists,
		"favorite_track_ids": user.FavoriteTrackIDs,
		"created_at":         user.CreatedAt,
		"updated_at":         user.UpdatedAt,
		"badges":             badges,
		"stats":              stats,
		"genre_stats":        genreStats,
		"favorite_albums":    favoriteAlbums,
		"favorite_tracks":    favoriteTracks,
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
	Criteria    string `json:"criteria"` // как получить звание (для подсказки в UI)
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
			Criteria:    "Учитываются только одобренные рецензии. Звание при 51 и более таких рецензиях.",
			Icon:        "👑",
			Priority:    1,
		})
	} else if totalReviews >= 21 {
		badges = append(badges, Badge{
			Name:        "Мастер рецензий",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Criteria:    "Учитываются только одобренные рецензии. Звание при 21–50 рецензиях включительно.",
			Icon:        "⭐",
			Priority:    2,
		})
	} else if totalReviews >= 6 {
		badges = append(badges, Badge{
			Name:        "Опытный критик",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Criteria:    "Учитываются только одобренные рецензии. Звание при 6–20 рецензиях включительно.",
			Icon:        "📝",
			Priority:    3,
		})
	} else if totalReviews >= 1 {
		badges = append(badges, Badge{
			Name:        "Начинающий критик",
			Description: fmt.Sprintf("%d рецензий", totalReviews),
			Criteria:    "Учитываются только одобренные рецензии. Звание с первой опубликованной и одобренной рецензии.",
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
				Criteria:    fmt.Sprintf("Не менее 5 одобренных рецензий, в которых указан жанр «%s» (альбом или трек).", genreName),
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
			Criteria:    "В одобренных рецензиях встречается не менее 5 разных жанров (по данным альбомов и треков).",
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
					Criteria:    fmt.Sprintf("Не менее 80%% одобренных рецензий относятся к жанру «%s».", genreName),
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
