package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"course_management/services"
)

// ServiceLayerHandlers demonstrates how to use the new service layer
type ServiceLayerHandlers struct {
	// No direct dependencies - services are injected via context
}

func NewServiceLayerHandlers() *ServiceLayerHandlers {
	return &ServiceLayerHandlers{}
}

// Example: Home handler using service layer
func (h *ServiceLayerHandlers) Home(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	courseService := GetCourseService(c)
	sessionService := GetSessionService(c)
	
	// Get user from session
	user := sessionService.GetUser(c)
	
	// Get all courses using service layer
	courses, err := courseService.GetAllCourses(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses")
	}
	
	// Build edit permissions
	editPermissions := make(map[int]bool)
	if user != nil {
		userID := sessionService.GetUserID(c)
		if userID != nil {
			for i, course := range courses {
				canEdit, err := courseService.CanEditCourse(ctx, uint(course.ID), *userID)
				if err == nil {
					editPermissions[i] = canEdit
				}
			}
		}
	}
	
	data := struct {
		Courses         []services.Course
		MapboxToken     string
		User            *services.GoogleUser
		EditPermissions map[int]bool
	}{
		Courses:         courses,
		MapboxToken:     getEnv("MAPBOX_ACCESS_TOKEN", ""),
		User:            user,
		EditPermissions: editPermissions,
	}
	
	return c.Render(http.StatusOK, "welcome", data)
}

// Example: Create course handler using service layer
func (h *ServiceLayerHandlers) CreateCourse(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	courseService := GetCourseService(c)
	sessionService := GetSessionService(c)
	
	// Require authentication
	_, err := sessionService.RequireAuth(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	
	userID := sessionService.GetUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "User not found")
	}
	
	// Parse form data
	form := c.Request().Form
	formMap := services.ConvertFormValues(form)
	
	holes, scores, err := courseService.ParseCourseForm(formMap)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}
	
	// Create course object
	course := services.Course{
		Name:          c.FormValue("name"),
		Description:   c.FormValue("description"),
		Address:       c.FormValue("address"),
		Review:        c.FormValue("review"),
		OverallRating: c.FormValue("overallRating"),
		Holes:         holes,
		Scores:        scores,
		// Rankings would be parsed from form as well
	}
	
	// Create course using service layer
	if err := courseService.CreateCourse(ctx, course, userID); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	
	return c.Redirect(http.StatusSeeOther, "/")
}

// Example: Update course handler using service layer
func (h *ServiceLayerHandlers) UpdateCourse(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	courseService := GetCourseService(c)
	sessionService := GetSessionService(c)
	
	// Require authentication
	_, err := sessionService.RequireAuth(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	
	userID := sessionService.GetUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "User not found")
	}
	
	// Get course ID from URL parameter
	courseIDParam := c.Param("id")
	courseIndex, err := strconv.Atoi(courseIDParam)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}
	
	// Get existing course
	existingCourse, err := courseService.GetCourseByIndex(ctx, courseIndex)
	if err != nil {
		return c.String(http.StatusNotFound, "Course not found")
	}
	
	// Check edit permissions
	canEdit, err := courseService.CanEditCourse(ctx, uint(existingCourse.ID), *userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to check permissions")
	}
	if !canEdit {
		return c.String(http.StatusForbidden, "You don't have permission to edit this course")
	}
	
	// Parse form data
	form := c.Request().Form
	formMap := services.ConvertFormValues(form)
	
	holes, scores, err := courseService.ParseCourseForm(formMap)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}
	
	// Update course object
	updatedCourse := *existingCourse
	updatedCourse.Name = c.FormValue("name")
	updatedCourse.Description = c.FormValue("description")
	updatedCourse.Address = c.FormValue("address")
	updatedCourse.Review = c.FormValue("review")
	updatedCourse.OverallRating = c.FormValue("overallRating")
	updatedCourse.Holes = holes
	updatedCourse.Scores = scores
	
	// Update course using service layer
	if err := courseService.UpdateCourse(ctx, updatedCourse, userID); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	
	return c.Redirect(http.StatusSeeOther, "/")
}

// Example: Delete course handler using service layer
func (h *ServiceLayerHandlers) DeleteCourse(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	courseService := GetCourseService(c)
	sessionService := GetSessionService(c)
	
	// Require authentication
	_, err := sessionService.RequireAuth(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	
	userID := sessionService.GetUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "User not found")
	}
	
	// Get course ID from URL parameter
	courseIDParam := c.Param("id")
	courseIndex, err := strconv.Atoi(courseIDParam)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}
	
	// Get existing course
	existingCourse, err := courseService.GetCourseByIndex(ctx, courseIndex)
	if err != nil {
		return c.String(http.StatusNotFound, "Course not found")
	}
	
	// Delete course using service layer
	if err := courseService.DeleteCourse(ctx, uint(existingCourse.ID), *userID); err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	
	return c.NoContent(http.StatusNoContent)
}

// Example: Create review handler using service layer
func (h *ServiceLayerHandlers) CreateReview(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	reviewService := GetReviewService(c)
	sessionService := GetSessionService(c)
	
	// Require authentication
	_, err := sessionService.RequireAuth(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	
	userID := sessionService.GetUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "User not found")
	}
	
	// Parse form data
	courseID, err := strconv.ParseUint(c.FormValue("course_id"), 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}
	
	
	// Create review object
	reviewText := c.FormValue("review")
	overallRating := c.FormValue("overall_rating")
	
	review := services.CourseReview{
		UserID:        *userID,
		CourseID:      uint(courseID),
		ReviewText:    &reviewText,
		OverallRating: &overallRating,
	}
	
	// Create review using service layer
	if err := reviewService.CreateReview(ctx, review); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	
	return c.Redirect(http.StatusSeeOther, "/")
}

// Example: Get user reviews handler using service layer
func (h *ServiceLayerHandlers) GetUserReviews(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	reviewService := GetReviewService(c)
	sessionService := GetSessionService(c)
	
	// Require authentication
	_, err := sessionService.RequireAuth(c)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	
	userID := sessionService.GetUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "User not found")
	}
	
	// Get user reviews using service layer
	reviews, err := reviewService.GetUserReviews(ctx, *userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load reviews")
	}
	
	return c.JSON(http.StatusOK, reviews)
}

// Example: Authentication handler using service layer
func (h *ServiceLayerHandlers) VerifyGoogleToken(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get services from context
	authService := GetAuthService(c)
	sessionService := GetSessionService(c)
	
	// Get token from request
	var tokenRequest struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&tokenRequest); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}
	
	// Verify token using service layer
	googleUser, err := authService.VerifyGoogleToken(ctx, tokenRequest.Token)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Invalid token")
	}
	
	// Create or update user using service layer
	user, err := authService.CreateOrUpdateUser(ctx, *googleUser)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create user")
	}
	
	// Set user in session
	if err := sessionService.SetUser(c, *user); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to set session")
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

// Helper function to get environment variable
func getEnv(key, defaultValue string) string {
	// This would normally use os.Getenv(key)
	_ = key // unused in this example
	return defaultValue
}

// Migration Guide Comments:
// 
// To migrate existing handlers to use the service layer:
// 1. Remove direct database service creation (NewDatabaseService())
// 2. Use services.GetCourseService(c) instead of direct database calls
// 3. Use services.GetSessionService(c) instead of NewSessionService()
// 4. Use services.GetAuthService(c) instead of NewAuthService()
// 5. Use services.GetReviewService(c) instead of NewReviewService()
// 6. Use ctx := c.Request().Context() for all service calls
// 7. Update main.go to use ServiceMiddleware
// 8. Initialize the service container in main.go
//
// Benefits:
// - Cleaner separation of concerns
// - Easier testing with interface mocking
// - Better error handling and validation
// - Consistent dependency injection
// - Centralized service management