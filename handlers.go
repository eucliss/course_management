package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	courses       *[]Course
	courseService *CourseService
}

func NewHandlers(courses *[]Course, courseService *CourseService) *Handlers {
	return &Handlers{
		courses:       courses,
		courseService: courseService,
	}
}

func (h *Handlers) Home(c echo.Context) error {
	sessionService := NewSessionService()
	user := sessionService.GetUser(c)

	data := struct {
		Courses     []Course
		MapboxToken string
		User        *GoogleUser
	}{
		Courses:     *h.courses,
		MapboxToken: os.Getenv("MAPBOX_ACCESS_TOKEN"),
		User:        user,
	}

	return c.Render(http.StatusOK, "welcome", data)
}

func (h *Handlers) Introduction(c echo.Context) error {
	return c.Render(http.StatusOK, "introduction", PageData{
		Courses: *h.courses,
	})
}

func (h *Handlers) Profile(c echo.Context) error {
	sessionService := NewSessionService()
	user := sessionService.GetUser(c)

	if user == nil {
		return c.Render(http.StatusOK, "authentication", map[string]string{
			"GoogleClientID": os.Getenv("GOOGLE_CLIENT_ID"),
		})
	}

	data := struct {
		*GoogleUser
		Courses []Course
	}{
		GoogleUser: user,
		Courses:    *h.courses,
	}

	return c.Render(http.StatusOK, "user-profile", data)
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
	data := struct {
		Course  Course
		Courses []Course
		IsEdit  bool
	}{
		Course:  Course{},
		Courses: *h.courses,
		IsEdit:  false,
	}

	return c.Render(http.StatusOK, "create-course", data)
}

func (h *Handlers) EditCourseForm(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt >= len(*h.courses) {
		return c.String(http.StatusNotFound, "Course not found")
	}

	course := (*h.courses)[idInt]

	data := struct {
		Course  Course
		Courses []Course
		IsEdit  bool
	}{
		Course:  course,
		Courses: *h.courses,
		IsEdit:  true,
	}

	return c.Render(http.StatusOK, "create-course", data)
}

func (h *Handlers) UpdateCourse(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt >= len(*h.courses) {
		return c.String(http.StatusNotFound, "Course not found")
	}

	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	course, err := h.parseFormToCourse(c, idInt)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	(*h.courses)[idInt] = course

	if err := h.courseService.UpdateCourse(course); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update course: "+err.Error())
	}

	return h.renderSuccessMessage(c, "Course Updated Successfully!", "has been updated and saved", course.Name)
}

func (h *Handlers) CreateCourse(c echo.Context) error {
	log.Printf("[CREATE_COURSE] Starting request from %s", c.RealIP())

	if err := c.Request().ParseForm(); err != nil {
		log.Printf("[CREATE_COURSE] ERROR: Failed to parse form: %v", err)
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	course, err := h.parseFormToCourse(c, 0)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err := h.courseService.SaveCourse(course); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save course: "+err.Error())
	}

	// Reload courses to include the new one
	if err := h.reloadCourses(); err != nil {
		log.Printf("Warning: failed to reload courses: %v", err)
	}

	return h.renderSuccessMessage(c, "Course Created Successfully!", "has been created and saved", course.Name)
}

func (h *Handlers) Map(c echo.Context) error {
	coursesJSON, err := json.Marshal(*h.courses)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to marshal courses to JSON: "+err.Error())
	}

	data := struct {
		Courses     []Course
		CoursesJSON template.JS
		MapboxToken string
	}{
		Courses:     *h.courses,
		CoursesJSON: template.JS(coursesJSON),
		MapboxToken: os.Getenv("MAPBOX_ACCESS_TOKEN"),
	}

	return c.Render(http.StatusOK, "map", data)
}

// Helper methods

func (h *Handlers) parseFormToCourse(c echo.Context, existingID int) (Course, error) {
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

	if name == "" || description == "" || overallRating == "" {
		return Course{}, fmt.Errorf("missing required fields")
	}

	course := Course{
		ID:            existingID,
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

	holes, scores, err := h.courseService.ParseFormData(c.Request().Form)
	if err != nil {
		return Course{}, err
	}

	course.Holes = holes
	course.Scores = scores

	return course, nil
}

func (h *Handlers) renderSuccessMessage(c echo.Context, title, message, courseName string) error {
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<div style="text-align: center; padding: 40px; color: #204606;">
			<h1 style="color: #204606; margin-bottom: 20px;">%s</h1>
			<p style="font-size: 18px; margin-bottom: 30px;">The course "<strong>%s</strong>" %s.</p>
			<button hx-get="/introduction" hx-target="#main-content" style="background-color: #204606; color: #FFFCE7; padding: 15px 30px; border: none; border-radius: 4px; cursor: pointer; font-size: 16px;">Return to Home</button>
		</div>
	`, title, courseName, message))
}

func (h *Handlers) reloadCourses() error {
	courses, err := h.courseService.LoadCourses()
	if err != nil {
		return err
	}
	*h.courses = courses
	return nil
}
