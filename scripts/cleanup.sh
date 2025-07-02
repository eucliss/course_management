#!/bin/bash

# Cleanup script for Course Management application
echo "🧹 Cleaning up temporary files..."

# Remove build artifacts
rm -f tmp/main tmp/course_management tmp/build-errors.log

# Clean up large log files (keep last 1000 lines)
if [ -f "logs/app.log" ]; then
    tail -n 1000 logs/app.log > logs/app.log.tmp && mv logs/app.log.tmp logs/app.log
    echo "✅ Cleaned up log files"
fi

# Remove any backup files
find . -name "*.bak" -delete
find . -name "*~" -delete

# Clean up Go build cache
go clean -cache

echo "✅ Cleanup complete!"

# cleanup.sh - Remove all entries from users and courses tables
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
cat > cleanup_db.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// Initialize database connection
	err := InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer CloseDB()

	fmt.Println("🔌 Connected to database")

	// Delete all courses
	fmt.Print("🗑️  Deleting all courses... ")
	result := DB.Exec("DELETE FROM course_dbs")
	if result.Error != nil {
		log.Fatalf("Failed to delete courses: %v", result.Error)
	}
	coursesDeleted := result.RowsAffected
	fmt.Printf("✅ %d courses deleted\n", coursesDeleted)

	// Delete all users
	fmt.Print("🗑️  Deleting all users... ")
	result = DB.Exec("DELETE FROM users")
	if result.Error != nil {
		log.Fatalf("Failed to delete users: %v", result.Error)
	}
	usersDeleted := result.RowsAffected
	fmt.Printf("✅ %d users deleted\n", usersDeleted)

	// Reset auto-increment sequences (optional)
	fmt.Print("🔄 Resetting ID sequences... ")
	DB.Exec("ALTER SEQUENCE course_dbs_id_seq RESTART WITH 1")
	DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
	fmt.Println("✅ Sequences reset")

	fmt.Println("")
	fmt.Println("🎉 Database cleanup completed successfully!")
	fmt.Printf("   📊 Total records removed: %d courses + %d users = %d total\n", 
		coursesDeleted, usersDeleted, coursesDeleted+usersDeleted)
}
EOF

# Run the cleanup script
echo "🚀 Executing cleanup..."
go run cleanup_db.go database.go config.go

# Clean up the temporary file
rm -f cleanup_db.go

echo ""
echo "✅ Cleanup script completed!"
echo "📝 Note: The database is now empty and ready for fresh data"