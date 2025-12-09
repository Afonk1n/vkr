package routes

import (
	"music-review-site/backend/controllers"
	"music-review-site/backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all routes
func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize controllers
	authController := &controllers.AuthController{DB: db}
	albumController := &controllers.AlbumController{DB: db}
	reviewController := &controllers.ReviewController{DB: db}
	genreController := &controllers.GenreController{DB: db}
	userController := &controllers.UserController{DB: db}
	trackController := &controllers.TrackController{DB: db}
	searchController := &controllers.SearchController{DB: db}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.GET("/me", middleware.AuthMiddleware(db), authController.GetMe)
		}

		// Genre routes
		genres := api.Group("/genres")
		{
			genres.GET("", genreController.GetGenres)
			genres.GET("/:id", genreController.GetGenre)
			genres.POST("", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), genreController.CreateGenre)
			genres.PUT("/:id", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), genreController.UpdateGenre)
			genres.DELETE("/:id", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), genreController.DeleteGenre)
		}

		// Album routes
		albums := api.Group("/albums")
		{
			albums.GET("", albumController.GetAlbums)
			// More specific routes must come before /:id
			albums.GET("/artist/:name", albumController.GetAlbumsByArtist)
			albums.GET("/:id/tracks", trackController.GetTracks)
			albums.GET("/:id", albumController.GetAlbum)
			albums.POST("", middleware.AuthMiddleware(db), albumController.CreateAlbum)
			albums.PUT("/:id", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), albumController.UpdateAlbum)
			albums.DELETE("/:id", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), albumController.DeleteAlbum)
			// Like routes
			albums.POST("/:id/like", middleware.AuthMiddleware(db), albumController.LikeAlbum)
			albums.DELETE("/:id/like", middleware.AuthMiddleware(db), albumController.UnlikeAlbum)
		}

		// Review routes
		reviews := api.Group("/reviews")
		{
			reviews.GET("", reviewController.GetReviews)
			reviews.GET("/popular", reviewController.GetPopularReviews)
			reviews.GET("/:id", reviewController.GetReview)
			reviews.POST("", middleware.AuthMiddleware(db), reviewController.CreateReview)
			reviews.PUT("/:id", middleware.AuthMiddleware(db), reviewController.UpdateReview)
			reviews.DELETE("/:id", middleware.AuthMiddleware(db), reviewController.DeleteReview)
			
			// Like routes
			reviews.POST("/:id/like", middleware.AuthMiddleware(db), reviewController.LikeReview)
			reviews.DELETE("/:id/like", middleware.AuthMiddleware(db), reviewController.UnlikeReview)
			
			// Moderation routes (admin only)
			reviews.POST("/:id/approve", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), reviewController.ApproveReview)
			reviews.POST("/:id/reject", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), reviewController.RejectReview)
		}

		// Track routes
		tracks := api.Group("/tracks")
		{
			tracks.GET("", trackController.GetAllTracks) // Must come before /:id
			tracks.GET("/popular", trackController.GetPopularTracks)
			tracks.GET("/:id", trackController.GetTrack)
			tracks.POST("", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), trackController.CreateTrack)
			tracks.PUT("/:id", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), trackController.UpdateTrack)
			tracks.DELETE("/:id", middleware.AuthMiddleware(db), middleware.AdminMiddleware(), trackController.DeleteTrack)
			// Like routes
			tracks.POST("/:id/like", middleware.AuthMiddleware(db), trackController.LikeTrack)
			tracks.DELETE("/:id/like", middleware.AuthMiddleware(db), trackController.UnlikeTrack)
		}

		// Search routes
		api.GET("/search", searchController.Search)

		// User routes
		users := api.Group("/users")
		{
			users.GET("/:id", userController.GetUser)
			users.GET("/:id/reviews", userController.GetUserReviews)
			users.PUT("/:id", middleware.AuthMiddleware(db), userController.UpdateUser)
			users.POST("/:id/avatar", middleware.AuthMiddleware(db), userController.UploadAvatar) // Must come before /:id
			users.DELETE("/:id", middleware.AuthMiddleware(db), userController.DeleteUser)
		}
	}
}

