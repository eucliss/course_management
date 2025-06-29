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
