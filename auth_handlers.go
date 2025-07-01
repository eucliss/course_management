package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

type AuthHandlers struct {
	sessionService *SessionService
	dbService      *DatabaseService
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
		dbService:      NewDatabaseService(),
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
	googleUser := &GoogleUser{
		ID:       payload.Subject,
		Email:    payload.Claims["email"].(string),
		Name:     payload.Claims["name"].(string),
		Picture:  payload.Claims["picture"].(string),
		Verified: payload.Claims["email_verified"].(bool),
	}

	// Verify email is verified
	if !googleUser.Verified {
		return c.String(http.StatusUnauthorized, "Email not verified")
	}

	// Create or update user in database
	var dbUser *User
	if DB != nil { // Only if database is available
		log.Printf("🔄 Database available, attempting to create/update user for: %s", googleUser.Email)

		existingUser, err := a.dbService.GetUserByGoogleID(googleUser.ID)
		if err != nil {
			log.Printf("❌ Database error when looking up user %s: %v", googleUser.Email, err)
			return c.String(http.StatusInternalServerError, "Database error: "+err.Error())
		}

		if existingUser == nil {
			// User doesn't exist, create new user
			log.Printf("📝 Creating new user for: %s", googleUser.Email)
			dbUser, err = a.dbService.CreateUser(googleUser)
			if err != nil {
				log.Printf("❌ Failed to create user %s: %v", googleUser.Email, err)
				return c.String(http.StatusInternalServerError, "Failed to create user: "+err.Error())
			}
			log.Printf("✅ Successfully created user %s with DB ID: %d", googleUser.Email, dbUser.ID)
		} else {
			// User exists, update their information in case it changed
			log.Printf("🔄 Updating existing user: %s (ID: %d)", googleUser.Email, existingUser.ID)
			err = a.dbService.UpdateUser(existingUser.ID, googleUser)
			if err != nil {
				log.Printf("❌ Failed to update user %s: %v", googleUser.Email, err)
				return c.String(http.StatusInternalServerError, "Failed to update user: "+err.Error())
			}
			dbUser = existingUser
			log.Printf("✅ Successfully updated user %s", googleUser.Email)
		}

		// Save user to session with database user ID
		if err := a.sessionService.SetDatabaseUser(c, googleUser, dbUser.ID); err != nil {
			log.Printf("❌ Failed to save session for user %s: %v", googleUser.Email, err)
			return c.String(http.StatusInternalServerError, "Failed to save session")
		}
		log.Printf("✅ Session saved for user %s with DB ID: %d", googleUser.Email, dbUser.ID)
	} else {
		log.Printf("⚠️ Database not available, using session-only authentication for: %s", googleUser.Email)
		// No database available, use regular session
		if err := a.sessionService.SetUser(c, googleUser); err != nil {
			log.Printf("❌ Failed to save session for user %s: %v", googleUser.Email, err)
			return c.String(http.StatusInternalServerError, "Failed to save session")
		}
		log.Printf("✅ Session-only authentication saved for user %s", googleUser.Email)
	}

	// Return HTMX-friendly response that redirects to main app
	return c.HTML(http.StatusOK, `
		<div hx-get="/" hx-target="body" hx-trigger="load">
			<div style="text-align: center; padding: 40px; color: #204606;">
				<h2>Welcome, `+googleUser.Name+`!</h2>
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
