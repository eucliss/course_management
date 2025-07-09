package api

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// ReviewHandler handles review-related API endpoints
type ReviewHandler struct {
	dbService ReviewDatabaseServiceInterface
}

// ReviewCreateRequest represents review creation request
type ReviewCreateRequest struct {
	CourseID          uint    `json:"course_id" validate:"required"`
	OverallRating     int     `json:"overall_rating" validate:"required,min=1,max=10"`
	ReviewText        *string `json:"review_text,omitempty" validate:"omitempty,max=2000"`
	Price             *string `json:"price,omitempty" validate:"omitempty,max=50"`
	HandicapDifficulty *int   `json:"handicap_difficulty,omitempty" validate:"omitempty,min=1,max=10"`
	HazardDifficulty  *int    `json:"hazard_difficulty,omitempty" validate:"omitempty,min=1,max=10"`
	Merch             *string `json:"merch,omitempty" validate:"omitempty,max=100"`
	Condition         *string `json:"condition,omitempty" validate:"omitempty,max=100"`
	EnjoymentRating   *int    `json:"enjoyment_rating,omitempty" validate:"omitempty,min=1,max=10"`
	Vibe              *string `json:"vibe,omitempty" validate:"omitempty,max=100"`
	Range             *string `json:"range,omitempty" validate:"omitempty,max=100"`
	Amenities         *string `json:"amenities,omitempty" validate:"omitempty,max=200"`
	Food              *string `json:"food,omitempty" validate:"omitempty,max=100"`
	Atmosphere        *string `json:"atmosphere,omitempty" validate:"omitempty,max=100"`
	Value             *int    `json:"value,omitempty" validate:"omitempty,min=1,max=10"`
	Maintenance       *int    `json:"maintenance,omitempty" validate:"omitempty,min=1,max=10"`
	Pace              *int    `json:"pace,omitempty" validate:"omitempty,min=1,max=10"`
	Staff             *int    `json:"staff,omitempty" validate:"omitempty,min=1,max=10"`
}

// ReviewUpdateRequest represents review update request
type ReviewUpdateRequest struct {
	OverallRating     *int    `json:"overall_rating,omitempty" validate:"omitempty,min=1,max=10"`
	ReviewText        *string `json:"review_text,omitempty" validate:"omitempty,max=2000"`
	Price             *string `json:"price,omitempty" validate:"omitempty,max=50"`
	HandicapDifficulty *int   `json:"handicap_difficulty,omitempty" validate:"omitempty,min=1,max=10"`
	HazardDifficulty  *int    `json:"hazard_difficulty,omitempty" validate:"omitempty,min=1,max=10"`
	Merch             *string `json:"merch,omitempty" validate:"omitempty,max=100"`
	Condition         *string `json:"condition,omitempty" validate:"omitempty,max=100"`
	EnjoymentRating   *int    `json:"enjoyment_rating,omitempty" validate:"omitempty,min=1,max=10"`
	Vibe              *string `json:"vibe,omitempty" validate:"omitempty,max=100"`
	Range             *string `json:"range,omitempty" validate:"omitempty,max=100"`
	Amenities         *string `json:"amenities,omitempty" validate:"omitempty,max=200"`
	Food              *string `json:"food,omitempty" validate:"omitempty,max=100"`
	Atmosphere        *string `json:"atmosphere,omitempty" validate:"omitempty,max=100"`
	Value             *int    `json:"value,omitempty" validate:"omitempty,min=1,max=10"`
	Maintenance       *int    `json:"maintenance,omitempty" validate:"omitempty,min=1,max=10"`
	Pace              *int    `json:"pace,omitempty" validate:"omitempty,min=1,max=10"`
	Staff             *int    `json:"staff,omitempty" validate:"omitempty,min=1,max=10"`
}

// ReviewResponse represents review data for API responses
type ReviewResponse struct {
	ID                uint    `json:"id"`
	CourseID          uint    `json:"course_id"`
	CourseName        string  `json:"course_name"`
	UserID            uint    `json:"user_id"`
	UserName          string  `json:"user_name"`
	UserDisplayName   *string `json:"user_display_name"`
	OverallRating     int     `json:"overall_rating"`
	ReviewText        *string `json:"review_text"`
	Price             *string `json:"price"`
	HandicapDifficulty *int   `json:"handicap_difficulty"`
	HazardDifficulty  *int    `json:"hazard_difficulty"`
	Merch             *string `json:"merch"`
	Condition         *string `json:"condition"`
	EnjoymentRating   *int    `json:"enjoyment_rating"`
	Vibe              *string `json:"vibe"`
	Range             *string `json:"range"`
	Amenities         *string `json:"amenities"`
	Food              *string `json:"food"`
	Atmosphere        *string `json:"atmosphere"`
	Value             *int    `json:"value"`
	Maintenance       *int    `json:"maintenance"`
	Pace              *int    `json:"pace"`
	Staff             *int    `json:"staff"`
	CreatedAt         int64   `json:"created_at"`
	UpdatedAt         int64   `json:"updated_at"`
	// Additional metadata
	CanEdit           bool    `json:"can_edit"`
	HelpfulCount      int     `json:"helpful_count"`
	IsHelpful         *bool   `json:"is_helpful,omitempty"` // null if not authenticated
}

// ReviewSummaryResponse represents aggregated review data for a course
type ReviewSummaryResponse struct {
	CourseID        uint              `json:"course_id"`
	TotalReviews    int               `json:"total_reviews"`
	AverageRating   float64           `json:"average_rating"`
	RatingBreakdown map[string]int    `json:"rating_breakdown"` // "1": 5, "2": 10, etc.
	Categories      map[string]float64 `json:"categories"`       // average ratings per category
	RecentReviews   []*ReviewResponse `json:"recent_reviews"`   // last 3 reviews
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(dbService ReviewDatabaseServiceInterface) *ReviewHandler {
	return &ReviewHandler{
		dbService: dbService,
	}
}

// GetCourseReviews returns reviews for a specific course
func (h *ReviewHandler) GetCourseReviews(c echo.Context) error {
	courseIDParam := c.Param("courseId")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid course ID")
	}

	// Get pagination parameters
	pagination := GetPagination(c)

	// Get sort parameters
	sortBy := c.QueryParam("sort_by") // "rating", "date", "helpful"
	if sortBy == "" {
		sortBy = "date"
	}
	
	sortOrder := c.QueryParam("sort_order") // "asc", "desc"
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sort parameters
	validSortFields := []string{"rating", "date", "helpful"}
	if !contains(validSortFields, sortBy) {
		return BadRequestError(c, "Invalid sort field")
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		return BadRequestError(c, "Sort order must be 'asc' or 'desc'")
	}

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	// Verify course exists
	courseExists, err := h.dbService.CourseExists(uint(courseID))
	if err != nil {
		return InternalServerError(c, "Failed to verify course")
	}
	if !courseExists {
		return NotFoundError(c, "Course")
	}

	// Get reviews
	reviews, total, err := h.dbService.GetCourseReviews(uint(courseID), userID, sortBy, sortOrder, pagination.Page, pagination.PerPage)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve reviews")
	}

	// Create paginated response
	meta := &APIMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
	}

	return SuccessResponseWithMeta(c, reviews, meta)
}

// GetCourseReviewSummary returns aggregated review data for a course
func (h *ReviewHandler) GetCourseReviewSummary(c echo.Context) error {
	courseIDParam := c.Param("courseId")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid course ID")
	}

	// Verify course exists
	courseExists, err := h.dbService.CourseExists(uint(courseID))
	if err != nil {
		return InternalServerError(c, "Failed to verify course")
	}
	if !courseExists {
		return NotFoundError(c, "Course")
	}

	summary, err := h.dbService.GetCourseReviewSummary(uint(courseID))
	if err != nil {
		return InternalServerError(c, "Failed to retrieve review summary")
	}

	return SuccessResponse(c, summary)
}

// CreateReview creates a new review for a course
func (h *ReviewHandler) CreateReview(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	var req ReviewCreateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate required fields
	validationErrors := make(map[string]string)
	
	if req.CourseID == 0 {
		validationErrors["course_id"] = "Course ID is required"
	}
	
	if req.OverallRating < 1 || req.OverallRating > 10 {
		validationErrors["overall_rating"] = "Overall rating must be between 1 and 10"
	}

	// Validate optional rating fields
	if req.HandicapDifficulty != nil && (*req.HandicapDifficulty < 1 || *req.HandicapDifficulty > 10) {
		validationErrors["handicap_difficulty"] = "Handicap difficulty must be between 1 and 10"
	}

	if req.HazardDifficulty != nil && (*req.HazardDifficulty < 1 || *req.HazardDifficulty > 10) {
		validationErrors["hazard_difficulty"] = "Hazard difficulty must be between 1 and 10"
	}

	if req.EnjoymentRating != nil && (*req.EnjoymentRating < 1 || *req.EnjoymentRating > 10) {
		validationErrors["enjoyment_rating"] = "Enjoyment rating must be between 1 and 10"
	}

	if req.Value != nil && (*req.Value < 1 || *req.Value > 10) {
		validationErrors["value"] = "Value rating must be between 1 and 10"
	}

	if req.Maintenance != nil && (*req.Maintenance < 1 || *req.Maintenance > 10) {
		validationErrors["maintenance"] = "Maintenance rating must be between 1 and 10"
	}

	if req.Pace != nil && (*req.Pace < 1 || *req.Pace > 10) {
		validationErrors["pace"] = "Pace rating must be between 1 and 10"
	}

	if req.Staff != nil && (*req.Staff < 1 || *req.Staff > 10) {
		validationErrors["staff"] = "Staff rating must be between 1 and 10"
	}

	if req.ReviewText != nil && len(*req.ReviewText) > 2000 {
		validationErrors["review_text"] = "Review text must be 2000 characters or less"
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

	// Check if user already has a review for this course
	hasReview, err := h.dbService.UserHasReviewForCourse(userID, req.CourseID)
	if err != nil {
		return InternalServerError(c, "Failed to check existing review")
	}
	if hasReview {
		return ConflictError(c, "You have already reviewed this course. Use PUT to update your review.")
	}

	// Create review
	review, err := h.dbService.CreateReview(userID, &req)
	if err != nil {
		return InternalServerError(c, "Failed to create review")
	}

	return CreatedResponse(c, review)
}

// UpdateReview updates an existing review
func (h *ReviewHandler) UpdateReview(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	reviewIDParam := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid review ID")
	}

	// Check if user owns the review
	isOwner, err := h.dbService.IsUserReviewOwner(userID, uint(reviewID))
	if err != nil {
		return NotFoundError(c, "Review")
	}
	if !isOwner {
		return ForbiddenError(c, "You can only edit your own reviews")
	}

	var req ReviewUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate fields if provided
	validationErrors := make(map[string]string)
	
	if req.OverallRating != nil && (*req.OverallRating < 1 || *req.OverallRating > 10) {
		validationErrors["overall_rating"] = "Overall rating must be between 1 and 10"
	}

	// Similar validation for other fields...
	if len(validationErrors) > 0 {
		return ValidationError(c, validationErrors)
	}

	// Update review
	review, err := h.dbService.UpdateReview(uint(reviewID), &req)
	if err != nil {
		return InternalServerError(c, "Failed to update review")
	}

	return SuccessResponse(c, review)
}

// DeleteReview deletes a review
func (h *ReviewHandler) DeleteReview(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	reviewIDParam := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid review ID")
	}

	// Check if user owns the review
	isOwner, err := h.dbService.IsUserReviewOwner(userID, uint(reviewID))
	if err != nil {
		return NotFoundError(c, "Review")
	}
	if !isOwner {
		return ForbiddenError(c, "You can only delete your own reviews")
	}

	// Delete review
	err = h.dbService.DeleteReview(uint(reviewID))
	if err != nil {
		return InternalServerError(c, "Failed to delete review")
	}

	return NoContentResponse(c)
}

// GetUserReviews returns reviews by the authenticated user
func (h *ReviewHandler) GetUserReviews(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	// Get pagination parameters
	pagination := GetPagination(c)

	reviews, total, err := h.dbService.GetUserReviews(userID, pagination.Page, pagination.PerPage)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve user reviews")
	}

	// Create paginated response
	meta := &APIMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
	}

	return SuccessResponseWithMeta(c, reviews, meta)
}

// MarkReviewHelpful marks a review as helpful or not helpful
func (h *ReviewHandler) MarkReviewHelpful(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	reviewIDParam := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid review ID")
	}

	var req struct {
		Helpful bool `json:"helpful"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Cannot mark own review as helpful
	isOwner, err := h.dbService.IsUserReviewOwner(userID, uint(reviewID))
	if err != nil {
		return NotFoundError(c, "Review")
	}
	if isOwner {
		return BadRequestError(c, "You cannot mark your own review as helpful")
	}

	err = h.dbService.SetReviewHelpfulness(userID, uint(reviewID), req.Helpful)
	if err != nil {
		return InternalServerError(c, "Failed to update review helpfulness")
	}

	return SuccessResponse(c, map[string]string{
		"message": "Review helpfulness updated",
	})
}

// RegisterRoutes registers review-related routes
func (h *ReviewHandler) RegisterRoutes(g *echo.Group, jwtService *JWTService) {
	// Public routes (optionally authenticated)
	g.GET("/courses/:courseId/reviews", h.GetCourseReviews, OptionalJWTMiddleware(jwtService))
	g.GET("/courses/:courseId/reviews/summary", h.GetCourseReviewSummary)
	
	// Protected routes (authentication required)
	g.POST("/reviews", h.CreateReview, JWTMiddleware(jwtService))
	g.PUT("/reviews/:id", h.UpdateReview, JWTMiddleware(jwtService))
	g.DELETE("/reviews/:id", h.DeleteReview, JWTMiddleware(jwtService))
	g.GET("/reviews/user", h.GetUserReviews, JWTMiddleware(jwtService))
	g.POST("/reviews/:id/helpful", h.MarkReviewHelpful, JWTMiddleware(jwtService))
}

// Extended database interface for review operations
type ReviewDatabaseServiceInterface interface {
	CoursesDatabaseServiceInterface
	GetCourseReviews(courseID uint, userID *uint, sortBy, sortOrder string, page, perPage int) ([]*ReviewResponse, int, error)
	GetCourseReviewSummary(courseID uint) (*ReviewSummaryResponse, error)
	CreateReview(userID uint, req *ReviewCreateRequest) (*ReviewResponse, error)
	UpdateReview(reviewID uint, req *ReviewUpdateRequest) (*ReviewResponse, error)
	DeleteReview(reviewID uint) error
	GetUserReviews(userID uint, page, perPage int) ([]*ReviewResponse, int, error)
	IsUserReviewOwner(userID, reviewID uint) (bool, error)
	UserHasReviewForCourse(userID, courseID uint) (bool, error)
	SetReviewHelpfulness(userID, reviewID uint, helpful bool) error
}