package api

import (
	"time"

	"github.com/labstack/echo/v4"
)

// APIRouter handles API route registration and configuration
type APIRouter struct {
	jwtService    *JWTService
	authHandler   *AuthHandler
	userHandler   *UserHandler
	courseHandler *CourseHandler
	reviewHandler *ReviewHandler
	mapHandler    *MapHandler
}

// NewAPIRouter creates a new API router with all handlers
func NewAPIRouter(
	jwtService *JWTService,
	authHandler *AuthHandler,
	userHandler *UserHandler,
	courseHandler *CourseHandler,
	reviewHandler *ReviewHandler,
	mapHandler *MapHandler,
) *APIRouter {
	return &APIRouter{
		jwtService:    jwtService,
		authHandler:   authHandler,
		userHandler:   userHandler,
		courseHandler: courseHandler,
		reviewHandler: reviewHandler,
		mapHandler:    mapHandler,
	}
}

// SetupRoutes configures all API routes with proper middleware
func (r *APIRouter) SetupRoutes(e *echo.Echo, apiConfig *APIConfig) {
	// Create API group with versioning
	apiGroup := e.Group("/api/v1")
	
	// Setup API middleware
	SetupAPIMiddleware(apiGroup, apiConfig)

	// Health check endpoint (no middleware needed)
	apiGroup.GET("/health", r.healthCheck)

	// Register handler routes
	r.authHandler.RegisterRoutes(apiGroup, r.jwtService)
	r.userHandler.RegisterRoutes(apiGroup, r.jwtService)
	r.courseHandler.RegisterRoutes(apiGroup, r.jwtService)
	r.reviewHandler.RegisterRoutes(apiGroup, r.jwtService)
	r.mapHandler.RegisterRoutes(apiGroup, r.jwtService)

	// Register additional API routes
	r.registerStatisticsRoutes(apiGroup)
	r.registerUtilityRoutes(apiGroup)
}

// healthCheck provides API health status
func (r *APIRouter) healthCheck(c echo.Context) error {
	return SuccessResponse(c, HealthResponse{
		Status:      "healthy",
		Version:     "1.0.0",
		Environment: "production", // TODO: Get from config
		Services: map[string]string{
			"database":    "connected",
			"auth":        "operational",
			"geocoding":   "operational",
		},
		Timestamp: time.Now().Unix(),
	})
}

// registerStatisticsRoutes registers statistics and analytics endpoints
func (r *APIRouter) registerStatisticsRoutes(g *echo.Group) {
	statsGroup := g.Group("/stats")
	
	// Public statistics
	statsGroup.GET("/courses", r.getCourseStatistics)
	statsGroup.GET("/reviews", r.getReviewStatistics)
	statsGroup.GET("/users", r.getUserStatistics)
	
	// Protected user-specific statistics
	statsGroup.GET("/user/dashboard", r.getUserDashboard, JWTMiddleware(r.jwtService))
}

// registerUtilityRoutes registers utility endpoints
func (r *APIRouter) registerUtilityRoutes(g *echo.Group) {
	utilGroup := g.Group("/utils")
	
	// Validation utilities
	utilGroup.POST("/validate/course", r.validateCourseData)
	utilGroup.POST("/validate/review", r.validateReviewData)
	
	// Search suggestions
	utilGroup.GET("/search/suggestions", r.getSearchSuggestions)
	
	// Export utilities (protected)
	utilGroup.GET("/export/user-data", r.exportUserData, JWTMiddleware(r.jwtService))
}

// Statistics handlers

func (r *APIRouter) getCourseStatistics(c echo.Context) error {
	// TODO: Implement course statistics
	stats := map[string]interface{}{
		"total_courses": 0,
		"courses_with_reviews": 0,
		"average_rating": 0.0,
		"most_popular_states": []string{},
	}
	return SuccessResponse(c, stats)
}

func (r *APIRouter) getReviewStatistics(c echo.Context) error {
	// TODO: Implement review statistics
	stats := map[string]interface{}{
		"total_reviews": 0,
		"average_rating": 0.0,
		"reviews_per_month": []int{},
		"rating_distribution": map[string]int{},
	}
	return SuccessResponse(c, stats)
}

func (r *APIRouter) getUserStatistics(c echo.Context) error {
	// TODO: Implement user statistics
	stats := map[string]interface{}{
		"total_users": 0,
		"active_users": 0,
		"new_users_this_month": 0,
		"average_handicap": 0.0,
	}
	return SuccessResponse(c, stats)
}

func (r *APIRouter) getUserDashboard(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	// TODO: Implement user dashboard data
	dashboard := map[string]interface{}{
		"user_id": userID,
		"recent_scores": []interface{}{},
		"recent_reviews": []interface{}{},
		"favorite_courses": []interface{}{},
		"achievements": []interface{}{},
		"statistics": map[string]interface{}{
			"total_rounds": 0,
			"average_score": 0.0,
			"handicap_trend": "stable",
			"courses_played": 0,
		},
	}
	return SuccessResponse(c, dashboard)
}

// Utility handlers

func (r *APIRouter) validateCourseData(c echo.Context) error {
	var courseData map[string]interface{}
	if err := c.Bind(&courseData); err != nil {
		return BadRequestError(c, "Invalid course data format")
	}

	// TODO: Implement course data validation
	validation := map[string]interface{}{
		"valid": true,
		"errors": []string{},
		"warnings": []string{},
	}
	
	return SuccessResponse(c, validation)
}

func (r *APIRouter) validateReviewData(c echo.Context) error {
	var reviewData map[string]interface{}
	if err := c.Bind(&reviewData); err != nil {
		return BadRequestError(c, "Invalid review data format")
	}

	// TODO: Implement review data validation
	validation := map[string]interface{}{
		"valid": true,
		"errors": []string{},
		"warnings": []string{},
	}
	
	return SuccessResponse(c, validation)
}

func (r *APIRouter) getSearchSuggestions(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return BadRequestError(c, "Search query is required")
	}

	// TODO: Implement search suggestions
	suggestions := map[string]interface{}{
		"courses": []string{},
		"locations": []string{},
		"users": []string{},
	}
	
	return SuccessResponse(c, suggestions)
}

func (r *APIRouter) exportUserData(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return UnauthorizedError(c, "Authentication required")
	}

	format := c.QueryParam("format")
	if format == "" {
		format = "json"
	}

	if format != "json" && format != "csv" {
		return BadRequestError(c, "Format must be 'json' or 'csv'")
	}

	// TODO: Implement user data export
	exportData := map[string]interface{}{
		"user_id": userID,
		"format": format,
		"export_url": "https://api.example.com/exports/user-123.json",
		"expires_at": "2024-01-01T00:00:00Z",
	}
	
	return SuccessResponse(c, exportData)
}

// APIFactory creates all API components
type APIFactory struct {
	dbService DatabaseServiceInterface
	config    *APIConfig
}

// NewAPIFactory creates a new API factory
func NewAPIFactory(dbService DatabaseServiceInterface, config *APIConfig) *APIFactory {
	return &APIFactory{
		dbService: dbService,
		config:    config,
	}
}

// CreateAPIRouter creates a fully configured API router
func (f *APIFactory) CreateAPIRouter() *APIRouter {
	// Create handlers
	authHandler := NewAuthHandler(f.config.JWTService, f.dbService, "", "") // TODO: Add Google config
	userHandler := NewUserHandler(f.dbService.(ExtendedDatabaseServiceInterface))
	courseHandler := NewCourseHandler(f.dbService.(CoursesDatabaseServiceInterface))
	reviewHandler := NewReviewHandler(f.dbService.(ReviewDatabaseServiceInterface))
	mapHandler := NewMapHandler(f.dbService.(MapDatabaseServiceInterface))

	return NewAPIRouter(
		f.config.JWTService,
		authHandler,
		userHandler,
		courseHandler,
		reviewHandler,
		mapHandler,
	)
}