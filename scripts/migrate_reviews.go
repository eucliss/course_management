package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Copy the models from main package for migration
type User struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	GoogleID    string   `gorm:"uniqueIndex" json:"google_id"`
	Email       string   `gorm:"uniqueIndex" json:"email"`
	Name        string   `json:"name"`
	DisplayName *string  `json:"display_name"`
	Picture     string   `json:"picture"`
	Handicap    *float64 `json:"handicap,omitempty"`
	CreatedAt   int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64    `gorm:"autoUpdateTime" json:"updated_at"`
}

type CourseDB struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Address    string `json:"address"`
	Hash       string `gorm:"uniqueIndex;not null" json:"hash"`
	CourseData string `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint  `json:"created_by"`
	UpdatedBy  *uint  `json:"updated_by"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at"`

	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

type CourseReview struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	OverallRating      *string `gorm:"type:varchar(1);check:overall_rating IN ('S','A','B','C','D','F')" json:"overall_rating"`
	Price              *string `gorm:"type:varchar(10)" json:"price"`
	HandicapDifficulty *int    `json:"handicap_difficulty"`
	HazardDifficulty   *int    `json:"hazard_difficulty"`
	Merch              *string `gorm:"type:varchar(1);check:merch IN ('S','A','B','C','D','F')" json:"merch"`
	Condition          *string `gorm:"type:varchar(1);check:condition IN ('S','A','B','C','D','F')" json:"condition"`
	EnjoymentRating    *string `gorm:"type:varchar(1);check:enjoyment_rating IN ('S','A','B','C','D','F')" json:"enjoyment_rating"`
	Vibe               *string `gorm:"type:varchar(1);check:vibe IN ('S','A','B','C','D','F')" json:"vibe"`
	RangeRating        *string `gorm:"type:varchar(1);check:range_rating IN ('S','A','B','C','D','F')" json:"range_rating"`
	Amenities          *string `gorm:"type:varchar(1);check:amenities IN ('S','A','B','C','D','F')" json:"amenities"`
	Glizzies           *string `gorm:"type:varchar(1);check:glizzies IN ('S','A','B','C','D','F')" json:"glizzies"`
	ReviewText         *string `gorm:"type:text" json:"review_text"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updated_at"`

	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type UserCourseScore struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	Score      int      `gorm:"not null" json:"score"`
	Handicap   *float64 `gorm:"type:decimal(4,2)" json:"handicap"`
	DatePlayed *string  `gorm:"type:date" json:"date_played"`
	OutScore   *int     `json:"out_score"`
	InScore    *int     `json:"in_score"`
	Notes      *string  `gorm:"type:text" json:"notes"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`

	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type UserActivity struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	UserID       uint   `gorm:"not null" json:"user_id"`
	ActivityType string `gorm:"type:varchar(50);not null" json:"activity_type"`
	CourseID     *uint  `json:"course_id"`
	TargetUserID *uint  `json:"target_user_id"`
	Data         string `gorm:"type:jsonb" json:"data"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`

	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course     *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	TargetUser *User     `gorm:"foreignKey:TargetUserID" json:"target_user,omitempty"`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectDatabase() (*gorm.DB, error) {
	// Load environment variables - try current directory first, then parent
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found in current or parent directory")
		}
	}

	config := struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:   getEnvOrDefault("DB_NAME", "course_management"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	if config.Password == "" {
		return nil, fmt.Errorf("database password not set in environment variables")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

func main() {
	fmt.Println("🔄 MULTI-USER REVIEW SYSTEM MIGRATION")
	fmt.Println("=====================================")
	fmt.Println()

	// Connect to database
	fmt.Println("🔌 Connecting to database...")
	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	fmt.Println("✅ Connected successfully!")
	fmt.Println()

	// Run migration
	fmt.Println("🔄 Running database migrations...")

	err = db.AutoMigrate(
		&User{},
		&CourseDB{},
		&CourseReview{},
		&UserCourseScore{},
		&UserActivity{},
	)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Database migration completed successfully!")
	fmt.Println()

	// Check if tables exist
	fmt.Println("🔍 Verifying tables...")

	tableNames := []string{"users", "course_dbs", "course_reviews", "user_course_scores", "user_activities"}
	for _, tableName := range tableNames {
		if db.Migrator().HasTable(tableName) {
			fmt.Printf("✅ Table '%s' exists\n", tableName)
		} else {
			fmt.Printf("❌ Table '%s' missing\n", tableName)
		}
	}

	fmt.Println()
	fmt.Println("🎉 Migration process completed!")
	fmt.Println()
	fmt.Println("💡 You can now:")
	fmt.Println("   1. Run the review listing script: go run list_reviews.go")
	fmt.Println("   2. Start the application and create reviews")
	fmt.Println("   3. Use the multi-user review system")
}
