package api

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// UserHandler handles user-related API endpoints
type UserHandler struct {
	dbService ExtendedDatabaseServiceInterface
}

// UserProfileUpdateRequest represents user profile update request
type UserProfileUpdateRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=50"`
}

// UserHandicapUpdateRequest represents handicap update request
type UserHandicapUpdateRequest struct {
	Handicap *float64 `json:"handicap" validate:"omitempty,min=0,max=54"`
}

// UserScoreCreateRequest represents score creation request
type UserScoreCreateRequest struct {
	CourseID    uint    `json:"course_id" validate:"required"`
	Score       int     `json:"score" validate:"required,min=18,max=200"`
	Handicap    float64 `json:"handicap" validate:"min=0,max=54"`
	PlayedAt    *int64  `json:"played_at,omitempty"`
	Notes       *string `json:"notes,omitempty" validate:"omitempty,max=500"`
	Weather     *string `json:"weather,omitempty" validate:"omitempty,max=100"`
	Conditions  *string `json:"conditions,omitempty" validate:"omitempty,max=100"`
}

// UserScoreResponse represents a user's score
type UserScoreResponse struct {
	ID         uint    `json:"id"`
	CourseID   uint    `json:"course_id"`
	CourseName string  `json:"course_name"`
	Score      int     `json:"score"`
	Handicap   float64 `json:"handicap"`
	PlayedAt   int64   `json:"played_at"`
	Notes      *string `json:"notes,omitempty"`
	Weather    *string `json:"weather,omitempty"`
	Conditions *string `json:"conditions,omitempty"`
	CreatedAt  int64   `json:"created_at"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	TotalRounds     int     `json:"total_rounds"`
	AverageScore    float64 `json:"average_score"`
	BestScore       int     `json:"best_score"`
	CurrentHandicap float64 `json:"current_handicap"`
	CoursesPlayed   int     `json:"courses_played"`
	RecentTrend     string  `json:"recent_trend"` // "improving", "declining", "stable"
}

// NewUserHandler creates a new user handler
func NewUserHandler(dbService ExtendedDatabaseServiceInterface) *UserHandler {
	return &UserHandler{
		dbService: dbService,
	}
}

// GetProfile returns the authenticated user's profile
func (h *UserHandler) GetProfile(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	user, err := h.dbService.GetUserByID(userID)
	if err != nil {
		return NotFoundError(c, "User")
	}

	return SuccessResponse(c, user)
}

// UpdateProfile updates the authenticated user's profile
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	var req UserProfileUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate display name
	if req.DisplayName != nil {
		if len(*req.DisplayName) == 0 {
			return ValidationError(c, map[string]string{
				"display_name": "Display name cannot be empty",
			})
		}
		if len(*req.DisplayName) > 50 {
			return ValidationError(c, map[string]string{
				"display_name": "Display name must be 50 characters or less",
			})
		}
	}

	// Update user profile
	user, err := h.dbService.UpdateUserProfile(userID, req.DisplayName)
	if err != nil {
		return InternalServerError(c, "Failed to update profile")
	}

	return SuccessResponse(c, user)
}

// UpdateHandicap updates the authenticated user's handicap
func (h *UserHandler) UpdateHandicap(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	var req UserHandicapUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate handicap
	if req.Handicap != nil {
		if *req.Handicap < 0 || *req.Handicap > 54 {
			return ValidationError(c, map[string]string{
				"handicap": "Handicap must be between 0 and 54",
			})
		}
	}

	// Update user handicap
	user, err := h.dbService.UpdateUserHandicap(userID, req.Handicap)
	if err != nil {
		return InternalServerError(c, "Failed to update handicap")
	}

	return SuccessResponse(c, user)
}

// GetScores returns the authenticated user's score history
func (h *UserHandler) GetScores(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	// Get pagination parameters
	pagination := GetPagination(c)

	// Get course ID filter if provided
	courseIDParam := c.QueryParam("course_id")
	var courseID *uint
	if courseIDParam != "" {
		id, err := strconv.ParseUint(courseIDParam, 10, 32)
		if err != nil {
			return BadRequestError(c, "Invalid course_id parameter")
		}
		courseIDUint := uint(id)
		courseID = &courseIDUint
	}

	scores, total, err := h.dbService.GetUserScores(userID, courseID, pagination.Page, pagination.PerPage)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve scores")
	}

	// Create paginated response
	meta := &APIMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
	}

	return SuccessResponseWithMeta(c, scores, meta)
}

// AddScore adds a new score for the authenticated user
func (h *UserHandler) AddScore(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	var req UserScoreCreateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate required fields
	validationErrors := make(map[string]string)
	
	if req.CourseID == 0 {
		validationErrors["course_id"] = "Course ID is required"
	}
	
	if req.Score < 18 || req.Score > 200 {
		validationErrors["score"] = "Score must be between 18 and 200"
	}
	
	if req.Handicap < 0 || req.Handicap > 54 {
		validationErrors["handicap"] = "Handicap must be between 0 and 54"
	}

	if req.Notes != nil && len(*req.Notes) > 500 {
		validationErrors["notes"] = "Notes must be 500 characters or less"
	}

	if len(validationErrors) > 0 {
		return ValidationError(c, validationErrors)
	}

	// Verify course exists
	courseExists, err := h.dbService.CourseExists(req.CourseID)
	if err != nil {
		return InternalServerError(c, "Failed to verify course")
	}
	if !courseExists {
		return NotFoundError(c, "Course")
	}

	// Create score
	score, err := h.dbService.CreateUserScore(userID, &req)
	if err != nil {
		return InternalServerError(c, "Failed to create score")
	}

	return CreatedResponse(c, score)
}

// DeleteScore deletes a user's score
func (h *UserHandler) DeleteScore(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	scoreIDParam := c.Param("scoreId")
	scoreID, err := strconv.ParseUint(scoreIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid score ID")
	}

	// Verify score belongs to user
	scoreOwner, err := h.dbService.GetScoreOwner(uint(scoreID))
	if err != nil {
		return NotFoundError(c, "Score")
	}

	if scoreOwner != userID {
		return ForbiddenError(c, "You can only delete your own scores")
	}

	// Delete score
	err = h.dbService.DeleteUserScore(uint(scoreID))
	if err != nil {
		return InternalServerError(c, "Failed to delete score")
	}

	return NoContentResponse(c)
}

// GetStats returns statistics for the authenticated user
func (h *UserHandler) GetStats(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	stats, err := h.dbService.GetUserStats(userID)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve user statistics")
	}

	return SuccessResponse(c, stats)
}

// RegisterRoutes registers user-related routes
func (h *UserHandler) RegisterRoutes(g *echo.Group, jwtService *JWTService) {
	// All user routes require authentication
	userGroup := g.Group("/user", JWTMiddleware(jwtService))
	
	// Profile management
	userGroup.GET("/profile", h.GetProfile)
	userGroup.PUT("/profile", h.UpdateProfile)
	userGroup.PUT("/handicap", h.UpdateHandicap)
	
	// Score management
	userGroup.GET("/scores", h.GetScores)
	userGroup.POST("/scores", h.AddScore)
	userGroup.DELETE("/scores/:scoreId", h.DeleteScore)
	
	// Statistics
	userGroup.GET("/stats", h.GetStats)
}

// Extended database interface for user operations
type ExtendedDatabaseServiceInterface interface {
	DatabaseServiceInterface
	GetUserByID(userID uint) (*UserResponse, error)
	UpdateUserProfile(userID uint, displayName *string) (*UserResponse, error)
	UpdateUserHandicap(userID uint, handicap *float64) (*UserResponse, error)
	GetUserScores(userID uint, courseID *uint, page, perPage int) ([]*UserScoreResponse, int, error)
	CreateUserScore(userID uint, req *UserScoreCreateRequest) (*UserScoreResponse, error)
	DeleteUserScore(scoreID uint) error
	GetScoreOwner(scoreID uint) (uint, error)
	GetUserStats(userID uint) (*UserStatsResponse, error)
	CourseExists(courseID uint) (bool, error)
}