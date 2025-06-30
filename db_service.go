package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type DatabaseService struct {
	db *gorm.DB
}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{
		db: GetDB(),
	}
}

// User operations
func (ds *DatabaseService) CreateUser(googleUser *GoogleUser) (*User, error) {
	if ds.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	user := &User{
		GoogleID: googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Picture:  googleUser.Picture,
	}

	result := ds.db.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}

	return user, nil
}

func (ds *DatabaseService) GetUserByGoogleID(googleID string) (*User, error) {
	if ds.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var user User
	result := ds.db.Where("google_id = ?", googleID).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, nil // User doesn't exist
		}
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}

	return &user, nil
}

func (ds *DatabaseService) GetUserByEmail(email string) (*User, error) {
	var user User
	result := ds.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}

	return &user, nil
}

func (ds *DatabaseService) GetUserByID(userID uint) (*User, error) {
	var user User
	result := ds.db.First(&user, userID)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}

	return &user, nil
}

func (ds *DatabaseService) UpdateUser(userID uint, googleUser *GoogleUser) error {
	var user User
	result := ds.db.First(&user, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to find user: %v", result.Error)
	}

	// Update user fields that might have changed in Google profile
	user.Email = googleUser.Email
	user.Name = googleUser.Name
	user.Picture = googleUser.Picture

	result = ds.db.Save(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %v", result.Error)
	}

	return nil
}

func (ds *DatabaseService) UpdateUserHandicap(userID uint, handicap float64) error {
	if ds.db == nil {
		return fmt.Errorf("database not connected")
	}

	var user User
	result := ds.db.First(&user, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to find user: %v", result.Error)
	}

	user.Handicap = &handicap

	result = ds.db.Save(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user handicap: %v", result.Error)
	}

	return nil
}

func (ds *DatabaseService) UpdateUserDisplayName(userID uint, displayName string) error {
	if ds.db == nil {
		return fmt.Errorf("database not connected")
	}

	var user User
	result := ds.db.First(&user, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to find user: %v", result.Error)
	}

	if displayName == "" {
		user.DisplayName = nil
	} else {
		user.DisplayName = &displayName
	}

	result = ds.db.Save(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user display name: %v", result.Error)
	}

	return nil
}

// Course operations
func (ds *DatabaseService) SaveCourseToDatabase(course Course, createdBy *uint) error {
	// Convert course struct to JSON for storage
	courseDataJSON, err := json.Marshal(course)
	if err != nil {
		return fmt.Errorf("failed to marshal course data: %v", err)
	}

	courseDB := &CourseDB{
		Name:       course.Name,
		Address:    course.Address,
		CourseData: string(courseDataJSON),
		CreatedBy:  createdBy,
	}

	result := ds.db.Create(courseDB)
	if result.Error != nil {
		return fmt.Errorf("failed to save course to database: %v", result.Error)
	}

	log.Printf("✅ Course '%s' saved to database with ID: %d", course.Name, courseDB.ID)
	return nil
}

func (ds *DatabaseService) GetAllCoursesFromDatabase() ([]Course, error) {
	var coursesDB []CourseDB
	result := ds.db.Find(&coursesDB)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch courses from database: %v", result.Error)
	}

	var courses []Course
	for i, courseDB := range coursesDB {
		var course Course
		if err := json.Unmarshal([]byte(courseDB.CourseData), &course); err != nil {
			log.Printf("Warning: failed to unmarshal course %d: %v", courseDB.ID, err)
			continue
		}

		// Set the ID to match the array index for backward compatibility
		course.ID = i
		courses = append(courses, course)
	}

	return courses, nil
}

func (ds *DatabaseService) UpdateCourseInDatabase(course Course) error {
	// Find the course by name (since we're transitioning from file-based system)
	var courseDB CourseDB
	result := ds.db.Where("name = ?", course.Name).First(&courseDB)

	if result.Error != nil {
		// If course doesn't exist in DB, create it
		if result.Error.Error() == "record not found" {
			return ds.SaveCourseToDatabase(course, nil)
		}
		return fmt.Errorf("failed to find course: %v", result.Error)
	}

	// Update the course data
	courseDataJSON, err := json.Marshal(course)
	if err != nil {
		return fmt.Errorf("failed to marshal course data: %v", err)
	}

	courseDB.CourseData = string(courseDataJSON)
	courseDB.Address = course.Address

	result = ds.db.Save(&courseDB)
	if result.Error != nil {
		return fmt.Errorf("failed to update course: %v", result.Error)
	}

	log.Printf("✅ Course '%s' updated in database", course.Name)
	return nil
}

// Migration helper functions
func (ds *DatabaseService) MigrateJSONFilesToDatabase(courses []Course) error {
	log.Printf("🔄 Migrating %d courses from JSON files to database...", len(courses))

	for _, course := range courses {
		// Check if course already exists
		var existingCourse CourseDB
		result := ds.db.Where("name = ?", course.Name).First(&existingCourse)

		if result.Error != nil && result.Error.Error() == "record not found" {
			// Course doesn't exist, create it
			if err := ds.SaveCourseToDatabase(course, nil); err != nil {
				log.Printf("Warning: failed to migrate course '%s': %v", course.Name, err)
				continue
			}
		} else if result.Error != nil {
			log.Printf("Warning: error checking course '%s': %v", course.Name, result.Error)
			continue
		} else {
			log.Printf("Course '%s' already exists in database, skipping", course.Name)
		}
	}

	log.Printf("✅ Course migration completed")
	return nil
}

func (ds *DatabaseService) GetDatabaseStats() (map[string]int, error) {
	stats := make(map[string]int)

	var userCount int64
	if err := ds.db.Model(&User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}
	stats["users"] = int(userCount)

	var courseCount int64
	if err := ds.db.Model(&CourseDB{}).Count(&courseCount).Error; err != nil {
		return nil, err
	}
	stats["courses"] = int(courseCount)

	return stats, nil
}
