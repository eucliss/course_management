package services

import (
	"context"
	"fmt"
	"strings"
)

type reviewService struct {
	reviewRepo ReviewRepository
	courseRepo CourseRepository
	userRepo   UserRepository
}

func NewReviewService(reviewRepo ReviewRepository, courseRepo CourseRepository, userRepo UserRepository) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
		courseRepo: courseRepo,
		userRepo:   userRepo,
	}
}

func (s *reviewService) CreateReview(ctx context.Context, review CourseReview) error {
	// Validate review data
	if err := s.validateReview(review); err != nil {
		return fmt.Errorf("review validation failed: %w", err)
	}

	// Check if user already has a review for this course
	existingReview, err := s.reviewRepo.GetByUserAndCourse(ctx, review.UserID, review.CourseID)
	if err == nil && existingReview != nil {
		return fmt.Errorf("user already has a review for this course")
	}

	// Verify course exists
	course, err := s.courseRepo.GetByID(ctx, review.CourseID)
	if err != nil {
		return fmt.Errorf("course not found: %w", err)
	}

	// Create the review
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	// Record activity
	if err := s.RecordActivity(ctx, review.UserID, "review_created", map[string]interface{}{
		"course_id":   review.CourseID,
		"course_name": course.Name,
		"rating":      review.OverallRating,
	}); err != nil {
		// Log but don't fail the review creation
		fmt.Printf("Warning: failed to record activity: %v\n", err)
	}

	return nil
}

func (s *reviewService) GetReview(ctx context.Context, id uint) (*CourseReview, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return review, nil
}

func (s *reviewService) GetUserReviews(ctx context.Context, userID uint) ([]CourseReview, error) {
	reviews, err := s.reviewRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reviews: %w", err)
	}

	return reviews, nil
}

func (s *reviewService) GetCourseReviews(ctx context.Context, courseID uint) ([]CourseReview, error) {
	reviews, err := s.reviewRepo.GetByCourse(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course reviews: %w", err)
	}

	return reviews, nil
}

func (s *reviewService) UpdateReview(ctx context.Context, review CourseReview) error {
	// Validate review data
	if err := s.validateReview(review); err != nil {
		return fmt.Errorf("review validation failed: %w", err)
	}

	// Check if review exists and belongs to user
	existingReview, err := s.reviewRepo.GetByID(ctx, review.ID)
	if err != nil {
		return fmt.Errorf("review not found: %w", err)
	}

	if existingReview.UserID != review.UserID {
		return fmt.Errorf("user does not have permission to update this review")
	}

	// Update the review
	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	// Get course info for activity logging
	course, err := s.courseRepo.GetByID(ctx, review.CourseID)
	if err != nil {
		return fmt.Errorf("course not found: %w", err)
	}

	// Record activity
	if err := s.RecordActivity(ctx, review.UserID, "review_updated", map[string]interface{}{
		"review_id":   review.ID,
		"course_id":   review.CourseID,
		"course_name": course.Name,
		"rating":      review.OverallRating,
	}); err != nil {
		// Log but don't fail the review update
		fmt.Printf("Warning: failed to record activity: %v\n", err)
	}

	return nil
}

func (s *reviewService) DeleteReview(ctx context.Context, id uint, userID uint) error {
	// Check if review exists and belongs to user
	existingReview, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("review not found: %w", err)
	}

	if existingReview.UserID != userID {
		return fmt.Errorf("user does not have permission to delete this review")
	}

	// Get course info for activity logging
	course, err := s.courseRepo.GetByID(ctx, existingReview.CourseID)
	if err != nil {
		return fmt.Errorf("course not found: %w", err)
	}

	// Delete the review
	if err := s.reviewRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	// Record activity
	if err := s.RecordActivity(ctx, userID, "review_deleted", map[string]interface{}{
		"review_id":   id,
		"course_id":   existingReview.CourseID,
		"course_name": course.Name,
	}); err != nil {
		// Log but don't fail the review deletion
		fmt.Printf("Warning: failed to record activity: %v\n", err)
	}

	return nil
}

func (s *reviewService) AddScore(ctx context.Context, userID uint, courseID uint, score Score) error {
	// Validate score
	if err := s.validateScore(score); err != nil {
		return fmt.Errorf("score validation failed: %w", err)
	}

	// Verify course exists
	_, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return fmt.Errorf("course not found: %w", err)
	}

	// Create score record
	courseScore := UserCourseScore{
		UserID:   userID,
		CourseID: courseID,
		Score:    score.Score,
		Handicap: score.Handicap,
	}

	if err := s.reviewRepo.AddScore(ctx, courseScore); err != nil {
		return fmt.Errorf("failed to add score: %w", err)
	}

	// Record activity
	if err := s.RecordActivity(ctx, userID, "score_added", map[string]interface{}{
		"course_id": courseID,
		"score":     score.Score,
		"handicap":  score.Handicap,
	}); err != nil {
		// Log but don't fail the score addition
		fmt.Printf("Warning: failed to record activity: %v\n", err)
	}

	return nil
}

func (s *reviewService) GetUserScores(ctx context.Context, userID uint) ([]UserCourseScore, error) {
	scores, err := s.reviewRepo.GetUserScores(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user scores: %w", err)
	}

	return scores, nil
}

func (s *reviewService) AddHoleScore(ctx context.Context, userID uint, courseID uint, holeNumber int, score int, par int) error {
	// Validate hole score
	if holeNumber < 1 || holeNumber > 18 {
		return fmt.Errorf("hole number must be between 1 and 18")
	}
	if score < 1 || score > 20 {
		return fmt.Errorf("score must be between 1 and 20")
	}
	if par < 3 || par > 6 {
		return fmt.Errorf("par must be between 3 and 6")
	}

	// Verify course exists
	_, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return fmt.Errorf("course not found: %w", err)
	}

	// Create hole score record
	holeScore := UserCourseHole{
		UserID:     userID,
		CourseID:   courseID,
		HoleNumber: holeNumber,
		Score:      score,
		Par:        par,
	}

	if err := s.reviewRepo.AddHoleScore(ctx, holeScore); err != nil {
		return fmt.Errorf("failed to add hole score: %w", err)
	}

	// Record activity
	if err := s.RecordActivity(ctx, userID, "hole_score_added", map[string]interface{}{
		"course_id":   courseID,
		"hole_number": holeNumber,
		"score":       score,
		"par":         par,
	}); err != nil {
		// Log but don't fail the hole score addition
		fmt.Printf("Warning: failed to record activity: %v\n", err)
	}

	return nil
}

func (s *reviewService) GetUserHoleScores(ctx context.Context, userID uint) ([]UserCourseHole, error) {
	holeScores, err := s.reviewRepo.GetUserHoleScores(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user hole scores: %w", err)
	}

	return holeScores, nil
}

func (s *reviewService) RecordActivity(ctx context.Context, userID uint, activityType string, details map[string]interface{}) error {
	// This is a placeholder for activity recording
	// In a real implementation, you would have an ActivityRepository
	// and proper activity tracking logic
	
	// For now, just log the activity
	fmt.Printf("Activity recorded: User %d performed %s with details: %+v\n", userID, activityType, details)
	
	return nil
}

func (s *reviewService) validateReview(review CourseReview) error {
	if review.UserID == 0 {
		return fmt.Errorf("user ID is required")
	}
	if review.CourseID == 0 {
		return fmt.Errorf("course ID is required")
	}
	if review.ReviewText == nil || strings.TrimSpace(*review.ReviewText) == "" {
		return fmt.Errorf("review text is required")
	}
	if review.OverallRating == nil {
		return fmt.Errorf("overall rating is required")
	}
	if len(*review.ReviewText) < 10 {
		return fmt.Errorf("review must be at least 10 characters long")
	}
	if len(*review.ReviewText) > 2000 {
		return fmt.Errorf("review must be less than 2000 characters")
	}

	return nil
}

func (s *reviewService) validateScore(score Score) error {
	if score.Score < 1 || score.Score > 20 {
		return fmt.Errorf("score must be between 1 and 20")
	}
	if score.Handicap < -5 || score.Handicap > 40 {
		return fmt.Errorf("handicap must be between -5 and 40")
	}

	return nil
}