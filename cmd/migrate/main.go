package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Import the main package functions (we'll need to adjust this)
// For now, let's create a standalone version

func main() {
	fmt.Println("🔄 Course Migration Tool")
	fmt.Println("========================")

	// Load environment variables from current directory or parent
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	// Find the project root (where courses directory is)
	if _, err := os.Stat("courses"); os.IsNotExist(err) {
		// Try going up directories to find courses
		if err := os.Chdir("../.."); err != nil {
			log.Fatalf("Failed to change directory: %v", err)
		}
	}

	fmt.Println("📁 Current directory:", getCurrentDir())
	fmt.Println("📂 Checking courses directory...")

	// Check if courses directory exists
	if _, err := os.Stat("courses"); os.IsNotExist(err) {
		log.Fatalf("❌ Courses directory not found")
	}

	// List course files
	files, err := os.ReadDir("courses")
	if err != nil {
		log.Fatalf("❌ Failed to read courses directory: %v", err)
	}

	var courseFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && file.Name() != "schema.json" {
			courseFiles = append(courseFiles, file.Name())
		}
	}

	fmt.Printf("📋 Found %d course files:\n", len(courseFiles))
	for _, file := range courseFiles {
		fmt.Printf("   - %s\n", file)
	}

	fmt.Println("\n🚀 To migrate these courses to your database:")
	fmt.Println("1. Make sure your database credentials are set in .env")
	fmt.Println("2. Start your application: go run .")
	fmt.Println("3. Run the migration script: ./migrate_courses.sh")
	fmt.Println("\nOr use the API directly:")
	fmt.Println("curl -X POST http://localhost:8080/api/migrate/courses")
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}
