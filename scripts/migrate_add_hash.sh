#!/bin/bash

# migrate_add_hash.sh - Add hash column to existing course_dbs table
# This migration adds the hash field and populates it for existing courses

echo "🔧 Database Migration: Add Hash Column"
echo "======================================"
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: Please run this script from the course_management directory"
    exit 1
fi

echo "🔄 This migration will:"
echo "   - Add 'hash' column to course_dbs table"
echo "   - Generate hashes for existing courses"
echo "   - Add unique constraint on hash column"
echo ""
read -p "Continue with migration? (type 'yes' to confirm): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "❌ Migration cancelled"
    exit 0
fi

echo ""
echo "🚀 Starting migration..."

# Create a temporary Go script for migration
cat > migrate_hash_temp.go << 'EOF'
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
	Hash       string `json:"hash"` // Don't add constraint yet
	CourseData string `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint  `json:"created_by"`
	UpdatedBy  *uint  `json:"updated_by"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// Course structure for parsing existing data
type Course struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	Description   string `json:"description"`
	OverallRating string `json:"overallRating"`
	Review        string `json:"review"`
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
	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔌 Connected to database")

	// Step 1: Add hash column if it doesn't exist
	fmt.Print("🔧 Adding hash column... ")
	err = db.Exec("ALTER TABLE course_dbs ADD COLUMN IF NOT EXISTS hash VARCHAR(16)").Error
	if err != nil {
		log.Fatalf("Failed to add hash column: %v", err)
	}
	fmt.Println("✅ Hash column added")

	// Step 2: Get all existing courses
	var courses []CourseDB
	result := db.Find(&courses)
	if result.Error != nil {
		log.Fatalf("Failed to fetch courses: %v", result.Error)
	}

	fmt.Printf("📊 Found %d existing courses to migrate\n", len(courses))

	// Step 3: Generate and update hashes for existing courses
	var updated int
	for _, course := range courses {
		// Skip if hash already exists
		if course.Hash != "" {
			continue
		}

		// Parse course data to get name and address
		var courseData Course
		name := course.Name
		address := course.Address

		// Try to get address from CourseData JSON if not in Address field
		if address == "" && course.CourseData != "" {
			if err := json.Unmarshal([]byte(course.CourseData), &courseData); err == nil {
				address = courseData.Address
			}
		}

		// Generate hash
		hash := GenerateCourseHash(name, address)

		// Update the course with the hash
		result := db.Model(&course).Update("hash", hash)
		if result.Error != nil {
			log.Printf("❌ Failed to update hash for course ID %d: %v", course.ID, result.Error)
			continue
		}

		updated++
		if updated%10 == 0 || updated <= 5 {
			fmt.Printf("✅ Updated: %s (hash: %s)\n", name, hash)
		}
	}

	// Step 4: Add unique constraint on hash column
	fmt.Print("🔧 Adding unique constraint on hash... ")
	err = db.Exec("ALTER TABLE course_dbs ADD CONSTRAINT IF NOT EXISTS course_dbs_hash_unique UNIQUE (hash)").Error
	if err != nil {
		log.Printf("⚠️  Warning: Could not add unique constraint (may already exist): %v", err)
	} else {
		fmt.Println("✅ Unique constraint added")
	}

	// Final summary
	fmt.Println("")
	fmt.Println("🎉 Migration completed!")
	fmt.Printf("   📊 Total courses: %d\n", len(courses))
	fmt.Printf("   ✅ Hashes generated: %d\n", updated)
	fmt.Println("   📝 Database is now ready for hash-based imports")
}
EOF

# Run the migration script
echo "🚀 Executing migration..."
go run migrate_hash_temp.go

# Clean up the temporary file
rm -f migrate_hash_temp.go

echo ""
echo "✅ Migration completed!"
echo "📝 You can now run ./import_courses.sh to import courses with duplicate detection" 