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
	sess, _ := session.Get("session", c)
	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		return nil
	}

	return &GoogleUser{
		ID:      sess.Values["user_id"].(string),
		Email:   sess.Values["user_email"].(string),
		Name:    sess.Values["user_name"].(string),
		Picture: sess.Values["user_picture"].(string),
	}
}

func (s *SessionService) Logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Values["authenticated"] = false
	return sess.Save(c.Request(), c.Response())
}

func (s *SessionService) IsAuthenticated(c echo.Context) bool {
	sess, _ := session.Get("session", c)
	if auth, ok := sess.Values["authenticated"].(bool); ok && auth {
		return true
	}
	return false
}
