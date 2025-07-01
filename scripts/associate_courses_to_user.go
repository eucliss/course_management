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

// Database models for migration
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
	CourseData string `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint  `json:"created_by"`
	UpdatedBy  *uint  `json:"updated_by"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:   getEnvOrDefault("DB_NAME", "course_management"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
}

func main() {
	fmt.Println("🏌️ Course Ownership Association Script")
	fmt.Println("======================================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	config := LoadDatabaseConfig()

	if config.Password == "" {
		log.Fatalf("❌ Database password not set. Please configure DB_PASSWORD environment variable.")
	}

	// Create connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Printf("✅ Connected to PostgreSQL database: %s\n", config.DBName)
	fmt.Println()

	// Miller Kinlin's Google ID
	millerGoogleID := "104280292232218166639"

	// Find Miller Kinlin by Google ID
	fmt.Printf("🔍 Looking for user with Google ID: %s\n", millerGoogleID)
	var miller User
	result := db.Where("google_id = ?", millerGoogleID).First(&miller)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Fatalf("❌ User with Google ID %s not found in database. Please make sure Miller Kinlin has logged in at least once.", millerGoogleID)
		} else {
			log.Fatalf("❌ Error finding user: %v", result.Error)
		}
	}

	fmt.Printf("✅ Found user: %s (%s) - Database ID: %d\n", miller.Name, miller.Email, miller.ID)
	fmt.Println()

	// Get all courses that don't have an owner (CreatedBy is NULL)
	fmt.Println("📋 Finding courses without owners...")
	var unownedCourses []CourseDB
	result = db.Where("created_by IS NULL").Find(&unownedCourses)

	if result.Error != nil {
		log.Fatalf("❌ Error fetching unowned courses: %v", result.Error)
	}

	if len(unownedCourses) == 0 {
		fmt.Println("✅ All courses already have owners!")
		fmt.Println()

		// Show current course ownership
		var allCourses []CourseDB
		db.Preload("Creator").Find(&allCourses)

		fmt.Println("📊 Current course ownership:")
		for _, course := range allCourses {
			ownerInfo := "No owner"
			if course.Creator != nil {
				ownerInfo = fmt.Sprintf("Owned by: %s (%s)", course.Creator.Name, course.Creator.Email)
			}
			fmt.Printf("   • %s - %s\n", course.Name, ownerInfo)
		}
		return
	}

	fmt.Printf("📋 Found %d courses without owners:\n", len(unownedCourses))
	for _, course := range unownedCourses {
		fmt.Printf("   • %s (ID: %d)\n", course.Name, course.ID)
	}
	fmt.Println()

	// Associate all unowned courses with Miller
	fmt.Printf("🔄 Associating %d courses with %s...\n", len(unownedCourses), miller.Name)

	result = db.Model(&CourseDB{}).Where("created_by IS NULL").Update("created_by", miller.ID)
	if result.Error != nil {
		log.Fatalf("❌ Error updating course ownership: %v", result.Error)
	}

	fmt.Printf("✅ Successfully associated %d courses with %s!\n", result.RowsAffected, miller.Name)
	fmt.Println()

	// Verify the changes
	fmt.Println("🔍 Verifying course ownership...")
	var millerCourses []CourseDB
	result = db.Where("created_by = ?", miller.ID).Find(&millerCourses)

	if result.Error != nil {
		log.Printf("⚠️ Error verifying courses: %v", result.Error)
	} else {
		fmt.Printf("✅ Miller Kinlin now owns %d courses:\n", len(millerCourses))
		for _, course := range millerCourses {
			fmt.Printf("   • %s (ID: %d)\n", course.Name, course.ID)
		}
	}

	fmt.Println()
	fmt.Println("🎉 Course ownership association completed successfully!")
}
