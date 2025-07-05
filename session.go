package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type SessionService struct{}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func (s *SessionService) SetUser(c echo.Context, user *GoogleUser) error {
	sess, err := session.Get("session", c)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return fmt.Errorf("failed to get session: %v", err)
	}
	sess.Values["user_id"] = user.ID
	sess.Values["user_email"] = user.Email
	sess.Values["user_name"] = user.Name
	sess.Values["user_picture"] = user.Picture
	sess.Values["authenticated"] = true
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		log.Printf("Failed to save session: %v", err)
		return fmt.Errorf("failed to save session: %v", err)
	}
	log.Printf("Session saved successfully for user: %s", user.Email)
	return nil
}

func (s *SessionService) SetDatabaseUser(c echo.Context, user *GoogleUser, dbUserID uint) error {
	sess, err := session.Get("session", c)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return fmt.Errorf("failed to get session: %v", err)
	}
	sess.Values["user_id"] = user.ID
	sess.Values["user_email"] = user.Email
	sess.Values["user_name"] = user.Name
	sess.Values["user_picture"] = user.Picture
	sess.Values["db_user_id"] = dbUserID
	sess.Values["authenticated"] = true
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		log.Printf("Failed to save session: %v", err)
		return fmt.Errorf("failed to save session: %v", err)
	}
	log.Printf("Session saved successfully for user: %s (DB ID: %d)", user.Email, dbUserID)
	return nil
}

func (s *SessionService) GetUser(c echo.Context) *GoogleUser {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil
	}
	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		return nil
	}

	// Type assertions with safety checks
	userID, ok1 := sess.Values["user_id"].(string)
	userEmail, ok2 := sess.Values["user_email"].(string)
	userName, ok3 := sess.Values["user_name"].(string)
	userPicture, ok4 := sess.Values["user_picture"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		log.Printf("Warning: Invalid session data types")
		return nil
	}

	return &GoogleUser{
		ID:      userID,
		Email:   userEmail,
		Name:    userName,
		Picture: userPicture,
	}
}

func (s *SessionService) GetDatabaseUserID(c echo.Context) *uint {
	sess, err := session.Get("session", c)
	if err != nil {
		log.Printf("🔍 GetDatabaseUserID: Failed to get session: %v", err)
		return nil
	}
	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		log.Printf("🔍 GetDatabaseUserID: User not authenticated")
		return nil
	}

	if dbUserID, ok := sess.Values["db_user_id"].(uint); ok {
		// Also log the user's email for debugging
		if userEmail, ok := sess.Values["user_email"].(string); ok {
			log.Printf("🔍 GetDatabaseUserID: Found DB User ID: %d for user: %s", dbUserID, userEmail)
		} else {
			log.Printf("🔍 GetDatabaseUserID: Found DB User ID: %d (no email in session)", dbUserID)
		}
		return &dbUserID
	}

	log.Printf("🔍 GetDatabaseUserID: No db_user_id in session. Session keys: %v", getSessionKeys(sess.Values))
	return nil
}

// Helper function to debug session contents
func getSessionKeys(values map[interface{}]interface{}) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		if str, ok := k.(string); ok {
			keys = append(keys, str)
		}
	}
	return keys
}

func (s *SessionService) Logout(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Values["authenticated"] = false
	// Clear all user data
	delete(sess.Values, "user_id")
	delete(sess.Values, "user_email")
	delete(sess.Values, "user_name")
	delete(sess.Values, "user_picture")
	delete(sess.Values, "db_user_id")
	return sess.Save(c.Request(), c.Response())
}

func (s *SessionService) IsAuthenticated(c echo.Context) bool {
	sess, err := session.Get("session", c)
	if err != nil {
		return false
	}
	if auth, ok := sess.Values["authenticated"].(bool); ok && auth {
		return true
	}
	return false
}
