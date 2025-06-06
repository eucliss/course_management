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
	"strconv"
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
		)),
	}
}

type Block struct {
	Id int
}

type Blocks struct {
	Start  int
	Next   int
	More   bool
	Blocks []Block
}

type Ranking struct {
	Price              string
	HandicapDifficulty int
	HazardDifficulty   int
	Merch              string
	Condition          string
	EnjoymentRating    string
	Vibe               string
	Range              string
	Amenities          string
	Glizzies           string
}

type Hole struct {
	Number      int
	Par         int
	Yardage     int
	Description string
}

type Course struct {
	Name          string
	ID            int
	Description   string
	Ranks         Ranking
	OverallRating string
	Review        string
	Holes         []Hole
	Scores        []Score
}

type Score struct {
	Score    int
	Handicap float64
}

type PageData struct {
	Courses []Course
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
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

func main() {
	// Get port from environment variable, default to 8080 if not set

	// Load courses from files
	courses, err := loadCourses()
	if err != nil {
		log.Printf("Warning: failed to load courses: %v", err)
		// Initialize with empty courses array if loading fails
		courses = []Course{}
	}

	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "welcome", PageData{
			Courses: courses,
		})
	})

	e.GET("/introduction", func(c echo.Context) error {
		return c.Render(http.StatusOK, "introduction", PageData{
			Courses: courses,
		})
	})

	e.GET("/course/:id", func(c echo.Context) error {
		id := c.Param("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return c.Render(http.StatusNotFound, "error", "Invalid course ID")
		}
		return c.Render(http.StatusOK, "course", courses[idInt])
	})

	e.Logger.Fatal(e.Start(":" + "8080"))
}

// e.GET("/blocks", func(c echo.Context) error {
//     startStr := c.QueryParam("start")
//     start, err := strconv.Atoi(startStr)
//     if err != nil {
//         start = 0
//     }

//     blocks := []Block{}
//     for i := start; i < start+10; i++ {
//         blocks = append(blocks, Block{Id: i})
//     }

//     template := "blocks"
//     if start == 0 {
//         template = "blocks-index"
//     }
//     return c.Render(http.StatusOK, template, Blocks{
//         Start:  start,
//         Next:   start + 10,
//         More:   start+10 < 100,
//         Blocks: blocks,
//     })
// })
