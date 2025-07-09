package api

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, "Bearer", tokens.TokenType)
	assert.Equal(t, int64(3600), tokens.ExpiresIn) // 1 hour
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate tokens
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Validate access token
	claims, err := jwtService.ValidateAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, googleID, claims.GoogleID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, name, claims.Name)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "course-management-api", claims.Issuer)
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate tokens
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Validate refresh token
	claims, err := jwtService.ValidateRefreshToken(tokens.RefreshToken)
	require.NoError(t, err)
	
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, googleID, claims.GoogleID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, name, claims.Name)
	assert.Equal(t, "refresh", claims.TokenType)
	assert.Equal(t, "course-management-api", claims.Issuer)
}

func TestJWTService_RefreshTokens(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate initial tokens
	initialTokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Refresh tokens
	newTokens, err := jwtService.RefreshTokens(initialTokens.RefreshToken)
	require.NoError(t, err)
	
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	assert.NotEqual(t, initialTokens.AccessToken, newTokens.AccessToken)
	assert.NotEqual(t, initialTokens.RefreshToken, newTokens.RefreshToken)
}

func TestJWTService_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	// Test invalid token
	_, err := jwtService.ValidateAccessToken("invalid-token")
	assert.Error(t, err)
	
	// Test empty token
	_, err = jwtService.ValidateAccessToken("")
	assert.Error(t, err)
}

func TestJWTService_ExpiredToken(t *testing.T) {
	// Create service with very short TTL for testing
	jwtService := &JWTService{
		accessSecret:  []byte("test-access-secret-very-long-key"),
		refreshSecret: []byte("test-refresh-secret-very-long-key"),
		accessTTL:     time.Nanosecond, // Immediate expiry
		refreshTTL:    time.Hour,
	}
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate tokens (will be expired immediately)
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Wait a bit to ensure expiry
	time.Sleep(time.Millisecond)

	// Validate expired token
	_, err = jwtService.ValidateAccessToken(tokens.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWTService_WrongTokenType(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate tokens
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Try to validate refresh token as access token
	// Note: This will fail with signature error since different secrets are used
	_, err = jwtService.ValidateAccessToken(tokens.RefreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature is invalid")

	// Try to validate access token as refresh token  
	// Note: This will fail with signature error since different secrets are used
	_, err = jwtService.ValidateRefreshToken(tokens.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestJWTService_WrongSecret(t *testing.T) {
	jwtService1 := NewJWTService("test-access-secret-1", "test-refresh-secret-1")
	jwtService2 := NewJWTService("test-access-secret-2", "test-refresh-secret-2")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate tokens with first service
	tokens, err := jwtService1.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Try to validate with second service (different secret)
	_, err = jwtService2.ValidateAccessToken(tokens.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestJWTService_generateJTI(t *testing.T) {
	jwtService := NewJWTService("test-access-secret", "test-refresh-secret")
	
	// Generate multiple JTIs
	jti1, err := jwtService.generateJTI()
	require.NoError(t, err)
	assert.NotEmpty(t, jti1)
	
	jti2, err := jwtService.generateJTI()
	require.NoError(t, err)
	assert.NotEmpty(t, jti2)
	
	// JTIs should be unique
	assert.NotEqual(t, jti1, jti2)
	
	// JTIs should be valid base64 URL encoding (may contain = padding)
	assert.Regexp(t, "^[A-Za-z0-9_=-]+$", jti1)
	assert.Regexp(t, "^[A-Za-z0-9_=-]+$", jti2)
}

func TestJWTService_TokenClaims(t *testing.T) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	// Generate tokens
	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	require.NoError(t, err)

	// Parse token manually to check all claims
	token, err := jwt.ParseWithClaims(tokens.AccessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtService.accessSecret, nil
	})
	require.NoError(t, err)

	claims, ok := token.Claims.(*JWTClaims)
	require.True(t, ok)

	// Check all claims
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, googleID, claims.GoogleID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, name, claims.Name)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "course-management-api", claims.Issuer)
	assert.Equal(t, jwt.ClaimStrings{"course-management-mobile"}, claims.Audience)
	assert.NotEmpty(t, claims.ID) // JTI
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	assert.NotNil(t, claims.ExpiresAt)
	
	// Check timing
	now := time.Now()
	assert.True(t, claims.IssuedAt.Time.Before(now) || claims.IssuedAt.Time.Equal(now))
	assert.True(t, claims.NotBefore.Time.Before(now) || claims.NotBefore.Time.Equal(now))
	assert.True(t, claims.ExpiresAt.Time.After(now))
}

// Benchmark tests
func BenchmarkJWTService_GenerateTokenPair(b *testing.B) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTService_ValidateAccessToken(b *testing.B) {
	jwtService := NewJWTService("test-access-secret-very-long-key", "test-refresh-secret-very-long-key")
	
	userID := uint(123)
	googleID := "google123"
	email := "test@example.com"
	name := "Test User"

	tokens, err := jwtService.GenerateTokenPair(userID, googleID, email, name)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtService.ValidateAccessToken(tokens.AccessToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}