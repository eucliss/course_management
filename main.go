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
			"views/create-course.html",
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

func sanitizeFilename(name string) string {
	// Replace spaces and special characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return strings.ToLower(reg.ReplaceAllString(name, "_"))
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

	// Course creation routes
	e.GET("/create-course", func(c echo.Context) error {
		return c.Render(http.StatusOK, "create-course", PageData{
			Courses: courses,
		})
	})

	e.POST("/create-course", func(c echo.Context) error {
		// Parse form data first
		if err := c.Request().ParseForm(); err != nil {
			return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
		}

		// Parse basic form data
		name := c.FormValue("name")
		description := c.FormValue("description")
		overallRating := c.FormValue("overallRating")
		price := c.FormValue("price")
		handicapDifficulty, _ := strconv.Atoi(c.FormValue("handicapDifficulty"))
		hazardDifficulty, _ := strconv.Atoi(c.FormValue("hazardDifficulty"))
		condition := c.FormValue("condition")
		merch := c.FormValue("merch")
		enjoymentRating := c.FormValue("enjoymentRating")
		vibe := c.FormValue("vibe")
		rangeRating := c.FormValue("range")
		amenities := c.FormValue("amenities")
		glizzies := c.FormValue("glizzies")
		review := c.FormValue("review")

		// Validate required fields
		if name == "" || description == "" || overallRating == "" {
			return c.String(http.StatusBadRequest, "Missing required fields")
		}

		// Create course structure
		course := Course{
			Name:          name,
			Description:   description,
			OverallRating: overallRating,
			Review:        review,
			Ranks: Ranking{
				Price:              price,
				HandicapDifficulty: handicapDifficulty,
				HazardDifficulty:   hazardDifficulty,
				Merch:              merch,
				Condition:          condition,
				EnjoymentRating:    enjoymentRating,
				Vibe:               vibe,
				Range:              rangeRating,
				Amenities:          amenities,
				Glizzies:           glizzies,
			},
			Holes:  []Hole{},
			Scores: []Score{},
		}

		// Parse holes data
		form := c.Request().Form
		holeMap := make(map[int]Hole)

		for key, values := range form {
			if strings.HasPrefix(key, "holes[") && len(values) > 0 {
				// Extract hole index and field name
				parts := strings.Split(key, "].")
				if len(parts) == 2 {
					indexStr := strings.TrimPrefix(parts[0], "holes[")
					fieldName := parts[1]
					index, err := strconv.Atoi(indexStr)
					if err != nil {
						continue
					}

					if _, exists := holeMap[index]; !exists {
						holeMap[index] = Hole{Number: index + 1}
					}

					hole := holeMap[index]
					switch fieldName {
					case "par":
						hole.Par, _ = strconv.Atoi(values[0])
					case "yardage":
						hole.Yardage, _ = strconv.Atoi(values[0])
					case "description":
						hole.Description = values[0]
					case "number":
						hole.Number, _ = strconv.Atoi(values[0])
					}
					holeMap[index] = hole
				}
			}
		}

		// Convert hole map to slice in order
		for i := 0; i < len(holeMap); i++ {
			if hole, exists := holeMap[i]; exists {
				course.Holes = append(course.Holes, hole)
			}
		}

		// Parse scores data
		scoreMap := make(map[int]Score)

		for key, values := range form {
			if strings.HasPrefix(key, "scores[") && len(values) > 0 {
				// Extract score index and field name
				parts := strings.Split(key, "].")
				if len(parts) == 2 {
					indexStr := strings.TrimPrefix(parts[0], "scores[")
					fieldName := parts[1]
					index, err := strconv.Atoi(indexStr)
					if err != nil {
						continue
					}

					if _, exists := scoreMap[index]; !exists {
						scoreMap[index] = Score{}
					}

					score := scoreMap[index]
					switch fieldName {
					case "score":
						score.Score, _ = strconv.Atoi(values[0])
					case "handicap":
						score.Handicap, _ = strconv.ParseFloat(values[0], 64)
					}
					scoreMap[index] = score
				}
			}
		}

		// Convert score map to slice in order
		for i := 0; i < len(scoreMap); i++ {
			if score, exists := scoreMap[i]; exists {
				course.Scores = append(course.Scores, score)
			}
		}

		// Generate filename
		filename := sanitizeFilename(name) + ".json"
		filepath := filepath.Join("courses", filename)

		// Check if file already exists
		if _, err := os.Stat(filepath); err == nil {
			return c.String(http.StatusBadRequest, "A course with this name already exists")
		}

		// Convert course to JSON
		courseJSON, err := json.MarshalIndent(course, "", "  ")
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to create course JSON: "+err.Error())
		}

		// Write to file
		if err := os.WriteFile(filepath, courseJSON, 0644); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to save course file: "+err.Error())
		}

		// Reload courses to include the new one
		courses, err = loadCourses()
		if err != nil {
			log.Printf("Warning: failed to reload courses: %v", err)
		}

		// Return success message and redirect to home
		return c.HTML(http.StatusOK, `
			<div style="text-align: center; padding: 40px; color: #204606;">
				<h1 style="color: #204606; margin-bottom: 20px;">Course Created Successfully!</h1>
				<p style="font-size: 18px; margin-bottom: 30px;">The course "<strong>`+name+`</strong>" has been created and saved.</p>
				<button hx-get="/introduction" hx-target="#main-content" style="background-color: #204606; color: #FFFCE7; padding: 15px 30px; border: none; border-radius: 4px; cursor: pointer; font-size: 16px;">Return to Home</button>
			</div>
		`)
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
