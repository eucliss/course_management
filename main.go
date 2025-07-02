package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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

// RequireOwnership middleware checks if user owns the course they're trying to edit
func RequireOwnership(sessionService *SessionService, courseService *CourseService, courses *[]Course) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// First check authentication
			if !sessionService.IsAuthenticated(c) {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			// Get course ID from URL parameter
			courseIDParam := c.Param("id")
			courseIndex, err := strconv.Atoi(courseIDParam)
			if err != nil {
				return c.String(http.StatusBadRequest, "Invalid course ID")
			}

			// Get user ID from session
			userID := sessionService.GetDatabaseUserID(c)
			if userID == nil {
				return c.String(http.StatusUnauthorized, "User not found in database")
			}

			// Check if database is available
			if DB == nil {
				return c.String(http.StatusServiceUnavailable, "Database not available")
			}

			// Check ownership using database service
			dbService := NewDatabaseService()

			// OPTIMIZED: Get course name from in-memory array and check ownership directly
			if courseIndex < 0 || courseIndex >= len(*courses) {
				return c.String(http.StatusNotFound, "Course not found")
			}

			courseName := (*courses)[courseIndex].Name
			isOwner, err := dbService.IsUserCourseOwner(*userID, courseName)
			if err != nil {
				log.Printf("❌ Error checking course ownership: %v", err)
				return c.String(http.StatusInternalServerError, "Error checking permissions")
			}

			if !isOwner {
				return c.String(http.StatusForbidden, "You don't have permission to edit this course")
			}

			// Store ownership info in context for handlers to use
			c.Set("userID", *userID)
			c.Set("courseIndex", courseIndex)
			c.Set("canEdit", true)

			return next(c)
		}
	}
}

// AddOwnershipContext middleware adds ownership information to all course-related routes
func AddOwnershipContext(sessionService *SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user ID from session if available
			userID := sessionService.GetDatabaseUserID(c)
			if userID != nil {
				c.Set("userID", *userID)
				c.Set("authenticated", true)
			} else {
				c.Set("authenticated", false)
			}

			return next(c)
		}
	}
}

func main() {
	config := LoadConfig()

	// Initialize database connection (optional)
	if err := InitDatabase(); err != nil {
		log.Printf("⚠️ Database initialization failed: %v", err)
		log.Printf("📁 Continuing with JSON file storage")
	} else {
		// Create performance indexes if database is available
		if err := CreatePerformanceIndexes(); err != nil {
			log.Printf("⚠️ Failed to create performance indexes: %v", err)
		}
	}

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

	// Application routes with ownership context
	e.GET("/", handlers.Home, AddOwnershipContext(sessionService))
	e.GET("/introduction", handlers.Introduction, AddOwnershipContext(sessionService))
	e.GET("/profile", handlers.Profile, AddOwnershipContext(sessionService))
	e.POST("/profile/handicap", handlers.UpdateHandicap, RequireAuth(sessionService))
	e.POST("/profile/display-name", handlers.UpdateDisplayName, RequireAuth(sessionService))
	e.GET("/course/:id", handlers.GetCourse, AddOwnershipContext(sessionService))
	e.GET("/create-course", handlers.CreateCourseForm, RequireAuth(sessionService))
	e.POST("/create-course", handlers.CreateCourse, RequireAuth(sessionService))
	e.GET("/map", handlers.Map, AddOwnershipContext(sessionService))

	// Protected edit routes with ownership verification
	e.GET("/edit-course/:id", handlers.EditCourseForm, RequireOwnership(sessionService, courseService, &courses))
	e.POST("/edit-course/:id", handlers.UpdateCourse, RequireOwnership(sessionService, courseService, &courses))
	e.DELETE("/delete-course/:id", handlers.DeleteCourse, RequireOwnership(sessionService, courseService, &courses))

	// API routes
	e.GET("/api/status/database", handlers.DatabaseStatus)
	e.POST("/api/migrate/courses", handlers.MigrateCourses)
	// Serve static files
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")

	log.Printf("Server starting on port %s", config.Port)
	e.Logger.Fatal(e.Start(":" + config.Port))
}
