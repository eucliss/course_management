package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models (matching the main app structure)
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
	ID         uint     `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"not null" json:"name"`
	Address    string   `json:"address"`
	Hash       string   `gorm:"uniqueIndex;not null" json:"hash"`
	CourseData string   `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint    `json:"created_by"`
	UpdatedBy  *uint    `json:"updated_by"`
	Latitude   *float64 `json:"latitude"`  // New field for geocoding
	Longitude  *float64 `json:"longitude"` // New field for geocoding
	CreatedAt  int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64    `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

// Mapbox Geocoding API response structures
type MapboxGeocodeResponse struct {
	Features []MapboxFeature `json:"features"`
}

type MapboxFeature struct {
	Center []float64 `json:"center"` // [longitude, latitude]
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Look for .env file in current directory and parent directories
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				log.Printf("⚠️ Warning: Error loading %s: %v", envPath, err)
			} else {
				log.Printf("✅ Loaded environment variables from %s", envPath)
				return
			}
		}
	}

	log.Printf("ℹ️ No .env file found, using system environment variables")
}

func loadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:   getEnvOrDefault("DB_NAME", "course_management_dev"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToDatabase() (*gorm.DB, error) {
	config := loadDatabaseConfig()

	var dsn string
	if config.Password == "" || config.Password == "trust_auth_no_password" {
		dsn = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
			config.Host, config.User, config.DBName, config.Port, config.SSLMode)
		log.Printf("🔄 Using trust authentication for database connection")
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)
	}

	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Printf("✅ Connected to PostgreSQL database: %s", config.DBName)
	return db, nil
}

func geocodeAddress(address, mapboxToken string) (float64, float64, error) {
	if address == "" {
		return 0, 0, fmt.Errorf("empty address")
	}

	// URL encode the address
	encodedAddress := url.QueryEscape(address)

	// Build the Mapbox Geocoding API URL
	apiURL := fmt.Sprintf("https://api.mapbox.com/geocoding/v5/mapbox.places/%s.json?access_token=%s", encodedAddress, mapboxToken)

	// Make the HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to make geocoding request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	// Parse the response
	var geocodeResponse MapboxGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geocodeResponse); err != nil {
		return 0, 0, fmt.Errorf("failed to decode geocoding response: %v", err)
	}

	// Check if we got results
	if len(geocodeResponse.Features) == 0 {
		return 0, 0, fmt.Errorf("no geocoding results found")
	}

	// Extract coordinates (Mapbox returns [longitude, latitude])
	center := geocodeResponse.Features[0].Center
	if len(center) < 2 {
		return 0, 0, fmt.Errorf("invalid coordinate format")
	}

	longitude := center[0]
	latitude := center[1]

	return latitude, longitude, nil
}

func addLatLongColumns(db *gorm.DB) error {
	log.Printf("🔄 Adding latitude and longitude columns to courses table...")

	// Check if columns already exist using a simpler approach
	var latitudeExists bool
	var longitudeExists bool

	// Check for latitude column
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'course_dbs' AND column_name = 'latitude'
		)
	`).Row().Scan(&latitudeExists)
	if err != nil {
		return fmt.Errorf("failed to check for latitude column: %v", err)
	}

	// Check for longitude column
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'course_dbs' AND column_name = 'longitude'
		)
	`).Row().Scan(&longitudeExists)
	if err != nil {
		return fmt.Errorf("failed to check for longitude column: %v", err)
	}

	// Add latitude column if it doesn't exist
	if !latitudeExists {
		if err := db.Exec("ALTER TABLE course_dbs ADD COLUMN latitude DOUBLE PRECISION").Error; err != nil {
			return fmt.Errorf("failed to add latitude column: %v", err)
		}
		log.Printf("✅ Added latitude column")
	} else {
		log.Printf("ℹ️ Latitude column already exists")
	}

	// Add longitude column if it doesn't exist
	if !longitudeExists {
		if err := db.Exec("ALTER TABLE course_dbs ADD COLUMN longitude DOUBLE PRECISION").Error; err != nil {
			return fmt.Errorf("failed to add longitude column: %v", err)
		}
		log.Printf("✅ Added longitude column")
	} else {
		log.Printf("ℹ️ Longitude column already exists")
	}

	return nil
}

func main() {
	log.Printf("🚀 Starting course geocoding script...")

	// Load environment variables from .env file
	loadEnvFile()

	// Check for required environment variables
	mapboxToken := os.Getenv("MAPBOX_ACCESS_TOKEN")
	if mapboxToken == "" {
		log.Fatal("❌ MAPBOX_ACCESS_TOKEN environment variable is required. Please set it in your .env file or environment.")
	}

	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Add latitude and longitude columns if they don't exist
	if err := addLatLongColumns(db); err != nil {
		log.Fatal("❌ Failed to add lat/long columns:", err)
	}

	// Get all courses from the database
	var courses []CourseDB
	if err := db.Find(&courses).Error; err != nil {
		log.Fatal("❌ Failed to fetch courses:", err)
	}

	log.Printf("📊 Found %d courses to geocode", len(courses))

	// Statistics
	var processed, successful, skipped, failed int
	var rateLimitDelay = 100 * time.Millisecond // 100ms delay between requests to respect rate limits

	for _, course := range courses {
		processed++
		log.Printf("🔄 Processing course %d/%d: %s", processed, len(courses), course.Name)

		// Skip if already geocoded
		if course.Latitude != nil && course.Longitude != nil {
			log.Printf("⏭️ Course already geocoded (lat: %.6f, lng: %.6f)", *course.Latitude, *course.Longitude)
			skipped++
			continue
		}

		// Skip if no address
		if course.Address == "" {
			log.Printf("⚠️ No address for course: %s", course.Name)
			skipped++
			continue
		}

		// Geocode the address
		latitude, longitude, err := geocodeAddress(course.Address, mapboxToken)
		if err != nil {
			log.Printf("❌ Failed to geocode %s: %v", course.Name, err)
			failed++

			// Continue to next course instead of failing completely
			time.Sleep(rateLimitDelay)
			continue
		}

		// Update the course with coordinates
		course.Latitude = &latitude
		course.Longitude = &longitude

		if err := db.Save(&course).Error; err != nil {
			log.Printf("❌ Failed to save coordinates for %s: %v", course.Name, err)
			failed++
		} else {
			log.Printf("✅ Geocoded %s: lat=%.6f, lng=%.6f", course.Name, latitude, longitude)
			successful++
		}

		// Rate limiting - respect Mapbox API limits
		time.Sleep(rateLimitDelay)

		// Progress update every 10 courses
		if processed%10 == 0 {
			log.Printf("📈 Progress: %d/%d processed (✅ %d successful, ⏭️ %d skipped, ❌ %d failed)",
				processed, len(courses), successful, skipped, failed)
		}
	}

	// Final statistics
	log.Printf("\n🎉 Geocoding complete!")
	log.Printf("📊 Final Statistics:")
	log.Printf("   Total courses: %d", len(courses))
	log.Printf("   Processed: %d", processed)
	log.Printf("   ✅ Successfully geocoded: %d", successful)
	log.Printf("   ⏭️ Skipped (already geocoded or no address): %d", skipped)
	log.Printf("   ❌ Failed: %d", failed)

	if failed > 0 {
		log.Printf("⚠️ %d courses failed to geocode. You may want to run the script again to retry failed courses.", failed)
	}
}
