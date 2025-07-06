package main

import (
	"fmt"
	"log"
	"strconv"

	"gorm.io/gorm"
)

type ReviewService struct {
	db *gorm.DB
}

func NewReviewService() *ReviewService {
	return &ReviewService{
		db: GetDB(),
	}
}

// CreateOrUpdateReview creates a new review or updates an existing one
func (rs *ReviewService) CreateOrUpdateReview(userID uint, courseID uint, formData ReviewFormData) (*CourseReview, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Check if review already exists
	var existingReview CourseReview
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&existingReview)

	review := &CourseReview{
		UserID:   userID,
		CourseID: courseID,
	}

	// Convert form data to review fields
	if formData.OverallRating != "" {
		review.OverallRating = &formData.OverallRating
	}
	if formData.Price != "" {
		review.Price = &formData.Price
	}
	if formData.HandicapDifficulty > 0 {
		review.HandicapDifficulty = &formData.HandicapDifficulty
	}
	if formData.HazardDifficulty > 0 {
		review.HazardDifficulty = &formData.HazardDifficulty
	}
	if formData.Merch != "" {
		review.Merch = &formData.Merch
	}
	if formData.Condition != "" {
		review.Condition = &formData.Condition
	}
	if formData.EnjoymentRating != "" {
		review.EnjoymentRating = &formData.EnjoymentRating
	}
	if formData.Vibe != "" {
		review.Vibe = &formData.Vibe
	}
	if formData.RangeRating != "" {
		review.RangeRating = &formData.RangeRating
	}
	if formData.Amenities != "" {
		review.Amenities = &formData.Amenities
	}
	if formData.Glizzies != "" {
		review.Glizzies = &formData.Glizzies
	}
	if formData.ReviewText != "" {
		review.ReviewText = &formData.ReviewText
	}

	if result.Error == nil {
		// Update existing review
		review.ID = existingReview.ID
		result = rs.db.Save(review)
		log.Printf("✅ Updated review for user %d, course %d", userID, courseID)
	} else {
		// Create new review
		result = rs.db.Create(review)
		log.Printf("✅ Created new review for user %d, course %d", userID, courseID)

		// Create activity record
		rs.createActivity(userID, "course_review", &courseID, nil)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to save review: %v", result.Error)
	}

	return review, nil
}

// GetUserReviewForCourse gets a user's review for a specific course
func (rs *ReviewService) GetUserReviewForCourse(userID uint, courseID uint) (*CourseReview, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	log.Printf("🔍 [REVIEW_SERVICE] Querying for review: user_id=%d, course_id=%d", userID, courseID)

	var review CourseReview
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&review)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("🔍 [REVIEW_SERVICE] No review found for user %d, course %d", userID, courseID)
			return nil, nil // No review found (not an error)
		}
		log.Printf("🔍 [REVIEW_SERVICE] Database error: %v", result.Error)
		return nil, fmt.Errorf("failed to get user review: %v", result.Error)
	}

	log.Printf("🔍 [REVIEW_SERVICE] Found review: ID=%d, user_id=%d, course_id=%d, rating=%s",
		review.ID, review.UserID, review.CourseID, func() string {
			if review.OverallRating == nil {
				return "nil"
			}
			return *review.OverallRating
		}())
	return &review, nil
}

// DeleteUserReview deletes a user's review for a specific course (does NOT delete the course)
func (rs *ReviewService) DeleteUserReview(userID uint, courseID uint) error {
	if rs.db == nil {
		return fmt.Errorf("database not connected")
	}

	// First, verify the review exists and belongs to the user
	var review CourseReview
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&review)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("review not found")
		}
		return fmt.Errorf("failed to find review: %v", result.Error)
	}

	// Delete the review (this only deletes the CourseReview record, NOT the CourseDB record)
	result = rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Delete(&CourseReview{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete review: %v", result.Error)
	}

	// Also delete associated scores and holes for this user/course
	// Delete scores
	result = rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Delete(&UserCourseScore{})
	if result.Error != nil {
		log.Printf("Warning: failed to delete user scores: %v", result.Error)
	}

	// Delete holes
	result = rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Delete(&UserCourseHole{})
	if result.Error != nil {
		log.Printf("Warning: failed to delete user holes: %v", result.Error)
	}

	log.Printf("✅ Deleted review and associated data for user %d, course %d", userID, courseID)
	return nil
}

// GetUserReviews gets all reviews by a specific user
func (rs *ReviewService) GetUserReviews(userID uint) ([]UserReviewWithCourse, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var reviews []UserReviewWithCourse
	result := rs.db.Table("course_reviews").
		Select("course_reviews.*, course_dbs.name as course_name, course_dbs.address as course_address").
		Joins("JOIN course_dbs ON course_reviews.course_id = course_dbs.id").
		Where("course_reviews.user_id = ?", userID).
		Order("course_reviews.created_at DESC").
		Scan(&reviews)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user reviews: %v", result.Error)
	}

	return reviews, nil
}

// GetCourseReviews gets all reviews for a specific course
func (rs *ReviewService) GetCourseReviews(courseID uint) ([]CourseReview, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var reviews []CourseReview
	result := rs.db.Preload("User").Where("course_id = ?", courseID).Order("created_at DESC").Find(&reviews)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get course reviews: %v", result.Error)
	}

	return reviews, nil
}

// GetCourseReviewSummary gets aggregated review data for a course
func (rs *ReviewService) GetCourseReviewSummary(courseID uint) (*CourseReviewSummary, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var totalReviews int64
	result := rs.db.Model(&CourseReview{}).Where("course_id = ?", courseID).Count(&totalReviews)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to count reviews: %v", result.Error)
	}

	// Get rating counts
	var ratingData []struct {
		OverallRating string
		Count         int64
	}

	result = rs.db.Model(&CourseReview{}).
		Select("overall_rating, COUNT(*) as count").
		Where("course_id = ? AND overall_rating IS NOT NULL", courseID).
		Group("overall_rating").
		Scan(&ratingData)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get rating counts: %v", result.Error)
	}

	ratingCounts := make(map[string]int)
	var mostCommonRating string
	var maxCount int64

	for _, data := range ratingData {
		ratingCounts[data.OverallRating] = int(data.Count)
		if data.Count > maxCount {
			maxCount = data.Count
			mostCommonRating = data.OverallRating
		}
	}

	summary := &CourseReviewSummary{
		CourseID:     courseID,
		TotalReviews: int(totalReviews),
		RatingCounts: ratingCounts,
	}

	if mostCommonRating != "" {
		summary.AverageRating = &mostCommonRating
	}

	return summary, nil
}

// AddScore adds a score for a user and course
func (rs *ReviewService) AddScore(userID uint, courseID uint, formData ScoreFormData) (*UserCourseScore, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	score := &UserCourseScore{
		UserID:   userID,
		CourseID: courseID,
		Score:    formData.Score,
	}

	if formData.Handicap > 0 {
		score.Handicap = &formData.Handicap
	}
	if formData.DatePlayed != "" {
		score.DatePlayed = &formData.DatePlayed
	}
	if formData.OutScore > 0 {
		score.OutScore = &formData.OutScore
	}
	if formData.InScore > 0 {
		score.InScore = &formData.InScore
	}
	if formData.Notes != "" {
		score.Notes = &formData.Notes
	}

	result := rs.db.Create(score)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to save score: %v", result.Error)
	}

	log.Printf("✅ Added score %d for user %d, course %d", formData.Score, userID, courseID)

	// Create activity record
	rs.createActivity(userID, "score_posted", &courseID, map[string]interface{}{
		"score": formData.Score,
	})

	return score, nil
}

// AddScores saves multiple scores for a user and course
func (rs *ReviewService) AddScores(userID uint, courseID uint, scores []ScoreFormData) error {
	if rs.db == nil {
		return fmt.Errorf("database not connected")
	}

	// Delete existing scores for this user/course
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Delete(&UserCourseScore{})
	if result.Error != nil {
		log.Printf("Warning: failed to delete existing scores: %v", result.Error)
	}

	// Add new scores
	for _, scoreData := range scores {
		if scoreData.Score <= 0 {
			continue // Skip invalid scores
		}

		score := &UserCourseScore{
			UserID:   userID,
			CourseID: courseID,
			Score:    scoreData.Score,
		}

		if scoreData.Handicap > 0 {
			score.Handicap = &scoreData.Handicap
		}
		if scoreData.DatePlayed != "" {
			score.DatePlayed = &scoreData.DatePlayed
		}
		if scoreData.OutScore > 0 {
			score.OutScore = &scoreData.OutScore
		}
		if scoreData.InScore > 0 {
			score.InScore = &scoreData.InScore
		}
		if scoreData.Notes != "" {
			score.Notes = &scoreData.Notes
		}

		result := rs.db.Create(score)
		if result.Error != nil {
			log.Printf("Warning: failed to save score %d: %v", scoreData.Score, result.Error)
		}
	}

	log.Printf("✅ Saved %d scores for user %d, course %d", len(scores), userID, courseID)

	// Create activity record for the first score
	if len(scores) > 0 {
		rs.createActivity(userID, "score_posted", &courseID, map[string]interface{}{
			"score": scores[0].Score,
		})
	}

	return nil
}

// GetUserScoresForCourse gets all scores for a user and course
func (rs *ReviewService) GetUserScoresForCourse(userID uint, courseID uint) ([]UserCourseScore, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var scores []UserCourseScore
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Order("created_at DESC").Find(&scores)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user scores: %v", result.Error)
	}

	return scores, nil
}

// AddHoles saves hole-by-hole data for a user and course
func (rs *ReviewService) AddHoles(userID uint, courseID uint, holes []HoleFormData) error {
	if rs.db == nil {
		return fmt.Errorf("database not connected")
	}

	// Delete existing holes for this user/course
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Delete(&UserCourseHole{})
	if result.Error != nil {
		log.Printf("Warning: failed to delete existing holes: %v", result.Error)
	}

	// Add new holes
	for _, holeData := range holes {
		if holeData.Number <= 0 {
			continue // Skip invalid holes
		}

		hole := &UserCourseHole{
			UserID:   userID,
			CourseID: courseID,
			Number:   holeData.Number,
		}

		if holeData.Par > 0 {
			hole.Par = &holeData.Par
		}
		if holeData.Yardage > 0 {
			hole.Yardage = &holeData.Yardage
		}
		if holeData.Description != "" {
			hole.Description = &holeData.Description
		}

		result := rs.db.Create(hole)
		if result.Error != nil {
			log.Printf("Warning: failed to save hole %d: %v", holeData.Number, result.Error)
		}
	}

	log.Printf("✅ Saved %d holes for user %d, course %d", len(holes), userID, courseID)
	return nil
}

// GetUserHolesForCourse gets all holes for a user and course
func (rs *ReviewService) GetUserHolesForCourse(userID uint, courseID uint) ([]UserCourseHole, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var holes []UserCourseHole
	result := rs.db.Where("user_id = ? AND course_id = ?", userID, courseID).Order("number ASC").Find(&holes)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user holes: %v", result.Error)
	}

	return holes, nil
}

// GetCoursesUserHasReviewed gets list of course IDs that a user has reviewed
func (rs *ReviewService) GetCoursesUserHasReviewed(userID uint) (map[uint]bool, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var courseIDs []uint
	result := rs.db.Model(&CourseReview{}).
		Select("DISTINCT course_id").
		Where("user_id = ?", userID).
		Pluck("course_id", &courseIDs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get reviewed courses: %v", result.Error)
	}

	reviewedCourses := make(map[uint]bool)
	for _, courseID := range courseIDs {
		reviewedCourses[courseID] = true
	}

	return reviewedCourses, nil
}

// GetAvailableCoursesForReview gets courses that a user hasn't reviewed yet
func (rs *ReviewService) GetAvailableCoursesForReview(userID uint) ([]CourseDB, error) {
	if rs.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Get course IDs that user has already reviewed
	reviewedCourses, err := rs.GetCoursesUserHasReviewed(userID)
	if err != nil {
		return nil, err
	}

	// Get all courses
	var allCourses []CourseDB
	result := rs.db.Find(&allCourses)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all courses: %v", result.Error)
	}

	// Filter out already reviewed courses
	var availableCourses []CourseDB
	for _, course := range allCourses {
		if !reviewedCourses[course.ID] {
			availableCourses = append(availableCourses, course)
		}
	}

	return availableCourses, nil
}

// Helper function to create activity records
func (rs *ReviewService) createActivity(userID uint, activityType string, courseID *uint, data interface{}) {
	if rs.db == nil {
		return
	}

	activity := &UserActivity{
		UserID:       userID,
		ActivityType: activityType,
		CourseID:     courseID,
	}

	// Convert data to JSON string if provided
	if data != nil {
		// For simplicity, we'll store as string - you might want to use proper JSON marshaling
		activity.Data = fmt.Sprintf("%v", data)
	}

	result := rs.db.Create(activity)
	if result.Error != nil {
		log.Printf("Warning: failed to create activity record: %v", result.Error)
	}
}

// ParseReviewFormData parses form data from HTTP request into ReviewFormData
func ParseReviewFormData(getFormValue func(string) string) ReviewFormData {
	// Helper function to parse int safely
	parseInt := func(s string) int {
		if s == "" {
			return 0
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return val
	}

	return ReviewFormData{
		OverallRating:      getFormValue("overall-rating"),
		Price:              getFormValue("price"),
		HandicapDifficulty: parseInt(getFormValue("handicap-difficulty")),
		HazardDifficulty:   parseInt(getFormValue("hazard-difficulty")),
		Merch:              getFormValue("merch"),
		Condition:          getFormValue("condition"),
		EnjoymentRating:    getFormValue("enjoyment-rating"),
		Vibe:               getFormValue("vibe"),
		RangeRating:        getFormValue("range"),
		Amenities:          getFormValue("amenities"),
		Glizzies:           getFormValue("glizzies"),
		ReviewText:         getFormValue("course-review"),
	}
}

// ParseScoreFormData parses score form data from HTTP request
func ParseScoreFormData(getFormValue func(string) string) []ScoreFormData {
	var scores []ScoreFormData

	// Helper function to parse int safely
	parseInt := func(s string) int {
		if s == "" {
			return 0
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return val
	}

	// Helper function to parse float safely
	parseFloat := func(s string) float64 {
		if s == "" {
			return 0
		}
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return val
	}

	// Parse scores array - look for scores[0].score, scores[1].score, etc.
	for i := 0; i < 10; i++ { // Max 10 scores (reasonable limit)
		scoreStr := getFormValue(fmt.Sprintf("scores[%d].score", i))
		if scoreStr == "" {
			continue // No more scores
		}

		scoreData := ScoreFormData{
			Score:      parseInt(scoreStr),
			Handicap:   parseFloat(getFormValue(fmt.Sprintf("scores[%d].handicap", i))),
			DatePlayed: getFormValue(fmt.Sprintf("scores[%d].date-played", i)),
			OutScore:   parseInt(getFormValue(fmt.Sprintf("scores[%d].out-score", i))),
			InScore:    parseInt(getFormValue(fmt.Sprintf("scores[%d].in-score", i))),
			Notes:      getFormValue(fmt.Sprintf("scores[%d].notes", i)),
		}

		if scoreData.Score > 0 {
			scores = append(scores, scoreData)
		}
	}

	return scores
}

// ParseHoleFormData parses hole form data from HTTP request
func ParseHoleFormData(getFormValue func(string) string) []HoleFormData {
	var holes []HoleFormData

	// Helper function to parse int safely
	parseInt := func(s string) int {
		if s == "" {
			return 0
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return val
	}

	// Parse holes array - look for holes[0].number, holes[1].number, etc.
	for i := 0; i < 18; i++ { // Max 18 holes
		numberStr := getFormValue(fmt.Sprintf("holes[%d].number", i))
		if numberStr == "" {
			continue // No more holes
		}

		holeData := HoleFormData{
			Number:      parseInt(numberStr),
			Par:         parseInt(getFormValue(fmt.Sprintf("holes[%d].par", i))),
			Yardage:     parseInt(getFormValue(fmt.Sprintf("holes[%d].yardage", i))),
			Description: getFormValue(fmt.Sprintf("holes[%d].description", i)),
		}

		if holeData.Number > 0 {
			holes = append(holes, holeData)
		}
	}

	return holes
}
