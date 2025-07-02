package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ScrapedCourse represents the structure from your Python scraper
type ScrapedCourse struct {
	Name    string `json:"course_name"`
	Address string `json:"address"`
}

// ProcessScrapedCourses processes a JSON file of scraped courses and adds them to the database
func ProcessScrapedCourses(filename string) error {
	// Read the JSON file from your Python scraper
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var scrapedCourses []ScrapedCourse
	if err := json.Unmarshal(data, &scrapedCourses); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	log.Printf("📂 Processing %d courses from %s", len(scrapedCourses), filename)

	// Initialize database connection
	if err := InitDatabase(); err != nil {
		log.Printf("⚠️ Database not available: %v", err)
		// Continue with hash generation even without database
	}

	// Get existing hashes to avoid duplicates
	existingHashes := make(map[string]bool)
	if DB != nil {
		hashes, err := GetAllCourseHashes()
		if err != nil {
			log.Printf("⚠️ Could not get existing hashes: %v", err)
		} else {
			existingHashes = hashes
			log.Printf("📊 Found %d existing courses in database", len(existingHashes))
		}
	}

	// Process each scraped course
	var newCourses, skippedCourses, errorCourses int

	for i, course := range scrapedCourses {
		// Validate and normalize data
		name, address, err := ValidateAndNormalizeCourseData(course.Name, course.Address)
		if err != nil {
			log.Printf("❌ [%d/%d] Invalid course data: %v", i+1, len(scrapedCourses), err)
			errorCourses++
			continue
		}

		// Generate hash for this course
		hash := GenerateCourseHash(name, address)

		// Check if course already exists
		if existingHashes[hash] {
			log.Printf("⏭️ [%d/%d] Skipping duplicate: %s (hash: %s)", i+1, len(scrapedCourses), name, hash)
			skippedCourses++
			continue
		}

		// Convert to JSON for storage
		courseJSON, err := json.Marshal(map[string]interface{}{
			"name":    name,
			"address": address,
			"source":  "web_scraper",
		})
		if err != nil {
			log.Printf("❌ [%d/%d] Error marshaling course data: %v", i+1, len(scrapedCourses), err)
			errorCourses++
			continue
		}

		// Add to database if available
		if DB != nil {
			_, err := CreateCourseWithHash(name, address, string(courseJSON), nil)
			if err != nil {
				log.Printf("❌ [%d/%d] Error saving to database: %v", i+1, len(scrapedCourses), err)
				errorCourses++
				continue
			}
			// Add to existing hashes to prevent duplicates in this batch
			existingHashes[hash] = true
			newCourses++
		} else {
			// Just show what would be created
			log.Printf("✅ [%d/%d] Would create: %s (hash: %s)", i+1, len(scrapedCourses), name, hash)
			newCourses++
		}
	}

	// Summary
	fmt.Println("\n📊 Processing Summary:")
	fmt.Printf("  ✅ New courses: %d\n", newCourses)
	fmt.Printf("  ⏭️ Skipped duplicates: %d\n", skippedCourses)
	fmt.Printf("  ❌ Errors: %d\n", errorCourses)
	fmt.Printf("  📂 Total processed: %d\n", len(scrapedCourses))

	return nil
}

// CreateSampleScrapedData creates a sample JSON file to demonstrate the integration
func CreateSampleScrapedData() error {
	sampleCourses := []ScrapedCourse{
		{"Pebble Beach Golf Links", "1700 17-Mile Drive, Pebble Beach, CA 93953"},
		{"Augusta National Golf Club", "2604 Washington Rd, Augusta, GA 30904"},
		{"TPC Sawgrass", "110 Championship Way, Ponte Vedra Beach, FL 32082"},
		{"Pine Valley Golf Club", "Pine Valley, NJ 08021"},
		{"Pine Valley G.C.", "Pine Valley, NJ 08021"},                             // Duplicate (different format)
		{"Pebble Beach Golf Links", "1700 17-Mile Drive, Pebble Beach, CA 93953"}, // Exact duplicate
	}

	data, err := json.MarshalIndent(sampleCourses, "", "  ")
	if err != nil {
		return err
	}

	filename := "sample_scraped_courses.json"
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	log.Printf("📄 Created sample file: %s", filename)
	return nil
}

// Example of how to use this with your Python scraper output
func ExampleScraperIntegration() {
	log.Println("🔗 Course Scraper Integration Example")
	log.Println("=" + fmt.Sprintf("%s", "=============================================="))

	// Create sample data (this would normally come from your Python scraper)
	if err := CreateSampleScrapedData(); err != nil {
		log.Printf("❌ Error creating sample data: %v", err)
		return
	}

	// Process the scraped courses
	if err := ProcessScrapedCourses("sample_scraped_courses.json"); err != nil {
		log.Printf("❌ Error processing courses: %v", err)
		return
	}

	log.Println("✅ Integration example completed!")
}
