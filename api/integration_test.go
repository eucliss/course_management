package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDatabaseService implements all database interfaces for testing
type MockDatabaseService struct {
	mock.Mock
}

// DatabaseServiceInterface methods
func (m *MockDatabaseService) CreateUser(googleID, email, name, picture string) (*UserResponse, error) {
	args := m.Called(googleID, email, name, picture)
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockDatabaseService) GetUserByGoogleID(googleID string) (*UserResponse, error) {
	args := m.Called(googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockDatabaseService) GetUserByEmail(email string) (*UserResponse, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

// ExtendedDatabaseServiceInterface methods
func (m *MockDatabaseService) GetUserByID(userID uint) (*UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockDatabaseService) UpdateUserProfile(userID uint, displayName *string) (*UserResponse, error) {
	args := m.Called(userID, displayName)
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockDatabaseService) UpdateUserHandicap(userID uint, handicap *float64) (*UserResponse, error) {
	args := m.Called(userID, handicap)
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockDatabaseService) GetUserScores(userID uint, courseID *uint, page, perPage int) ([]*UserScoreResponse, int, error) {
	args := m.Called(userID, courseID, page, perPage)
	return args.Get(0).([]*UserScoreResponse), args.Int(1), args.Error(2)
}

func (m *MockDatabaseService) CreateUserScore(userID uint, req *UserScoreCreateRequest) (*UserScoreResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*UserScoreResponse), args.Error(1)
}

func (m *MockDatabaseService) DeleteUserScore(scoreID uint) error {
	args := m.Called(scoreID)
	return args.Error(0)
}

func (m *MockDatabaseService) GetScoreOwner(scoreID uint) (uint, error) {
	args := m.Called(scoreID)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockDatabaseService) GetUserStats(userID uint) (*UserStatsResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(*UserStatsResponse), args.Error(1)
}

func (m *MockDatabaseService) CourseExists(courseID uint) (bool, error) {
	args := m.Called(courseID)
	return args.Bool(0), args.Error(1)
}

// CoursesDatabaseServiceInterface methods
func (m *MockDatabaseService) GetCourses(search *CourseSearchRequest, userID *uint, page, perPage int) ([]*CourseResponse, int, error) {
	args := m.Called(search, userID, page, perPage)
	return args.Get(0).([]*CourseResponse), args.Int(1), args.Error(2)
}

func (m *MockDatabaseService) GetCourseByID(courseID uint, userID *uint) (*CourseResponse, error) {
	args := m.Called(courseID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CourseResponse), args.Error(1)
}

func (m *MockDatabaseService) CreateCourse(userID uint, req *CourseCreateRequest) (*CourseResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*CourseResponse), args.Error(1)
}

func (m *MockDatabaseService) UpdateCourse(courseID uint, req *CourseUpdateRequest) (*CourseResponse, error) {
	args := m.Called(courseID, req)
	return args.Get(0).(*CourseResponse), args.Error(1)
}

func (m *MockDatabaseService) DeleteCourse(courseID uint) error {
	args := m.Called(courseID)
	return args.Error(0)
}

func (m *MockDatabaseService) SearchCourses(search *CourseSearchRequest, userID *uint, page, perPage int) ([]*CourseResponse, int, error) {
	args := m.Called(search, userID, page, perPage)
	return args.Get(0).([]*CourseResponse), args.Int(1), args.Error(2)
}

func (m *MockDatabaseService) GetNearbyCoures(lat, lng, radius float64, userID *uint, page, perPage int) ([]*CourseResponse, int, error) {
	args := m.Called(lat, lng, radius, userID, page, perPage)
	return args.Get(0).([]*CourseResponse), args.Int(1), args.Error(2)
}

func (m *MockDatabaseService) CourseExistsByNameAndAddress(name, address string) (bool, error) {
	args := m.Called(name, address)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabaseService) IsUserCourseOwner(userID, courseID uint) (bool, error) {
	args := m.Called(userID, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabaseService) CourseHasAssociatedData(courseID uint) (bool, error) {
	args := m.Called(courseID)
	return args.Bool(0), args.Error(1)
}

// ReviewDatabaseServiceInterface methods
func (m *MockDatabaseService) GetCourseReviews(courseID uint, userID *uint, sortBy, sortOrder string, page, perPage int) ([]*ReviewResponse, int, error) {
	args := m.Called(courseID, userID, sortBy, sortOrder, page, perPage)
	return args.Get(0).([]*ReviewResponse), args.Int(1), args.Error(2)
}

func (m *MockDatabaseService) GetCourseReviewSummary(courseID uint) (*ReviewSummaryResponse, error) {
	args := m.Called(courseID)
	return args.Get(0).(*ReviewSummaryResponse), args.Error(1)
}

func (m *MockDatabaseService) CreateReview(userID uint, req *ReviewCreateRequest) (*ReviewResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*ReviewResponse), args.Error(1)
}

func (m *MockDatabaseService) UpdateReview(reviewID uint, req *ReviewUpdateRequest) (*ReviewResponse, error) {
	args := m.Called(reviewID, req)
	return args.Get(0).(*ReviewResponse), args.Error(1)
}

func (m *MockDatabaseService) DeleteReview(reviewID uint) error {
	args := m.Called(reviewID)
	return args.Error(0)
}

func (m *MockDatabaseService) GetUserReviews(userID uint, page, perPage int) ([]*ReviewResponse, int, error) {
	args := m.Called(userID, page, perPage)
	return args.Get(0).([]*ReviewResponse), args.Int(1), args.Error(2)
}

func (m *MockDatabaseService) IsUserReviewOwner(userID, reviewID uint) (bool, error) {
	args := m.Called(userID, reviewID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabaseService) UserHasReviewForCourse(userID, courseID uint) (bool, error) {
	args := m.Called(userID, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabaseService) SetReviewHelpfulness(userID, reviewID uint, helpful bool) error {
	args := m.Called(userID, reviewID, helpful)
	return args.Error(0)
}

// MapDatabaseServiceInterface methods
func (m *MockDatabaseService) GetMapCourses(userID *uint) ([]*MapCourseResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]*MapCourseResponse), args.Error(1)
}

func (m *MockDatabaseService) GetCoursesInBounds(bounds *BoundsRequest, userID *uint) ([]*MapCourseResponse, error) {
	args := m.Called(bounds, userID)
	return args.Get(0).([]*MapCourseResponse), args.Error(1)
}

func (m *MockDatabaseService) GetClusteredCourses(bounds *BoundsRequest, userID *uint, zoomLevel, maxClusterSize int) ([]*CourseClusterResponse, error) {
	args := m.Called(bounds, userID, zoomLevel, maxClusterSize)
	return args.Get(0).([]*CourseClusterResponse), args.Error(1)
}

func (m *MockDatabaseService) GeocodeAddress(address string) (*GeocodeResponse, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GeocodeResponse), args.Error(1)
}

func (m *MockDatabaseService) ReverseGeocode(lat, lng float64) (*GeocodeResponse, error) {
	args := m.Called(lat, lng)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GeocodeResponse), args.Error(1)
}

func (m *MockDatabaseService) GetCourseLocation(courseID uint) (*MapCourseResponse, error) {
	args := m.Called(courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MapCourseResponse), args.Error(1)
}

func (m *MockDatabaseService) GetRoute(fromLat, fromLng, toLat, toLng float64) (*RouteResponse, error) {
	args := m.Called(fromLat, fromLng, toLat, toLng)
	return args.Get(0).(*RouteResponse), args.Error(1)
}

func (m *MockDatabaseService) GetMapStatistics() (*MapStatisticsResponse, error) {
	args := m.Called()
	return args.Get(0).(*MapStatisticsResponse), args.Error(1)
}

// Integration Test Setup
func setupTestAPI() (*echo.Echo, *MockDatabaseService, *JWTService) {
	e := echo.New()
	
	mockDB := new(MockDatabaseService)
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	// Setup API config
	config := &APIConfig{
		JWTService:    jwtService,
		RateLimit:     100,
		RequestSizeKB: 1024,
	}
	
	// Create API factory and router
	factory := NewAPIFactory(mockDB, config)
	router := factory.CreateAPIRouter()
	router.SetupRoutes(e, config)
	
	return e, mockDB, jwtService
}

func createTestUser() *UserResponse {
	return &UserResponse{
		ID:       123,
		GoogleID: "google123",
		Email:    "test@example.com",
		Name:     "Test User",
		Picture:  "https://example.com/picture.jpg",
	}
}

func generateTestTokens(jwtService *JWTService, user *UserResponse) (*TokenResponse, error) {
	return jwtService.GenerateTokenPair(user.ID, user.GoogleID, user.Email, user.Name)
}

// Integration Tests

func TestAPI_HealthCheck(t *testing.T) {
	e, _, _ := setupTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response.Success)
	
	healthData := response.Data.(map[string]interface{})
	assert.Equal(t, "healthy", healthData["status"])
	assert.Equal(t, "1.0.0", healthData["version"])
}

func TestAPI_AuthFlow(t *testing.T) {
	e, mockDB, jwtService := setupTestAPI()
	user := createTestUser()

	// Mock Google token verification (would need to mock HTTP client in real implementation)
	mockDB.On("GetUserByGoogleID", user.GoogleID).Return(nil, fmt.Errorf("user not found"))
	mockDB.On("GetUserByEmail", user.Email).Return(nil, fmt.Errorf("user not found"))
	mockDB.On("CreateUser", user.GoogleID, user.Email, user.Name, user.Picture).Return(user, nil)

	// Test authentication endpoint (simplified - would need proper Google token)
	tokens, err := generateTestTokens(jwtService, user)
	require.NoError(t, err)

	// Test auth status with valid token
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestAPI_UserProfile(t *testing.T) {
	e, mockDB, jwtService := setupTestAPI()
	user := createTestUser()
	tokens, err := generateTestTokens(jwtService, user)
	require.NoError(t, err)

	// Mock database calls
	mockDB.On("GetUserByID", user.ID).Return(user, nil)

	// Test get profile
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	
	// Verify user data structure
	userData := response.Data.(map[string]interface{})
	assert.Equal(t, float64(user.ID), userData["id"])
	assert.Equal(t, user.Email, userData["email"])
}

func TestAPI_CourseManagement(t *testing.T) {
	e, mockDB, jwtService := setupTestAPI()
	user := createTestUser()
	tokens, err := generateTestTokens(jwtService, user)
	require.NoError(t, err)

	// Test create course
	courseReq := CourseCreateRequest{
		Name:    "Test Golf Course",
		Address: "123 Golf Course Rd, Test City, TX 12345",
	}

	course := &CourseResponse{
		ID:        1,
		Name:      courseReq.Name,
		Address:   courseReq.Address,
		CreatedBy: &user.ID,
		CanEdit:   true,
	}

	mockDB.On("CourseExistsByNameAndAddress", courseReq.Name, courseReq.Address).Return(false, nil)
	mockDB.On("CreateCourse", user.ID, &courseReq).Return(course, nil)

	reqBody, _ := json.Marshal(courseReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	
	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	
	courseData := response.Data.(map[string]interface{})
	assert.Equal(t, courseReq.Name, courseData["name"])
	assert.Equal(t, courseReq.Address, courseData["address"])
}

func TestAPI_ReviewManagement(t *testing.T) {
	e, mockDB, jwtService := setupTestAPI()
	user := createTestUser()
	tokens, err := generateTestTokens(jwtService, user)
	require.NoError(t, err)

	courseID := uint(1)
	reviewReq := ReviewCreateRequest{
		CourseID:      courseID,
		OverallRating: 8,
		ReviewText:    stringPtr("Great course!"),
	}

	review := &ReviewResponse{
		ID:            1,
		CourseID:      courseID,
		CourseName:    "Test Course",
		UserID:        user.ID,
		UserName:      user.Name,
		OverallRating: reviewReq.OverallRating,
		ReviewText:    reviewReq.ReviewText,
		CanEdit:       true,
	}

	mockDB.On("CourseExists", courseID).Return(true, nil)
	mockDB.On("UserHasReviewForCourse", user.ID, courseID).Return(false, nil)
	mockDB.On("CreateReview", user.ID, &reviewReq).Return(review, nil)

	reqBody, _ := json.Marshal(reviewReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	
	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestAPI_MapEndpoints(t *testing.T) {
	e, mockDB, _ := setupTestAPI()

	mapCourses := []*MapCourseResponse{
		{
			ID:        1,
			Name:      "Test Course",
			Address:   "123 Golf St",
			Latitude:  floatPtr(40.7128),
			Longitude: floatPtr(-74.0060),
		},
	}

	mockDB.On("GetMapCourses", (*uint)(nil)).Return(mapCourses, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/map/courses", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

func TestAPI_ErrorHandling(t *testing.T) {
	e, _, _ := setupTestAPI()

	// Test unauthorized access
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	
	// Just check that the response contains the expected error message
	assert.Contains(t, rec.Body.String(), "authentication_required")

	// Test invalid token - create fresh server to avoid rate limiting
	e2, _, _ := setupTestAPI()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	req2.Header.Set("Authorization", "Bearer invalid-token")
	rec2 := httptest.NewRecorder()
	e2.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
}

func TestAPI_ValidationErrors(t *testing.T) {
	e, _, jwtService := setupTestAPI()
	user := createTestUser()
	tokens, err := generateTestTokens(jwtService, user)
	require.NoError(t, err)

	// Test invalid course creation
	invalidCourseReq := CourseCreateRequest{
		Name:    "ab", // Too short
		Address: "123", // Too short
	}

	reqBody, _ := json.Marshal(invalidCourseReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	
	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "validation_error", response.Error.Error)
	assert.NotNil(t, response.Error.Details)
}

func TestAPI_PaginationHeaders(t *testing.T) {
	e, mockDB, _ := setupTestAPI()

	courses := []*CourseResponse{}
	total := 50

	mockDB.On("GetCourses", mock.AnythingOfType("*api.CourseSearchRequest"), (*uint)(nil), 2, 10).Return(courses, total, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?page=2&per_page=10", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Meta)
	assert.Equal(t, 2, response.Meta.Page)
	assert.Equal(t, 10, response.Meta.PerPage)
	assert.Equal(t, 50, response.Meta.Total)
	assert.Equal(t, 5, response.Meta.TotalPages) // ceil(50/10)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

// Benchmark test for API performance
func BenchmarkAPI_AuthenticatedRequest(b *testing.B) {
	e, mockDB, jwtService := setupTestAPI()
	user := createTestUser()
	tokens, err := generateTestTokens(jwtService, user)
	require.NoError(b, err)

	mockDB.On("GetUserByID", user.ID).Return(user, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", rec.Code)
		}
	}
}