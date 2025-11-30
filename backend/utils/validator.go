package utils

import (
	"fmt"
	"music-review-site/backend/models"
	"regexp"
)

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must be at most 50 characters long")
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}
	return nil
}

// ValidateRating validates rating value (1-10)
func ValidateRating(rating int) error {
	if rating < 1 || rating > 10 {
		return fmt.Errorf("rating must be between 1 and 10")
	}
	return nil
}

// ValidateAtmosphereRating validates atmosphere rating (1-10)
func ValidateAtmosphereRating(rating int) error {
	if rating < 1 || rating > 10 {
		return fmt.Errorf("atmosphere rating must be between 1 and 10")
	}
	return nil
}

// ValidateAtmosphereMultiplier validates atmosphere multiplier (1.0000-1.6072)
// This is kept for backward compatibility with stored data
func ValidateAtmosphereMultiplier(multiplier float64) error {
	if multiplier < 1.0000 || multiplier > 1.6072 {
		return fmt.Errorf("atmosphere multiplier must be between 1.0000 and 1.6072")
	}
	return nil
}

// ValidateReview validates review data
func ValidateReview(review *models.Review) error {
	// Either album_id or track_id must be set, but not both
	if review.AlbumID == nil && review.TrackID == nil {
		return fmt.Errorf("either album_id or track_id must be provided")
	}
	if review.AlbumID != nil && review.TrackID != nil {
		return fmt.Errorf("only one of album_id or track_id can be provided")
	}
	if err := ValidateRating(review.RatingRhymes); err != nil {
		return fmt.Errorf("rating_rhymes: %w", err)
	}
	if err := ValidateRating(review.RatingStructure); err != nil {
		return fmt.Errorf("rating_structure: %w", err)
	}
	if err := ValidateRating(review.RatingImplementation); err != nil {
		return fmt.Errorf("rating_implementation: %w", err)
	}
	if err := ValidateRating(review.RatingIndividuality); err != nil {
		return fmt.Errorf("rating_individuality: %w", err)
	}
	if err := ValidateAtmosphereMultiplier(review.AtmosphereMultiplier); err != nil {
		return fmt.Errorf("atmosphere_multiplier: %w", err)
	}
	return nil
}

