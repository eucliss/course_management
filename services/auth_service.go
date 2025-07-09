package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type authService struct {
	userRepo UserRepository
	config   AuthConfig
}

func NewAuthService(userRepo UserRepository, config AuthConfig) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *authService) VerifyGoogleToken(ctx context.Context, token string) (*GoogleUser, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Verify token with Google
	googleUser, err := s.verifyWithGoogle(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	return googleUser, nil
}

func (s *authService) GetAuthConfig() AuthConfig {
	return s.config
}

func (s *authService) CreateOrUpdateUser(ctx context.Context, googleUser GoogleUser) (*GoogleUser, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByGoogleID(ctx, googleUser.ID)
	if err == nil && existingUser != nil {
		// User exists, update their information
		existingUser.Name = googleUser.Name
		existingUser.Email = googleUser.Email
		existingUser.Picture = googleUser.Picture
		
		// Keep existing display name and handicap if present
		if googleUser.DisplayName != nil {
			existingUser.DisplayName = googleUser.DisplayName
		}
		if googleUser.Handicap != nil {
			existingUser.Handicap = googleUser.Handicap
		}

		if err := s.userRepo.Update(ctx, *existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return existingUser, nil
	}

	// User doesn't exist, create new user
	createdUser, err := s.userRepo.Create(ctx, googleUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *authService) GetUserByGoogleID(ctx context.Context, googleID string) (*GoogleUser, error) {
	user, err := s.userRepo.GetByGoogleID(ctx, googleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *authService) verifyWithGoogle(ctx context.Context, token string) (*GoogleUser, error) {
	// Create request to Google's tokeninfo endpoint
	tokenInfoURL := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", url.QueryEscape(token))
	
	req, err := http.NewRequestWithContext(ctx, "GET", tokenInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set timeout for the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token with Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google token verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Google's response
	var tokenInfo struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Locale        string `json:"locale"`
		Iss           string `json:"iss"`
		Exp           string `json:"exp"`
		Iat           string `json:"iat"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode Google response: %w", err)
	}

	// Verify the token is for our application
	if tokenInfo.Aud != s.config.GoogleClientID {
		return nil, fmt.Errorf("token audience mismatch")
	}

	// Verify the token issuer
	if tokenInfo.Iss != "https://accounts.google.com" && tokenInfo.Iss != "accounts.google.com" {
		return nil, fmt.Errorf("token issuer mismatch")
	}

	// Verify email is verified
	if tokenInfo.EmailVerified != "true" {
		return nil, fmt.Errorf("email not verified")
	}

	// Create GoogleUser from token info
	googleUser := &GoogleUser{
		ID:      tokenInfo.Sub,
		Email:   tokenInfo.Email,
		Name:    tokenInfo.Name,
		Picture: tokenInfo.Picture,
	}

	return googleUser, nil
}

// SessionService implementation
type sessionService struct {
	userRepo UserRepository
}

func NewSessionService(userRepo UserRepository) SessionService {
	return &sessionService{
		userRepo: userRepo,
	}
}

func (s *sessionService) GetUser(c echo.Context) *GoogleUser {
	user, ok := c.Get("user").(*GoogleUser)
	if !ok {
		return nil
	}
	return user
}

func (s *sessionService) GetUserID(c echo.Context) *uint {
	user := s.GetUser(c)
	if user == nil {
		return nil
	}

	// Convert Google ID to database ID
	ctx := c.Request().Context()
	_, err := s.userRepo.GetByGoogleID(ctx, user.ID)
	if err != nil {
		return nil
	}

	// We need to get the database ID from the user record
	// This is a limitation of the current interface design
	// In a real implementation, you'd want to store the database ID in the session
	return nil // Placeholder - would need to be fixed with proper session management
}

func (s *sessionService) IsAuthenticated(c echo.Context) bool {
	return s.GetUser(c) != nil
}

func (s *sessionService) SetUser(c echo.Context, user GoogleUser) error {
	c.Set("user", &user)
	return nil
}

func (s *sessionService) ClearSession(c echo.Context) error {
	c.Set("user", nil)
	return nil
}

func (s *sessionService) GetDatabaseUserID(c echo.Context) *uint {
	user := s.GetUser(c)
	if user == nil {
		return nil
	}

	// Convert Google ID to database ID
	ctx := c.Request().Context()
	_, err := s.userRepo.GetByGoogleID(ctx, user.ID)
	if err != nil {
		return nil
	}

	// This is a placeholder - the actual implementation would need to be
	// fixed to properly return the database ID
	return nil
}

func (s *sessionService) RequireAuth(c echo.Context) (*GoogleUser, error) {
	user := s.GetUser(c)
	if user == nil {
		return nil, fmt.Errorf("authentication required")
	}

	return user, nil
}

// Utility functions for authentication
func GenerateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func ValidateState(provided, expected string) bool {
	return provided == expected && provided != ""
}