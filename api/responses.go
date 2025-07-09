package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Standard API response structures

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// APIError represents an API error response
type APIError struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// APIMeta contains metadata for API responses (pagination, etc.)
type APIMeta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Pagination parameters for list endpoints
type PaginationParams struct {
	Page    int `query:"page" validate:"min=1"`
	PerPage int `query:"per_page" validate:"min=1,max=100"`
}

// GetPagination returns pagination parameters with defaults
func GetPagination(c echo.Context) PaginationParams {
	params := PaginationParams{
		Page:    1,
		PerPage: 20,
	}

	// Bind query parameters
	if err := c.Bind(&params); err != nil {
		// Use defaults on error
		return PaginationParams{Page: 1, PerPage: 20}
	}

	// Apply bounds
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	return params
}

// Success response helpers

// SuccessResponse returns a successful API response
func SuccessResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessResponseWithMeta returns a successful API response with metadata
func SuccessResponseWithMeta(c echo.Context, data interface{}, meta *APIMeta) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().Unix(),
	})
}

// CreatedResponse returns a 201 Created response
func CreatedResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// NoContentResponse returns a 204 No Content response
func NoContentResponse(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Error response helpers

// ErrorResponse returns an error API response
func ErrorResponse(c echo.Context, statusCode int, error, message, code string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Error:   error,
			Message: message,
			Code:    code,
		},
		Timestamp: time.Now().Unix(),
	})
}

// ErrorResponseWithDetails returns an error API response with additional details
func ErrorResponseWithDetails(c echo.Context, statusCode int, error, message, code string, details map[string]string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Error:   error,
			Message: message,
			Code:    code,
			Details: details,
		},
		Timestamp: time.Now().Unix(),
	})
}

// Common error responses

// BadRequestError returns a 400 Bad Request error
func BadRequestError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusBadRequest, "bad_request", message, "REQ_001")
}

// ValidationError returns a 400 Bad Request error for validation failures
func ValidationError(c echo.Context, details map[string]string) error {
	return ErrorResponseWithDetails(c, http.StatusBadRequest, 
		"validation_error", "Request validation failed", "VAL_001", details)
}

// UnauthorizedError returns a 401 Unauthorized error
func UnauthorizedError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, "unauthorized", message, "AUTH_001")
}

// ForbiddenError returns a 403 Forbidden error
func ForbiddenError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusForbidden, "forbidden", message, "AUTH_002")
}

// NotFoundError returns a 404 Not Found error
func NotFoundError(c echo.Context, resource string) error {
	return ErrorResponse(c, http.StatusNotFound, "not_found", 
		resource+" not found", "RES_001")
}

// ConflictError returns a 409 Conflict error
func ConflictError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusConflict, "conflict", message, "RES_002")
}

// InternalServerError returns a 500 Internal Server Error
func InternalServerError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusInternalServerError, 
		"internal_server_error", message, "SYS_001")
}

// ServiceUnavailableError returns a 503 Service Unavailable error
func ServiceUnavailableError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusServiceUnavailable, 
		"service_unavailable", message, "SYS_002")
}

// Database and business logic error mapping

// MapDatabaseError maps database errors to appropriate API responses
func MapDatabaseError(c echo.Context, err error) error {
	switch {
	case err.Error() == "record not found":
		return NotFoundError(c, "Resource")
	case err.Error() == "duplicate key value violates unique constraint":
		return ConflictError(c, "Resource already exists")
	default:
		return InternalServerError(c, "Database operation failed")
	}
}

// Health check response structure
type HealthResponse struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Services    map[string]string `json:"services"`
	Timestamp   int64             `json:"timestamp"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items []interface{} `json:"items"`
	Meta  *APIMeta      `json:"meta"`
}

// NewListResponse creates a paginated list response
func NewListResponse(items []interface{}, page, perPage, total int) *ListResponse {
	totalPages := (total + perPage - 1) / perPage

	return &ListResponse{
		Items: items,
		Meta: &APIMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}