package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) HomeOptimized(c echo.Context) error {
	sessionService := NewSessionService()
	user := sessionService.GetUser(c)

	// Get user ID from middleware context if available
	var userID *uint
	if uid, ok := c.Get("userID").(uint); ok {
		userID = &uid
	}

	// Load courses with pagination (first 50 courses)
	const pageSize = 50
	courses, totalCount, err := h.loadCoursesOptimized(0, pageSize, userID)
	if err != nil {
		log.Printf("Error loading courses: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to load courses")
	}

	data := struct {
		Courses         []Course
		MapboxToken     string
		User            *GoogleUser
		EditPermissions map[int]bool
		TotalCourses    int64
		ShowPagination  bool
	}{
		Courses:         courses,
		MapboxToken:     os.Getenv("MAPBOX_ACCESS_TOKEN"),
		User:            user,
		EditPermissions: h.buildEditPermissions(courses, userID),
		TotalCourses:    totalCount,
		ShowPagination:  totalCount > pageSize,
	}

	return c.Render(http.StatusOK, "welcome", data)
}

func (h *Handlers) loadCoursesOptimized(offset, limit int, userID *uint) ([]Course, int64, error) {
	// Use optimized database loading
	dbService := NewDatabaseService()
	coursesDB, totalCount, err := dbService.GetCoursesWithPagination(offset, limit, false)
	if err != nil {
		return nil, 0, err
	}

	var courses []Course
	for i, courseDB := range coursesDB {
		var course Course
		if err := json.Unmarshal([]byte(courseDB.CourseData), &course); err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}
		course.ID = offset + i // Maintain consistent indexing
		courses = append(courses, course)
	}

	log.Printf("✅ Loaded %d courses from database (page %d-%d of %d total)",
		len(courses), offset, offset+limit, totalCount)
	return courses, totalCount, nil
}

func (h *Handlers) buildEditPermissions(courses []Course, userID *uint) map[int]bool {
	editPermissions := make(map[int]bool)
	if userID == nil {
		return editPermissions
	}

	for i := range courses {
		editPermissions[i] = h.CanEditCourse(i, userID)
	}
	return editPermissions
}

// Optimized version of CanEditCourse
func (h *Handlers) CanEditCourseOptimized(courseID uint, userID *uint) bool {
	if userID == nil {
		return false
	}

	dbService := NewDatabaseService()
	canEdit, err := dbService.CanEditCourseOptimized(courseID, *userID)
	if err != nil {
		log.Printf("Error checking course ownership: %v", err)
		return false
	}
	return canEdit
}
