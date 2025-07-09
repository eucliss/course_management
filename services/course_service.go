package services

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type courseService struct {
	courseRepo CourseRepository
	userRepo   UserRepository
}

func NewCourseService(courseRepo CourseRepository, userRepo UserRepository) CourseService {
	return &courseService{
		courseRepo: courseRepo,
		userRepo:   userRepo,
	}
}

func (s *courseService) CreateCourse(ctx context.Context, course Course, createdBy *uint) error {
	// Validate course data
	if err := s.ValidateCourse(course); err != nil {
		return fmt.Errorf("course validation failed: %w", err)
	}

	// Check if course already exists
	exists, err := s.courseRepo.Exists(ctx, course.Name, course.Address)
	if err != nil {
		return fmt.Errorf("failed to check course existence: %w", err)
	}
	if exists {
		return fmt.Errorf("course with name '%s' and address '%s' already exists", course.Name, course.Address)
	}

	// Create the course
	if err := s.courseRepo.Create(ctx, course, createdBy); err != nil {
		return fmt.Errorf("failed to create course: %w", err)
	}

	return nil
}

func (s *courseService) GetCourse(ctx context.Context, id uint) (*Course, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	return course, nil
}

func (s *courseService) GetCourseByIndex(ctx context.Context, index int) (*Course, error) {
	course, err := s.courseRepo.GetByIndex(ctx, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get course by index: %w", err)
	}

	return course, nil
}

func (s *courseService) GetAllCourses(ctx context.Context) ([]Course, error) {
	courses, err := s.courseRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all courses: %w", err)
	}

	return courses, nil
}

func (s *courseService) UpdateCourse(ctx context.Context, course Course, updatedBy *uint) error {
	// Validate course data
	if err := s.ValidateCourse(course); err != nil {
		return fmt.Errorf("course validation failed: %w", err)
	}

	// Check if user can edit this course
	if updatedBy != nil {
		canEdit, err := s.courseRepo.CanEdit(ctx, uint(course.ID), *updatedBy)
		if err != nil {
			return fmt.Errorf("failed to check edit permissions: %w", err)
		}
		if !canEdit {
			return fmt.Errorf("user does not have permission to edit this course")
		}
	}

	// Update the course
	if err := s.courseRepo.Update(ctx, course, updatedBy); err != nil {
		return fmt.Errorf("failed to update course: %w", err)
	}

	return nil
}

func (s *courseService) DeleteCourse(ctx context.Context, id uint, userID uint) error {
	// Check if user can edit this course
	canEdit, err := s.courseRepo.CanEdit(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("failed to check edit permissions: %w", err)
	}
	if !canEdit {
		return fmt.Errorf("user does not have permission to delete this course")
	}

	// Delete the course
	if err := s.courseRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}

	return nil
}

func (s *courseService) CanEditCourse(ctx context.Context, courseID uint, userID uint) (bool, error) {
	return s.courseRepo.CanEdit(ctx, courseID, userID)
}

func (s *courseService) CanEditCourseByIndex(ctx context.Context, index int, userID uint) (bool, error) {
	return s.courseRepo.CanEditByIndex(ctx, index, userID)
}

func (s *courseService) GetUserCourses(ctx context.Context, userID uint) ([]Course, error) {
	courses, err := s.courseRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user courses: %w", err)
	}

	return courses, nil
}

func (s *courseService) GetCoursesWithPagination(ctx context.Context, offset, limit int) ([]Course, int64, error) {
	courses, total, err := s.courseRepo.GetWithPagination(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get courses with pagination: %w", err)
	}

	return courses, total, nil
}

func (s *courseService) GetAvailableCoursesForReview(ctx context.Context, userID uint) ([]Course, error) {
	courses, err := s.courseRepo.GetAvailableForReview(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available courses for review: %w", err)
	}

	return courses, nil
}

func (s *courseService) FindCourseByNameAndAddress(ctx context.Context, name, address string) (*Course, error) {
	course, err := s.courseRepo.GetByNameAndAddress(ctx, name, address)
	if err != nil {
		return nil, fmt.Errorf("failed to find course: %w", err)
	}

	return course, nil
}

func (s *courseService) ParseCourseForm(form map[string][]string) ([]Hole, []Score, error) {
	holes := s.parseHoles(form)
	scores := s.parseScores(form)
	return holes, scores, nil
}

func (s *courseService) ValidateCourse(course Course) error {
	if strings.TrimSpace(course.Name) == "" {
		return fmt.Errorf("course name is required")
	}
	if strings.TrimSpace(course.Address) == "" {
		return fmt.Errorf("course address is required")
	}
	if len(course.Name) < 3 || len(course.Name) > 100 {
		return fmt.Errorf("course name must be between 3 and 100 characters")
	}
	if len(course.Address) < 10 || len(course.Address) > 200 {
		return fmt.Errorf("course address must be between 10 and 200 characters")
	}

	// Validate holes if present
	for i, hole := range course.Holes {
		if hole.Par < 3 || hole.Par > 6 {
			return fmt.Errorf("hole %d: par must be between 3 and 6", i+1)
		}
		if hole.Yardage < 0 || hole.Yardage > 800 {
			return fmt.Errorf("hole %d: yardage must be between 0 and 800", i+1)
		}
	}

	// Validate scores if present
	for i, score := range course.Scores {
		if score.Score < 1 || score.Score > 20 {
			return fmt.Errorf("score %d: score must be between 1 and 20", i+1)
		}
		if score.Handicap < -5 || score.Handicap > 40 {
			return fmt.Errorf("score %d: handicap must be between -5 and 40", i+1)
		}
	}

	return nil
}

func (s *courseService) parseHoles(form map[string][]string) []Hole {
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
	holes := make([]Hole, 0)
	for i := 0; i < len(holeMap); i++ {
		if hole, exists := holeMap[i]; exists {
			holes = append(holes, hole)
		}
	}

	return holes
}

func (s *courseService) parseScores(form map[string][]string) []Score {
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
	scores := make([]Score, 0)
	for i := 0; i < len(scoreMap); i++ {
		if score, exists := scoreMap[i]; exists {
			scores = append(scores, score)
		}
	}

	return scores
}

// Helper function to convert url.Values to map[string][]string for compatibility
func ConvertFormValues(form url.Values) map[string][]string {
	converted := make(map[string][]string)
	for key, values := range form {
		converted[key] = values
	}
	return converted
}