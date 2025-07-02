package main

import (
	"fmt"
	"log"
	"strings"
)

// TestCourseHashing demonstrates the hash functionality
func TestCourseHashing() {
	fmt.Println("🧪 Testing Course Hash Generation")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Test cases with various course names and addresses
	testCases := []struct {
		name    string
		address string
	}{
		{"Pebble Beach Golf Links", "1700 17-Mile Drive, Pebble Beach, CA 93953"},
		{"Augusta National Golf Club", "2604 Washington Rd, Augusta, GA 30904"},
		{"St. Andrews Links - Old Course", "Pilmour Cottage, St Andrews KY16 9SF, Scotland"},
		{"TPC Sawgrass", "110 Championship Way, Ponte Vedra Beach, FL 32082"},
		{"Bethpage Black", "99 Quaker Meeting House Rd, Farmingdale, NY 11735"},
		// Test case with variations to show normalization
		{"Pine Valley Golf Club", "Pine Valley, NJ 08021"},
		{"Pine Valley Golf Club", "Pine Valley, New Jersey 08021"}, // Should produce same hash
		{"Pine Valley G.C.", "Pine Valley, NJ 08021"},              // Should produce same hash
	}

	fmt.Printf("%-40s %-50s %s\n", "Course Name", "Address", "Generated Hash")
	fmt.Println(strings.Repeat("-", 110))

	hashMap := make(map[string][]string) // Track duplicate hashes

	for _, tc := range testCases {
		hash := GenerateCourseHash(tc.name, tc.address)
		fmt.Printf("%-40s %-50s %s\n", tc.name, tc.address, hash)

		// Track for duplicate detection
		hashMap[hash] = append(hashMap[hash], tc.name)
	}

	// Check for duplicates (this demonstrates normalization working)
	fmt.Println("\n🔍 Duplicate Hash Detection:")
	for hash, names := range hashMap {
		if len(names) > 1 {
			fmt.Printf("Hash %s appears %d times:\n", hash, len(names))
			for _, name := range names {
				fmt.Printf("  - %s\n", name)
			}
		}
	}

	fmt.Println("\n✅ Hash testing complete!")
}

// TestHashNormalization shows how the normalization works
func TestHashNormalization() {
	fmt.Println("\n🔧 Testing String Normalization")
	fmt.Println("=" + strings.Repeat("=", 50))

	testStrings := []string{
		"Pebble Beach Golf Course",
		"PEBBLE BEACH GOLF COURSE",
		"Pebble Beach Golf Club",
		"Pebble Beach G.C.",
		"Pebble  Beach   Golf Course", // Extra spaces
		"123 Main Street",
		"123 Main St.",
		"123 Main St",
		"North Carolina Golf Club",
		"N. Carolina Golf Club",
	}

	fmt.Printf("%-35s -> %s\n", "Original", "Normalized")
	fmt.Println(strings.Repeat("-", 70))

	for _, str := range testStrings {
		normalized := normalizeString(str)
		fmt.Printf("%-35s -> %s\n", str, normalized)
	}
}

// RunHashTests runs all hash-related tests
func RunHashTests() {
	log.Println("🚀 Starting course hash tests...")
	TestCourseHashing()
	TestHashNormalization()
	log.Println("🏁 Course hash tests completed!")
}
