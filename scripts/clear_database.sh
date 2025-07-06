#!/bin/bash

# clear_database.sh - Remove all entries from users and courses tables
# This script will completely clear both tables

echo "🧹 Database Cleanup Script"
echo "=========================="
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: Please run this script from the course_management directory"
    exit 1
fi

# Ask for confirmation
echo "⚠️  WARNING: This will delete ALL data from the database!"
echo "   - All users will be removed"
echo "   - All courses will be removed"
echo "   - This action cannot be undone"
echo ""
read -p "Are you sure you want to continue? (type 'yes' to confirm): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "❌ Operation cancelled"
    exit 0
fi

echo ""
echo "🗑️  Starting database cleanup..."

# Create a temporary Go script to clean the database
cat > cleanup_db_temp.go << 'EOF'
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

func main() {
	// Connect to database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔌 Connected to database")

	// Delete all courses
	fmt.Print("🗑️  Deleting all courses... ")
	result := db.Exec("DELETE FROM course_dbs")
	if result.Error != nil {
		log.Fatalf("Failed to delete courses: %v", result.Error)
	}
	coursesDeleted := result.RowsAffected
	fmt.Printf("✅ %d courses deleted\n", coursesDeleted)

	// Delete all users
	fmt.Print("🗑️  Deleting all users... ")
	result = db.Exec("DELETE FROM users")
	if result.Error != nil {
		log.Fatalf("Failed to delete users: %v", result.Error)
	}
	usersDeleted := result.RowsAffected
	fmt.Printf("✅ %d users deleted\n", usersDeleted)

	// Reset auto-increment sequences (optional)
	fmt.Print("🔄 Resetting ID sequences... ")
	db.Exec("ALTER SEQUENCE course_dbs_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
	fmt.Println("✅ Sequences reset")

	fmt.Println("")
	fmt.Println("🎉 Database cleanup completed successfully!")
	fmt.Printf("   📊 Total records removed: %d courses + %d users = %d total\n", 
		coursesDeleted, usersDeleted, coursesDeleted+usersDeleted)
}
EOF

# Run the cleanup script
echo "🚀 Executing cleanup..."
go run cleanup_db_temp.go

# Clean up the temporary file
rm -f cleanup_db_temp.go

echo ""
echo "✅ Database cleanup completed!"
echo "📝 Note: The database is now empty and ready for fresh data" 