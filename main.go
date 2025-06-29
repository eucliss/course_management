package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseFiles(
			"views/welcome.html",
			"views/course.html",
			"views/introduction.html",
			"views/create-course.html",
			"views/map.html",
			"views/authentication.html",
			"views/sidebar.html",
		)),
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}

	// Open log file
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	// Log to both file and console
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)
			status := c.Response().Status

			if err != nil {
				log.Printf("[REQUEST] %s %s - ERROR: %v (Duration: %v)",
					c.Request().Method, c.Request().URL.Path, err, duration)
			} else {
				log.Printf("[REQUEST] %s %s - %d (Duration: %v)",
					c.Request().Method, c.Request().URL.Path, status, duration)
			}

			return err
		}
	}
}

func RequireAuth(sessionService *SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !sessionService.IsAuthenticated(c) {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(c)
		}
	}
}

func main() {
	config := LoadConfig()
	courseService := NewCourseService()
	sessionService := NewSessionService()

	courses, err := courseService.LoadCourses()
	if err != nil {
		log.Printf("Warning: failed to load courses: %v", err)
		courses = []Course{}
	}

	handlers := NewHandlers(&courses, courseService)

	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(RequestLogger())

	// Session middleware with proper secret handling
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Printf("Warning: SESSION_SECRET not set, using default (not secure for production)")
		sessionSecret = "development-secret-key-please-change-in-production-32chars"
	}
	if len(sessionSecret) < 32 {
		log.Printf("Warning: SESSION_SECRET should be at least 32 characters")
	}
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	// Auth handlers
	authHandlers := NewAuthHandlers()

	// Auth routes
	e.POST("/auth/google/verify", authHandlers.VerifyGoogleToken)
	e.POST("/auth/logout", authHandlers.Logout)
	e.GET("/login", authHandlers.GetAuthStatus)

	// Application routes
	e.GET("/", handlers.Home)
	e.GET("/introduction", handlers.Introduction)
	e.GET("/profile", handlers.Profile)
	e.GET("/course/:id", handlers.GetCourse)
	e.GET("/create-course", handlers.CreateCourseForm, RequireAuth(sessionService))
	e.POST("/create-course", handlers.CreateCourse)
	e.GET("/map", handlers.Map)
	e.GET("/edit-course/:id", handlers.EditCourseForm)
	e.POST("/edit-course/:id", handlers.UpdateCourse)
	e.Static("/favicon.ico", "static/favicon.ico")

	log.Printf("Server starting on port %s", config.Port)
	e.Logger.Fatal(e.Start(":" + config.Port))
}
