package main

import (
	"fmt"
)

// CourseListItem for lightweight course listing
type CourseListItem struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Creator *struct {
		Name        string  `json:"name"`
		DisplayName *string `json:"display_name"`
	} `json:"creator,omitempty"`
	CreatedAt int64 `json:"created_at"`
	CanEdit   bool  `json:"can_edit"`
}

// Add these optimized methods to your DatabaseService

// GetCourseByIDOptimized - Direct lookup by ID with optional preloading
func (ds *DatabaseService) GetCourseByIDOptimized(courseID uint, preloadRelations bool) (*CourseDB, error) {
	if ds.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var courseDB CourseDB
	query := ds.db

	if preloadRelations {
		query = query.Preload("Creator").Preload("Updater")
	}

	result := query.First(&courseDB, courseID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find course: %v", result.Error)
	}

	return &courseDB, nil
}

// GetCoursesWithPagination - Paginated course loading
func (ds *DatabaseService) GetCoursesWithPagination(offset, limit int, preloadRelations bool) ([]CourseDB, int64, error) {
	if ds.db == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}

	var courses []CourseDB
	var totalCount int64

	// Get total count for pagination info
	if err := ds.db.Model(&CourseDB{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %v", err)
	}

	query := ds.db.Offset(offset).Limit(limit).Order("created_at DESC")

	if preloadRelations {
		query = query.Preload("Creator").Preload("Updater")
	}

	if err := query.Find(&courses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch courses: %v", err)
	}

	return courses, totalCount, nil
}

// CanEditCourseOptimized - Direct ownership check without loading all courses
func (ds *DatabaseService) CanEditCourseOptimized(courseID uint, userID uint) (bool, error) {
	if ds.db == nil {
		return false, fmt.Errorf("database not connected")
	}

	var count int64
	err := ds.db.Model(&CourseDB{}).Where("id = ? AND created_by = ?", courseID, userID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check course ownership: %v", err)
	}

	return count > 0, nil
}

// GetCoursesByUserOptimized - Get user courses with pagination
func (ds *DatabaseService) GetCoursesByUserOptimized(userID uint, offset, limit int) ([]CourseDB, int64, error) {
	if ds.db == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}

	var courses []CourseDB
	var totalCount int64

	// Get total count
	if err := ds.db.Model(&CourseDB{}).Where("created_by = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user courses: %v", err)
	}

	// Get paginated results
	err := ds.db.Where("created_by = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&courses).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch user courses: %v", err)
	}

	return courses, totalCount, nil
}

// GetCoursesLightweight - Get courses without JSON data unmarshaling for lists
func (ds *DatabaseService) GetCoursesLightweight(offset, limit int, userID *uint) ([]CourseListItem, int64, error) {
	if ds.db == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}

	var courses []CourseListItem
	var totalCount int64

	// Build query
	query := ds.db.Table("course_dbs").
		Select("course_dbs.id, course_dbs.name, course_dbs.address, course_dbs.created_at, course_dbs.created_by").
		Joins("LEFT JOIN users ON course_dbs.created_by = users.id")

	// Add user filter if specified
	if userID != nil {
		query = query.Where("course_dbs.created_by = ?", *userID)
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %v", err)
	}

	// Get paginated data
	rows, err := query.Offset(offset).Limit(limit).Order("course_dbs.created_at DESC").Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch courses: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var course CourseListItem
		var createdBy *uint
		var creatorName *string
		var creatorDisplayName *string

		err := rows.Scan(&course.ID, &course.Name, &course.Address, &course.CreatedAt, &createdBy, &creatorName, &creatorDisplayName)
		if err != nil {
			continue
		}

		if createdBy != nil && creatorName != nil {
			course.Creator = &struct {
				Name        string  `json:"name"`
				DisplayName *string `json:"display_name"`
			}{
				Name:        *creatorName,
				DisplayName: creatorDisplayName,
			}
		}

		// Check if user can edit
		if userID != nil && createdBy != nil && *createdBy == *userID {
			course.CanEdit = true
		}

		courses = append(courses, course)
	}

	return courses, totalCount, nil
}
