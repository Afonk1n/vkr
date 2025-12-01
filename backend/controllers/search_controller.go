package controllers

import (
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchController struct {
	DB *gorm.DB
}

// ArtistSearchResult represents artist search result
type ArtistSearchResult struct {
	Name  string `json:"name"`
	Count int    `json:"count"` // Number of albums
}

// SearchResponse represents search results
type SearchResponse struct {
	Artists []ArtistSearchResult `json:"artists"`
	Albums  []models.Album       `json:"albums"`
	Tracks  []TrackSearchResult  `json:"tracks"`
}

// TrackSearchResult represents track with album info for search
type TrackSearchResult struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	AlbumID        uint   `json:"album_id"`
	AlbumTitle     string `json:"album_title"`
	Artist         string `json:"artist"`
	CoverImagePath string `json:"cover_image_path"`
}

// Search performs search across albums and tracks
func (sc *SearchController) Search(c *gin.Context) {
	query := c.Query("q")
	limit := 5 // Limit results for autocomplete

	if query == "" {
		c.JSON(http.StatusOK, SearchResponse{
			Artists: []ArtistSearchResult{},
			Albums:  []models.Album{},
			Tracks:  []TrackSearchResult{},
		})
		return
	}

	// Search for unique artists
	var artistResults []struct {
		Artist string
		Count  int64
	}
	artistQuery := sc.DB.Model(&models.Album{}).
		Select("artist, COUNT(*) as count").
		Where("artist ILIKE ?", "%"+query+"%").
		Group("artist").
		Order("count DESC").
		Limit(limit)

	if err := artistQuery.Scan(&artistResults).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to search artists",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	artists := make([]ArtistSearchResult, len(artistResults))
	for i, result := range artistResults {
		artists[i] = ArtistSearchResult{
			Name:  result.Artist,
			Count: int(result.Count),
		}
	}

	var albums []models.Album
	albumQuery := sc.DB.Model(&models.Album{}).
		Preload("Genre").
		Where("title ILIKE ? OR artist ILIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Order("created_at DESC")

	if err := albumQuery.Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to search albums",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var tracks []models.Track
	trackQuery := sc.DB.Model(&models.Track{}).
		Preload("Album").
		Joins("JOIN albums ON tracks.album_id = albums.id").
		Where("tracks.title ILIKE ? OR albums.title ILIKE ? OR albums.artist ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Order("tracks.created_at DESC")

	if err := trackQuery.Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to search tracks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert tracks to search results
	trackResults := make([]TrackSearchResult, len(tracks))
	for i, track := range tracks {
		// Use track cover if available, otherwise use album cover
		coverImagePath := track.CoverImagePath
		if coverImagePath == "" {
			coverImagePath = track.Album.CoverImagePath
		}

		trackResults[i] = TrackSearchResult{
			ID:             track.ID,
			Title:          track.Title,
			AlbumID:        track.AlbumID,
			AlbumTitle:     track.Album.Title,
			Artist:         track.Album.Artist,
			CoverImagePath: coverImagePath,
		}
	}

	c.JSON(http.StatusOK, SearchResponse{
		Artists: artists,
		Albums:  albums,
		Tracks:  trackResults,
	})
}
