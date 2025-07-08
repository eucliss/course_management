package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models (same as export_geojson.go)
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
		log.Printf("Warning: .env file not found")
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "course_management")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	if password == "" {
		return nil, fmt.Errorf("database password not set")
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

func checkTippecanoe() error {
	_, err := exec.LookPath("tippecanoe")
	if err != nil {
		return fmt.Errorf("tippecanoe not found. Please install it:\n" +
			"  macOS: brew install tippecanoe\n" +
			"  Ubuntu: sudo apt-get install tippecanoe\n" +
			"  Or build from source: https://github.com/mapbox/tippecanoe")
	}
	return nil
}

func main() {
	fmt.Println("🗺️ GOLF COURSE VECTOR TILE GENERATOR")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	// Check if tippecanoe is installed
	if err := checkTippecanoe(); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Tippecanoe found")

	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connected to database")

	// Get all courses with coordinates
	var courses []CourseDB
	result := db.Where("latitude IS NOT NULL AND longitude IS NOT NULL").Find(&courses)
	if result.Error != nil {
		fmt.Printf("❌ Failed to fetch courses: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("📍 Found %d courses with coordinates\n", len(courses))

	if len(courses) == 0 {
		fmt.Println("⚠️ No courses with coordinates found")
		os.Exit(1)
	}

	// Create GeoJSON
	geoJSON := GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]GeoJSONFeature, 0, len(courses)),
	}

	for _, courseDB := range courses {
		var course Course
		if err := json.Unmarshal([]byte(courseDB.CourseData), &course); err != nil {
			continue
		}

		feature := GeoJSONFeature{
			Type: "Feature",
			Geometry: GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{*courseDB.Longitude, *courseDB.Latitude},
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
			},
		}

		geoJSON.Features = append(geoJSON.Features, feature)
	}

	// Create output directories
	os.MkdirAll("tiles", 0755)

	// Write temporary GeoJSON file
	tempGeoJSON := "temp_golf_courses.geojson"
	file, err := os.Create(tempGeoJSON)
	if err != nil {
		fmt.Printf("❌ Failed to create temp file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tempGeoJSON)

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(geoJSON); err != nil {
		fmt.Printf("❌ Failed to write GeoJSON: %v\n", err)
		file.Close()
		os.Exit(1)
	}
	file.Close()

	fmt.Println("✅ Created temporary GeoJSON file")

	// Generate vector tiles using tippecanoe
	fmt.Println("🔄 Generating vector tiles...")

	outputMBTiles := "golf_courses.mbtiles"
	cmd := exec.Command("tippecanoe",
		"-o", outputMBTiles,
		"-z", "14", // max zoom
		"-Z", "3", // min zoom
		"-r", "1", // simplification
		"-l", "golf", // layer name
		"--drop-densest-as-needed",
		"--extend-zooms-if-still-dropping",
		tempGeoJSON)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Tippecanoe failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated vector tiles: %s\n", outputMBTiles)

	// Extract tiles to directory structure
	fmt.Println("🔄 Extracting tiles to directory structure...")

	tilesDir := "tiles"
	cmd = exec.Command("tile-join",
		"--no-tile-compression",
		"--no-tile-size-limit",
		"--output-to-directory", tilesDir,
		outputMBTiles)

	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ tile-join failed (this is optional): %v\n", err)
		fmt.Println("You can use the .mbtiles file directly or extract manually")
	} else {
		fmt.Printf("✅ Extracted tiles to %s directory\n", tilesDir)
	}

	fmt.Println()
	fmt.Println("🎉 Vector tile generation complete!")
	fmt.Printf("📄 MBTiles file: %s\n", outputMBTiles)
	if _, err := os.Stat(tilesDir); err == nil {
		fmt.Printf("📁 Tiles directory: %s\n", tilesDir)
	}
	fmt.Println()
	fmt.Println("📋 Next steps:")
	fmt.Println("   1. Upload tiles to your CDN (DigitalOcean Spaces, AWS S3, etc.)")
	fmt.Println("   2. Update your map code to use vector tiles")
	fmt.Println("   3. Set the correct tile URL in your frontend")
	fmt.Println()
	fmt.Println("💡 Example tile URL pattern:")
	fmt.Println("   https://your-space.nyc3.cdn.digitaloceanspaces.com/tiles/{z}/{x}/{y}.pbf")
}
