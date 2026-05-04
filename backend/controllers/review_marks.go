package controllers

import (
	"music-review-site/backend/models"

	"gorm.io/gorm"
)

type artistReviewMark struct {
	ReviewID uint
	Username string
}

func annotateArtistMarks(db *gorm.DB, reviews []models.Review) {
	if len(reviews) == 0 {
		return
	}

	reviewIDs := make([]uint, 0, len(reviews))
	for _, review := range reviews {
		if review.ID != 0 {
			reviewIDs = append(reviewIDs, review.ID)
		}
	}
	if len(reviewIDs) == 0 {
		return
	}

	var marks []artistReviewMark
	if err := db.Table("review_likes").
		Select("review_likes.review_id, users.username").
		Joins("JOIN users ON users.id = review_likes.user_id").
		Where("review_likes.review_id IN ? AND users.is_verified_artist = ?", reviewIDs, true).
		Scan(&marks).Error; err != nil {
		return
	}

	marksByReview := make(map[uint][]string)
	for _, mark := range marks {
		marksByReview[mark.ReviewID] = append(marksByReview[mark.ReviewID], mark.Username)
	}

	for i := range reviews {
		if usernames, ok := marksByReview[reviews[i].ID]; ok {
			reviews[i].HasArtistMark = true
			reviews[i].ArtistMarkUsernames = usernames
		}
	}
}

func annotateArtistMark(db *gorm.DB, review *models.Review) {
	if review == nil || review.ID == 0 {
		return
	}

	reviews := []models.Review{*review}
	annotateArtistMarks(db, reviews)
	*review = reviews[0]
}
