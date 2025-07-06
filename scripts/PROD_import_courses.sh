#!/bin/bash

# import_courses.sh - Import courses from JSON file with duplicate detection
# This script will import courses and skip duplicates based on hash

echo "📥 Course Import Script"
echo "======================"
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: Please run this script from the course_management directory"
    exit 1
fi

# Check which JSON file to use - prioritize the smaller file
JSON_FILE=""
if [ -f "scripts/course_details.json" ]; then
    JSON_FILE="scripts/course_details.json"
    echo "📁 Found: course_details.json (small dataset)"
elif [ -f "scripts/all_course_details.json" ]; then
    JSON_FILE="scripts/all_course_details.json"
    echo "📁 Found: all_course_details.json (large dataset)"
else
    echo "❌ Error: No course_details.json file found in scripts/ directory"
    exit 1
fi

# Show file info
echo "📊 File: $JSON_FILE"
COURSE_COUNT=$(jq length "$JSON_FILE" 2>/dev/null || echo "unknown")
echo "📊 Courses in file: $COURSE_COUNT"
echo ""

# Ask for confirmation
echo "🔄 This will import courses into the database:"
echo "   - Duplicate courses (same name+address) will be skipped"
echo "   - New courses will be added with generated hashes"
echo "   - Existing database courses will be preserved"
echo ""
read -p "Continue with import? (type 'yes' to confirm): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "❌ Import cancelled"
    exit 0
fi

echo ""
echo "📥 Starting course import..."

# Create a temporary Go script to import courses
cat > import_courses_temp.go << 'EOF'
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models
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
}

// JSON structure from scraped data
type ScrapedCourse struct {
	CourseName  string `json:"course_name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Destination string `json:"destination"`
	CityURL     string `json:"city_url"`
}

// Course structure for JSON storage
type Course struct {
	Name          string `json:"name"`
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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env.prod"); err != nil {
			log.Printf("Warning: .env file not found in current or parent directory")
		}
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5433")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "course_management_dev")
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

// GenerateCourseHash creates a deterministic hash from course name and address
func GenerateCourseHash(name, address string) string {
	normalizedName := normalizeString(name)
	normalizedAddress := normalizeString(address)
	combined := normalizedName + normalizedAddress
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])[:16]
}

// normalizeString cleans and standardizes a string for hashing
func normalizeString(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`[.,\-#]`).ReplaceAllString(s, "")

	replacements := map[string]string{
		" golf course":  " gc",
		" golf club":    " gc",
		" country club": " cc",
		" golf links":   " gl",
		" golf resort":  " gr",
		" street":       " st",
		" avenue":       " ave",
		" drive":        " dr",
		" road":         " rd",
		" boulevard":    " blvd",
		" north":        " n",
		" south":        " s",
		" east":         " e",
		" west":         " w",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	return s
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run import_courses_temp.go <json_file>")
	}

	jsonFile := os.Args[1]

	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔌 Connected to database")

	// Read JSON file
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	var scrapedCourses []ScrapedCourse
	if err := json.Unmarshal(data, &scrapedCourses); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("📊 Found %d courses in JSON file\n", len(scrapedCourses))

	// Track statistics
	var imported, skipped, errors int

	// Process each course
	for i, scraped := range scrapedCourses {
		// Generate hash for duplicate detection
		hash := GenerateCourseHash(scraped.CourseName, scraped.Address)

		// Check if course already exists
		var existingCourse CourseDB
		result := db.Where("hash = ?", hash).First(&existingCourse)
		
		if result.Error == nil {
			// Course exists, skip it
			skipped++
			if i%100 == 0 || i < 10 {
				fmt.Printf("⏭️  Skipped: %s (duplicate hash: %s)\n", scraped.CourseName, hash)
			}
			continue
		}

		// Create course data JSON
		courseData := Course{
			Name:          scraped.CourseName,
			Description:   fmt.Sprintf("Golf course in %s, %s", scraped.City, scraped.Destination),
			OverallRating: "", // No default rating - leave unset
			Review:        fmt.Sprintf("Located in %s.", scraped.Destination),
			Address:       scraped.Address,
		}

		courseDataJSON, err := json.Marshal(courseData)
		if err != nil {
			log.Printf("❌ Failed to marshal course data for %s: %v", scraped.CourseName, err)
			errors++
			continue
		}

		// Create database record
		course := CourseDB{
			Name:       scraped.CourseName,
			Address:    scraped.Address,
			Hash:       hash,
			CourseData: string(courseDataJSON),
			CreatedBy:  nil, // System import
		}

		// Insert into database
		if err := db.Create(&course).Error; err != nil {
			log.Printf("❌ Failed to insert %s: %v", scraped.CourseName, err)
			errors++
			continue
		}

		imported++
		
		// Progress indicator
		if i%100 == 0 || i < 10 {
			fmt.Printf("✅ Imported: %s (hash: %s)\n", scraped.CourseName, hash)
		}
	}

	// Final summary
	fmt.Println("")
	fmt.Println("🎉 Import completed!")
	fmt.Printf("   📊 Total processed: %d courses\n", len(scrapedCourses))
	fmt.Printf("   ✅ Successfully imported: %d courses\n", imported)
	fmt.Printf("   ⏭️  Skipped (duplicates): %d courses\n", skipped)
	fmt.Printf("   ❌ Errors: %d courses\n", errors)
	fmt.Printf("   📈 Success rate: %.1f%%\n", float64(imported)/float64(len(scrapedCourses))*100)
}
EOF

# Run the import script
echo "🚀 Executing import..."
go run import_courses_temp.go "$JSON_FILE"

# Clean up the temporary file
rm -f import_courses_temp.go

echo ""
echo "✅ Import script completed!"
echo "📝 Check the database with ./list_db.sh to see imported courses"
echo ""
echo "🔄 To see imported courses in the web app sidebar:"
echo "   Option 1: Restart the application with: ./restart_app.sh"
echo "   Option 2: Navigate to http://localhost:8080/api/migrate/courses to refresh" 