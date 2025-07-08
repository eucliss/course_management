package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models
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
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	CreatedBy  *uint    `json:"created_by"`
	UpdatedBy  *uint    `json:"updated_by"`
	CreatedAt  int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64    `gorm:"autoUpdateTime" json:"updated_at"`
	Creator    *User    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater    *User    `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

type Course struct {
	Name          string `json:"name"`
	ID            int    `json:"ID"`
	Description   string `json:"description"`
	OverallRating string `json:"overallRating"`
	Review        string `json:"review"`
	Address       string `json:"address"`
}

// GeoJSON structures
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToDatabase() (*gorm.DB, error) {
	// Try to load .env from current directory, parent, or grandparent
	envPaths := []string{".env", "../.env", "../../.env"}
	var envLoaded bool
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("📄 Loaded environment variables from %s\n", path)
			envLoaded = true
			break
		}
	}
	if !envLoaded {
		log.Printf("Warning: .env file not found in current, parent, or grandparent directory")
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

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "Unknown"
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

func main() {
	fmt.Println("🗺️ GOLF COURSE GEOJSON EXPORTER")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		fmt.Println("\n💡 Setup instructions:")
		fmt.Println("   1. Ensure .env file exists with database credentials")
		fmt.Println("   2. Ensure PostgreSQL is running")
		os.Exit(1)
	}

	fmt.Println("✅ Connected to database successfully!")

	// Get all courses with coordinates
	var courses []CourseDB
	result := db.Where("latitude IS NOT NULL AND longitude IS NOT NULL").Find(&courses)
	if result.Error != nil {
		fmt.Printf("❌ Failed to fetch courses: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("📍 Found %d courses with coordinates\n", len(courses))

	if len(courses) == 0 {
		fmt.Println("⚠️ No courses with coordinates found. Run the geocoding script first.")
		os.Exit(1)
	}

	// Create GeoJSON structure
	geoJSON := GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]GeoJSONFeature, 0, len(courses)),
	}

	// Convert each course to GeoJSON feature
	successCount := 0
	for _, courseDB := range courses {
		// Parse course data JSON
		var course Course
		if err := json.Unmarshal([]byte(courseDB.CourseData), &course); err != nil {
			fmt.Printf("⚠️ Failed to parse course data for %s: %v\n", courseDB.Name, err)
			continue
		}

		// Create GeoJSON feature
		feature := GeoJSONFeature{
			Type: "Feature",
			Geometry: GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{*courseDB.Longitude, *courseDB.Latitude}, // [lng, lat] order for GeoJSON
			},
			Properties: map[string]interface{}{
				"name":    course.Name,
				"address": course.Address,
				"rating": func() string {
					if course.OverallRating == "" {
						return "-"
					}
					return course.OverallRating
				}(),
				"created": formatTime(courseDB.CreatedAt),
				"id":      courseDB.ID,
				"hash":    courseDB.Hash,
			},
		}

		geoJSON.Features = append(geoJSON.Features, feature)
		successCount++
	}

	fmt.Printf("✅ Successfully processed %d courses\n", successCount)

	// Write to file
	outputFile := "golf_courses.geojson"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("❌ Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Write JSON with pretty formatting
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(geoJSON); err != nil {
		fmt.Printf("❌ Failed to write GeoJSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🎉 GeoJSON export complete!\n")
	fmt.Printf("📄 Output file: %s\n", outputFile)
	fmt.Printf("📊 Total features: %d\n", len(geoJSON.Features))
	fmt.Println()
	fmt.Println("💡 You can now:")
	fmt.Println("   • Upload to QGIS, ArcGIS, or other GIS software")
	fmt.Println("   • Visualize on GitHub (supports GeoJSON)")
	fmt.Println("   • Use with Leaflet, Mapbox, or other web mapping libraries")
	fmt.Println("   • Import into Google Earth or Google My Maps")
}
