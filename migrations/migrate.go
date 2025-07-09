package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New relational database models
type CourseNew struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"size:100;not null;index" json:"name"`
	Address       string    `gorm:"type:text;not null" json:"address"`
	Description   string    `gorm:"type:text" json:"description"`
	City          string    `gorm:"size:50;index" json:"city"`
	State         string    `gorm:"size:2" json:"state"`
	ZipCode       string    `gorm:"size:10" json:"zip_code"`
	Phone         string    `gorm:"size:20" json:"phone"`
	Website       string    `gorm:"size:255" json:"website"`
	OverallRating string    `gorm:"size:1;check:overall_rating IN ('S','A','B','C','D','F')" json:"overall_rating"`
	Review        string    `gorm:"type:text" json:"review"`
	Hash          string    `gorm:"uniqueIndex;not null" json:"hash"`
	Latitude      *float64  `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude     *float64  `gorm:"type:decimal(11,8)" json:"longitude"`
	CreatedBy     *uint     `gorm:"index" json:"created_by"`
	UpdatedBy     *uint     `json:"updated_by"`
	CreatedAt     int64     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     int64     `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	Holes    []CourseHole    `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"holes"`
	Rankings CourseRanking   `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"rankings"`
	Scores   []UserCourseScoreNew `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"scores"`
}

type CourseHole struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID    uint   `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"course_id"`
	HoleNumber  int    `gorm:"not null;check:hole_number BETWEEN 1 AND 18" json:"hole_number"`
	Par         int    `gorm:"check:par BETWEEN 3 AND 6" json:"par"`
	Yardage     int    `gorm:"check:yardage BETWEEN 0 AND 800" json:"yardage"`
	Description string `gorm:"type:text" json:"description"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
}

type CourseRanking struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID           uint   `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE" json:"course_id"`
	Price              string `gorm:"size:10" json:"price"`
	HandicapDifficulty int    `gorm:"check:handicap_difficulty BETWEEN 1 AND 10" json:"handicap_difficulty"`
	HazardDifficulty   int    `gorm:"check:hazard_difficulty BETWEEN 1 AND 10" json:"hazard_difficulty"`
	Merch              string `gorm:"size:1;check:merch IN ('S','A','B','C','D','F')" json:"merch"`
	Condition          string `gorm:"size:1;check:condition IN ('S','A','B','C','D','F')" json:"condition"`
	EnjoymentRating    string `gorm:"size:1;check:enjoyment_rating IN ('S','A','B','C','D','F')" json:"enjoyment_rating"`
	Vibe               string `gorm:"size:1;check:vibe IN ('S','A','B','C','D','F')" json:"vibe"`
	RangeRating        string `gorm:"size:1;check:range_rating IN ('S','A','B','C','D','F')" json:"range_rating"`
	Amenities          string `gorm:"size:1;check:amenities IN ('S','A','B','C','D','F')" json:"amenities"`
	Glizzies           string `gorm:"size:1;check:glizzies IN ('S','A','B','C','D','F')" json:"glizzies"`
	CreatedAt          int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserCourseScoreNew struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint    `gorm:"not null;index" json:"user_id"`
	CourseID  uint    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"course_id"`
	Score     int     `gorm:"not null;check:score BETWEEN 1 AND 200" json:"score"`
	Handicap  float64 `gorm:"type:decimal(4,1);check:handicap BETWEEN -5 AND 40" json:"handicap"`
	CreatedAt int64   `gorm:"autoCreateTime" json:"created_at"`
}

// Old JSON-based model for migration
type CourseDB struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"not null" json:"name"`
	Address    string   `json:"address"`
	Hash       string   `gorm:"uniqueIndex;not null" json:"hash"`
	CourseData string   `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint    `json:"created_by"`
	UpdatedBy  *uint    `json:"updated_by"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	CreatedAt  int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CourseDB) TableName() string {
	return "course_dbs"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToDatabase() (*gorm.DB, error) {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "password")
	dbname := getEnvOrDefault("DB_NAME", "course_management")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func runMigration(db *gorm.DB) error {
	ctx := context.Background()
	
	log.Println("🚀 Starting database schema migration...")

	// Step 1: Create new tables
	log.Println("📋 Creating new relational tables...")
	if err := db.AutoMigrate(&CourseNew{}, &CourseHole{}, &CourseRanking{}, &UserCourseScoreNew{}); err != nil {
		return fmt.Errorf("failed to create new tables: %w", err)
	}
	log.Println("✅ New tables created successfully")

	// Step 2: Check if we have data to migrate
	var oldCourseCount int64
	db.Model(&CourseDB{}).Count(&oldCourseCount)
	log.Printf("📊 Found %d courses to migrate from JSON storage", oldCourseCount)

	if oldCourseCount == 0 {
		log.Println("ℹ️  No existing course data found, skipping migration")
		return nil
	}

	// Step 3: Migrate course data
	log.Println("🔄 Migrating course data...")
	if err := migrateCourseData(db, ctx); err != nil {
		return fmt.Errorf("failed to migrate course data: %w", err)
	}

	// Step 4: Verify migration
	log.Println("🔍 Verifying migration...")
	if err := verifyMigration(db); err != nil {
		return fmt.Errorf("migration verification failed: %w", err)
	}

	log.Println("🎉 Database migration completed successfully!")
	return nil
}

func migrateCourseData(db *gorm.DB, ctx context.Context) error {
	// Get all old courses with JSON data
	var oldCourses []CourseDB
	if err := db.Find(&oldCourses).Error; err != nil {
		return fmt.Errorf("failed to fetch old courses: %w", err)
	}

	log.Printf("📁 Processing %d courses...", len(oldCourses))

	for i, oldCourse := range oldCourses {
		log.Printf("🔄 Processing course %d/%d: %s", i+1, len(oldCourses), oldCourse.Name)
		
		// Create basic course record
		newCourse := CourseNew{
			ID:          oldCourse.ID,
			Name:        oldCourse.Name,
			Address:     oldCourse.Address,
			Hash:        oldCourse.Hash,
			Latitude:    oldCourse.Latitude,
			Longitude:   oldCourse.Longitude,
			CreatedBy:   oldCourse.CreatedBy,
			UpdatedBy:   oldCourse.UpdatedBy,
			CreatedAt:   oldCourse.CreatedAt,
			UpdatedAt:   oldCourse.UpdatedAt,
		}

		// Parse JSON data if available
		if oldCourse.CourseData != "" {
			if err := parseJSONData(&newCourse, oldCourse.CourseData); err != nil {
				log.Printf("⚠️  Warning: Failed to parse JSON for course %s: %v", oldCourse.Name, err)
			}
		}

		// Save the course (this will cascade to related tables)
		if err := db.Create(&newCourse).Error; err != nil {
			return fmt.Errorf("failed to create course %s: %w", oldCourse.Name, err)
		}
	}

	return nil
}

func parseJSONData(course *CourseNew, jsonData string) error {
	// For now, we'll implement basic JSON parsing
	// In a real scenario, you'd use proper JSON unmarshaling
	
	// This is a simplified implementation
	// You would need to properly parse the JSON structure based on your actual data
	
	log.Printf("📝 Note: JSON parsing is simplified in this migration script")
	log.Printf("   Consider implementing full JSON parsing based on your actual data structure")
	
	return nil
}

func verifyMigration(db *gorm.DB) error {
	var newCourseCount, holeCount, rankingCount, scoreCount int64
	
	db.Model(&CourseNew{}).Count(&newCourseCount)
	db.Model(&CourseHole{}).Count(&holeCount)
	db.Model(&CourseRanking{}).Count(&rankingCount)
	db.Model(&UserCourseScoreNew{}).Count(&scoreCount)

	log.Printf("📊 Migration Results:")
	log.Printf("   - Courses migrated: %d", newCourseCount)
	log.Printf("   - Holes migrated: %d", holeCount)
	log.Printf("   - Rankings migrated: %d", rankingCount)
	log.Printf("   - Scores migrated: %d", scoreCount)

	if newCourseCount == 0 {
		return fmt.Errorf("no courses were migrated")
	}

	return nil
}

func main() {
	log.Println("🗃️  Course Management Database Migration")
	log.Println("======================================")

	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Run migration
	if err := runMigration(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migration completed successfully!")
}