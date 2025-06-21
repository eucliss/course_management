package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	courses *[]Course
}

func NewHandlers(courses *[]Course) *Handlers {
	return &Handlers{courses: courses}
}

func (h *Handlers) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "welcome", PageData{
		Courses: *h.courses,
	})
}

func (h *Handlers) Introduction(c echo.Context) error {
	return c.Render(http.StatusOK, "introduction", PageData{
		Courses: *h.courses,
	})
}

func (h *Handlers) GetCourse(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt >= len(*h.courses) {
		return c.String(http.StatusNotFound, "Course not found")
	}
	return c.Render(http.StatusOK, "course", (*h.courses)[idInt])
}

func (h *Handlers) CreateCourseForm(c echo.Context) error {
	return c.Render(http.StatusOK, "create-course", PageData{
		Courses: *h.courses,
	})
}

func (h *Handlers) CreateCourse(c echo.Context) error {
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
	address := c.FormValue("address")
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
		Address:       address,
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

	// Use the service to parse complex data
	courseService := NewCourseService()
	holes, scores, err := courseService.ParseFormData(c.Request().Form)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	course.Holes = holes
	course.Scores = scores

	// Use the service to save
	if err := courseService.SaveCourse(course); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save course: "+err.Error())
	}

	// Reload courses to include the new one
	courses, err := loadCourses()
	if err != nil {
		log.Printf("Warning: failed to reload courses: %v", err)
	} else {
		*h.courses = courses // Update the courses slice
	}

	// Return success message and redirect to home
	return c.HTML(http.StatusOK, `
		<div style="text-align: center; padding: 40px; color: #204606;">
			<h1 style="color: #204606; margin-bottom: 20px;">Course Created Successfully!</h1>
			<p style="font-size: 18px; margin-bottom: 30px;">The course "<strong>`+name+`</strong>" has been created and saved.</p>
			<button hx-get="/introduction" hx-target="#main-content" style="background-color: #204606; color: #FFFCE7; padding: 15px 30px; border: none; border-radius: 4px; cursor: pointer; font-size: 16px;">Return to Home</button>
		</div>
	`)
}

func (h *Handlers) Welcome(c echo.Context) error {
	courses, err := loadCourses()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses: "+err.Error())
	}

	// Convert courses to JSON
	coursesJSON, err := json.Marshal(courses)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to marshal courses to JSON: "+err.Error())
	}

	data := struct {
		Courses     []Course
		MapboxToken string
		CoursesJSON template.JS
	}{
		Courses:     courses,
		MapboxToken: os.Getenv("MAPBOX_ACCESS_TOKEN"),
		CoursesJSON: template.JS(coursesJSON),
	}

	return c.Render(http.StatusOK, "welcome", data)
}

func (h *Handlers) Map(c echo.Context) error {
	data := struct {
		MapboxToken string
	}{
		MapboxToken: os.Getenv("MAPBOX_ACCESS_TOKEN"),
	}

	return c.Render(http.StatusOK, "map", data)
}
