package main

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"course_management/config"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates(viewsDir string) *Templates {
	templateFiles := []string{
		filepath.Join(viewsDir, "welcome.html"),
		filepath.Join(viewsDir, "course.html"),
		filepath.Join(viewsDir, "introduction.html"),
		filepath.Join(viewsDir, "review-landing.html"),
		filepath.Join(viewsDir, "map.html"),
		filepath.Join(viewsDir, "authentication.html"),
		filepath.Join(viewsDir, "sidebar.html"),
		filepath.Join(viewsDir, "review-course.html"),
	}
	
	return &Templates{
		templates: template.Must(template.ParseFiles(templateFiles...)),
	}
}

func setupLogging(cfg *config.Config) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}

	// Only log to file if not using stdout
	if cfg.Logging.Output != "stdout" {
		logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file: %v", err)
			return
		}

		// Log to both file and console
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	// Set log format
	if cfg.Logging.Format == "json" {
		log.SetFlags(0) // JSON logging handles timestamps
	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}
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
func RequireOwnership(sessionService *SessionService) echo.MiddlewareFunc {
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

			// Check ownership using database service
			dbService := NewDatabaseService()

			// Get all courses from database to get the course name
			allCourses, err := dbService.GetAllCoursesFromDatabase()
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to load courses")
			}

			if courseIndex < 0 || courseIndex >= len(allCourses) {
				return c.String(http.StatusNotFound, "Course not found")
			}

			courseName := allCourses[courseIndex].Name
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
	// Determine environment
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "development"
	}

	// Load configuration
	cfg, err := config.LoadConfigFromFile(environment)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging based on configuration
	setupLogging(cfg)

	log.Printf("🚀 Starting Course Management System")
	log.Printf("📊 Environment: %s", cfg.Environment)
	log.Printf("🌐 Server: %s", cfg.GetServerAddress())
	log.Printf("🗄️ Database: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	// Validate secrets
	secretManager := config.NewSecretManager(cfg)
	if err := secretManager.ValidateSecrets(); err != nil {
		if cfg.IsProduction() {
			log.Fatalf("❌ Secret validation failed in production: %v", err)
		} else {
			log.Printf("⚠️ Secret validation warning: %v", err)
		}
	}

	// Initialize database connection
	if err := InitDatabase(); err != nil {
		log.Fatalf("❌ Database initialization failed: %v", err)
	} else {
		log.Printf("✅ Database connected successfully")
		
		// Create performance indexes if database is available
		if err := CreatePerformanceIndexes(); err != nil {
			log.Printf("⚠️ Failed to create performance indexes: %v", err)
		}
	}

	sessionService := NewSessionService()
	handlers := NewHandlers()

	// Create Echo instance
	e := echo.New()
	e.Renderer = NewTemplates(cfg.Paths.ViewsDir)

	// Configure middleware based on environment
	if cfg.IsDevelopment() {
		e.Use(middleware.Logger())
	}
	e.Use(middleware.Recover())
	e.Use(RequestLogger())

	// Request size limiting
	e.Use(middleware.BodyLimit(strconv.FormatInt(cfg.Server.MaxRequestSize, 10)))

	// Security middleware
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		HSTSPreloadEnabled:    false,
	}))

	// Rate limiting middleware
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(cfg.Security.RateLimitPerMin))))

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Session middleware with configuration-based settings
	store := sessions.NewCookieStore([]byte(cfg.Security.SessionSecret))
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = cfg.Security.SecureCookies
	store.Options.SameSite = http.SameSiteStrictMode
	store.MaxAge(int(cfg.Security.SessionTimeout.Seconds()))
	e.Use(session.Middleware(store))

	// Auth handlers
	authHandlers := NewAuthHandlers()

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":      "healthy",
			"environment": cfg.Environment,
			"version":     "1.0.0",
			"timestamp":   time.Now().Unix(),
		})
	})

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
	e.POST("/profile/add-score", handlers.AddScore, RequireAuth(sessionService))
	e.GET("/course/:id", handlers.GetCourse, AddOwnershipContext(sessionService))
	e.GET("/review-landing", handlers.CreateCourseForm, RequireAuth(sessionService))
	e.GET("/review-course/:id", handlers.ReviewSpecificCourseForm, RequireAuth(sessionService))
	e.POST("/create-course", handlers.CreateCourse, RequireAuth(sessionService))
	e.GET("/map", handlers.Map, AddOwnershipContext(sessionService))

	// Protected edit routes with ownership verification
	e.GET("/edit-course/:id", handlers.EditCourseForm, RequireOwnership(sessionService))
	e.POST("/edit-course/:id", handlers.UpdateCourse, RequireOwnership(sessionService))
	e.DELETE("/delete-course/:id", handlers.DeleteCourse, RequireOwnership(sessionService))

	// Review management routes
	e.DELETE("/delete-review/:id", handlers.DeleteReview, RequireAuth(sessionService))

	// API routes
	e.GET("/api/status/database", handlers.DatabaseStatus)
	e.POST("/api/migrate/courses", handlers.MigrateCourses)
	e.GET("/api/courses/all", handlers.GetAllCoursesAPI, AddOwnershipContext(sessionService))
	e.GET("/api/courses/review", handlers.GetReviewCoursesAPI, RequireAuth(sessionService))

	// Serve static files
	e.Static("/static", cfg.Paths.StaticDir)
	e.File("/favicon.ico", filepath.Join(cfg.Paths.StaticDir, "favicon.ico"))

	// Start server with graceful shutdown
	startServer(e, cfg)
}

func startServer(e *echo.Echo, cfg *config.Config) {
	// Configure server timeouts
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	// Start server in a goroutine
	go func() {
		log.Printf("🌟 Server starting on %s", cfg.GetServerAddress())
		if err := e.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Printf("🛑 Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Gracefully shutdown the server
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	} else {
		log.Printf("✅ Server gracefully stopped")
	}
}
