package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models (copied from main package)
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

type Course struct {
	Name          string `json:"name"`
	ID            int    `json:"ID"`
	Description   string `json:"description"`
	OverallRating string `json:"overallRating"`
	Review        string `json:"review"`
	Address       string `json:"address"`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToDatabase() (*gorm.DB, error) {
	// Load environment variables - try current directory first, then parent
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found in current or parent directory")
		}
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "course_management")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	if password == "" {
		return nil, fmt.Errorf("database password not set in environment variables")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

func main() {
	fmt.Println("🏌️ Golf Course Management - Database Contents")
	fmt.Println("=" + fmt.Sprintf("%*s", 48, "="))
	fmt.Println()

	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		fmt.Printf("❌ Cannot connect to database: %v\n", err)
		fmt.Println()
		fmt.Println("💡 To set up database connection:")
		fmt.Println("   1. Create a .env file in the parent directory")
		fmt.Println("   2. Add your database credentials:")
		fmt.Println("      DB_HOST=localhost")
		fmt.Println("      DB_PORT=5432")
		fmt.Println("      DB_USER=postgres")
		fmt.Println("      DB_PASSWORD=your_password")
		fmt.Println("      DB_NAME=course_management")
		fmt.Println("      DB_SSLMODE=disable")
		fmt.Println()
		return
	}

	fmt.Println("✅ Connected to database successfully!")
	fmt.Println()

	// List Users
	fmt.Println("👥 USERS:")
	fmt.Println("---------")

	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		log.Printf("❌ Failed to fetch users: %v", result.Error)
	} else if len(users) == 0 {
		fmt.Println("   No users found")
	} else {
		for i, user := range users {
			handicapStr := "Not set"
			if user.Handicap != nil {
				handicapStr = fmt.Sprintf("%.1f", *user.Handicap)
			}

			fmt.Printf("   %d. %s (%s)\n", i+1, user.Name, user.Email)
			fmt.Printf("      ID: %d | GoogleID: %s | Handicap: %s\n",
				user.ID, user.GoogleID, handicapStr)
			fmt.Println()
		}
	}

	// List Courses
	fmt.Println("⛳ COURSES:")
	fmt.Println("----------")

	var courses []CourseDB
	result = db.Preload("Creator").Preload("Updater").Find(&courses)
	if result.Error != nil {
		log.Printf("❌ Failed to fetch courses: %v", result.Error)
	} else if len(courses) == 0 {
		fmt.Println("   No courses found")
	} else {
		for i, courseDB := range courses {
			// Parse course data
			var course Course
			if err := json.Unmarshal([]byte(courseDB.CourseData), &course); err != nil {
				fmt.Printf("   %d. %s (Failed to parse course data)\n", i+1, courseDB.Name)
				continue
			}

			createdBy := "System"
			if courseDB.Creator != nil {
				displayName := courseDB.Creator.Name
				if courseDB.Creator.DisplayName != nil && *courseDB.Creator.DisplayName != "" {
					displayName = *courseDB.Creator.DisplayName
				}
				createdBy = displayName
			}

			updatedBy := ""
			if courseDB.Updater != nil {
				displayName := courseDB.Updater.Name
				if courseDB.Updater.DisplayName != nil && *courseDB.Updater.DisplayName != "" {
					displayName = *courseDB.Updater.DisplayName
				}
				updatedBy = fmt.Sprintf(" | Last edited by: %s", displayName)
			}

			fmt.Printf("   %d. %s\n", i+1, course.Name)
			fmt.Printf("      Address: %s\n", course.Address)

			rating := course.OverallRating
			if rating == "" {
				rating = "-"
			}
			fmt.Printf("      Rating: %s | Created by: %s%s\n", rating, createdBy, updatedBy)
			if course.Review != "" {
				reviewPreview := course.Review
				if len(reviewPreview) > 80 {
					reviewPreview = reviewPreview[:80] + "..."
				}
				fmt.Printf("      Review: %s\n", reviewPreview)
			}
			fmt.Println()
		}
	}

	// Summary
	fmt.Println("📊 SUMMARY:")
	fmt.Println("-----------")
	fmt.Printf("   Total Users: %d\n", len(users))
	fmt.Printf("   Total Courses: %d\n", len(courses))
	fmt.Println()
}
