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
	dbUserID := sessionService.GetDatabaseUserID(c)

	// Check which courses the user can edit
	editPermissions := make(map[int]bool)
	if dbUserID != nil {
		for i := range *h.courses {
			editPermissions[i] = h.CanEditCourse(i, dbUserID)
		}
	}

	data := struct {
		Courses         []Course
		MapboxToken     string
		User            *GoogleUser
		EditPermissions map[int]bool
	}{
		Courses:         *h.courses,
		MapboxToken:     os.Getenv("MAPBOX_ACCESS_TOKEN"),
		User:            user,
		EditPermissions: editPermissions,
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

	// Get user's handicap from database if available
	var handicap *float64
	var displayName *string
	dbUserID := sessionService.GetDatabaseUserID(c)
	log.Printf("🔍 Profile request for user: %s, DB User ID: %v, DB available: %t",
		user.Email, dbUserID, DB != nil)

	if DB != nil {
		dbService := NewDatabaseService()
		var dbUser *User
		var err error

		if dbUserID != nil {
			// Try to get user by database ID first
			dbUser, err = dbService.GetUserByID(*dbUserID)
			if err != nil {
				log.Printf("❌ Error fetching user %d from database: %v", *dbUserID, err)
			}
		}

		// Fallback: if no dbUserID in session or user not found, try to find by Google ID
		if dbUser == nil && user != nil {
			log.Printf("🔄 Fallback: Looking up user by Google ID: %s", user.ID)
			dbUser, err = dbService.GetUserByGoogleID(user.ID)
			if err != nil {
				log.Printf("❌ Error fetching user by Google ID %s: %v", user.ID, err)
			}
		}

		if dbUser != nil {
			handicap = dbUser.Handicap
			displayName = dbUser.DisplayName
			if handicap != nil {
				log.Printf("✅ Found user in database - ID: %d, Handicap: %.1f", dbUser.ID, *handicap)
			} else {
				log.Printf("✅ Found user in database - ID: %d, Handicap: nil", dbUser.ID)
			}
			if displayName != nil {
				log.Printf("✅ Display name: %s", *displayName)
			} else {
				log.Printf("✅ No display name set")
			}

			// Update session with database user ID if it was missing
			if dbUserID == nil {
				log.Printf("🔄 Updating session with missing DB User ID: %d", dbUser.ID)
				if err := sessionService.SetDatabaseUser(c, user, dbUser.ID); err != nil {
					log.Printf("⚠️ Failed to update session with DB User ID: %v", err)
				}
			}
		} else {
			log.Printf("⚠️ User not found in database")
		}
	} else {
		log.Printf("⚠️ Database not available")
	}

	// Filter courses to only show ones the user can edit (owns)
	var userCourses []Course
	editPermissions := make(map[int]bool)
	if dbUserID != nil {
		for i, course := range *h.courses {
			canEdit := h.CanEditCourse(i, dbUserID)
			if canEdit {
				userCourses = append(userCourses, course)
				// Map the new index to edit permission (always true for user's courses)
				editPermissions[len(userCourses)-1] = true
			}
		}
	}

	data := struct {
		*GoogleUser
		Courses         []Course
		Handicap        *float64
		DisplayName     *string
		EditPermissions map[int]bool
	}{
		GoogleUser:      user,
		Courses:         userCourses,
		Handicap:        handicap,
		DisplayName:     displayName,
		EditPermissions: editPermissions,
	}

	if handicap != nil {
		log.Printf("📊 Rendering profile with handicap: %.1f", *handicap)
	} else {
		log.Printf("📊 Rendering profile with handicap: nil")
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

	// Get authenticated user ID
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "You must be logged in to edit a course")
	}

	// Check ownership if database is available
	if DB != nil {
		dbService := NewDatabaseService()
		canEdit, courseDB, err := dbService.CanEditCourseByIndex(idInt, *userID)
		if err != nil {
			log.Printf("Error checking course ownership: %v", err)
			return c.String(http.StatusInternalServerError, "Error verifying course ownership")
		}

		if !canEdit {
			return c.String(http.StatusForbidden, "You don't have permission to edit this course")
		}

		log.Printf("✅ User %d authorized to edit course at index %d (DB ID: %d)", *userID, idInt, courseDB.ID)
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

	// Get authenticated user ID
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "You must be logged in to edit a course")
	}

	// Check ownership and get course from database if available
	var courseDB *CourseDB
	if DB != nil {
		dbService := NewDatabaseService()
		canEdit, dbCourse, err := dbService.CanEditCourseByIndex(idInt, *userID)
		if err != nil {
			log.Printf("Error checking course ownership: %v", err)
			return c.String(http.StatusInternalServerError, "Error verifying course ownership")
		}

		if !canEdit {
			return c.String(http.StatusForbidden, "You don't have permission to edit this course")
		}

		courseDB = dbCourse
		log.Printf("✅ User %d authorized to update course at index %d (DB ID: %d)", *userID, idInt, courseDB.ID)
	}

	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	course, err := h.parseFormToCourse(c, idInt)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Update in memory array
	(*h.courses)[idInt] = course

	// Update in database with ownership tracking if available
	if DB != nil && courseDB != nil {
		dbService := NewDatabaseService()
		if err := dbService.UpdateCourseWithOwnership(courseDB, course, *userID); err != nil {
			log.Printf("Failed to update course in database: %v", err)
			return c.String(http.StatusInternalServerError, "Failed to update course in database: "+err.Error())
		}
	}

	// Also update via course service for backward compatibility
	if err := h.courseService.UpdateCourse(course); err != nil {
		log.Printf("Warning: failed to update via course service: %v", err)
	}

	return h.renderSuccessMessage(c, "Course Updated Successfully!", "has been updated and saved", course.Name)
}

func (h *Handlers) CreateCourse(c echo.Context) error {
	log.Printf("[CREATE_COURSE] Starting request from %s", c.RealIP())

	// Get authenticated user ID
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		log.Printf("[CREATE_COURSE] ERROR: User not authenticated")
		return c.String(http.StatusUnauthorized, "You must be logged in to create a course")
	}

	log.Printf("[CREATE_COURSE] User ID %d creating course", *userID)

	if err := c.Request().ParseForm(); err != nil {
		log.Printf("[CREATE_COURSE] ERROR: Failed to parse form: %v", err)
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	course, err := h.parseFormToCourse(c, 0)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Save course with user ownership
	if err := h.courseService.SaveCourseWithOwner(course, userID); err != nil {
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

func (h *Handlers) UpdateHandicap(c echo.Context) error {
	sessionService := NewSessionService()
	dbUserID := sessionService.GetDatabaseUserID(c)

	if dbUserID == nil {
		return c.String(http.StatusUnauthorized, "User not authenticated with database")
	}

	if DB == nil {
		return c.String(http.StatusServiceUnavailable, "Database not available")
	}

	// Parse handicap from form
	handicapStr := c.FormValue("handicap")
	if handicapStr == "" {
		return c.String(http.StatusBadRequest, "Handicap value required")
	}

	handicap, err := strconv.ParseFloat(handicapStr, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid handicap value")
	}

	if handicap < 0 || handicap > 54 {
		return c.String(http.StatusBadRequest, "Handicap must be between 0 and 54")
	}

	// Update handicap in database
	dbService := NewDatabaseService()
	if err := dbService.UpdateUserHandicap(*dbUserID, handicap); err != nil {
		log.Printf("Failed to update handicap for user %d: %v", *dbUserID, err)
		return c.String(http.StatusInternalServerError, "Failed to update handicap")
	}

	log.Printf("✅ Updated handicap to %.1f for user ID %d", handicap, *dbUserID)

	// Return success response
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<div style="color: #204606; padding: 10px; text-align: center; font-weight: bold;">
			Handicap updated to %.1f!
		</div>
	`, handicap))
}

func (h *Handlers) UpdateDisplayName(c echo.Context) error {
	sessionService := NewSessionService()
	dbUserID := sessionService.GetDatabaseUserID(c)

	if dbUserID == nil {
		return c.String(http.StatusUnauthorized, "User not authenticated with database")
	}

	if DB == nil {
		return c.String(http.StatusServiceUnavailable, "Database not available")
	}

	// Parse display name from form
	displayName := c.FormValue("display_name")
	// Allow empty display name to clear it

	// Update display name in database
	dbService := NewDatabaseService()
	if err := dbService.UpdateUserDisplayName(*dbUserID, displayName); err != nil {
		log.Printf("Failed to update display name for user %d: %v", *dbUserID, err)
		return c.String(http.StatusInternalServerError, "Failed to update display name")
	}

	log.Printf("✅ Updated display name to '%s' for user ID %d", displayName, *dbUserID)

	// Return success response
	if displayName == "" {
		return c.HTML(http.StatusOK, `
			<div style="color: #204606; padding: 10px; text-align: center; font-weight: bold;">
				Display name cleared!
			</div>
		`)
	} else {
		return c.HTML(http.StatusOK, fmt.Sprintf(`
			<div style="color: #204606; padding: 10px; text-align: center; font-weight: bold;">
				Display name updated to %s!
			</div>
		`, displayName))
	}
}

// Helper methods

func (h *Handlers) CanEditCourse(courseIndex int, userID *uint) bool {
	if userID == nil || DB == nil {
		return false
	}

	dbService := NewDatabaseService()
	canEdit, _, err := dbService.CanEditCourseByIndex(courseIndex, *userID)
	if err != nil {
		log.Printf("Error checking course edit permission: %v", err)
		return false
	}

	return canEdit
}

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

func (h *Handlers) DatabaseStatus(c echo.Context) error {
	status := map[string]interface{}{
		"database_connected": false,
		"message":            "Database not available",
	}

	if GetDB() != nil {
		dbService := NewDatabaseService()
		stats, err := dbService.GetDatabaseStats()
		if err != nil {
			status["message"] = fmt.Sprintf("Database error: %v", err)
		} else {
			status["database_connected"] = true
			status["message"] = "Database connected successfully"
			status["stats"] = stats
		}
	}

	return c.JSON(http.StatusOK, status)
}

func (h *Handlers) MigrateCourses(c echo.Context) error {
	if GetDB() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Database not available",
		})
	}

	dbService := NewDatabaseService()

	// Load courses from JSON files
	courses, err := h.courseService.LoadCoursesFromJSON()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to load courses from JSON: %v", err),
		})
	}

	log.Printf("🔄 Starting migration of %d courses to database...", len(courses))

	// Migrate courses to database
	if err := dbService.MigrateJSONFilesToDatabase(courses); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Migration failed: %v", err),
		})
	}

	// Get updated stats
	stats, err := dbService.GetDatabaseStats()
	if err != nil {
		log.Printf("Warning: failed to get stats after migration: %v", err)
		stats = map[string]int{"courses": len(courses)}
	}

	// Reload courses in memory from database
	if err := h.reloadCourses(); err != nil {
		log.Printf("Warning: failed to reload courses after migration: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":          "Migration completed successfully",
		"migrated_courses": len(courses),
		"database_stats":   stats,
	})
}
