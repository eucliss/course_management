package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// DemonstrateHashGeneration shows the complete SHA-256 hash process
func DemonstrateHashGeneration() {
	fmt.Println("🔐 SHA-256 Hash Generation Demonstration")
	fmt.Println("=" + strings.Repeat("=", 60))

	testCases := []struct {
		name    string
		address string
	}{
		{"Pebble Beach Golf Links", "1700 17-Mile Drive, Pebble Beach, CA 93953"},
		{"Augusta National Golf Club", "2604 Washington Rd, Augusta, GA 30904"},
		{"Pine Valley Golf Club", "Pine Valley, NJ 08021"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n📍 Course: %s\n", tc.name)
		fmt.Printf("📍 Address: %s\n", tc.address)

		// Step 1: Show normalization
		normalizedName := normalizeString(tc.name)
		normalizedAddress := normalizeString(tc.address)
		fmt.Printf("🔧 Normalized Name: '%s'\n", normalizedName)
		fmt.Printf("🔧 Normalized Address: '%s'\n", normalizedAddress)

		// Step 2: Show combined input
		combined := normalizedName + normalizedAddress
		fmt.Printf("🔗 Combined Input: '%s'\n", combined)

		// Step 3: Generate full SHA-256 hash
		hashBytes := sha256.Sum256([]byte(combined))
		fullHash := hex.EncodeToString(hashBytes[:])
		fmt.Printf("🔐 Full SHA-256 Hash (64 chars): %s\n", fullHash)

		// Step 4: Show truncated version (current system)
		shortHash := fullHash[:16]
		fmt.Printf("✂️  Truncated Hash (16 chars): %s\n", shortHash)

		// Step 5: Show different length options
		fmt.Printf("📏 Hash Length Options:\n")
		fmt.Printf("   8 chars (32-bit):  %s\n", fullHash[:8])
		fmt.Printf("   16 chars (64-bit): %s\n", fullHash[:16])
		fmt.Printf("   32 chars (128-bit): %s\n", fullHash[:32])
		fmt.Printf("   64 chars (256-bit): %s\n", fullHash)

		fmt.Println(strings.Repeat("-", 60))
	}
}

// GenerateHashWithLength allows you to specify the hash length
func GenerateHashWithLength(name, address string, length int) string {
	normalizedName := normalizeString(name)
	normalizedAddress := normalizeString(address)
	combined := normalizedName + "|" + normalizedAddress

	hash := sha256.Sum256([]byte(combined))
	fullHex := hex.EncodeToString(hash[:])

	// Ensure length doesn't exceed full hash length
	if length > len(fullHex) {
		length = len(fullHex)
	}

	return fullHex[:length]
}

// CompareHashLengths shows collision probability for different lengths
func CompareHashLengths() {
	fmt.Println("\n📊 Hash Length Security Analysis")
	fmt.Println("=" + strings.Repeat("=", 50))

	lengths := []struct {
		chars int
		bits  int
		desc  string
	}{
		{8, 32, "Very Fast, Higher Collision Risk"},
		{16, 64, "Fast, Good for Most Use Cases"},
		{32, 128, "Secure, Enterprise Level"},
		{64, 256, "Maximum Security, Full SHA-256"},
	}

	fmt.Printf("%-6s %-8s %-15s %s\n", "Chars", "Bits", "Combinations", "Description")
	fmt.Println(strings.Repeat("-", 70))

	for _, l := range lengths {
		combinations := fmt.Sprintf("2^%d", l.bits)
		fmt.Printf("%-6d %-8d %-15s %s\n", l.chars, l.bits, combinations, l.desc)
	}

	fmt.Println("\n🎯 For golf courses (~50K worldwide):")
	fmt.Println("   • 8 chars: Risk of collisions")
	fmt.Println("   • 16 chars: Excellent choice (current)")
	fmt.Println("   • 32+ chars: Overkill but ultra-secure")
}

// func main() {
// 	DemonstrateHashGeneration()
// 	CompareHashLengths()

// 	fmt.Println("\n🔍 Quick Test - Generate hashes with different lengths:")
// 	name := "Test Golf Course"
// 	address := "123 Main St, City, State"

// 	for _, length := range []int{8, 16, 32, 64} {
// 		hash := GenerateHashWithLength(name, address, length)
// 		fmt.Printf("  %d chars: %s\n", length, hash)
// 	}
// }
