package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
}

func loadCourses() ([]Course, error) {
	var courses []Course

	// Read all files from courses directory
	files, err := os.ReadDir("courses")
	if err != nil {
		return nil, fmt.Errorf("failed to read courses directory: %v", err)
	}

	courseID := 0
	// Load each JSON file
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Skip schema files
		if strings.Contains(file.Name(), "schema") {
			continue
		}

		data, err := os.ReadFile(filepath.Join("courses", file.Name()))
		if err != nil {
			log.Printf("Warning: failed to read course file %s: %v", file.Name(), err)
			continue
		}

		var course Course
		if err := json.Unmarshal(data, &course); err != nil {
			log.Printf("Warning: failed to parse course file %s: %v", file.Name(), err)
			continue
		}

		// Assign unique ID
		course.ID = courseID
		courseID++

		courses = append(courses, course)
	}

	if len(courses) == 0 {
		return nil, fmt.Errorf("no course files found in courses directory")
	}

	return courses, nil
}

func sanitizeFilename(name string) string {
	// Replace spaces and special characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return strings.ToLower(reg.ReplaceAllString(name, "_"))
}

func main() {
	config := LoadConfig()

	courseService := NewCourseService()
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

	// Routes
	e.GET("/", handlers.Home)
	e.GET("/introduction", handlers.Introduction)
	e.GET("/course/:id", handlers.GetCourse)
	e.GET("/create-course", handlers.CreateCourseForm)
	e.POST("/create-course", handlers.CreateCourse)
	e.GET("/map", handlers.Map)
	e.Static("/favicon.ico", "static/favicon.ico")

	log.Printf("Server starting on port %s", config.Port)
	e.Logger.Fatal(e.Start(":" + config.Port))
}
