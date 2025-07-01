package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type CourseService struct {
	dbService *DatabaseService
	useDB     bool
}

func NewCourseService() *CourseService {
	// Check if database is available
	useDB := true
	var dbService *DatabaseService

	if GetDB() != nil {
		dbService = NewDatabaseService()
	} else {
		useDB = false
		log.Printf("Warning: Database not available, using JSON files only")
	}

	return &CourseService{
		dbService: dbService,
		useDB:     useDB,
	}
}

func (cs *CourseService) LoadCourses() ([]Course, error) {
	// Try to load from database first if available
	if cs.useDB && cs.dbService != nil {
		dbCourses, err := cs.dbService.GetAllCoursesFromDatabase()
		if err == nil && len(dbCourses) > 0 {
			log.Printf("✅ Loaded %d courses from database", len(dbCourses))
			return dbCourses, nil
		}
		log.Printf("Warning: failed to load from database or no courses found: %v", err)
	}

	// Fallback to JSON files
	courses, err := cs.LoadCoursesFromJSON()
	if err != nil {
		return nil, err
	}

	// If database is available and we loaded from JSON, migrate the data
	if cs.useDB && cs.dbService != nil && len(courses) > 0 {
		log.Printf("📤 Migrating courses from JSON files to database...")
		if err := cs.dbService.MigrateJSONFilesToDatabase(courses); err != nil {
			log.Printf("Warning: failed to migrate courses to database: %v", err)
		}
	}

	return courses, nil
}

func (cs *CourseService) LoadCoursesFromJSON() ([]Course, error) {
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

	log.Printf("✅ Loaded %d courses from JSON files", len(courses))
	return courses, nil
}

func (cs *CourseService) SaveCourse(course Course) error {
	return cs.SaveCourseWithOwner(course, nil)
}

func (cs *CourseService) SaveCourseWithOwner(course Course, createdBy *uint) error {
	// Save to database if available
	if cs.useDB && cs.dbService != nil {
		if err := cs.dbService.SaveCourseToDatabase(course, createdBy); err != nil {
			log.Printf("Warning: failed to save to database: %v", err)
			// Continue to save to file as backup
		} else {
			if createdBy != nil {
				log.Printf("✅ Course saved to database with owner ID %d", *createdBy)
			} else {
				log.Printf("✅ Course saved to database")
			}
		}
	}

	// Also save to JSON file for backup/compatibility
	filename := cs.sanitizeFilename(course.Name) + ".json"
	filepath := filepath.Join("courses", filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("course with this name already exists")
	}

	courseJSON, err := json.MarshalIndent(course, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create course JSON: %v", err)
	}

	if err := os.WriteFile(filepath, courseJSON, 0644); err != nil {
		return fmt.Errorf("failed to save course file: %v", err)
	}

	log.Printf("✅ Course saved to JSON file")
	return nil
}

func (cs *CourseService) UpdateCourse(course Course) error {
	// Update in database if available
	if cs.useDB && cs.dbService != nil {
		if err := cs.dbService.UpdateCourseInDatabase(course); err != nil {
			log.Printf("Warning: failed to update in database: %v", err)
			// Continue to update file as backup
		} else {
			log.Printf("✅ Course updated in database")
		}
	}

	// Also update JSON file for backup/compatibility
	filename := cs.sanitizeFilename(course.Name) + ".json"
	filepath := filepath.Join("courses", filename)

	courseJSON, err := json.MarshalIndent(course, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create course JSON: %v", err)
	}

	if err := os.WriteFile(filepath, courseJSON, 0644); err != nil {
		return fmt.Errorf("failed to update course file: %v", err)
	}

	log.Printf("✅ Course updated in JSON file")
	return nil
}

func (cs *CourseService) sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return strings.ToLower(reg.ReplaceAllString(name, "_"))
}

func (cs *CourseService) ParseFormData(form url.Values) ([]Hole, []Score, error) {
	holes := cs.parseHoles(form)
	scores := cs.parseScores(form)
	return holes, scores, nil
}

func (cs *CourseService) parseHoles(form url.Values) []Hole {
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
	holes := make([]Hole, 0) // Initialize as empty slice, not nil
	for i := 0; i < len(holeMap); i++ {
		if hole, exists := holeMap[i]; exists {
			holes = append(holes, hole)
		}
	}

	return holes
}

func (cs *CourseService) parseScores(form url.Values) []Score {
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
	scores := make([]Score, 0) // Initialize as empty slice, not nil
	for i := 0; i < len(scoreMap); i++ {
		if score, exists := scoreMap[i]; exists {
			scores = append(scores, score)
		}
	}

	return scores
}
