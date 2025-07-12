package main

// CourseReview represents a user's review of a specific course
type CourseReview struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	// Overall rating
	OverallRating *string `gorm:"type:varchar(1);check:overall_rating IN ('S','A','B','C','D','F')" json:"overall_rating"`

	// Individual ratings (matching your current Ranking struct)
	Price              *string `gorm:"type:varchar(10)" json:"price"`
	HandicapDifficulty *int    `json:"handicap_difficulty"`
	HazardDifficulty   *int    `json:"hazard_difficulty"`
	Merch              *string `gorm:"type:varchar(1);check:merch IN ('S','A','B','C','D','F')" json:"merch"`
	Condition          *string `gorm:"type:varchar(1);check:condition IN ('S','A','B','C','D','F')" json:"condition"`
	EnjoymentRating    *string `gorm:"type:varchar(1);check:enjoyment_rating IN ('S','A','B','C','D','F')" json:"enjoyment_rating"`
	Vibe               *string `gorm:"type:varchar(1);check:vibe IN ('S','A','B','C','D','F')" json:"vibe"`
	RangeRating        *string `gorm:"type:varchar(1);check:range_rating IN ('S','A','B','C','D','F')" json:"range_rating"`
	Amenities          *string `gorm:"type:varchar(1);check:amenities IN ('S','A','B','C','D','F')" json:"amenities"`
	Glizzies           *string `gorm:"type:varchar(1);check:glizzies IN ('S','A','B','C','D','F')" json:"glizzies"`
	Walkability        *string `gorm:"type:varchar(1);check:walkability IN ('S','A','B','C','D','F')" json:"walkability"`

	// Review text
	ReviewText *string `gorm:"type:text" json:"review_text"`

	// Timestamps
	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserCourseScore represents a user's score for a specific course
type UserCourseScore struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	// Score data
	Score      int      `gorm:"not null" json:"score"`
	Handicap   *float64 `gorm:"type:decimal(4,2)" json:"handicap"`
	DatePlayed *string  `gorm:"type:date" json:"date_played"`

	// Additional score details
	OutScore *int    `json:"out_score"`
	InScore  *int    `json:"in_score"`
	Notes    *string `gorm:"type:text" json:"notes"`

	// Timestamps
	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserCourseHole represents a user's hole-by-hole data for a specific course
type UserCourseHole struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	// Hole data
	Number      int     `gorm:"not null" json:"number"`
	Par         *int    `json:"par"`
	Yardage     *int    `json:"yardage"`
	Description *string `gorm:"type:text" json:"description"`

	// Timestamps
	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserActivity represents activities for social feed
type UserActivity struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	UserID       uint   `gorm:"not null" json:"user_id"`
	ActivityType string `gorm:"type:varchar(50);not null" json:"activity_type"` // 'course_review', 'score_posted', 'course_added'
	CourseID     *uint  `json:"course_id"`
	TargetUserID *uint  `json:"target_user_id"`         // For following/friend activities
	Data         string `gorm:"type:jsonb" json:"data"` // Additional activity-specific data

	// Timestamps
	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course     *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	TargetUser *User     `gorm:"foreignKey:TargetUserID" json:"target_user,omitempty"`
}

// CourseReviewSummary represents aggregated review data for a course
type CourseReviewSummary struct {
	CourseID      uint           `json:"course_id"`
	TotalReviews  int            `json:"total_reviews"`
	AverageRating *string        `json:"average_rating"` // Most common overall rating
	RatingCounts  map[string]int `json:"rating_counts"`  // Count for each rating (S, A, B, C, D, F)
}

// UserReviewWithCourse represents a user's review joined with course data for display
type UserReviewWithCourse struct {
	CourseReview
	CourseName    string `json:"course_name"`
	CourseAddress string `json:"course_address"`
}

// CourseWithUserReview represents a course with the current user's review (if any)
type CourseWithUserReview struct {
	CourseDB
	UserReview *CourseReview `json:"user_review,omitempty"`
	HasReview  bool          `json:"has_review"`
}

// ReviewFormData represents the form data for creating/updating reviews
type ReviewFormData struct {
	CourseID           uint   `json:"course_id"`
	OverallRating      string `json:"overall_rating"`
	Price              string `json:"price"`
	HandicapDifficulty int    `json:"handicap_difficulty"`
	HazardDifficulty   int    `json:"hazard_difficulty"`
	Merch              string `json:"merch"`
	Condition          string `json:"condition"`
	EnjoymentRating    string `json:"enjoyment_rating"`
	Vibe               string `json:"vibe"`
	RangeRating        string `json:"range"`
	Amenities          string `json:"amenities"`
	Glizzies           string `json:"glizzies"`
	Walkability        string `json:"walkability"`
	ReviewText         string `json:"review_text"`
}

// ScoreFormData represents the form data for adding scores
type ScoreFormData struct {
	CourseID   uint    `json:"course_id"`
	Score      int     `json:"score"`
	Handicap   float64 `json:"handicap"`
	DatePlayed string  `json:"date_played"`
	OutScore   int     `json:"out_score"`
	InScore    int     `json:"in_score"`
	Notes      string  `json:"notes"`
}

// HoleFormData represents the form data for hole information
type HoleFormData struct {
	Number      int    `json:"number"`
	Par         int    `json:"par"`
	Yardage     int    `json:"yardage"`
	Description string `json:"description"`
}
