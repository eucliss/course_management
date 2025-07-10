package testing

import (
	"log"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB represents a test database instance
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB creates a new test database with in-memory SQLite
func NewTestDB(t *testing.T) *TestDB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable SQL logging during tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	testDB := &TestDB{DB: db}
	testDB.migrate(t)
	return testDB
}

// migrate runs all necessary migrations for test database
func (tdb *TestDB) migrate(t *testing.T) {
	// Import the actual database models from the main package
	// Note: In a real scenario, you'd define these models in a shared package
	
	// User model for testing - matching main database schema
	type UserDB struct {
		ID          uint     `gorm:"primaryKey"`
		GoogleID    string   `gorm:"uniqueIndex"`
		Email       string   `gorm:"uniqueIndex"`
		Name        string   
		DisplayName *string  
		Picture     string   
		Handicap    *float64 
		CreatedAt   int64    `gorm:"autoCreateTime"`
		UpdatedAt   int64    `gorm:"autoUpdateTime"`
	}

	// Course model for testing - matching main database schema
	type CourseDB struct {
		ID         uint     `gorm:"primaryKey"`
		Name       string   `gorm:"not null"`
		Address    string   
		Hash       string   `gorm:"uniqueIndex;not null"`
		CourseData string   `gorm:"type:text"`
		CreatedBy  *uint    
		UpdatedBy  *uint    
		Latitude   *float64 
		Longitude  *float64 
		CreatedAt  int64    `gorm:"autoCreateTime"`
		UpdatedAt  int64    `gorm:"autoUpdateTime"`
	}

	// CourseReview model for testing - matching main database schema
	type CourseReviewDB struct {
		ID                 uint    `gorm:"primaryKey"`
		CourseID           uint    `gorm:"not null"`
		UserID             uint    `gorm:"not null"`
		OverallRating      *string `gorm:"type:varchar(1)"`
		Price              *string `gorm:"type:varchar(10)"`
		HandicapDifficulty *int    
		HazardDifficulty   *int    
		Merch              *string `gorm:"type:varchar(1)"`
		Condition          *string `gorm:"type:varchar(1)"`
		EnjoymentRating    *string `gorm:"type:varchar(1)"`
		Vibe               *string `gorm:"type:varchar(1)"`
		RangeRating        *string `gorm:"type:varchar(1)"`
		Amenities          *string `gorm:"type:varchar(1)"`
		Glizzies           *string `gorm:"type:varchar(1)"`
		ReviewText         *string `gorm:"type:text"`
		CreatedAt          int64   `gorm:"autoCreateTime"`
		UpdatedAt          int64   `gorm:"autoUpdateTime"`
	}

	// UserCourseScore model for testing
	type UserCourseScoreDB struct {
		ID         uint     `gorm:"primaryKey"`
		CourseID   uint     `gorm:"not null"`
		UserID     uint     `gorm:"not null"`
		Score      int      `gorm:"not null"`
		Handicap   *float64 `gorm:"type:decimal(4,2)"`
		DatePlayed *string  `gorm:"type:date"`
		OutScore   *int     
		InScore    *int     
		Notes      *string  `gorm:"type:text"`
		CreatedAt  int64    `gorm:"autoCreateTime"`
	}

	// UserCourseHole model for testing
	type UserCourseHoleDB struct {
		ID          uint    `gorm:"primaryKey"`
		CourseID    uint    `gorm:"not null"`
		UserID      uint    `gorm:"not null"`
		Number      int     `gorm:"not null"`
		Par         *int    
		Yardage     *int    
		Description *string `gorm:"type:text"`
		CreatedAt   int64   `gorm:"autoCreateTime"`
		UpdatedAt   int64   `gorm:"autoUpdateTime"`
	}

	// UserActivity model for testing
	type UserActivityDB struct {
		ID           uint   `gorm:"primaryKey"`
		UserID       uint   `gorm:"not null"`
		ActivityType string `gorm:"type:varchar(50);not null"`
		CourseID     *uint  
		TargetUserID *uint  
		Data         string `gorm:"type:text"`
		CreatedAt    int64  `gorm:"autoCreateTime"`
	}

	err := tdb.DB.AutoMigrate(
		&UserDB{},
		&CourseDB{},
		&CourseReviewDB{},
		&UserCourseScoreDB{},
		&UserCourseHoleDB{},
		&UserActivityDB{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Debug: show created tables
	var tables []string
	rows, err := tdb.DB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err == nil {
				tables = append(tables, tableName)
			}
		}
	}
	log.Printf("✅ Test database migrated successfully. Tables: %v", tables)
}

// Close closes the test database connection
func (tdb *TestDB) Close() error {
	sqlDB, err := tdb.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CleanupTables truncates all tables for test isolation
func (tdb *TestDB) CleanupTables(t *testing.T) {
	// Get list of tables that exist
	var tables []string
	rows, err := tdb.DB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Rows()
	if err != nil {
		t.Logf("Warning: failed to get table list: %v", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		// Skip SQLite system tables
		if tableName != "sqlite_sequence" && tableName != "sqlite_master" {
			tables = append(tables, tableName)
		}
	}
	
	for _, table := range tables {
		if err := tdb.DB.Exec("DELETE FROM " + table).Error; err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
	
	// Reset auto-increment sequences for consistent test IDs
	if err := tdb.DB.Exec("DELETE FROM sqlite_sequence").Error; err != nil {
		t.Logf("Warning: failed to reset auto-increment sequences: %v", err)
	}
}

// SeedTestData adds common test data for use across tests
func (tdb *TestDB) SeedTestData(t *testing.T) *TestFixtures {
	fixtures := &TestFixtures{}

	// Create test users
	testUser1 := map[string]interface{}{
		"google_id":    "test-user-1",
		"email":        "test1@example.com",
		"name":         "Test User 1",
		"display_name": "TestUser1",
		"picture":      "https://example.com/pic1.jpg",
		"handicap":     15.5,
		"created_at":   1704067200, // 2024-01-01
		"updated_at":   1704067200,
	}

	testUser2 := map[string]interface{}{
		"google_id":    "test-user-2",
		"email":        "test2@example.com",
		"name":         "Test User 2",
		"display_name": "TestUser2",
		"picture":      "https://example.com/pic2.jpg",
		"handicap":     8.0,
		"created_at":   1704067200,
		"updated_at":   1704067200,
	}

	result1 := tdb.DB.Table("user_dbs").Create(testUser1)
	if result1.Error != nil {
		t.Fatalf("Failed to seed test user 1: %v", result1.Error)
	}
	fixtures.User1ID = 1

	result2 := tdb.DB.Table("user_dbs").Create(testUser2)
	if result2.Error != nil {
		t.Fatalf("Failed to seed test user 2: %v", result2.Error)
	}
	fixtures.User2ID = 2

	// Create test courses
	testCourse1 := map[string]interface{}{
		"name":         "Test Golf Course 1",
		"address":      "123 Golf Lane, Test City, TX 12345",
		"hash":         "test-course-hash-1",
		"course_data":  `{"name":"Test Golf Course 1","ID":1,"description":"A beautiful test golf course","ranks":{"price":"$50","handicapDifficulty":7,"hazardDifficulty":6,"merch":"B","condition":"A","enjoymentRating":"A","vibe":"A","range":"B","amenities":"A","glizzies":"C"},"overallRating":"A","review":"Great course for testing","holes":[{"number":1,"par":4,"yardage":350}],"scores":[{"hole":1,"score":4}],"address":"123 Golf Lane, Test City, TX 12345"}`,
		"latitude":     30.2672,
		"longitude":    -97.7431,
		"created_by":   fixtures.User1ID,
		"created_at":   1704067200,
		"updated_at":   1704067200,
	}

	testCourse2 := map[string]interface{}{
		"name":         "Test Golf Course 2",
		"address":      "456 Fairway Drive, Test Town, CA 54321",
		"hash":         "test-course-hash-2",
		"course_data":  `{"name":"Test Golf Course 2","ID":2,"description":"Another test golf course","ranks":{"price":"$40","handicapDifficulty":5,"hazardDifficulty":4,"merch":"A","condition":"B","enjoymentRating":"B","vibe":"B","range":"A","amenities":"B","glizzies":"B"},"overallRating":"B","review":"Good for practice rounds","holes":[{"number":1,"par":3,"yardage":180}],"scores":[{"hole":1,"score":3}],"address":"456 Fairway Drive, Test Town, CA 54321"}`,
		"latitude":     34.0522,
		"longitude":    -118.2437,
		"created_by":   fixtures.User2ID,
		"created_at":   1704067200,
		"updated_at":   1704067200,
	}

	result3 := tdb.DB.Table("course_dbs").Create(testCourse1)
	if result3.Error != nil {
		t.Fatalf("Failed to seed test course 1: %v", result3.Error)
	}
	fixtures.Course1ID = 1

	result4 := tdb.DB.Table("course_dbs").Create(testCourse2)
	if result4.Error != nil {
		t.Fatalf("Failed to seed test course 2: %v", result4.Error)
	}
	fixtures.Course2ID = 2

	// Create test reviews
	testReview1 := map[string]interface{}{
		"course_id":        fixtures.Course1ID,
		"user_id":          fixtures.User1ID,
		"overall_rating":   "A",
		"review_text":      "Excellent course layout",
		"price":            "$50",
		"handicap_difficulty": 7,
		"hazard_difficulty": 6,
		"condition":        "A",
		"created_at":       1704067200,
		"updated_at":       1704067200,
	}

	result5 := tdb.DB.Table("course_review_dbs").Create(testReview1)
	if result5.Error != nil {
		t.Fatalf("Failed to seed test review: %v", result5.Error)
	}
	fixtures.Review1ID = 1

	log.Printf("✅ Test data seeded: users=%d, courses=%d, reviews=%d", 2, 2, 1)
	return fixtures
}

// TestFixtures holds references to test data IDs
type TestFixtures struct {
	User1ID   uint
	User2ID   uint
	Course1ID uint
	Course2ID uint
	Review1ID uint
}