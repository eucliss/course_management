package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// CourseHandler handles course-related API endpoints
type CourseHandler struct {
	dbService CoursesDatabaseServiceInterface
}

// CourseCreateRequest represents course creation request
type CourseCreateRequest struct {
	Name        string     `json:"name" validate:"required,min=3,max=100"`
	Address     string     `json:"address" validate:"required,min=10,max=200"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	Phone       *string    `json:"phone,omitempty" validate:"omitempty,max=20"`
	Website     *string    `json:"website,omitempty" validate:"omitempty,url,max=200"`
	Holes       []HoleData `json:"holes,omitempty" validate:"dive"`
}

// CourseUpdateRequest represents course update request
type CourseUpdateRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Address     *string    `json:"address,omitempty" validate:"omitempty,min=10,max=200"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	Phone       *string    `json:"phone,omitempty" validate:"omitempty,max=20"`
	Website     *string    `json:"website,omitempty" validate:"omitempty,url,max=200"`
	Holes       []HoleData `json:"holes,omitempty" validate:"dive"`
}

// HoleData represents hole information
type HoleData struct {
	Number      int     `json:"number" validate:"required,min=1,max=18"`
	Par         int     `json:"par" validate:"required,min=3,max=6"`
	Yardage     int     `json:"yardage" validate:"required,min=50,max=800"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=200"`
}

// CourseResponse represents course data for API responses
type CourseResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Description *string   `json:"description"`
	Phone       *string   `json:"phone"`
	Website     *string   `json:"website"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	Holes       []HoleData `json:"holes"`
	CreatedBy   *uint     `json:"created_by"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
	// Additional fields for authenticated users
	CanEdit     bool               `json:"can_edit"`
	UserReview  *UserReviewSummary `json:"user_review,omitempty"`
	Stats       *CourseStats       `json:"stats,omitempty"`
}

// UserReviewSummary represents user's review summary for a course
type UserReviewSummary struct {
	OverallRating int     `json:"overall_rating"`
	LastPlayed    *int64  `json:"last_played"`
	TimesPlayed   int     `json:"times_played"`
	BestScore     *int    `json:"best_score"`
	AverageScore  *float64 `json:"average_score"`
}

// CourseStats represents course statistics
type CourseStats struct {
	TotalReviews    int     `json:"total_reviews"`
	AverageRating   float64 `json:"average_rating"`
	TotalRounds     int     `json:"total_rounds"`
	AverageScore    float64 `json:"average_score"`
	DifficultyLevel string  `json:"difficulty_level"` // "Easy", "Medium", "Hard"
}

// CourseSearchRequest represents search parameters
type CourseSearchRequest struct {
	Query     string   `query:"q"`
	Latitude  *float64 `query:"lat"`
	Longitude *float64 `query:"lng"`
	Radius    *float64 `query:"radius"` // in kilometers
	MinRating *float64 `query:"min_rating"`
	MaxRating *float64 `query:"max_rating"`
	SortBy    string   `query:"sort_by"` // "name", "rating", "distance", "created_at"
	SortOrder string   `query:"sort_order"` // "asc", "desc"
}

// NewCourseHandler creates a new course handler
func NewCourseHandler(dbService CoursesDatabaseServiceInterface) *CourseHandler {
	return &CourseHandler{
		dbService: dbService,
	}
}

// GetCourses returns a paginated list of courses
func (h *CourseHandler) GetCourses(c echo.Context) error {
	// Get pagination parameters
	pagination := GetPagination(c)
	
	// Get search parameters
	var search CourseSearchRequest
	if err := c.Bind(&search); err != nil {
		return BadRequestError(c, "Invalid search parameters")
	}

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	// Get courses from database
	courses, total, err := h.dbService.GetCourses(&search, userID, pagination.Page, pagination.PerPage)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve courses")
	}

	// Create paginated response
	meta := &APIMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
	}

	return SuccessResponseWithMeta(c, courses, meta)
}

// GetCourse returns a specific course by ID
func (h *CourseHandler) GetCourse(c echo.Context) error {
	courseIDParam := c.Param("id")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid course ID")
	}

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	course, err := h.dbService.GetCourseByID(uint(courseID), userID)
	if err != nil {
		return NotFoundError(c, "Course")
	}

	return SuccessResponse(c, course)
}

// CreateCourse creates a new course
func (h *CourseHandler) CreateCourse(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	var req CourseCreateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate required fields
	validationErrors := make(map[string]string)
	
	if strings.TrimSpace(req.Name) == "" {
		validationErrors["name"] = "Course name is required"
	} else if len(req.Name) < 3 || len(req.Name) > 100 {
		validationErrors["name"] = "Course name must be between 3 and 100 characters"
	}
	
	if strings.TrimSpace(req.Address) == "" {
		validationErrors["address"] = "Course address is required"
	} else if len(req.Address) < 10 || len(req.Address) > 200 {
		validationErrors["address"] = "Course address must be between 10 and 200 characters"
	}

	if req.Website != nil && *req.Website != "" && !isValidURL(*req.Website) {
		validationErrors["website"] = "Invalid website URL format"
	}

	// Validate holes if provided
	if len(req.Holes) > 0 {
		if len(req.Holes) > 18 {
			validationErrors["holes"] = "Maximum 18 holes allowed"
		}
		
		for i, hole := range req.Holes {
			if hole.Number < 1 || hole.Number > 18 {
				validationErrors[fmt.Sprintf("holes[%d].number", i)] = "Hole number must be between 1 and 18"
			}
			if hole.Par < 3 || hole.Par > 6 {
				validationErrors[fmt.Sprintf("holes[%d].par", i)] = "Par must be between 3 and 6"
			}
			if hole.Yardage < 50 || hole.Yardage > 800 {
				validationErrors[fmt.Sprintf("holes[%d].yardage", i)] = "Yardage must be between 50 and 800"
			}
		}
	}

	if len(validationErrors) > 0 {
		return ValidationError(c, validationErrors)
	}

	// Check if course already exists
	exists, err := h.dbService.CourseExistsByNameAndAddress(req.Name, req.Address)
	if err != nil {
		return InternalServerError(c, "Failed to check course existence")
	}
	if exists {
		return ConflictError(c, "A course with this name and address already exists")
	}

	// Create course
	course, err := h.dbService.CreateCourse(userID, &req)
	if err != nil {
		return InternalServerError(c, "Failed to create course")
	}

	return CreatedResponse(c, course)
}

// UpdateCourse updates an existing course
func (h *CourseHandler) UpdateCourse(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	courseIDParam := c.Param("id")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid course ID")
	}

	// Check if user owns the course
	isOwner, err := h.dbService.IsUserCourseOwner(userID, uint(courseID))
	if err != nil {
		return NotFoundError(c, "Course")
	}
	if !isOwner {
		return ForbiddenError(c, "You can only edit courses you created")
	}

	var req CourseUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate fields if provided
	validationErrors := make(map[string]string)
	
	if req.Name != nil {
		if len(*req.Name) < 3 || len(*req.Name) > 100 {
			validationErrors["name"] = "Course name must be between 3 and 100 characters"
		}
	}
	
	if req.Address != nil {
		if len(*req.Address) < 10 || len(*req.Address) > 200 {
			validationErrors["address"] = "Course address must be between 10 and 200 characters"
		}
	}

	if req.Website != nil && *req.Website != "" && !isValidURL(*req.Website) {
		validationErrors["website"] = "Invalid website URL format"
	}

	if len(validationErrors) > 0 {
		return ValidationError(c, validationErrors)
	}

	// Update course
	course, err := h.dbService.UpdateCourse(uint(courseID), &req)
	if err != nil {
		return InternalServerError(c, "Failed to update course")
	}

	return SuccessResponse(c, course)
}

// DeleteCourse deletes a course
func (h *CourseHandler) DeleteCourse(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	courseIDParam := c.Param("id")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return BadRequestError(c, "Invalid course ID")
	}

	// Check if user owns the course
	isOwner, err := h.dbService.IsUserCourseOwner(userID, uint(courseID))
	if err != nil {
		return NotFoundError(c, "Course")
	}
	if !isOwner {
		return ForbiddenError(c, "You can only delete courses you created")
	}

	// Check if course has associated data (scores, reviews)
	hasData, err := h.dbService.CourseHasAssociatedData(uint(courseID))
	if err != nil {
		return InternalServerError(c, "Failed to check course data")
	}
	if hasData {
		return ConflictError(c, "Cannot delete course with existing reviews or scores")
	}

	// Delete course
	err = h.dbService.DeleteCourse(uint(courseID))
	if err != nil {
		return InternalServerError(c, "Failed to delete course")
	}

	return NoContentResponse(c)
}

// SearchCourses performs course search with various filters
func (h *CourseHandler) SearchCourses(c echo.Context) error {
	// Get pagination parameters
	pagination := GetPagination(c)
	
	// Get search parameters
	var search CourseSearchRequest
	if err := c.Bind(&search); err != nil {
		return BadRequestError(c, "Invalid search parameters")
	}

	// Validate search parameters
	if search.Radius != nil && (*search.Radius < 0 || *search.Radius > 1000) {
		return BadRequestError(c, "Radius must be between 0 and 1000 kilometers")
	}

	if search.MinRating != nil && (*search.MinRating < 0 || *search.MinRating > 10) {
		return BadRequestError(c, "Minimum rating must be between 0 and 10")
	}

	if search.MaxRating != nil && (*search.MaxRating < 0 || *search.MaxRating > 10) {
		return BadRequestError(c, "Maximum rating must be between 0 and 10")
	}

	// Validate sort parameters
	validSortFields := []string{"name", "rating", "distance", "created_at"}
	if search.SortBy != "" && !contains(validSortFields, search.SortBy) {
		return BadRequestError(c, "Invalid sort field")
	}

	if search.SortOrder != "" && search.SortOrder != "asc" && search.SortOrder != "desc" {
		return BadRequestError(c, "Sort order must be 'asc' or 'desc'")
	}

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	// Perform search
	courses, total, err := h.dbService.SearchCourses(&search, userID, pagination.Page, pagination.PerPage)
	if err != nil {
		return InternalServerError(c, "Search failed")
	}

	// Create paginated response
	meta := &APIMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
	}

	return SuccessResponseWithMeta(c, courses, meta)
}

// GetNearbyCourses returns courses near a location
func (h *CourseHandler) GetNearbyCourses(c echo.Context) error {
	latParam := c.QueryParam("lat")
	lngParam := c.QueryParam("lng")
	radiusParam := c.QueryParam("radius")

	if latParam == "" || lngParam == "" {
		return BadRequestError(c, "Latitude and longitude are required")
	}

	latitude, err := strconv.ParseFloat(latParam, 64)
	if err != nil {
		return BadRequestError(c, "Invalid latitude")
	}

	longitude, err := strconv.ParseFloat(lngParam, 64)
	if err != nil {
		return BadRequestError(c, "Invalid longitude")
	}

	radius := 10.0 // Default 10km
	if radiusParam != "" {
		r, err := strconv.ParseFloat(radiusParam, 64)
		if err != nil {
			return BadRequestError(c, "Invalid radius")
		}
		if r > 0 && r <= 100 {
			radius = r
		}
	}

	// Get pagination parameters
	pagination := GetPagination(c)

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	courses, total, err := h.dbService.GetNearbyCoures(latitude, longitude, radius, userID, pagination.Page, pagination.PerPage)
	if err != nil {
		return InternalServerError(c, "Failed to find nearby courses")
	}

	// Create paginated response
	meta := &APIMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
	}

	return SuccessResponseWithMeta(c, courses, meta)
}

// RegisterRoutes registers course-related routes
func (h *CourseHandler) RegisterRoutes(g *echo.Group, jwtService *JWTService) {
	// Public routes (optionally authenticated)
	g.GET("/courses", h.GetCourses, OptionalJWTMiddleware(jwtService))
	g.GET("/courses/search", h.SearchCourses, OptionalJWTMiddleware(jwtService))
	g.GET("/courses/nearby", h.GetNearbyCourses, OptionalJWTMiddleware(jwtService))
	g.GET("/courses/:id", h.GetCourse, OptionalJWTMiddleware(jwtService))
	
	// Protected routes (authentication required)
	g.POST("/courses", h.CreateCourse, JWTMiddleware(jwtService))
	g.PUT("/courses/:id", h.UpdateCourse, JWTMiddleware(jwtService))
	g.DELETE("/courses/:id", h.DeleteCourse, JWTMiddleware(jwtService))
}

// Helper functions

// isValidURL performs basic URL validation
func isValidURL(url string) bool {
	return strings.HasPrefix(strings.ToLower(url), "http://") || 
		   strings.HasPrefix(strings.ToLower(url), "https://")
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Extended database interface for course operations
type CoursesDatabaseServiceInterface interface {
	ExtendedDatabaseServiceInterface
	GetCourses(search *CourseSearchRequest, userID *uint, page, perPage int) ([]*CourseResponse, int, error)
	GetCourseByID(courseID uint, userID *uint) (*CourseResponse, error)
	CreateCourse(userID uint, req *CourseCreateRequest) (*CourseResponse, error)
	UpdateCourse(courseID uint, req *CourseUpdateRequest) (*CourseResponse, error)
	DeleteCourse(courseID uint) error
	SearchCourses(search *CourseSearchRequest, userID *uint, page, perPage int) ([]*CourseResponse, int, error)
	GetNearbyCoures(lat, lng, radius float64, userID *uint, page, perPage int) ([]*CourseResponse, int, error)
	CourseExistsByNameAndAddress(name, address string) (bool, error)
	IsUserCourseOwner(userID, courseID uint) (bool, error)
	CourseHasAssociatedData(courseID uint) (bool, error)
}