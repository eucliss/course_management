package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthHandler handles authentication-related API endpoints
type AuthHandler struct {
	jwtService       *JWTService
	dbService        DatabaseServiceInterface
	config           *oauth2.Config
	allowedClientIDs []string // Support multiple client IDs
}

// DatabaseServiceInterface defines the interface for database operations
type DatabaseServiceInterface interface {
	CreateUser(googleID, email, name, picture string) (*UserResponse, error)
	GetUserByGoogleID(googleID string) (*UserResponse, error)
	GetUserByEmail(email string) (*UserResponse, error)
}

// UserResponse represents user data for API responses
type UserResponse struct {
	ID          uint     `json:"id"`
	GoogleID    string   `json:"google_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	DisplayName *string  `json:"display_name"`
	Picture     string   `json:"picture"`
	Handicap    *float64 `json:"handicap"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

// GoogleTokenRequest represents Google OAuth token verification request
type GoogleTokenRequest struct {
	IDToken     string `json:"id_token" validate:"required"`
	AccessToken string `json:"access_token,omitempty"`
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthStatusResponse represents authentication status
type AuthStatusResponse struct {
	Authenticated bool          `json:"authenticated"`
	User          *UserResponse `json:"user,omitempty"`
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(jwtService *JWTService, dbService DatabaseServiceInterface, googleClientID, googleClientSecret, iosClientID, redirectURL string) *AuthHandler {
	config := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	// Support both web and iOS client IDs
	allowedClientIDs := []string{googleClientID}
	if iosClientID != "" && iosClientID != googleClientID {
		allowedClientIDs = append(allowedClientIDs, iosClientID)
	}

	return &AuthHandler{
		jwtService:       jwtService,
		dbService:        dbService,
		config:           config,
		allowedClientIDs: allowedClientIDs,
	}
}

// VerifyGoogleToken verifies Google OAuth token and returns JWT tokens
func (h *AuthHandler) VerifyGoogleToken(c echo.Context) error {
	var req GoogleTokenRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	// Validate required fields
	if req.IDToken == "" {
		return ValidationError(c, map[string]string{
			"id_token": "ID token is required",
		})
	}

	// Verify Google ID token and get user info
	userInfo, err := h.verifyGoogleIDToken(req.IDToken)
	if err != nil {
		return UnauthorizedError(c, "Invalid Google token")
	}

	// Create or get user from database
	user, err := h.getOrCreateUser(userInfo)
	if err != nil {
		return InternalServerError(c, "Failed to process user authentication")
	}

	// Generate JWT token pair
	tokens, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.GoogleID,
		user.Email,
		user.Name,
	)
	if err != nil {
		return InternalServerError(c, "Failed to generate authentication tokens")
	}

	// Return tokens and user info
	return SuccessResponse(c, map[string]interface{}{
		"tokens": tokens,
		"user":   user,
	})
}

// RefreshToken generates new access token using refresh token
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	if req.RefreshToken == "" {
		return ValidationError(c, map[string]string{
			"refresh_token": "Refresh token is required",
		})
	}

	// Generate new token pair using refresh token
	tokens, err := h.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return UnauthorizedError(c, "Invalid or expired refresh token")
	}

	return SuccessResponse(c, map[string]interface{}{
		"tokens": tokens,
	})
}

// GetAuthStatus returns current authentication status
func (h *AuthHandler) GetAuthStatus(c echo.Context) error {
	// Check if user is authenticated (set by optional JWT middleware)
	isAuth := c.Get("authenticated")
	if isAuth == nil || !isAuth.(bool) {
		return SuccessResponse(c, AuthStatusResponse{
			Authenticated: false,
		})
	}

	// Get user claims
	claims, err := GetUserClaims(c)
	if err != nil {
		return SuccessResponse(c, AuthStatusResponse{
			Authenticated: false,
		})
	}

	// Get current user data from database
	user, err := h.dbService.GetUserByGoogleID(claims.GoogleID)
	if err != nil {
		return SuccessResponse(c, AuthStatusResponse{
			Authenticated: false,
		})
	}

	return SuccessResponse(c, AuthStatusResponse{
		Authenticated: true,
		User:          user,
	})
}

// GoogleCallback handles OAuth2 callback from Google
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	// Get authorization code from query params
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	
	if code == "" {
		return BadRequestError(c, "Authorization code is required")
	}
	
	// Exchange authorization code for tokens
	token, err := h.config.Exchange(c.Request().Context(), code)
	if err != nil {
		return UnauthorizedError(c, "Failed to exchange authorization code")
	}
	
	// Extract ID token from the OAuth2 token
	idToken, ok := token.Extra("id_token").(string)
	if !ok || idToken == "" {
		return UnauthorizedError(c, "No ID token received from Google")
	}
	
	// Verify the ID token and get user info
	userInfo, err := h.verifyGoogleIDToken(idToken)
	if err != nil {
		return UnauthorizedError(c, "Invalid ID token")
	}
	
	// Create or get user from database
	user, err := h.getOrCreateUser(userInfo)
	if err != nil {
		return InternalServerError(c, "Failed to process user authentication")
	}
	
	// Generate JWT token pair
	tokens, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.GoogleID,
		user.Email,
		user.Name,
	)
	if err != nil {
		return InternalServerError(c, "Failed to generate authentication tokens")
	}
	
	// Handle different states/redirect scenarios
	if state == "ios_app" {
		// For iOS app, return a redirect with tokens as URL parameters
		// In production, you'd want to use a custom URL scheme
		redirectURL := fmt.Sprintf("golfapp://auth/success?access_token=%s&refresh_token=%s", 
			tokens.AccessToken, tokens.RefreshToken)
		return c.Redirect(302, redirectURL)
	}
	
	// Default: return JSON response
	return SuccessResponse(c, map[string]interface{}{
		"tokens": tokens,
		"user":   user,
	})
}

// Logout invalidates the current token (placeholder for token blacklisting)
func (h *AuthHandler) Logout(c echo.Context) error {
	// In a production system, you would add the token to a blacklist
	// For now, we'll just return success and let the client discard the token
	
	// Log the logout event
	userID, _ := GetUserID(c)
	if userID > 0 {
		// Could log to audit trail here
		fmt.Printf("User %d logged out\n", userID)
	}

	return SuccessResponse(c, map[string]string{
		"message": "Successfully logged out",
	})
}

// Helper methods

// verifyGoogleIDToken verifies Google ID token and returns user info
func (h *AuthHandler) verifyGoogleIDToken(idToken string) (*GoogleUserInfo, error) {
	// Create HTTP client to verify token with Google
	client := &http.Client{}
	
	// Google's token verification endpoint
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token with Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token verification failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verification response: %w", err)
	}

	var tokenInfo struct {
		Aud     string `json:"aud"`
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse token info: %w", err)
	}

	// Verify the token is for one of our allowed client IDs
	validClientID := false
	for _, allowedID := range h.allowedClientIDs {
		if tokenInfo.Aud == allowedID {
			validClientID = true
			break
		}
	}
	if !validClientID {
		return nil, fmt.Errorf("token audience mismatch: received %s, allowed %v", tokenInfo.Aud, h.allowedClientIDs)
	}

	return &GoogleUserInfo{
		ID:      tokenInfo.Sub,
		Email:   tokenInfo.Email,
		Name:    tokenInfo.Name,
		Picture: tokenInfo.Picture,
	}, nil
}

// getOrCreateUser gets existing user or creates new one
func (h *AuthHandler) getOrCreateUser(userInfo *GoogleUserInfo) (*UserResponse, error) {
	// Try to get existing user by Google ID
	user, err := h.dbService.GetUserByGoogleID(userInfo.ID)
	if err == nil {
		return user, nil
	}

	// If user doesn't exist, try by email (in case they had an account before Google auth)
	user, err = h.dbService.GetUserByEmail(userInfo.Email)
	if err == nil {
		// Update Google ID for existing user
		// This would require an UpdateUser method in the database service
		return user, nil
	}

	// Create new user
	return h.dbService.CreateUser(
		userInfo.ID,
		userInfo.Email,
		userInfo.Name,
		userInfo.Picture,
	)
}

// validateEmail performs basic email validation
func validateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(g *echo.Group, jwtService *JWTService) {
	// Public routes (no authentication required)
	g.POST("/auth/google/verify", h.VerifyGoogleToken)
	g.GET("/auth/google/callback", h.GoogleCallback)
	g.POST("/auth/refresh", h.RefreshToken)
	
	// Routes that use optional authentication (can work with or without token)
	g.GET("/auth/status", h.GetAuthStatus, OptionalJWTMiddleware(jwtService))
	
	// Protected routes (authentication required)
	g.POST("/auth/logout", h.Logout, JWTMiddleware(jwtService))
}