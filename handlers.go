package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	// Database-only handlers - no JSON dependencies
}

func NewHandlers() *Handlers {
	return &Handlers{}
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *Handlers) Home(c echo.Context) error {
	sessionService := NewSessionService()
	user := sessionService.GetUser(c)

	// Get user ID from middleware context if available
	var userID *uint
	if uid, ok := c.Get("userID").(uint); ok {
		userID = &uid
	}

	// Get courses from database only
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	log.Printf("✅ Home handler: Using %d courses from database", len(allCourses))

	// Debug: Check first few courses for coordinates
	for i := 0; i < min(3, len(allCourses)); i++ {
		course := allCourses[i]
		if course.Latitude != nil && course.Longitude != nil {
			log.Printf("🔍 Home handler: Course %d '%s' has coordinates: lat=%f, lng=%f", i, course.Name, *course.Latitude, *course.Longitude)
		} else {
			log.Printf("⚠️ Home handler: Course %d '%s' missing coordinates", i, course.Name)
		}
	}

	// Default to showing user's courses if logged in, all courses if not
	var coursesToShow []Course
	editPermissions := make(map[int]bool)
	allCoursesEditPermissions := make(map[int]bool) // Edit permissions for all courses

	if userID != nil {
		// Get courses the user has reviewed using the new review system
		reviewService := NewReviewService()
		userReviews, err := reviewService.GetUserReviews(*userID)
		if err != nil {
			log.Printf("Warning: failed to get user reviews: %v", err)
			// Return error if user reviews can't be loaded
			return c.String(http.StatusInternalServerError, "Failed to load user reviews")
		} else {
			log.Printf("✅ Found %d reviews for user %d in Home handler", len(userReviews), *userID)
			// Debug: Print review details
			for i, review := range userReviews {
				log.Printf("   Review %d: %s", i+1, review.CourseName)
			}

			// Get all courses owned by this user for edit permissions
			userOwnedCourses, err := dbService.GetCoursesByUser(*userID)
			userOwnedCourseNames := make(map[string]bool)
			if err == nil {
				for _, course := range userOwnedCourses {
					userOwnedCourseNames[course.Name] = true
				}
			}

			// Build edit permissions for ALL courses (for frontend filtering)
			for i, course := range allCourses {
				if userOwnedCourseNames[course.Name] {
					allCoursesEditPermissions[i] = true
				}
			}

			// Convert each review to a Course struct that the template expects
			for _, reviewWithCourse := range userReviews {
				// Find the corresponding course in the all courses array to get the correct index and coordinates
				var courseArrayIndex int = -1
				var baseCourse Course
				for idx, course := range allCourses {
					if course.Name == reviewWithCourse.CourseName {
						courseArrayIndex = idx
						baseCourse = course
						break
					}
				}

				// If we can't find the course in the all courses array, skip it
				if courseArrayIndex == -1 {
					log.Printf("Warning: Course '%s' from review not found in all courses array", reviewWithCourse.CourseName)
					continue
				}

				course := Course{
					ID:            courseArrayIndex, // Use the all courses array index for compatibility
					Name:          reviewWithCourse.CourseName,
					Description:   baseCourse.Description, // Use the actual course description
					OverallRating: safeStringValue(reviewWithCourse.OverallRating),
					Address:       reviewWithCourse.CourseAddress,
					Latitude:      baseCourse.Latitude,  // Include coordinates from database
					Longitude:     baseCourse.Longitude, // Include coordinates from database
					Ranks: Ranking{
						Price:              safeStringValue(reviewWithCourse.Price),
						HandicapDifficulty: safeIntValue(reviewWithCourse.HandicapDifficulty),
						HazardDifficulty:   safeIntValue(reviewWithCourse.HazardDifficulty),
						Merch:              safeStringValue(reviewWithCourse.Merch),
						Condition:          safeStringValue(reviewWithCourse.Condition),
						EnjoymentRating:    safeStringValue(reviewWithCourse.EnjoymentRating),
						Vibe:               safeStringValue(reviewWithCourse.Vibe),
						Range:              safeStringValue(reviewWithCourse.RangeRating),
						Amenities:          safeStringValue(reviewWithCourse.Amenities),
						Glizzies:           safeStringValue(reviewWithCourse.Glizzies),
					},
				}

				// Add review text if available
				if reviewWithCourse.ReviewText != nil {
					course.Review = *reviewWithCourse.ReviewText
				}

				coursesToShow = append(coursesToShow, course)

				// Check if user owns this course (for edit permissions)
				editPermissions[len(coursesToShow)-1] = userOwnedCourseNames[course.Name]
			}

			// If user has no reviewed courses, show all courses instead
			if len(coursesToShow) == 0 {
				coursesToShow = allCourses
				editPermissions = allCoursesEditPermissions // Use the all courses edit permissions
			}
		}
	} else {
		// Not logged in, show all courses
		coursesToShow = allCourses
	}

	// Debug: Log what we're sending to the template
	log.Printf("🎯 Home handler sending to template:")
	log.Printf("   - Courses to show: %d", len(coursesToShow))
	log.Printf("   - All courses: %d", len(allCourses))
	log.Printf("   - User logged in: %t", userID != nil)
	log.Printf("   - Default filter: %s", func() string {
		if userID != nil {
			return "my"
		}
		return "all"
	}())

	// Course list prepared for display

	data := struct {
		Courses                   []Course
		AllCourses                []Course // Include all courses for frontend filtering
		MapboxToken               string
		User                      *GoogleUser
		EditPermissions           map[int]bool
		AllCoursesEditPermissions map[int]bool // Edit permissions for all courses
		AllCoursesReviewStatus    map[int]bool // NEW: Track which courses have been reviewed
		DefaultFilter             string       // Add default filter indication
	}{
		Courses:                   coursesToShow,
		AllCourses:                allCourses, // Use courses with coordinates
		MapboxToken:               os.Getenv("MAPBOX_ACCESS_TOKEN"),
		User:                      user,
		EditPermissions:           editPermissions,
		AllCoursesEditPermissions: allCoursesEditPermissions,
		AllCoursesReviewStatus:    make(map[int]bool), // Will be populated below
		DefaultFilter: func() string {
			if userID != nil {
				return "my"
			}
			return "all"
		}(),
	}

	// Populate review status for all courses
	if userID != nil {
		reviewService := NewReviewService()
		userReviews, err := reviewService.GetUserReviews(*userID)
		if err == nil {
			// Create a map of reviewed course IDs
			reviewedCourseIDs := make(map[uint]bool)
			for _, review := range userReviews {
				reviewedCourseIDs[review.CourseID] = true
			}

			// Get all database courses with their IDs to avoid N+1 queries
			dbCourses, err := dbService.GetAllCourses()
			if err == nil {
				// Create a map from course name+address to database ID
				courseToIDMap := make(map[string]uint)
				for _, dbCourse := range dbCourses {
					key := dbCourse.Name + "|" + dbCourse.Address
					courseToIDMap[key] = dbCourse.ID
				}

				// Mark courses as reviewed using the pre-loaded course ID map
				for i, course := range allCourses {
					key := course.Name + "|" + course.Address
					if courseID, exists := courseToIDMap[key]; exists {
						if reviewedCourseIDs[courseID] {
							data.AllCoursesReviewStatus[i] = true
						}
					}
				}
			}
		}
	}

	return c.Render(http.StatusOK, "welcome", data)
}

func (h *Handlers) Introduction(c echo.Context) error {
	dbService := NewDatabaseService()
	courses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	return c.Render(http.StatusOK, "introduction", PageData{
		Courses: courses,
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

	// Get user ID from middleware context if available
	var dbUserID *uint
	if uid, ok := c.Get("userID").(uint); ok {
		dbUserID = &uid
	}

	// Get user's handicap from database if available
	var handicap *float64
	var displayName *string
	log.Printf("🔍 Profile request for user: %s, DB User ID: %v, DB available: %t",
		user.Email, dbUserID, DB != nil)

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
			dbUserID = &dbUser.ID
		}
	} else {
		log.Printf("⚠️ User not found in database")
	}

	// Get courses the user has reviewed using the new review system
	var userCourses []Course
	editPermissions := make(map[int]bool)

	if dbUserID != nil {
		// Use the review service to get user's reviews
		reviewService := NewReviewService()
		userReviews, err := reviewService.GetUserReviews(*dbUserID)
		if err != nil {
			log.Printf("Warning: failed to get user reviews: %v", err)
		} else {
			log.Printf("✅ Found %d reviews for user %d", len(userReviews), *dbUserID)

			// Get all courses from database for lookups
			allCourses, err := dbService.GetAllCoursesFromDatabase()
			if err != nil {
				log.Printf("Warning: failed to get all courses: %v", err)
				allCourses = []Course{}
			}

			// Convert each review to a Course struct that the template expects
			for _, reviewWithCourse := range userReviews {
				// Find the corresponding course in the database array to get the correct index
				var courseArrayIndex int = -1
				var originalCourse Course
				for idx, dbCourse := range allCourses {
					if dbCourse.Name == reviewWithCourse.CourseName {
						courseArrayIndex = idx
						originalCourse = dbCourse
						break
					}
				}

				// If we can't find the course in the database array, skip it
				if courseArrayIndex == -1 {
					log.Printf("Warning: Course '%s' from review not found in database array", reviewWithCourse.CourseName)
					continue
				}

				course := Course{
					ID:            courseArrayIndex, // Use the database array index for compatibility
					Name:          reviewWithCourse.CourseName,
					Description:   originalCourse.Description, // Use the actual course description
					OverallRating: safeStringValue(reviewWithCourse.OverallRating),
					Address:       reviewWithCourse.CourseAddress,
					Ranks: Ranking{
						Price:              safeStringValue(reviewWithCourse.Price),
						HandicapDifficulty: safeIntValue(reviewWithCourse.HandicapDifficulty),
						HazardDifficulty:   safeIntValue(reviewWithCourse.HazardDifficulty),
						Merch:              safeStringValue(reviewWithCourse.Merch),
						Condition:          safeStringValue(reviewWithCourse.Condition),
						EnjoymentRating:    safeStringValue(reviewWithCourse.EnjoymentRating),
						Vibe:               safeStringValue(reviewWithCourse.Vibe),
						Range:              safeStringValue(reviewWithCourse.RangeRating),
						Amenities:          safeStringValue(reviewWithCourse.Amenities),
						Glizzies:           safeStringValue(reviewWithCourse.Glizzies),
					},
				}

				// Add review text if available
				if reviewWithCourse.ReviewText != nil {
					course.Review = *reviewWithCourse.ReviewText
				}

				userCourses = append(userCourses, course)

				// Users can always edit their own reviews, but we need to check if they own the course
				isOwner, err := dbService.IsUserCourseOwner(*dbUserID, course.Name)
				if err != nil {
					log.Printf("Warning: failed to check course ownership: %v", err)
					editPermissions[len(userCourses)-1] = false
				} else {
					editPermissions[len(userCourses)-1] = isOwner
				}
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
	log.Printf("📊 Rendering profile with %d reviewed courses", len(userCourses))
	return c.Render(http.StatusOK, "user-profile", data)
}

func (h *Handlers) GetCourse(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if idInt >= len(allCourses) {
		return c.String(http.StatusNotFound, "Course not found")
	}

	// Get the base course data from the database array
	baseCourse := allCourses[idInt]

	// Get user context from middleware if available
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)

	// DEBUG: Log user information for course access
	user := sessionService.GetUser(c)
	if user != nil {
		log.Printf("🔍 [GETCOURSE] User %s (DB ID: %v) accessing course '%s'", user.Email, userID, baseCourse.Name)
	} else {
		log.Printf("🔍 [GETCOURSE] Anonymous user accessing course '%s'", baseCourse.Name)
	}

	// Start with the base course data
	courseToDisplay := baseCourse
	var canEdit bool
	var hasUserReview bool

	// If user is logged in, get their specific review
	if userID != nil {
		// Check if they own this course (for edit permissions)
		isOwner, err := dbService.IsUserCourseOwner(*userID, baseCourse.Name)
		if err != nil {
			log.Printf("Warning: failed to check course ownership: %v", err)
		} else {
			canEdit = isOwner
		}

		// Get the user's review for this course
		reviewService := NewReviewService()

		// First, find the database course ID by name and address
		dbCourse, err := dbService.GetCourseByNameAndAddress(baseCourse.Name, baseCourse.Address)
		if err == nil && dbCourse != nil {
			log.Printf("🔍 [GETCOURSE] Looking for review by user %d for course %d (%s)", *userID, dbCourse.ID, baseCourse.Name)
			userReview, err := reviewService.GetUserReviewForCourse(*userID, dbCourse.ID)
			if err != nil {
				log.Printf("Warning: failed to get user review: %v", err)
			} else if userReview != nil {
				log.Printf("✅ [GETCOURSE] Found user %d's review for course %d - Rating: %s", *userID, dbCourse.ID, safeStringValue(userReview.OverallRating))
			} else {
				log.Printf("ℹ️ [GETCOURSE] No review found for user %d on course %d", *userID, dbCourse.ID)
			}

			if userReview != nil {
				// User has a review for this course - use their review data
				hasUserReview = true

				// Get user's holes and scores for this course
				userHoles, err := reviewService.GetUserHolesForCourse(*userID, dbCourse.ID)
				if err != nil {
					log.Printf("Warning: failed to get user holes: %v", err)
				}

				userScores, err := reviewService.GetUserScoresForCourse(*userID, dbCourse.ID)
				if err != nil {
					log.Printf("Warning: failed to get user scores: %v", err)
				}

				// Convert database holes to Course.Holes format
				var holes []Hole
				for _, dbHole := range userHoles {
					hole := Hole{
						Number: dbHole.Number,
					}
					if dbHole.Par != nil {
						hole.Par = *dbHole.Par
					}
					if dbHole.Yardage != nil {
						hole.Yardage = *dbHole.Yardage
					}
					if dbHole.Description != nil {
						hole.Description = *dbHole.Description
					}
					holes = append(holes, hole)
				}

				// Convert database scores to Course.Scores format
				var scores []Score
				for _, dbScore := range userScores {
					score := Score{
						Score: dbScore.Score,
					}
					if dbScore.Handicap != nil {
						score.Handicap = *dbScore.Handicap
					}
					scores = append(scores, score)
				}

				courseToDisplay = Course{
					ID:            baseCourse.ID, // Keep the array index for routing
					Name:          baseCourse.Name,
					Description:   baseCourse.Description,
					Address:       baseCourse.Address,
					OverallRating: safeStringValue(userReview.OverallRating),
					Ranks: Ranking{
						Price:              safeStringValue(userReview.Price),
						HandicapDifficulty: safeIntValue(userReview.HandicapDifficulty),
						HazardDifficulty:   safeIntValue(userReview.HazardDifficulty),
						Merch:              safeStringValue(userReview.Merch),
						Condition:          safeStringValue(userReview.Condition),
						EnjoymentRating:    safeStringValue(userReview.EnjoymentRating),
						Vibe:               safeStringValue(userReview.Vibe),
						Range:              safeStringValue(userReview.RangeRating),
						Amenities:          safeStringValue(userReview.Amenities),
						Glizzies:           safeStringValue(userReview.Glizzies),
					},
					Holes:  holes,  // Use user's saved holes
					Scores: scores, // Use user's saved scores
				}

				// Add the user's review text if available
				if userReview.ReviewText != nil {
					courseToDisplay.Review = *userReview.ReviewText
				}

				log.Printf("✅ Displaying user %d's review for course '%s'", *userID, baseCourse.Name)
			} else {
				log.Printf("ℹ️  User %d has no review for course '%s', showing base course data", *userID, baseCourse.Name)
			}
		}
	}

	// Add context to course data
	courseData := struct {
		Course
		CanEdit       bool
		HasUserReview bool
		IsLoggedIn    bool
	}{
		Course:        courseToDisplay,
		CanEdit:       canEdit,
		HasUserReview: hasUserReview,
		IsLoggedIn:    userID != nil,
	}

	return c.Render(http.StatusOK, "course", courseData)
}

func (h *Handlers) CreateCourseForm(c echo.Context) error {
	// Fast loading - just render the page shell, courses will be loaded via API
	data := struct {
		IsEdit           bool
		IsReviewMode     bool
		UsePagination    bool
	}{
		IsEdit:           false,
		IsReviewMode:     true,
		UsePagination:    true, // Signal to use pagination
	}

	return c.Render(http.StatusOK, "review-landing", data)
}

func (h *Handlers) EditCourseForm(c echo.Context) error {
	// Get course index from middleware context (already validated)
	courseIndex, ok := c.Get("courseIndex").(int)
	if !ok {
		return c.String(http.StatusInternalServerError, "Course index not found in context")
	}

	// Get user ID from middleware context (already validated)
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return c.String(http.StatusInternalServerError, "User ID not found in context")
	}

	// Ownership already verified by middleware
	log.Printf("✅ User %d authorized to edit course at index %d", userID, courseIndex)

	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if courseIndex >= len(allCourses) {
		return c.String(http.StatusNotFound, "Course not found")
	}

	course := allCourses[courseIndex]

	data := struct {
		Course       Course
		Courses      []Course
		IsEdit       bool
		IsReviewMode bool
	}{
		Course:       course,
		Courses:      allCourses,
		IsEdit:       true,
		IsReviewMode: false,
	}

	return c.Render(http.StatusOK, "review-landing", data)
}

func (h *Handlers) UpdateCourse(c echo.Context) error {
	// Get course index from middleware context (already validated)
	courseIndex, ok := c.Get("courseIndex").(int)
	if !ok {
		return c.String(http.StatusInternalServerError, "Course index not found in context")
	}

	// Get user ID from middleware context (already validated)
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return c.String(http.StatusInternalServerError, "User ID not found in context")
	}

	// Get all courses from database to find the course to update
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if courseIndex >= len(allCourses) {
		return c.String(http.StatusNotFound, "Course not found")
	}

	// Get course from database by name for ownership verification
	courseName := allCourses[courseIndex].Name
	dbCourse, err := dbService.GetCourseWithOwnershipByName(courseName)
	if err != nil {
		log.Printf("Error getting course from database: %v", err)
		return c.String(http.StatusInternalServerError, "Error accessing course data")
	}

	if dbCourse == nil {
		return c.String(http.StatusNotFound, "Course not found in database")
	}

	log.Printf("✅ User %d authorized to update course at index %d (DB ID: %d)", userID, courseIndex, dbCourse.ID)

	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	course, err := h.parseFormToCourse(c, courseIndex)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Update in database with ownership tracking
	if err := dbService.UpdateCourseWithOwnership(dbCourse, course, userID); err != nil {
		log.Printf("Failed to update course in database: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to update course in database: "+err.Error())
	}

	return h.renderSuccessMessage(c, "Course Updated Successfully!", "has been updated and saved", course.Name)
}

func (h *Handlers) DeleteCourse(c echo.Context) error {
	// Get course index from middleware context (already validated)
	courseIndex, ok := c.Get("courseIndex").(int)
	if !ok {
		return c.String(http.StatusInternalServerError, "Course index not found in context")
	}

	// Get user ID from middleware context (already validated)
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return c.String(http.StatusInternalServerError, "User ID not found in context")
	}

	// Get all courses from database to find the course to delete
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if courseIndex >= len(allCourses) {
		return c.String(http.StatusNotFound, "Course not found")
	}

	courseName := allCourses[courseIndex].Name

	// Get course from database by name for deletion
	dbCourse, err := dbService.GetCourseWithOwnershipByName(courseName)
	if err != nil {
		log.Printf("Error getting course from database: %v", err)
		return c.String(http.StatusInternalServerError, "Error accessing course data")
	}

	if dbCourse == nil {
		return c.String(http.StatusNotFound, "Course not found in database")
	}

	if err := dbService.DeleteCourse(dbCourse.ID); err != nil {
		log.Printf("Failed to delete course from database: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to delete course from database: "+err.Error())
	}

	log.Printf("✅ User %d deleted course '%s' (DB ID: %d)", userID, courseName, dbCourse.ID)

	return h.renderSuccessMessage(c, "Course Deleted Successfully!", "has been deleted", courseName)
}

func (h *Handlers) DeleteReview(c echo.Context) error {
	log.Printf("[DELETE_REVIEW] Starting request from %s", c.RealIP())

	// Get authenticated user ID
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		log.Printf("[DELETE_REVIEW] ERROR: User not authenticated")
		return c.String(http.StatusUnauthorized, "You must be logged in to delete a review")
	}

	// Get course index from URL parameter (this is the database array index)
	courseIndexParam := c.Param("id")
	courseIndex, err := strconv.Atoi(courseIndexParam)
	if err != nil {
		log.Printf("[DELETE_REVIEW] ERROR: Invalid course index: %s", courseIndexParam)
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if courseIndex >= len(allCourses) {
		log.Printf("[DELETE_REVIEW] ERROR: Course index %d out of range", courseIndex)
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get the course from the database array
	course := allCourses[courseIndex]
	courseName := course.Name
	courseAddress := course.Address

	// Find the database course by name and address
	dbCourse, err := dbService.GetCourseByNameAndAddress(courseName, courseAddress)
	if err != nil {
		log.Printf("[DELETE_REVIEW] ERROR: Course not found in database: %v", err)
		return c.String(http.StatusNotFound, "Course not found")
	}

	// Verify the user has a review for this course
	reviewService := NewReviewService()
	existingReview, err := reviewService.GetUserReviewForCourse(*userID, dbCourse.ID)
	if err != nil {
		log.Printf("[DELETE_REVIEW] ERROR: Failed to check existing review: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to check existing review")
	}

	if existingReview == nil {
		log.Printf("[DELETE_REVIEW] ERROR: User %d has no review for course %d", *userID, dbCourse.ID)
		return c.String(http.StatusNotFound, "You have no review for this course")
	}

	// Delete the review (this will NOT delete the course, only the user's review)
	err = reviewService.DeleteUserReview(*userID, dbCourse.ID)
	if err != nil {
		log.Printf("[DELETE_REVIEW] ERROR: Failed to delete review: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to delete review: "+err.Error())
	}

	log.Printf("[DELETE_REVIEW] ✅ Review deleted successfully for user %d, course %d (%s)", *userID, dbCourse.ID, courseName)

	// Return success message
	return h.renderSuccessMessage(c, "Review Deleted Successfully!", "review has been deleted", courseName)
}

func (h *Handlers) CreateCourse(c echo.Context) error {
	log.Printf("[REVIEW_COURSE] Starting request from %s", c.RealIP())

	// Get authenticated user ID
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		log.Printf("[REVIEW_COURSE] ERROR: User not authenticated")
		return c.String(http.StatusUnauthorized, "You must be logged in to review a course")
	}

	log.Printf("[REVIEW_COURSE] User ID %d reviewing course", *userID)

	if err := c.Request().ParseForm(); err != nil {
		log.Printf("[REVIEW_COURSE] ERROR: Failed to parse form: %v", err)
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	// Get selected course ID from form - this is now the database course ID
	selectedCourseID := c.FormValue("selectedCourseId")
	if selectedCourseID == "" {
		return c.String(http.StatusBadRequest, "No course selected")
	}

	// Convert to database course ID
	courseID, err := strconv.ParseUint(selectedCourseID, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Validate that the course exists in the database
	if DB == nil {
		return c.String(http.StatusServiceUnavailable, "Database not available")
	}

	dbService := NewDatabaseService()
	dbCourse, err := dbService.GetCourseByID(uint(courseID))
	if err != nil {
		log.Printf("[REVIEW_COURSE] ERROR: Course not found in database: %v", err)
		return c.String(http.StatusNotFound, "Course not found")
	}

	// Parse review form data using the new service
	reviewService := NewReviewService()
	formData := ParseReviewFormData(func(key string) string {
		return c.FormValue(key)
	})

	// Create or update the review
	_, err = reviewService.CreateOrUpdateReview(*userID, dbCourse.ID, formData)
	if err != nil {
		log.Printf("[REVIEW_COURSE] ERROR: Failed to save review: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to save review: "+err.Error())
	}

	log.Printf("[REVIEW_COURSE] ✅ Review saved successfully for user %d, course %d", *userID, dbCourse.ID)

	// Also save any score data if provided
	scoreFormData := ParseScoreFormData(func(key string) string {
		return c.FormValue(key)
	})

	if len(scoreFormData) > 0 {
		err := reviewService.AddScores(*userID, dbCourse.ID, scoreFormData)
		if err != nil {
			log.Printf("[REVIEW_COURSE] Warning: Failed to save scores: %v", err)
		} else {
			log.Printf("[REVIEW_COURSE] ✅ %d scores saved for user %d, course %d", len(scoreFormData), *userID, dbCourse.ID)
		}
	}

	// Also save any hole data if provided
	holeFormData := ParseHoleFormData(func(key string) string {
		return c.FormValue(key)
	})

	if len(holeFormData) > 0 {
		err := reviewService.AddHoles(*userID, dbCourse.ID, holeFormData)
		if err != nil {
			log.Printf("[REVIEW_COURSE] Warning: Failed to save holes: %v", err)
		} else {
			log.Printf("[REVIEW_COURSE] ✅ %d holes saved for user %d, course %d", len(holeFormData), *userID, dbCourse.ID)
		}
	}

	return h.renderSuccessMessage(c, "Course Review Created Successfully!", "review has been created and saved", dbCourse.Name)
}

// AddScore handles adding a single score from the profile page
func (h *Handlers) AddScore(c echo.Context) error {
	log.Printf("[ADD_SCORE] Starting request from %s", c.RealIP())

	// Get authenticated user ID
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		log.Printf("[ADD_SCORE] ERROR: User not authenticated")
		return c.String(http.StatusUnauthorized, "You must be logged in to add a score")
	}

	if err := c.Request().ParseForm(); err != nil {
		log.Printf("[ADD_SCORE] ERROR: Failed to parse form: %v", err)
		return c.String(http.StatusBadRequest, "Failed to parse form data: "+err.Error())
	}

	// Get course ID from form
	courseIDStr := c.FormValue("courseId")
	if courseIDStr == "" {
		return c.String(http.StatusBadRequest, "No course ID provided")
	}

	// Convert course ID - this is the database array index
	courseIndex, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if courseIndex >= len(allCourses) {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get the course from the database array
	course := allCourses[courseIndex]
	courseName := course.Name
	courseAddress := course.Address

	// Find the database course by name and address
	dbCourse, err := dbService.GetCourseByNameAndAddress(courseName, courseAddress)
	if err != nil {
		log.Printf("[ADD_SCORE] ERROR: Course not found in database: %v", err)
		return c.String(http.StatusNotFound, "Course not found")
	}

	// Parse and validate score data
	validator := NewValidator()
	scoreResult, validationErrors := validator.ValidateScore(
		c.FormValue("totalScore"),
		c.FormValue("outScore"),
		c.FormValue("inScore"),
		c.FormValue("handicap"),
	)
	
	if len(validationErrors) > 0 {
		return c.String(http.StatusBadRequest, validationErrors.Error())
	}

	// Create score data
	scoreData := ScoreFormData{
		Score:    scoreResult.TotalScore,
		Handicap: scoreResult.Handicap,
		OutScore: scoreResult.OutScore,
		InScore:  scoreResult.InScore,
	}

	// Save the score
	reviewService := NewReviewService()
	_, err = reviewService.AddScore(*userID, dbCourse.ID, scoreData)
	if err != nil {
		log.Printf("[ADD_SCORE] ERROR: Failed to save score: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to save score: "+err.Error())
	}

	log.Printf("[ADD_SCORE] ✅ Score %d saved for user %d, course %d", scoreResult.TotalScore, *userID, dbCourse.ID)

	// Return success response
	return c.String(http.StatusOK, "Score added successfully!")
}


func (h *Handlers) Map(c echo.Context) error {
	// Get user information for ownership context
	sessionService := NewSessionService()
	user := sessionService.GetUser(c)

	// Get user ID from middleware context if available
	var userID *uint
	if uid, ok := c.Get("userID").(uint); ok {
		userID = &uid
	}

	// Get courses from database only
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}
	
	log.Printf("🔍 Map handler: Using %d courses from database", len(allCourses))

	// Default to showing user's courses if logged in, all courses if not
	var coursesToShow []Course
	editPermissions := make(map[int]bool)
	allCoursesEditPermissions := make(map[int]bool) // Edit permissions for all courses

	if userID != nil {
		// Get courses the user has reviewed using the new review system
		reviewService := NewReviewService()
		userReviews, err := reviewService.GetUserReviews(*userID)
		if err != nil {
			log.Printf("Warning: failed to get user reviews: %v", err)
			// Return error if user reviews can't be loaded
			return c.String(http.StatusInternalServerError, "Failed to load user reviews")
		} else {
			log.Printf("✅ Found %d reviews for user %d in Map handler", len(userReviews), *userID)

			// Get all courses owned by this user for edit permissions
			userOwnedCourses, err := dbService.GetCoursesByUser(*userID)
			userOwnedCourseNames := make(map[string]bool)
			if err == nil {
				for _, course := range userOwnedCourses {
					userOwnedCourseNames[course.Name] = true
				}
			}

			// Build edit permissions for ALL courses (for frontend filtering)
			for i, course := range allCourses {
				if userOwnedCourseNames[course.Name] {
					allCoursesEditPermissions[i] = true
				}
			}

			// Convert each review to a Course struct that the template expects
			for _, reviewWithCourse := range userReviews {
				// Find the corresponding course in the all courses array to get the correct index and coordinates
				var courseArrayIndex int = -1
				var baseCourse Course
				for idx, course := range allCourses {
					if course.Name == reviewWithCourse.CourseName {
						courseArrayIndex = idx
						baseCourse = course
						break
					}
				}

				// If we can't find the course in the all courses array, skip it
				if courseArrayIndex == -1 {
					log.Printf("Warning: Course '%s' from review not found in all courses array", reviewWithCourse.CourseName)
					continue
				}

				course := Course{
					ID:            courseArrayIndex, // Use the all courses array index for compatibility
					Name:          reviewWithCourse.CourseName,
					Description:   baseCourse.Description, // Use the actual course description
					OverallRating: safeStringValue(reviewWithCourse.OverallRating),
					Address:       reviewWithCourse.CourseAddress,
					Latitude:      baseCourse.Latitude,  // Include coordinates from database
					Longitude:     baseCourse.Longitude, // Include coordinates from database
					Ranks: Ranking{
						Price:              safeStringValue(reviewWithCourse.Price),
						HandicapDifficulty: safeIntValue(reviewWithCourse.HandicapDifficulty),
						HazardDifficulty:   safeIntValue(reviewWithCourse.HazardDifficulty),
						Merch:              safeStringValue(reviewWithCourse.Merch),
						Condition:          safeStringValue(reviewWithCourse.Condition),
						EnjoymentRating:    safeStringValue(reviewWithCourse.EnjoymentRating),
						Vibe:               safeStringValue(reviewWithCourse.Vibe),
						Range:              safeStringValue(reviewWithCourse.RangeRating),
						Amenities:          safeStringValue(reviewWithCourse.Amenities),
						Glizzies:           safeStringValue(reviewWithCourse.Glizzies),
					},
				}

				// Add review text if available
				if reviewWithCourse.ReviewText != nil {
					course.Review = *reviewWithCourse.ReviewText
				}

				coursesToShow = append(coursesToShow, course)

				// Check if user owns this course (for edit permissions)
				editPermissions[len(coursesToShow)-1] = userOwnedCourseNames[course.Name]
			}

			// If user has no reviewed courses, show all courses instead
			if len(coursesToShow) == 0 {
				coursesToShow = allCourses
				editPermissions = allCoursesEditPermissions // Use the all courses edit permissions
			}
		}
	} else {
		// Not logged in, show all courses
		coursesToShow = allCourses
	}

	coursesJSON, err := json.Marshal(coursesToShow)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to marshal courses to JSON: "+err.Error())
	}

	// Also include all courses JSON for frontend filtering
	allCoursesJSON, err := json.Marshal(allCourses)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to marshal all courses to JSON: "+err.Error())
	}

	data := struct {
		Courses                   []Course
		AllCourses                []Course
		CoursesJSON               template.JS
		AllCoursesJSON            template.JS
		MapboxToken               string
		VectorTileUrl             string
		User                      *GoogleUser
		EditPermissions           map[int]bool
		AllCoursesEditPermissions map[int]bool
		AllCoursesReviewStatus    map[int]bool
		DefaultFilter             string
	}{
		Courses:                   coursesToShow,
		AllCourses:                allCourses, // Use courses with coordinates
		CoursesJSON:               template.JS(coursesJSON),
		AllCoursesJSON:            template.JS(allCoursesJSON),
		MapboxToken:               os.Getenv("MAPBOX_ACCESS_TOKEN"),
		VectorTileUrl:             os.Getenv("VECTOR_TILE_URL"),
		User:                      user,
		EditPermissions:           editPermissions,
		AllCoursesEditPermissions: allCoursesEditPermissions,
		AllCoursesReviewStatus:    make(map[int]bool),
		DefaultFilter: func() string {
			if userID != nil {
				return "my"
			}
			return "all"
		}(),
	}

	// Populate review status for all courses
	if userID != nil {
		reviewService := NewReviewService()
		userReviews, err := reviewService.GetUserReviews(*userID)
		if err == nil {
			// Create a map of reviewed course IDs
			reviewedCourseIDs := make(map[uint]bool)
			for _, review := range userReviews {
				reviewedCourseIDs[review.CourseID] = true
			}

			// Get all database courses with their IDs to avoid N+1 queries
			dbCourses, err := dbService.GetAllCourses()
			if err == nil {
				// Create a map from course name+address to database ID
				courseToIDMap := make(map[string]uint)
				for _, dbCourse := range dbCourses {
					key := dbCourse.Name + "|" + dbCourse.Address
					courseToIDMap[key] = dbCourse.ID
				}

				// Mark courses as reviewed using the pre-loaded course ID map
				for i, course := range allCourses {
					key := course.Name + "|" + course.Address
					if courseID, exists := courseToIDMap[key]; exists {
						if reviewedCourseIDs[courseID] {
							data.AllCoursesReviewStatus[i] = true
						}
					}
				}
			}
		}
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

	// Parse and validate handicap from form
	handicapStr := c.FormValue("handicap")
	validator := NewValidator()
	
	handicap, validationErr := validator.ValidateHandicap(handicapStr)
	if validationErr != nil {
		return c.String(http.StatusBadRequest, validationErr.Message)
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

	// Parse and validate display name from form
	displayName := c.FormValue("display_name")
	validator := NewValidator()
	
	if validationErr := validator.ValidateDisplayName(displayName); validationErr != nil {
		return c.String(http.StatusBadRequest, validationErr.Message)
	}

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
	if userID == nil {
		return false
	}

	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil || courseIndex < 0 || courseIndex >= len(allCourses) {
		return false
	}

	courseName := allCourses[courseIndex].Name
	isOwner, err := dbService.IsUserCourseOwner(*userID, courseName)
	if err != nil {
		log.Printf("Error checking course edit permission: %v", err)
		return false
	}

	return isOwner
}

func (h *Handlers) parseFormToCourse(c echo.Context, existingID int) (Course, error) {
	validator := NewValidator()
	
	// Validate course data using the new validation system
	courseData, validationErrors := validator.ValidateCourseData(c)
	if len(validationErrors) > 0 {
		return Course{}, fmt.Errorf("validation failed: %s", validationErrors.Error())
	}

	course := Course{
		ID:            existingID,
		Name:          courseData.Name,
		Description:   courseData.Description,
		OverallRating: courseData.OverallRating,
		Review:        courseData.Review,
		Address:       courseData.Address,
		Ranks: Ranking{
			Price:              courseData.Price,
			HandicapDifficulty: courseData.HandicapDifficulty,
			HazardDifficulty:   courseData.HazardDifficulty,
			Merch:              courseData.Merch,
			Condition:          courseData.Condition,
			EnjoymentRating:    courseData.EnjoymentRating,
			Vibe:               courseData.Vibe,
			Range:              courseData.Range,
			Amenities:          courseData.Amenities,
			Glizzies:           courseData.Glizzies,
		},
		Holes:  []Hole{},
		Scores: []Score{},
	}

	courseService := NewCourseService()
	holes, scores, err := courseService.ParseFormData(c.Request().Form)
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

// reloadCourses method removed - no longer needed with database-only approach

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

	// JSON migration no longer supported - database is the single source of truth
	dbService := NewDatabaseService()
	stats, err := dbService.GetDatabaseStats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get database stats: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":        "Migration endpoint deprecated - system now uses database exclusively",
		"database_stats": stats,
	})
}


// Helper functions for safely converting nullable values to template-friendly formats
func safeStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func safeIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func (h *Handlers) ReviewSpecificCourseForm(c echo.Context) error {
	// Get course ID from URL parameter
	courseIDParam := c.Param("id")
	courseIndex, err := strconv.Atoi(courseIDParam)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load courses from database")
	}

	if courseIndex >= len(allCourses) {
		return c.String(http.StatusBadRequest, "Invalid course ID")
	}

	// Get the course from the database array
	course := allCourses[courseIndex]

	// Get authenticated user to verify they haven't already reviewed this course
	sessionService := NewSessionService()
	userID := sessionService.GetDatabaseUserID(c)
	if userID == nil {
		return c.String(http.StatusUnauthorized, "You must be logged in to review a course")
	}

	// Get the real database course ID if available
	var dbCourseID uint
	if DB != nil {
		dbService := NewDatabaseService()
		// Use name + address for unique identification
		dbCourse, err := dbService.GetCourseByNameAndAddress(course.Name, course.Address)
		if err == nil && dbCourse != nil {
			dbCourseID = dbCourse.ID
			log.Printf("✅ [REVIEW_COURSE] Found unique course: '%s' at '%s' with ID %d", course.Name, course.Address, dbCourseID)
		} else {
			// Fallback to a computed ID if course not in database
			dbCourseID = uint(courseIndex + 1)
			log.Printf("⚠️ [REVIEW_COURSE] Course '%s' at '%s' not found in database, using fallback ID %d", course.Name, course.Address, dbCourseID)
		}
	} else {
		dbCourseID = uint(courseIndex + 1)
	}

	// Convert the JSON course to a CourseDB format for the template
	courseDB := CourseDB{
		ID:      dbCourseID,
		Name:    course.Name,
		Address: course.Address,
	}

	// Get user's existing review, scores, and holes for this course
	var userReview *CourseReview
	var userScores []UserCourseScore
	var userHoles []UserCourseHole

	if DB != nil {
		dbService := NewDatabaseService()
		dbCourse, err := dbService.GetCourseByNameAndAddress(course.Name, course.Address)
		if err == nil && dbCourse != nil {
			reviewService := NewReviewService()

			// Get existing review
			userReview, err = reviewService.GetUserReviewForCourse(*userID, dbCourse.ID)
			if err != nil {
				log.Printf("Warning: failed to get user review: %v", err)
			}

			// Get user's scores
			userScores, err = reviewService.GetUserScoresForCourse(*userID, dbCourse.ID)
			if err != nil {
				log.Printf("Warning: failed to get user scores: %v", err)
			}

			// Get user's holes
			userHoles, err = reviewService.GetUserHolesForCourse(*userID, dbCourse.ID)
			if err != nil {
				log.Printf("Warning: failed to get user holes: %v", err)
			}
		}
	}

	// Prepare data for review-course template
	// Convert UserReview to template-friendly format if it exists
	type TemplateReview struct {
		OverallRating      string
		Price              string
		HandicapDifficulty int
		HazardDifficulty   int
		Merch              string
		Condition          string
		EnjoymentRating    string
		Vibe               string
		RangeRating        string
		Amenities          string
		Glizzies           string
		Walkability        string
		ReviewText         string
	}

	var templateReview *TemplateReview
	if userReview != nil {
		templateReview = &TemplateReview{
			OverallRating:      safeStringValue(userReview.OverallRating),
			Price:              safeStringValue(userReview.Price),
			HandicapDifficulty: safeIntValue(userReview.HandicapDifficulty),
			HazardDifficulty:   safeIntValue(userReview.HazardDifficulty),
			Merch:              safeStringValue(userReview.Merch),
			Condition:          safeStringValue(userReview.Condition),
			EnjoymentRating:    safeStringValue(userReview.EnjoymentRating),
			Vibe:               safeStringValue(userReview.Vibe),
			RangeRating:        safeStringValue(userReview.RangeRating),
			Amenities:          safeStringValue(userReview.Amenities),
			Glizzies:           safeStringValue(userReview.Glizzies),
			Walkability:        safeStringValue(userReview.Walkability),
			ReviewText:         safeStringValue(userReview.ReviewText),
		}
	}

	data := struct {
		Course     *CourseDB
		UserReview *TemplateReview
		UserScores []UserCourseScore
		UserHoles  []UserCourseHole
	}{
		Course:     &courseDB,
		UserReview: templateReview,
		UserScores: userScores,
		UserHoles:  userHoles,
	}

	return c.Render(http.StatusOK, "review-course", data)
}

// GetAllCoursesAPI returns paginated courses for ultra-fast sidebar loading
func (h *Handlers) GetAllCoursesAPI(c echo.Context) error {
	// Parse pagination parameters
	page := 1
	limit := 20
	search := ""
	filter := "all"
	
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	if s := c.QueryParam("search"); s != "" {
		search = s
	}
	
	if f := c.QueryParam("filter"); f != "" {
		filter = f
	}
	
	log.Printf("🚀 GetAllCoursesAPI: page=%d, limit=%d, search='%s', filter='%s'", page, limit, search, filter)
	
	// Get ownership context that was added by middleware
	editPermissions := c.Get("EditPermissions").([]bool)
	reviewStatus := c.Get("ReviewStatus").([]bool)
	
	// Get all courses from database
	dbService := NewDatabaseService()
	allCourses, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to load courses from database",
		})
	}

	// Filter courses based on search and filter criteria
	var filteredCourses []Course
	var filteredEditPermissions []bool
	var filteredReviewStatus []bool
	
	for i, course := range allCourses {
		// Apply filter criteria
		matchesFilter := true
		if filter == "my" && i < len(reviewStatus) {
			matchesFilter = reviewStatus[i]
		}
		
		// Apply search criteria
		matchesSearch := true
		if search != "" {
			matchesSearch = false
			if course.Name != "" && strings.Contains(strings.ToLower(course.Name), strings.ToLower(search)) {
				matchesSearch = true
			}
		}
		
		if matchesFilter && matchesSearch {
			filteredCourses = append(filteredCourses, course)
			if i < len(editPermissions) {
				filteredEditPermissions = append(filteredEditPermissions, editPermissions[i])
			} else {
				filteredEditPermissions = append(filteredEditPermissions, false)
			}
			if i < len(reviewStatus) {
				filteredReviewStatus = append(filteredReviewStatus, reviewStatus[i])
			} else {
				filteredReviewStatus = append(filteredReviewStatus, false)
			}
		}
	}
	
	// Calculate pagination
	totalItems := len(filteredCourses)
	totalPages := (totalItems + limit - 1) / limit // Ceiling division
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	
	if startIndex > totalItems {
		startIndex = totalItems
	}
	if endIndex > totalItems {
		endIndex = totalItems
	}
	
	// Get page slice
	var pageCourses []Course
	var pageEditPermissions []bool
	var pageReviewStatus []bool
	
	if startIndex < totalItems {
		pageCourses = filteredCourses[startIndex:endIndex]
		pageEditPermissions = filteredEditPermissions[startIndex:endIndex]
		pageReviewStatus = filteredReviewStatus[startIndex:endIndex]
	}
	
	response := map[string]interface{}{
		"courses":          pageCourses,
		"editPermissions":  pageEditPermissions,
		"reviewStatus":     pageReviewStatus,
		"pagination": map[string]interface{}{
			"currentPage":  page,
			"totalPages":   totalPages,
			"totalItems":   totalItems,
			"itemsPerPage": limit,
			"hasNext":      page < totalPages,
			"hasPrev":      page > 1,
		},
	}
	
	log.Printf("✅ GetAllCoursesAPI: Returning page %d/%d (%d items, %d total)", page, totalPages, len(pageCourses), totalItems)
	return c.JSON(http.StatusOK, response)
}

// GetReviewCoursesAPI returns paginated courses for the review page
func (h *Handlers) GetReviewCoursesAPI(c echo.Context) error {
	// Parse pagination parameters
	page := 1
	limit := 20
	search := ""
	
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	if s := c.QueryParam("search"); s != "" {
		search = s
	}
	
	log.Printf("🚀 GetReviewCoursesAPI: page=%d, limit=%d, search='%s'", page, limit, search)
	
	// Create a struct that includes both database info and database array index
	type CourseWithIndex struct {
		CourseDB
		DatabaseIndex int `json:"databaseIndex"`
	}
	
	var allAvailableCourses []CourseWithIndex
	
	// Get ALL courses from database - user should be able to review any course
	dbService := NewDatabaseService()
	allDBCourses, err := dbService.GetAllCourses()
	if err != nil {
		log.Printf("Warning: failed to get all courses: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load courses from database"})
	}
	
	// Get all courses from database with coordinates for mapping
	allCoursesWithCoords, err := dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load courses from database"})
	}
	
	log.Printf("📊 GetReviewCoursesAPI: Retrieved %d courses from database", len(allDBCourses))
	
	// Map database courses to database array indices
	for _, dbCourse := range allDBCourses {
		// Find the corresponding index in the database array
		databaseIndex := -1
		for i, course := range allCoursesWithCoords {
			if course.Name == dbCourse.Name && course.Address == dbCourse.Address {
				databaseIndex = i
				break
			}
		}
		
		if databaseIndex != -1 {
			allAvailableCourses = append(allAvailableCourses, CourseWithIndex{
				CourseDB:      dbCourse,
				DatabaseIndex: databaseIndex,
			})
		}
	}
		
	log.Printf("📊 GetReviewCoursesAPI: Mapped %d courses to database indices", len(allAvailableCourses))
	
	// Apply search filter
	var filteredCourses []CourseWithIndex
	if search != "" {
		for _, course := range allAvailableCourses {
			if strings.Contains(strings.ToLower(course.Name), strings.ToLower(search)) {
				filteredCourses = append(filteredCourses, course)
			}
		}
	} else {
		filteredCourses = allAvailableCourses
	}
	
	// Calculate pagination
	totalItems := len(filteredCourses)
	totalPages := (totalItems + limit - 1) / limit
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	
	if startIndex > totalItems {
		startIndex = totalItems
	}
	if endIndex > totalItems {
		endIndex = totalItems
	}
	
	// Get page slice
	var pageCourses []CourseWithIndex
	if startIndex < totalItems {
		pageCourses = filteredCourses[startIndex:endIndex]
	}
	
	response := map[string]interface{}{
		"courses": pageCourses,
		"pagination": map[string]interface{}{
			"currentPage":  page,
			"totalPages":   totalPages,
			"totalItems":   totalItems,
			"itemsPerPage": limit,
			"hasNext":      page < totalPages,
			"hasPrev":      page > 1,
		},
	}
	
	log.Printf("✅ GetReviewCoursesAPI: Returning page %d/%d (%d items, %d total)", page, totalPages, len(pageCourses), totalItems)
	return c.JSON(http.StatusOK, response)
}
