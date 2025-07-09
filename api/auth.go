package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWT Claims structure
type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	GoogleID  string `json:"google_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Token response structure
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// JWT service for mobile authentication
type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewJWTService creates a new JWT service with secure defaults
func NewJWTService(accessSecret, refreshSecret string) *JWTService {
	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     time.Hour * 1,      // 1 hour for access tokens
		refreshTTL:    time.Hour * 24 * 7, // 7 days for refresh tokens
	}
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (j *JWTService) GenerateTokenPair(userID uint, googleID, email, name string) (*TokenResponse, error) {
	// Generate access token
	accessToken, err := j.generateToken(userID, googleID, email, name, "access", j.accessTTL, j.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(userID, googleID, email, name, "refresh", j.refreshTTL, j.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessTTL.Seconds()),
	}, nil
}

// generateToken creates a JWT token with specified parameters
func (j *JWTService) generateToken(userID uint, googleID, email, name, tokenType string, ttl time.Duration, secret []byte) (string, error) {
	// Create unique JTI for token tracking
	jti, err := j.generateJTI()
	if err != nil {
		return "", fmt.Errorf("failed to generate JTI: %w", err)
	}

	now := time.Now()
	claims := &JWTClaims{
		UserID:    userID,
		GoogleID:  googleID,
		Email:     email,
		Name:      name,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "course-management-api",
			Audience:  []string{"course-management-mobile"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateAccessToken validates and extracts claims from an access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, "access", j.accessSecret)
}

// ValidateRefreshToken validates and extracts claims from a refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, "refresh", j.refreshSecret)
}

// validateToken validates a JWT token and returns claims
func (j *JWTService) validateToken(tokenString, expectedType string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validate token type
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Validate issuer
	if claims.Issuer != "course-management-api" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}

// RefreshTokens generates new token pair using a valid refresh token
func (j *JWTService) RefreshTokens(refreshTokenString string) (*TokenResponse, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.GoogleID, claims.Email, claims.Name)
}

// ExtractTokenFromHeader extracts Bearer token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	// Check for Bearer prefix
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// generateJTI creates a cryptographically secure random token ID
func (j *JWTService) generateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GetUserClaims extracts user claims from echo context (set by middleware)
func GetUserClaims(c echo.Context) (*JWTClaims, error) {
	claims, ok := c.Get("user_claims").(*JWTClaims)
	if !ok {
		return nil, errors.New("user claims not found in context")
	}
	return claims, nil
}

// GetUserID extracts user ID from echo context
func GetUserID(c echo.Context) (uint, error) {
	claims, err := GetUserClaims(c)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}