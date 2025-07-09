package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPagination(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name     string
		query    string
		expected PaginationParams
	}{
		{
			name:     "default values",
			query:    "",
			expected: PaginationParams{Page: 1, PerPage: 20},
		},
		{
			name:     "custom values",
			query:    "page=3&per_page=50",
			expected: PaginationParams{Page: 3, PerPage: 50},
		},
		{
			name:     "invalid page defaults to 1",
			query:    "page=0&per_page=25",
			expected: PaginationParams{Page: 1, PerPage: 25},
		},
		{
			name:     "invalid per_page defaults to 20",
			query:    "page=2&per_page=0",
			expected: PaginationParams{Page: 2, PerPage: 20},
		},
		{
			name:     "per_page over limit capped at 100",
			query:    "page=1&per_page=150",
			expected: PaginationParams{Page: 1, PerPage: 100},
		},
		{
			name:     "negative values corrected",
			query:    "page=-1&per_page=-5",
			expected: PaginationParams{Page: 1, PerPage: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			result := GetPagination(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]string{"message": "success"}
	err := SuccessResponse(c, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	// Check data content as JSON doesn't preserve exact types
	dataMap := response.Data.(map[string]interface{})
	assert.Equal(t, "success", dataMap["message"])
	assert.Nil(t, response.Error)
	assert.Greater(t, response.Timestamp, int64(0))
}

func TestSuccessResponseWithMeta(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := []string{"item1", "item2"}
	meta := &APIMeta{
		Page:       1,
		PerPage:    20,
		Total:      2,
		TotalPages: 1,
	}

	err := SuccessResponseWithMeta(c, data, meta)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	// Check data content as JSON doesn't preserve exact types
	dataSlice := response.Data.([]interface{})
	assert.Len(t, dataSlice, 2)
	assert.Equal(t, "item1", dataSlice[0])
	assert.Equal(t, "item2", dataSlice[1])
	assert.Equal(t, meta, response.Meta)
	assert.Nil(t, response.Error)
}

func TestCreatedResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]interface{}{"id": 123, "name": "test"}
	err := CreatedResponse(c, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	// Check data content as JSON doesn't preserve exact types
	dataMap := response.Data.(map[string]interface{})
	assert.Equal(t, float64(123), dataMap["id"])  // JSON numbers become float64
	assert.Equal(t, "test", dataMap["name"])
}

func TestNoContentResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := NoContentResponse(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestErrorResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ErrorResponse(c, http.StatusBadRequest, "bad_request", "Invalid input", "REQ_001")

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	require.NotNil(t, response.Error)
	assert.Equal(t, "bad_request", response.Error.Error)
	assert.Equal(t, "Invalid input", response.Error.Message)
	assert.Equal(t, "REQ_001", response.Error.Code)
}

func TestErrorResponseWithDetails(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	details := map[string]string{
		"name":  "Name is required",
		"email": "Invalid email format",
	}

	err := ErrorResponseWithDetails(c, http.StatusBadRequest, "validation_error", "Validation failed", "VAL_001", details)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "validation_error", response.Error.Error)
	assert.Equal(t, "Validation failed", response.Error.Message)
	assert.Equal(t, "VAL_001", response.Error.Code)
	assert.Equal(t, details, response.Error.Details)
}

func TestBadRequestError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := BadRequestError(c, "Invalid JSON format")

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "bad_request", response.Error.Error)
	assert.Equal(t, "Invalid JSON format", response.Error.Message)
	assert.Equal(t, "REQ_001", response.Error.Code)
}

func TestValidationError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	details := map[string]string{
		"email": "Email is required",
		"age":   "Age must be a positive number",
	}

	err := ValidationError(c, details)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "validation_error", response.Error.Error)
	assert.Equal(t, "Request validation failed", response.Error.Message)
	assert.Equal(t, "VAL_001", response.Error.Code)
	assert.Equal(t, details, response.Error.Details)
}

func TestUnauthorizedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := UnauthorizedError(c, "Authentication required")

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "unauthorized", response.Error.Error)
	assert.Equal(t, "Authentication required", response.Error.Message)
	assert.Equal(t, "AUTH_001", response.Error.Code)
}

func TestForbiddenError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ForbiddenError(c, "Insufficient permissions")

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "forbidden", response.Error.Error)
	assert.Equal(t, "Insufficient permissions", response.Error.Message)
	assert.Equal(t, "AUTH_002", response.Error.Code)
}

func TestNotFoundError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := NotFoundError(c, "User")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "not_found", response.Error.Error)
	assert.Equal(t, "User not found", response.Error.Message)
	assert.Equal(t, "RES_001", response.Error.Code)
}

func TestConflictError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ConflictError(c, "Email already exists")

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "conflict", response.Error.Error)
	assert.Equal(t, "Email already exists", response.Error.Message)
	assert.Equal(t, "RES_002", response.Error.Code)
}

func TestInternalServerError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := InternalServerError(c, "Database connection failed")

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "internal_server_error", response.Error.Error)
	assert.Equal(t, "Database connection failed", response.Error.Message)
	assert.Equal(t, "SYS_001", response.Error.Code)
}

func TestServiceUnavailableError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ServiceUnavailableError(c, "Service temporarily unavailable")

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	require.NotNil(t, response.Error)
	assert.Equal(t, "service_unavailable", response.Error.Error)
	assert.Equal(t, "Service temporarily unavailable", response.Error.Message)
	assert.Equal(t, "SYS_002", response.Error.Code)
}

func TestNewListResponse(t *testing.T) {
	items := []interface{}{"item1", "item2", "item3"}
	page := 2
	perPage := 10
	total := 25

	response := NewListResponse(items, page, perPage, total)

	assert.Equal(t, items, response.Items)
	require.NotNil(t, response.Meta)
	assert.Equal(t, page, response.Meta.Page)
	assert.Equal(t, perPage, response.Meta.PerPage)
	assert.Equal(t, total, response.Meta.Total)
	assert.Equal(t, 3, response.Meta.TotalPages) // ceil(25/10) = 3
}

func TestNewListResponse_ExactDivision(t *testing.T) {
	items := []interface{}{"item1", "item2"}
	page := 1
	perPage := 5
	total := 10

	response := NewListResponse(items, page, perPage, total)

	assert.Equal(t, 2, response.Meta.TotalPages) // 10/5 = 2
}

func TestNewListResponse_EmptyList(t *testing.T) {
	items := []interface{}{}
	page := 1
	perPage := 10
	total := 0

	response := NewListResponse(items, page, perPage, total)

	assert.Empty(t, response.Items)
	assert.Equal(t, 0, response.Meta.TotalPages)
}