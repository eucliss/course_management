package services

import (
	"context"

	"github.com/labstack/echo/v4"
)

// Core domain models - these should be moved to a separate package eventually
type Course struct {
	Name          string   `json:"name"`
	ID            int      `json:"ID"`
	Description   string   `json:"description"`
	Ranks         Ranking  `json:"ranks"`
	OverallRating string   `json:"overallRating"`
	Review        string   `json:"review"`
	Holes         []Hole   `json:"holes"`
	Scores        []Score  `json:"scores"`
	Address       string   `json:"address"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

type Ranking struct {
	Price              string `json:"price"`
	HandicapDifficulty int    `json:"handicapDifficulty"`
	HazardDifficulty   int    `json:"hazardDifficulty"`
	Merch              string `json:"merch"`
	Condition          string `json:"condition"`
	EnjoymentRating    string `json:"enjoymentRating"`
	Vibe               string `json:"vibe"`
	Range              string `json:"range"`
	Amenities          string `json:"amenities"`
	Glizzies           string `json:"glizzies"`
}

type Hole struct {
	Number      int    `json:"number"`
	Par         int    `json:"par"`
	Yardage     int    `json:"yardage"`
	Description string `json:"description"`
}

type Score struct {
	Score    int     `json:"score"`
	Handicap float64 `json:"handicap"`
}

type GoogleUser struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Name        string  `json:"name"`
	Picture     string  `json:"picture"`
	DisplayName *string `json:"display_name"`
	Handicap    *float64 `json:"handicap"`
}

// Repository interfaces for data access
type CourseRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, course Course, createdBy *uint) error
	GetByID(ctx context.Context, id uint) (*Course, error)
	GetByIndex(ctx context.Context, index int) (*Course, error)
	GetByName(ctx context.Context, name string) (*Course, error)
	GetByNameAndAddress(ctx context.Context, name, address string) (*Course, error)
	GetAll(ctx context.Context) ([]Course, error)
	Update(ctx context.Context, course Course, updatedBy *uint) error
	Delete(ctx context.Context, id uint) error

	// Ownership and permissions
	GetByUser(ctx context.Context, userID uint) ([]Course, error)
	CanEdit(ctx context.Context, courseID uint, userID uint) (bool, error)
	CanEditByIndex(ctx context.Context, index int, userID uint) (bool, error)
	IsOwner(ctx context.Context, userID uint, courseName string) (bool, error)

	// Pagination and filtering
	GetWithPagination(ctx context.Context, offset, limit int) ([]Course, int64, error)
	GetByUserWithPagination(ctx context.Context, userID uint, offset, limit int) ([]Course, int64, error)
	Count(ctx context.Context) (int64, error)

	// Utility methods
	Exists(ctx context.Context, name, address string) (bool, error)
	GetAvailableForReview(ctx context.Context, userID uint) ([]Course, error)
}

type UserRepository interface {
	Create(ctx context.Context, user GoogleUser) (*GoogleUser, error)
	GetByID(ctx context.Context, id uint) (*GoogleUser, error)
	GetByGoogleID(ctx context.Context, googleID string) (*GoogleUser, error)
	GetByEmail(ctx context.Context, email string) (*GoogleUser, error)
	Update(ctx context.Context, user GoogleUser) error
	UpdateHandicap(ctx context.Context, userID uint, handicap float64) error
	UpdateDisplayName(ctx context.Context, userID uint, displayName string) error
	Delete(ctx context.Context, id uint) error
}

type ReviewRepository interface {
	Create(ctx context.Context, review CourseReview) error
	GetByID(ctx context.Context, id uint) (*CourseReview, error)
	GetByUser(ctx context.Context, userID uint) ([]CourseReview, error)
	GetByCourse(ctx context.Context, courseID uint) ([]CourseReview, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID uint) (*CourseReview, error)
	Update(ctx context.Context, review CourseReview) error
	Delete(ctx context.Context, id uint) error
	
	// Score-specific operations
	AddScore(ctx context.Context, score UserCourseScore) error
	GetUserScores(ctx context.Context, userID uint) ([]UserCourseScore, error)
	GetCourseScores(ctx context.Context, courseID uint) ([]UserCourseScore, error)
	
	// Hole-specific operations
	AddHoleScore(ctx context.Context, holeScore UserCourseHole) error
	GetUserHoleScores(ctx context.Context, userID uint) ([]UserCourseHole, error)
}

// Service interfaces for business logic
type CourseService interface {
	// Course management
	CreateCourse(ctx context.Context, course Course, createdBy *uint) error
	GetCourse(ctx context.Context, id uint) (*Course, error)
	GetCourseByIndex(ctx context.Context, index int) (*Course, error)
	GetAllCourses(ctx context.Context) ([]Course, error)
	UpdateCourse(ctx context.Context, course Course, updatedBy *uint) error
	DeleteCourse(ctx context.Context, id uint, userID uint) error

	// Course permissions
	CanEditCourse(ctx context.Context, courseID uint, userID uint) (bool, error)
	CanEditCourseByIndex(ctx context.Context, index int, userID uint) (bool, error)
	GetUserCourses(ctx context.Context, userID uint) ([]Course, error)

	// Course search and filtering
	GetCoursesWithPagination(ctx context.Context, offset, limit int) ([]Course, int64, error)
	GetAvailableCoursesForReview(ctx context.Context, userID uint) ([]Course, error)
	FindCourseByNameAndAddress(ctx context.Context, name, address string) (*Course, error)

	// Form parsing and validation
	ParseCourseForm(form map[string][]string) ([]Hole, []Score, error)
	ValidateCourse(course Course) error
}

type AuthService interface {
	// Authentication
	VerifyGoogleToken(ctx context.Context, token string) (*GoogleUser, error)
	GetAuthConfig() AuthConfig
	
	// User management
	CreateOrUpdateUser(ctx context.Context, googleUser GoogleUser) (*GoogleUser, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*GoogleUser, error)
}

type SessionService interface {
	// Session management
	GetUser(c echo.Context) *GoogleUser
	GetUserID(c echo.Context) *uint
	IsAuthenticated(c echo.Context) bool
	SetUser(c echo.Context, user GoogleUser) error
	ClearSession(c echo.Context) error
	
	// Helper methods
	GetDatabaseUserID(c echo.Context) *uint
	RequireAuth(c echo.Context) (*GoogleUser, error)
}

type ReviewService interface {
	// Review management
	CreateReview(ctx context.Context, review CourseReview) error
	GetReview(ctx context.Context, id uint) (*CourseReview, error)
	GetUserReviews(ctx context.Context, userID uint) ([]CourseReview, error)
	GetCourseReviews(ctx context.Context, courseID uint) ([]CourseReview, error)
	UpdateReview(ctx context.Context, review CourseReview) error
	DeleteReview(ctx context.Context, id uint, userID uint) error
	
	// Score management
	AddScore(ctx context.Context, userID uint, courseID uint, score Score) error
	GetUserScores(ctx context.Context, userID uint) ([]UserCourseScore, error)
	
	// Hole management
	AddHoleScore(ctx context.Context, userID uint, courseID uint, holeNumber int, score int, par int) error
	GetUserHoleScores(ctx context.Context, userID uint) ([]UserCourseHole, error)
	
	// Activity tracking
	RecordActivity(ctx context.Context, userID uint, activityType string, details map[string]interface{}) error
}

// Supporting types for services
type AuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
}

type CourseReview struct {
	ID         uint     `json:"id"`
	UserID     uint     `json:"user_id"`
	CourseID   uint     `json:"course_id"`
	CourseName string   `json:"course_name"`
	Review     string   `json:"review"`
	Rating     int      `json:"rating"`
	CreatedAt  int64    `json:"created_at"`
	UpdatedAt  int64    `json:"updated_at"`
}

type UserCourseScore struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id"`
	CourseID  uint    `json:"course_id"`
	Score     int     `json:"score"`
	Handicap  float64 `json:"handicap"`
	CreatedAt int64   `json:"created_at"`
}

type UserCourseHole struct {
	ID        uint `json:"id"`
	UserID    uint `json:"user_id"`
	CourseID  uint `json:"course_id"`
	HoleNumber int `json:"hole_number"`
	Score     int  `json:"score"`
	Par       int  `json:"par"`
	CreatedAt int64 `json:"created_at"`
}

// Service configuration
type ServiceConfig struct {
	DatabaseURL string
	RedisURL    string
	AuthConfig  AuthConfig
}

// Service container interface
type ServiceContainer interface {
	CourseService() CourseService
	AuthService() AuthService
	SessionService() SessionService
	ReviewService() ReviewService
	
	// Repository access (for advanced use cases)
	CourseRepository() CourseRepository
	UserRepository() UserRepository
	ReviewRepository() ReviewRepository
	
	// Lifecycle management
	Close() error
}