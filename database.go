package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Database models for GORM
type User struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	GoogleID    string   `gorm:"uniqueIndex" json:"google_id"`
	Email       string   `gorm:"uniqueIndex" json:"email"`
	Name        string   `json:"name"`         // Google name
	DisplayName *string  `json:"display_name"` // Custom display name
	Picture     string   `json:"picture"`
	Handicap    *float64 `json:"handicap,omitempty"`
	CreatedAt   int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64    `gorm:"autoUpdateTime" json:"updated_at"`
}

type CourseDB struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"not null" json:"name"`
	Address    string   `json:"address"`
	Hash       string   `gorm:"uniqueIndex;not null" json:"hash"` // Unique hash based on name + address
	CourseData string   `gorm:"type:jsonb" json:"course_data"`    // Store existing JSON structure
	CreatedBy  *uint    `json:"created_by"`
	UpdatedBy  *uint    `json:"updated_by"`
	Latitude   *float64 `json:"latitude"`  // Geocoded latitude
	Longitude  *float64 `json:"longitude"` // Geocoded longitude
	CreatedAt  int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64    `gorm:"autoUpdateTime" json:"updated_at"`

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

func LoadDatabaseConfig() *DatabaseConfig {
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

func InitDatabase() error {
	config := LoadDatabaseConfig()

	// Create connection string - handle trust authentication (no password required)
	var dsn string
	if config.Password == "" || config.Password == "trust_auth_no_password" {
		// Trust authentication - no password needed
		dsn = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
			config.Host, config.User, config.DBName, config.Port, config.SSLMode)
		log.Printf("🔄 Using trust authentication for database connection")
	} else {
		// Password authentication
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Connect to database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		// Don't fail the application, just log the error
		log.Printf("⚠️ Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Printf("✅ Connected to PostgreSQL database: %s", config.DBName)

	// Auto-migrate the schema
	if err := AutoMigrate(); err != nil {
		log.Printf("⚠️ Failed to migrate database: %v", err)
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

func AutoMigrate() error {
	log.Printf("🔄 Running database migrations...")

	err := DB.AutoMigrate(
		&User{},
		&CourseDB{},
		&CourseReview{},
		&UserCourseScore{},
		&UserCourseHole{},
		&UserActivity{},
	)

	if err != nil {
		return err
	}

	log.Printf("✅ Database migration completed")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

// GenerateCourseHash creates a deterministic hash from course name and address
func GenerateCourseHash(name, address string) string {
	// Normalize the input strings
	normalizedName := normalizeString(name)
	normalizedAddress := normalizeString(address)

	// Combine name and address without separator
	combined := normalizedName + normalizedAddress

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(combined))

	// Return first 16 characters of hex string (64-bit hash equivalent)
	return hex.EncodeToString(hash[:])[:16]
}

// normalizeString cleans and standardizes a string for hashing
func normalizeString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove extra whitespace
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

	// Remove common punctuation that might vary
	s = regexp.MustCompile(`[.,\-#]`).ReplaceAllString(s, "")

	// Replace common abbreviations to standardize
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

// BeforeCreate GORM hook to automatically generate hash before saving
func (c *CourseDB) BeforeCreate(tx *gorm.DB) error {
	if c.Hash == "" {
		c.Hash = GenerateCourseHash(c.Name, c.Address)
	}
	return nil
}

// BeforeUpdate GORM hook to regenerate hash if name or address changes
func (c *CourseDB) BeforeUpdate(tx *gorm.DB) error {
	// Only regenerate hash if name or address changed
	if tx.Statement.Changed("Name") || tx.Statement.Changed("Address") {
		c.Hash = GenerateCourseHash(c.Name, c.Address)
	}
	return nil
}
