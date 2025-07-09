package testing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContext creates a context with timeout for tests
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// AssertNoError is a helper for asserting no error with additional context
func AssertNoError(t *testing.T, err error, msg string) {
	require.NoError(t, err, msg)
}

// AssertError is a helper for asserting error existence with additional context
func AssertError(t *testing.T, err error, msg string) {
	require.Error(t, err, msg)
}

// AssertEqual is a helper for asserting equality with additional context
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	assert.Equal(t, expected, actual, msg)
}

// AssertNotNil is a helper for asserting non-nil values
func AssertNotNil(t *testing.T, value interface{}, msg string) {
	assert.NotNil(t, value, msg)
}

// AssertNil is a helper for asserting nil values
func AssertNil(t *testing.T, value interface{}, msg string) {
	assert.Nil(t, value, msg)
}

// AssertTrue is a helper for asserting true values
func AssertTrue(t *testing.T, value bool, msg string) {
	assert.True(t, value, msg)
}

// AssertFalse is a helper for asserting false values
func AssertFalse(t *testing.T, value bool, msg string) {
	assert.False(t, value, msg)
}

// AssertContains is a helper for asserting string contains
func AssertContains(t *testing.T, haystack, needle string, msg string) {
	assert.Contains(t, haystack, needle, msg)
}

// AssertLen is a helper for asserting slice/map length
func AssertLen(t *testing.T, object interface{}, length int, msg string) {
	assert.Len(t, object, length, msg)
}

// LogTestStart logs the start of a test with clear formatting
func LogTestStart(t *testing.T, testName string) {
	t.Logf("\n🧪 Starting test: %s", testName)
}

// LogTestEnd logs the end of a test with clear formatting
func LogTestEnd(t *testing.T, testName string) {
	t.Logf("✅ Completed test: %s\n", testName)
}

// CreateTestCourse creates a basic test course for use in tests
func CreateTestCourse(name, address string, userID uint) map[string]interface{} {
	return map[string]interface{}{
		"name":              name,
		"address":           address,
		"description":       "Test course description",
		"review":            "Test review",
		"overall_rating":    "A",
		"latitude":          30.2672,
		"longitude":         -97.7431,
		"created_by":        userID,
		"course_data_json":  `{"holes":[{"number":1,"par":4,"yardage":350}],"scores":[{"hole":1,"score":4}]}`,
		"created_at":        time.Now().Unix(),
		"updated_at":        time.Now().Unix(),
	}
}

// CreateTestUser creates a basic test user for use in tests
func CreateTestUser(googleID, email, name string) map[string]interface{} {
	return map[string]interface{}{
		"google_id":  googleID,
		"email":      email,
		"name":       name,
		"picture":    "https://example.com/pic.jpg",
		"handicap":   15.0,
		"created_at": time.Now().Unix(),
		"updated_at": time.Now().Unix(),
	}
}

// CreateTestReview creates a basic test review for use in tests
func CreateTestReview(userID uint, courseName, courseAddress string) map[string]interface{} {
	return map[string]interface{}{
		"user_id":         userID,
		"course_name":     courseName,
		"course_address":  courseAddress,
		"review":          "Test review text",
		"rating":          4,
		"score_data_json": `{"totalScore":80,"holeScores":[{"hole":1,"score":4,"par":4}]}`,
		"created_at":      time.Now().Unix(),
		"updated_at":      time.Now().Unix(),
	}
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SetupTestEnvironment sets up the global test environment
func SetupTestEnvironment() {
	// Set test environment variables
	os.Setenv("GO_ENV", "test")
	os.Setenv("LOG_LEVEL", "error")
}

// CleanupTestEnvironment cleans up the global test environment
func CleanupTestEnvironment() {
	// Clean up test environment variables
	os.Unsetenv("GO_ENV")
	os.Unsetenv("LOG_LEVEL")
}