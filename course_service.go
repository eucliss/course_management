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

type CourseService struct{}

func NewCourseService() *CourseService {
	return &CourseService{}
}

func (cs *CourseService) LoadCourses() ([]Course, error) {
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

func (cs *CourseService) SaveCourse(course Course) error {
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

	return os.WriteFile(filepath, courseJSON, 0644)
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
	var holes []Hole
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
	var scores []Score
	for i := 0; i < len(scoreMap); i++ {
		if score, exists := scoreMap[i]; exists {
			scores = append(scores, score)
		}
	}

	return scores
}
