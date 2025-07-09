package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// APIConfig holds configuration for API middleware
type APIConfig struct {
	JWTService    *JWTService
	RateLimit     int
	RequestSizeKB int64
}

// JWTMiddleware validates JWT tokens for API routes
func JWTMiddleware(jwtService *JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from header
			tokenString, err := jwtService.ExtractTokenFromHeader(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, APIError{
					Error:   "authentication_required",
					Message: "Valid authentication token required",
					Code:    "AUTH_001",
				})
			}

			// Validate access token
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, APIError{
					Error:   "invalid_token",
					Message: "Token is invalid or expired",
					Code:    "AUTH_002",
				})
			}

			// Store claims in context for handlers
			c.Set("user_claims", claims)
			c.Set("user_id", claims.UserID)
			c.Set("authenticated", true)

			return next(c)
		}
	}
}

// OptionalJWTMiddleware validates JWT tokens but allows unauthenticated requests
func OptionalJWTMiddleware(jwtService *JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to extract and validate token
			tokenString, err := jwtService.ExtractTokenFromHeader(c)
			if err == nil {
				claims, err := jwtService.ValidateAccessToken(tokenString)
				if err == nil {
					// Valid token found
					c.Set("user_claims", claims)
					c.Set("user_id", claims.UserID)
					c.Set("authenticated", true)
				}
			}

			// Continue regardless of authentication status
			if c.Get("authenticated") == nil {
				c.Set("authenticated", false)
			}

			return next(c)
		}
	}
}

// APIResponseMiddleware sets standard API response headers
func APIResponseMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set standard API headers
			c.Response().Header().Set("Content-Type", "application/json")
			c.Response().Header().Set("X-API-Version", "1.0")
			c.Response().Header().Set("X-Response-Time", time.Now().Format(time.RFC3339))

			return next(c)
		}
	}
}

// APISecurityMiddleware adds security headers for API endpoints
func APISecurityMiddleware() echo.MiddlewareFunc {
	return middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	})
}

// APICORSMiddleware configures CORS for mobile apps
func APICORSMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	if len(allowedOrigins) == 0 {
		// Default origins for development
		allowedOrigins = []string{"*"}
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Requested-With",
			"X-API-Key",
		},
		ExposeHeaders: []string{
			"X-API-Version",
			"X-Response-Time",
			"X-Total-Count",
			"X-Page-Count",
		},
		AllowCredentials: false, // Don't allow credentials for mobile API
		MaxAge:           86400, // 24 hours
	})
}

// APIRateLimitMiddleware applies rate limiting for API endpoints
func APIRateLimitMiddleware(requestsPerMinute int) echo.MiddlewareFunc {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 120 // Default: 120 requests per minute for mobile
	}

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(rate.Limit(float64(requestsPerMinute) / 60)), // Convert to per-second rate
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Use combination of IP and user ID for rate limiting
			identifier := c.RealIP()
			
			// If authenticated, include user ID for more precise limiting
			if userID := c.Get("user_id"); userID != nil {
				identifier += "_" + string(rune(userID.(uint)))
			}
			
			return identifier, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, APIError{
				Error:   "rate_limit_exceeded",
				Message: "Too many requests. Please slow down.",
				Code:    "RATE_001",
			})
		},
	})
}

// RequestValidationMiddleware validates request content type for POST/PUT requests
func RequestValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			
			// Validate content type for body-containing requests
			if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
				contentType := c.Request().Header.Get("Content-Type")
				if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
					return c.JSON(http.StatusUnsupportedMediaType, APIError{
						Error:   "invalid_content_type",
						Message: "Content-Type must be application/json",
						Code:    "REQ_001",
					})
				}
			}

			return next(c)
		}
	}
}

// APILoggingMiddleware logs API requests with structured format
func APILoggingMiddleware() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","method":"${method}","uri":"${uri}","status":${status},"latency":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out},"user_agent":"${user_agent}","remote_ip":"${remote_ip}","error":"${error}"}` + "\n",
		Output: middleware.DefaultLoggerConfig.Output,
	})
}

// SetupAPIMiddleware configures all middleware for API routes
func SetupAPIMiddleware(e *echo.Group, config *APIConfig) {
	// Basic middleware stack
	e.Use(APIResponseMiddleware())
	e.Use(APISecurityMiddleware())
	e.Use(APICORSMiddleware(nil)) // Allow all origins for development
	e.Use(APILoggingMiddleware())
	e.Use(RequestValidationMiddleware())
	
	// Rate limiting
	if config.RateLimit > 0 {
		e.Use(APIRateLimitMiddleware(config.RateLimit))
	}
	
	// Request size limiting
	if config.RequestSizeKB > 0 {
		e.Use(middleware.BodyLimit(fmt.Sprintf("%dKB", config.RequestSizeKB)))
	}
}