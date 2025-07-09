package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware_ValidToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	// Generate a valid token
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"
	
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a simple handler that returns user info
	handler := func(c echo.Context) error {
		claims, err := GetUserClaims(c)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": claims.UserID,
			"email":   claims.Email,
		})
	}

	// Apply middleware
	middlewareHandler := JWTMiddleware(jwtService)(handler)
	err = middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Check that user claims were set in context
	assert.Equal(t, userID, c.Get("user_id"))
	assert.Equal(t, true, c.Get("authenticated"))
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middlewareHandler := JWTMiddleware(jwtService)(handler)
	err := middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "authentication_required")
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middlewareHandler := JWTMiddleware(jwtService)(handler)
	err := middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_token")
}

func TestJWTMiddleware_MalformedAuthHeader(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")

	tests := []string{
		"invalid-format",
		"Basic token123",
		"Bearer",
		"",
	}

	for _, authHeader := range tests {
		t.Run(fmt.Sprintf("auth_header_%s", authHeader), func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			}

			middlewareHandler := JWTMiddleware(jwtService)(handler)
			err := middlewareHandler(c)

			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func TestOptionalJWTMiddleware_ValidToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	// Generate a valid token
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"
	
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		isAuth := c.Get("authenticated").(bool)
		if isAuth {
			claims, _ := GetUserClaims(c)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"authenticated": true,
				"user_id":      claims.UserID,
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
	}

	middlewareHandler := OptionalJWTMiddleware(jwtService)(handler)
	err = middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"authenticated\":true")
}

func TestOptionalJWTMiddleware_NoToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		isAuth := c.Get("authenticated").(bool)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"authenticated": isAuth,
		})
	}

	middlewareHandler := OptionalJWTMiddleware(jwtService)(handler)
	err := middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"authenticated\":false")
}

func TestOptionalJWTMiddleware_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		isAuth := c.Get("authenticated").(bool)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"authenticated": isAuth,
		})
	}

	middlewareHandler := OptionalJWTMiddleware(jwtService)(handler)
	err := middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"authenticated\":false")
}

func TestAPIResponseMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middlewareHandler := APIResponseMiddleware()(handler)
	err := middlewareHandler(c)

	require.NoError(t, err)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "1.0", rec.Header().Get("X-API-Version"))
	assert.NotEmpty(t, rec.Header().Get("X-Response-Time"))
}

func TestRequestValidationMiddleware_ValidContentType(t *testing.T) {
	e := echo.New()
	
	tests := []struct {
		method      string
		contentType string
		shouldPass  bool
	}{
		{http.MethodGet, "", true},       // GET doesn't need content type
		{http.MethodPost, "application/json", true},
		{http.MethodPost, "application/json; charset=utf-8", true},
		{http.MethodPut, "application/json", true},
		{http.MethodPatch, "application/json", true},
		{http.MethodPost, "text/plain", false},
		{http.MethodPost, "application/xml", false},
		{http.MethodPut, "text/html", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.method, tt.contentType), func(t *testing.T) {
			body := ""
			if tt.method != http.MethodGet {
				body = `{"test": "data"}`
			}
			
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			}

			middlewareHandler := RequestValidationMiddleware()(handler)
			err := middlewareHandler(c)

			require.NoError(t, err)
			if tt.shouldPass {
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
				assert.Contains(t, rec.Body.String(), "invalid_content_type")
			}
		})
	}
}

func TestGetUserClaims_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set mock claims
	claims := &JWTClaims{
		UserID:   123,
		GoogleID: "google123",
		Email:    "test@example.com",
		Name:     "Test User",
	}
	c.Set("user_claims", claims)

	result, err := GetUserClaims(c)
	require.NoError(t, err)
	assert.Equal(t, claims, result)
}

func TestGetUserClaims_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// No claims set
	_, err := GetUserClaims(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user claims not found")
}

func TestGetUserID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &JWTClaims{
		UserID: 123,
	}
	c.Set("user_claims", claims)

	userID, err := GetUserID(c)
	require.NoError(t, err)
	assert.Equal(t, uint(123), userID)
}

func TestGetUserID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := GetUserID(c)
	assert.Error(t, err)
}

func TestAPISecurityMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middlewareHandler := APISecurityMiddleware()(handler)
	err := middlewareHandler(c)

	require.NoError(t, err)
	
	// Check security headers
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	// HSTS header is only set for HTTPS requests in production
	hstsHeader := rec.Header().Get("Strict-Transport-Security")
	if hstsHeader != "" {
		assert.Contains(t, hstsHeader, "max-age=31536000")
	}
}

// Integration test for multiple middleware layers
func TestMiddlewareStack(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	// Generate a valid token
	tokens, err := jwtService.GenerateTokenPair(123, "google123", "test@example.com", "Test User")
	require.NoError(t, err)

	e := echo.New()
	
	// Setup middleware stack
	g := e.Group("/api")
	config := &APIConfig{
		JWTService:    jwtService,
		RateLimit:     60,
		RequestSizeKB: 1024,
	}
	SetupAPIMiddleware(g, config)

	// Add a protected endpoint
	g.GET("/protected", func(c echo.Context) error {
		claims, err := GetUserClaims(c)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": claims.UserID,
			"message": "success",
		})
	}, JWTMiddleware(jwtService))

	// Test with valid token
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"user_id\":123")
	
	// Check that all middleware set their headers
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "1.0", rec.Header().Get("X-API-Version"))
	assert.NotEmpty(t, rec.Header().Get("X-Response-Time"))
}