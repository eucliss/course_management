package main

import (
	"fmt"
	"log"
	"os"

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
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Address    string `json:"address"`
	CourseData string `gorm:"type:jsonb" json:"course_data"` // Store existing JSON structure
	CreatedBy  *uint  `json:"created_by"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
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
		DBName:   getEnvOrDefault("DB_NAME", "course_management"),
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

	// Skip database initialization if no password is set (development mode)
	if config.Password == "" {
		log.Printf("🔄 No database password set, skipping database initialization")
		return fmt.Errorf("database credentials not configured")
	}

	// Create connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

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
