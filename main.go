package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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
		)),
	}
}

// Add proper error types
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

// Use consistent error handling
func handleError(c echo.Context, err error) error {
	if appErr, ok := err.(*AppError); ok {
		return c.String(appErr.Code, appErr.Message)
	}
	return c.String(http.StatusInternalServerError, "Internal server error")
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

			log.Printf("[REQUEST] %s %s - Started", c.Request().Method, c.Request().URL.Path)

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

func main() {
	config := LoadConfig()
	courseService := NewCourseService()

	// Remove the duplicate loadCourses() function from main.go
	courses, err := courseService.LoadCourses()
	if err != nil {
		log.Printf("Warning: failed to load courses: %v", err)
		courses = []Course{}
	}

	handlers := NewHandlers(&courses)

	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(RequestLogger())

	// Routes
	e.GET("/", handlers.Home)
	e.GET("/introduction", handlers.Introduction)
	e.GET("/course/:id", handlers.GetCourse)
	e.GET("/create-course", handlers.CreateCourseForm)
	e.POST("/create-course", handlers.CreateCourse)
	e.GET("/map", handlers.Map)
	e.GET("/login", handlers.LoginForm)
	e.POST("/login", handlers.Login)
	e.GET("/edit-course/:id", handlers.EditCourseForm)
	e.POST("/edit-course/:id", handlers.UpdateCourse)
	e.Static("/favicon.ico", "static/favicon.ico")

	log.Printf("Server starting on port %s", config.Port)
	e.Logger.Fatal(e.Start(":" + config.Port))
}
