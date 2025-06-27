package main

import (
	"context"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

type AuthHandlers struct {
	sessionService *SessionService
}

type GoogleUser struct {
	ID       string `json:"sub"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Verified bool   `json:"email_verified"`
}

func NewAuthHandlers() *AuthHandlers {
	return &AuthHandlers{
		sessionService: NewSessionService(),
	}
}

func (a *AuthHandlers) VerifyGoogleToken(c echo.Context) error {
	// Get the JWT credential from the request
	credential := c.FormValue("credential")
	if credential == "" {
		return c.String(http.StatusBadRequest, "Missing credential")
	}

	// Verify the JWT token with Google
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	payload, err := idtoken.Validate(context.Background(), credential, clientID)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Invalid token: "+err.Error())
	}

	// Extract user information from the token
	user := &GoogleUser{
		ID:       payload.Subject,
		Email:    payload.Claims["email"].(string),
		Name:     payload.Claims["name"].(string),
		Picture:  payload.Claims["picture"].(string),
		Verified: payload.Claims["email_verified"].(bool),
	}

	// Verify email is verified
	if !user.Verified {
		return c.String(http.StatusUnauthorized, "Email not verified")
	}

	// Save user to session
	if err := a.sessionService.SetUser(c, user); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save session")
	}

	// Return HTMX-friendly response that redirects to main app
	return c.HTML(http.StatusOK, `
		<div hx-get="/introduction" hx-target="body" hx-trigger="load">
			<div style="text-align: center; padding: 40px; color: #204606;">
				<h2>Welcome, `+user.Name+`!</h2>
				<p>Redirecting to dashboard...</p>
			</div>
		</div>
	`)
}

func (a *AuthHandlers) Logout(c echo.Context) error {
	a.sessionService.Logout(c)

	// Return the login form
	return c.Render(http.StatusOK, "authentication", map[string]string{
		"GoogleClientID": os.Getenv("GOOGLE_CLIENT_ID"),
	})
}

func (a *AuthHandlers) GetAuthStatus(c echo.Context) error {
	user := a.sessionService.GetUser(c)
	if user == nil {
		return c.Render(http.StatusOK, "authentication", map[string]string{
			"GoogleClientID": os.Getenv("GOOGLE_CLIENT_ID"),
		})
	}

	return c.Render(http.StatusOK, "user-profile", user)
}
