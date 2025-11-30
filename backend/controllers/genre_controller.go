package controllers

import (
	"music-review-site/backend/models"
	"music-review-site/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GenreController struct {
	DB *gorm.DB
}

// CreateGenreRequest represents genre creation request
type CreateGenreRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateGenreRequest represents genre update request
type UpdateGenreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetGenres retrieves list of all genres
func (gc *GenreController) GetGenres(c *gin.Context) {
	var genres []models.Genre

	if err := gc.DB.Find(&genres).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch genres",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, genres)
}

// GetGenre retrieves genre by ID
func (gc *GenreController) GetGenre(c *gin.Context) {
	id := c.Param("id")
	var genre models.Genre

	if err := gc.DB.First(&genre, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Genre not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, genre)
}

// CreateGenre creates a new genre
func (gc *GenreController) CreateGenre(c *gin.Context) {
	var req CreateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	genre := models.Genre{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := gc.DB.Create(&genre).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create genre",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, genre)
}

// UpdateGenre updates a genre
func (gc *GenreController) UpdateGenre(c *gin.Context) {
	id := c.Param("id")
	var genre models.Genre

	if err := gc.DB.First(&genre, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Genre not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	var req UpdateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Update fields
	if req.Name != "" {
		genre.Name = req.Name
	}
	if req.Description != "" {
		genre.Description = req.Description
	}

	if err := gc.DB.Save(&genre).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update genre",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, genre)
}

// DeleteGenre deletes a genre
func (gc *GenreController) DeleteGenre(c *gin.Context) {
	id := c.Param("id")
	var genre models.Genre

	if err := gc.DB.First(&genre, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error:   "Not Found",
			Message: "Genre not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	if err := gc.DB.Delete(&genre).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete genre",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Genre deleted successfully",
	})
}

