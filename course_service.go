package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type CourseService struct {
	dbService *DatabaseService
}

func NewCourseService() *CourseService {
	// Database is required - fail if not available
	if GetDB() == nil {
		log.Fatal("Database connection required - JSON fallback removed")
	}

	return &CourseService{
		dbService: NewDatabaseService(),
	}
}

func (cs *CourseService) LoadCourses() ([]Course, error) {
	// Load from database only
	dbCourses, err := cs.dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load courses from database: %v", err)
	}

	log.Printf("✅ Loaded %d courses from database", len(dbCourses))
	return dbCourses, nil
}


func (cs *CourseService) SaveCourse(course Course) error {
	return cs.SaveCourseWithOwner(course, nil)
}

func (cs *CourseService) SaveCourseWithOwner(course Course, createdBy *uint) error {
	// Save to database only
	if err := cs.dbService.SaveCourseToDatabase(course, createdBy); err != nil {
		return fmt.Errorf("failed to save course to database: %v", err)
	}

	if createdBy != nil {
		log.Printf("✅ Course saved to database with owner ID %d", *createdBy)
	} else {
		log.Printf("✅ Course saved to database")
	}

	return nil
}

func (cs *CourseService) UpdateCourse(course Course) error {
	// Update in database only
	if err := cs.dbService.UpdateCourseInDatabase(course); err != nil {
		return fmt.Errorf("failed to update course in database: %v", err)
	}

	log.Printf("✅ Course updated in database")
	return nil
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
