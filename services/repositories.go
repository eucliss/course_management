package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Database models (these should eventually be moved to a separate package)
type CourseDB struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"not null" json:"name"`
	Address    string   `json:"address"`
	Hash       string   `gorm:"uniqueIndex;not null" json:"hash"`
	CourseData string   `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint    `json:"created_by"`
	UpdatedBy  *uint    `json:"updated_by"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	CreatedAt  int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64    `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserDB struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	GoogleID    string   `gorm:"uniqueIndex" json:"google_id"`
	Email       string   `gorm:"uniqueIndex" json:"email"`
	Name        string   `json:"name"`
	DisplayName *string  `json:"display_name"`
	Picture     string   `json:"picture"`
	Handicap    *float64 `json:"handicap,omitempty"`
	CreatedAt   int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64    `gorm:"autoUpdateTime" json:"updated_at"`
}

type CourseReviewDB struct {
	ID                 uint    `gorm:"primaryKey" json:"id"`
	CourseID           uint    `gorm:"not null" json:"course_id"`
	UserID             uint    `gorm:"not null" json:"user_id"`
	OverallRating      *string `gorm:"type:varchar(1)" json:"overall_rating"`
	Price              *string `gorm:"type:varchar(10)" json:"price"`
	HandicapDifficulty *int    `json:"handicap_difficulty"`
	HazardDifficulty   *int    `json:"hazard_difficulty"`
	Merch              *string `gorm:"type:varchar(1)" json:"merch"`
	Condition          *string `gorm:"type:varchar(1)" json:"condition"`
	EnjoymentRating    *string `gorm:"type:varchar(1)" json:"enjoyment_rating"`
	Vibe               *string `gorm:"type:varchar(1)" json:"vibe"`
	RangeRating        *string `gorm:"type:varchar(1)" json:"range_rating"`
	Amenities          *string `gorm:"type:varchar(1)" json:"amenities"`
	Glizzies           *string `gorm:"type:varchar(1)" json:"glizzies"`
	ReviewText         *string `gorm:"type:text" json:"review_text"`
	CreatedAt          int64   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          int64   `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserCourseScoreDB struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	UserID    uint    `gorm:"not null;index" json:"user_id"`
	CourseID  uint    `gorm:"not null;index" json:"course_id"`
	Score     int     `json:"score"`
	Handicap  float64 `json:"handicap"`
	CreatedAt int64   `gorm:"autoCreateTime" json:"created_at"`
}

type UserCourseHoleDB struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	UserID     uint `gorm:"not null;index" json:"user_id"`
	CourseID   uint `gorm:"not null;index" json:"course_id"`
	HoleNumber int  `gorm:"not null" json:"hole_number"`
	Score      int  `json:"score"`
	Par        int  `json:"par"`
	CreatedAt  int64 `gorm:"autoCreateTime" json:"created_at"`
}

// CourseRepository implementation
type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) Create(ctx context.Context, course Course, createdBy *uint) error {
	courseData, err := json.Marshal(course)
	if err != nil {
		return fmt.Errorf("failed to marshal course data: %w", err)
	}

	// Generate hash for uniqueness
	hash := fmt.Sprintf("%s|%s", course.Name, course.Address)

	courseDB := CourseDB{
		Name:       course.Name,
		Address:    course.Address,
		Hash:       hash,
		CourseData: string(courseData),
		CreatedBy:  createdBy,
		Latitude:   course.Latitude,
		Longitude:  course.Longitude,
	}

	return r.db.WithContext(ctx).Create(&courseDB).Error
}

func (r *courseRepository) GetByID(ctx context.Context, id uint) (*Course, error) {
	var courseDB CourseDB
	if err := r.db.WithContext(ctx).First(&courseDB, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("course not found")
		}
		return nil, err
	}

	return r.dbToCourse(courseDB)
}

func (r *courseRepository) GetByIndex(ctx context.Context, index int) (*Course, error) {
	var courseIDs []uint
	if err := r.db.WithContext(ctx).Model(&CourseDB{}).Select("id").Order("created_at ASC").Find(&courseIDs).Error; err != nil {
		return nil, err
	}

	if index < 0 || index >= len(courseIDs) {
		return nil, fmt.Errorf("course index out of range")
	}

	return r.GetByID(ctx, courseIDs[index])
}

func (r *courseRepository) GetByName(ctx context.Context, name string) (*Course, error) {
	var courseDB CourseDB
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&courseDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("course not found")
		}
		return nil, err
	}

	return r.dbToCourse(courseDB)
}

func (r *courseRepository) GetByNameAndAddress(ctx context.Context, name, address string) (*Course, error) {
	var courseDB CourseDB
	if err := r.db.WithContext(ctx).Where("name = ? AND address = ?", name, address).First(&courseDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("course not found")
		}
		return nil, err
	}

	return r.dbToCourse(courseDB)
}

func (r *courseRepository) GetAll(ctx context.Context) ([]Course, error) {
	var coursesDB []CourseDB
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, 0, len(coursesDB))
	for _, courseDB := range coursesDB {
		course, err := r.dbToCourse(courseDB)
		if err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}
		course.ID = courseDB.ID // Use actual database ID
		courses = append(courses, *course)
	}

	return courses, nil
}

func (r *courseRepository) Update(ctx context.Context, course Course, updatedBy *uint) error {
	courseData, err := json.Marshal(course)
	if err != nil {
		return fmt.Errorf("failed to marshal course data: %w", err)
	}

	updates := map[string]interface{}{
		"course_data": string(courseData),
		"updated_by":  updatedBy,
		"latitude":    course.Latitude,
		"longitude":   course.Longitude,
	}

	result := r.db.WithContext(ctx).Model(&CourseDB{}).Where("id = ?", course.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("course not found")
	}

	return nil
}

func (r *courseRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&CourseDB{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("course not found")
	}

	return nil
}

func (r *courseRepository) GetByUser(ctx context.Context, userID uint) ([]Course, error) {
	var coursesDB []CourseDB
	if err := r.db.WithContext(ctx).Where("created_by = ?", userID).Order("created_at ASC").Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, 0, len(coursesDB))
	for _, courseDB := range coursesDB {
		course, err := r.dbToCourse(courseDB)
		if err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}
		courses = append(courses, *course)
	}

	return courses, nil
}

func (r *courseRepository) CanEdit(ctx context.Context, courseID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseDB{}).Where("id = ? AND created_by = ?", courseID, userID).Count(&count).Error
	return count > 0, err
}

func (r *courseRepository) CanEditByIndex(ctx context.Context, index int, userID uint) (bool, error) {
	course, err := r.GetByIndex(ctx, index)
	if err != nil {
		return false, err
	}

	return r.CanEdit(ctx, uint(course.ID), userID)
}

func (r *courseRepository) IsOwner(ctx context.Context, userID uint, courseName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseDB{}).Where("name = ? AND created_by = ?", courseName, userID).Count(&count).Error
	return count > 0, err
}

func (r *courseRepository) GetWithPagination(ctx context.Context, offset, limit int) ([]Course, int64, error) {
	var coursesDB []CourseDB
	var totalCount int64

	if err := r.db.WithContext(ctx).Model(&CourseDB{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Order("created_at ASC").Offset(offset).Limit(limit).Find(&coursesDB).Error; err != nil {
		return nil, 0, err
	}

	courses := make([]Course, 0, len(coursesDB))
	for i, courseDB := range coursesDB {
		course, err := r.dbToCourse(courseDB)
		if err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}
		course.ID = uint(offset + i) // Maintain consistent indexing
		courses = append(courses, *course)
	}

	return courses, totalCount, nil
}

func (r *courseRepository) GetByUserWithPagination(ctx context.Context, userID uint, offset, limit int) ([]Course, int64, error) {
	var coursesDB []CourseDB
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&CourseDB{}).Where("created_by = ?", userID)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&coursesDB).Error; err != nil {
		return nil, 0, err
	}

	courses := make([]Course, 0, len(coursesDB))
	for _, courseDB := range coursesDB {
		course, err := r.dbToCourse(courseDB)
		if err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}
		courses = append(courses, *course)
	}

	return courses, totalCount, nil
}

func (r *courseRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseDB{}).Count(&count).Error
	return count, err
}

func (r *courseRepository) Exists(ctx context.Context, name, address string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseDB{}).Where("name = ? AND address = ?", name, address).Count(&count).Error
	return count > 0, err
}

func (r *courseRepository) GetAvailableForReview(ctx context.Context, userID uint) ([]Course, error) {
	// Get all courses that the user hasn't reviewed yet
	var coursesDB []CourseDB
	subQuery := r.db.Select("course_id").Where("user_id = ?", userID).Table("course_reviews")
	
	if err := r.db.WithContext(ctx).Where("id NOT IN (?)", subQuery).Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, 0, len(coursesDB))
	for _, courseDB := range coursesDB {
		course, err := r.dbToCourse(courseDB)
		if err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}
		courses = append(courses, *course)
	}

	return courses, nil
}

// Helper method to convert database model to domain model
func (r *courseRepository) dbToCourse(courseDB CourseDB) (*Course, error) {
	var course Course
	if err := json.Unmarshal([]byte(courseDB.CourseData), &course); err != nil {
		return nil, fmt.Errorf("failed to unmarshal course data: %w", err)
	}

	// Ensure database fields take precedence
	course.Latitude = courseDB.Latitude
	course.Longitude = courseDB.Longitude

	return &course, nil
}

// UserRepository implementation
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user GoogleUser) (*GoogleUser, error) {
	userDB := UserDB{
		GoogleID:    user.ID,
		Email:       user.Email,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Picture:     user.Picture,
		Handicap:    user.Handicap,
	}

	if err := r.db.WithContext(ctx).Create(&userDB).Error; err != nil {
		return nil, err
	}

	return r.dbToUser(userDB), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*GoogleUser, error) {
	var userDB UserDB
	if err := r.db.WithContext(ctx).First(&userDB, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return r.dbToUser(userDB), nil
}

func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*GoogleUser, error) {
	var userDB UserDB
	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&userDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return r.dbToUser(userDB), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*GoogleUser, error) {
	var userDB UserDB
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&userDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return r.dbToUser(userDB), nil
}

func (r *userRepository) Update(ctx context.Context, user GoogleUser) error {
	updates := map[string]interface{}{
		"name":         user.Name,
		"display_name": user.DisplayName,
		"picture":      user.Picture,
		"handicap":     user.Handicap,
	}

	// Convert string ID to uint for database query
	result := r.db.WithContext(ctx).Model(&UserDB{}).Where("google_id = ?", user.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdateHandicap(ctx context.Context, userID uint, handicap float64) error {
	result := r.db.WithContext(ctx).Model(&UserDB{}).Where("id = ?", userID).Update("handicap", handicap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdateDisplayName(ctx context.Context, userID uint, displayName string) error {
	result := r.db.WithContext(ctx).Model(&UserDB{}).Where("id = ?", userID).Update("display_name", displayName)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&UserDB{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) dbToUser(userDB UserDB) *GoogleUser {
	return &GoogleUser{
		ID:          userDB.GoogleID,
		Email:       userDB.Email,
		Name:        userDB.Name,
		DisplayName: userDB.DisplayName,
		Picture:     userDB.Picture,
		Handicap:    userDB.Handicap,
	}
}

// ReviewRepository implementation
type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review CourseReview) error {
	reviewDB := CourseReviewDB{
		CourseID:           review.CourseID,
		UserID:             review.UserID,
		OverallRating:      review.OverallRating,
		Price:              review.Price,
		HandicapDifficulty: review.HandicapDifficulty,
		HazardDifficulty:   review.HazardDifficulty,
		Merch:              review.Merch,
		Condition:          review.Condition,
		EnjoymentRating:    review.EnjoymentRating,
		Vibe:               review.Vibe,
		RangeRating:        review.RangeRating,
		Amenities:          review.Amenities,
		Glizzies:           review.Glizzies,
		ReviewText:         review.ReviewText,
	}

	return r.db.WithContext(ctx).Create(&reviewDB).Error
}

func (r *reviewRepository) GetByID(ctx context.Context, id uint) (*CourseReview, error) {
	var reviewDB CourseReviewDB
	if err := r.db.WithContext(ctx).First(&reviewDB, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("review not found")
		}
		return nil, err
	}

	return r.dbToReview(reviewDB), nil
}

func (r *reviewRepository) GetByUser(ctx context.Context, userID uint) ([]CourseReview, error) {
	var reviewsDB []CourseReviewDB
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&reviewsDB).Error; err != nil {
		return nil, err
	}

	reviews := make([]CourseReview, len(reviewsDB))
	for i, reviewDB := range reviewsDB {
		reviews[i] = *r.dbToReview(reviewDB)
	}

	return reviews, nil
}

func (r *reviewRepository) GetByCourse(ctx context.Context, courseID uint) ([]CourseReview, error) {
	var reviewsDB []CourseReviewDB
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Find(&reviewsDB).Error; err != nil {
		return nil, err
	}

	reviews := make([]CourseReview, len(reviewsDB))
	for i, reviewDB := range reviewsDB {
		reviews[i] = *r.dbToReview(reviewDB)
	}

	return reviews, nil
}

func (r *reviewRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uint) (*CourseReview, error) {
	var reviewDB CourseReviewDB
	if err := r.db.WithContext(ctx).Where("user_id = ? AND course_id = ?", userID, courseID).First(&reviewDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("review not found")
		}
		return nil, err
	}

	return r.dbToReview(reviewDB), nil
}

func (r *reviewRepository) Update(ctx context.Context, review CourseReview) error {
	updates := map[string]interface{}{
		"overall_rating":      review.OverallRating,
		"price":               review.Price,
		"handicap_difficulty": review.HandicapDifficulty,
		"hazard_difficulty":   review.HazardDifficulty,
		"merch":               review.Merch,
		"condition":           review.Condition,
		"enjoyment_rating":    review.EnjoymentRating,
		"vibe":                review.Vibe,
		"range_rating":        review.RangeRating,
		"amenities":           review.Amenities,
		"glizzies":            review.Glizzies,
		"review_text":         review.ReviewText,
	}

	result := r.db.WithContext(ctx).Model(&CourseReviewDB{}).Where("id = ?", review.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&CourseReviewDB{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

func (r *reviewRepository) AddScore(ctx context.Context, score UserCourseScore) error {
	scoreDB := UserCourseScoreDB{
		UserID:   score.UserID,
		CourseID: score.CourseID,
		Score:    score.Score,
		Handicap: score.Handicap,
	}

	return r.db.WithContext(ctx).Create(&scoreDB).Error
}

func (r *reviewRepository) GetUserScores(ctx context.Context, userID uint) ([]UserCourseScore, error) {
	var scoresDB []UserCourseScoreDB
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&scoresDB).Error; err != nil {
		return nil, err
	}

	scores := make([]UserCourseScore, len(scoresDB))
	for i, scoreDB := range scoresDB {
		scores[i] = UserCourseScore{
			ID:        scoreDB.ID,
			UserID:    scoreDB.UserID,
			CourseID:  scoreDB.CourseID,
			Score:     scoreDB.Score,
			Handicap:  scoreDB.Handicap,
			CreatedAt: scoreDB.CreatedAt,
		}
	}

	return scores, nil
}

func (r *reviewRepository) GetCourseScores(ctx context.Context, courseID uint) ([]UserCourseScore, error) {
	var scoresDB []UserCourseScoreDB
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Find(&scoresDB).Error; err != nil {
		return nil, err
	}

	scores := make([]UserCourseScore, len(scoresDB))
	for i, scoreDB := range scoresDB {
		scores[i] = UserCourseScore{
			ID:        scoreDB.ID,
			UserID:    scoreDB.UserID,
			CourseID:  scoreDB.CourseID,
			Score:     scoreDB.Score,
			Handicap:  scoreDB.Handicap,
			CreatedAt: scoreDB.CreatedAt,
		}
	}

	return scores, nil
}

func (r *reviewRepository) AddHoleScore(ctx context.Context, holeScore UserCourseHole) error {
	holeScoreDB := UserCourseHoleDB{
		UserID:     holeScore.UserID,
		CourseID:   holeScore.CourseID,
		HoleNumber: holeScore.HoleNumber,
		Score:      holeScore.Score,
		Par:        holeScore.Par,
	}

	return r.db.WithContext(ctx).Create(&holeScoreDB).Error
}

func (r *reviewRepository) GetUserHoleScores(ctx context.Context, userID uint) ([]UserCourseHole, error) {
	var holeScoresDB []UserCourseHoleDB
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&holeScoresDB).Error; err != nil {
		return nil, err
	}

	holeScores := make([]UserCourseHole, len(holeScoresDB))
	for i, holeScoreDB := range holeScoresDB {
		holeScores[i] = UserCourseHole{
			ID:         holeScoreDB.ID,
			UserID:     holeScoreDB.UserID,
			CourseID:   holeScoreDB.CourseID,
			HoleNumber: holeScoreDB.HoleNumber,
			Score:      holeScoreDB.Score,
			Par:        holeScoreDB.Par,
			CreatedAt:  holeScoreDB.CreatedAt,
		}
	}

	return holeScores, nil
}

func (r *reviewRepository) dbToReview(reviewDB CourseReviewDB) *CourseReview {
	return &CourseReview{
		ID:                 reviewDB.ID,
		CourseID:           reviewDB.CourseID,
		UserID:             reviewDB.UserID,
		OverallRating:      reviewDB.OverallRating,
		Price:              reviewDB.Price,
		HandicapDifficulty: reviewDB.HandicapDifficulty,
		HazardDifficulty:   reviewDB.HazardDifficulty,
		Merch:              reviewDB.Merch,
		Condition:          reviewDB.Condition,
		EnjoymentRating:    reviewDB.EnjoymentRating,
		Vibe:               reviewDB.Vibe,
		RangeRating:        reviewDB.RangeRating,
		Amenities:          reviewDB.Amenities,
		Glizzies:           reviewDB.Glizzies,
		ReviewText:         reviewDB.ReviewText,
		CreatedAt:          reviewDB.CreatedAt,
		UpdatedAt:          reviewDB.UpdatedAt,
	}
}