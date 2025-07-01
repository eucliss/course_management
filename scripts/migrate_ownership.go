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
	fmt.Println("🔄 Running Course Ownership Migration")
	fmt.Println("====================================")

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

	// Run migration
	fmt.Println("🔄 Running migration to add ownership fields...")

	err = db.AutoMigrate(
		&User{},
		&CourseDB{},
	)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("📋 Added fields:")
	fmt.Println("   - CourseDB.UpdatedBy (foreign key to User)")
	fmt.Println("   - CourseDB.Updater relationship")
	fmt.Println("")
	fmt.Println("🎯 Next steps:")
	fmt.Println("   1. Update course creation handlers to set CreatedBy")
	fmt.Println("   2. Update course edit handlers to set UpdatedBy")
	fmt.Println("   3. Add ownership validation methods")
}
