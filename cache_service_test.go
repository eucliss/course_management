package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheService_MemoryCache(t *testing.T) {
	// Create cache service with memory cache only
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false, // Disable Redis for testing
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	require.NotNil(t, cache)

	// Test Set and Get
	testKey := "test:key"
	testData := []byte("test data")
	
	err := cache.Set(testKey, testData, 5*time.Minute)
	assert.NoError(t, err)

	retrievedData, err := cache.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, testData, retrievedData)

	// Test cache miss
	_, err = cache.Get("nonexistent:key")
	assert.Error(t, err)

	// Test Delete
	err = cache.Delete(testKey)
	assert.NoError(t, err)

	_, err = cache.Get(testKey)
	assert.Error(t, err)
}

func TestCacheService_JSONOperations(t *testing.T) {
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	require.NotNil(t, cache)

	// Test JSON operations
	testData := map[string]any{
		"id":   1,
		"name": "Test Course",
		"lat":  40.7128,
		"lng":  -74.0060,
	}

	err := cache.SetJSON("test:json", testData, 5*time.Minute)
	assert.NoError(t, err)

	var retrievedData map[string]any
	err = cache.GetJSON("test:json", &retrievedData)
	assert.NoError(t, err)
	assert.Equal(t, testData["name"], retrievedData["name"])
	assert.Equal(t, testData["lat"], retrievedData["lat"])
}

func TestCacheService_TTLExpiration(t *testing.T) {
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	require.NotNil(t, cache)

	// Set data with very short TTL
	testKey := "test:ttl"
	testData := []byte("test data with short ttl")
	
	err := cache.Set(testKey, testData, 100*time.Millisecond)
	assert.NoError(t, err)

	// Should be available immediately
	retrievedData, err := cache.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, testData, retrievedData)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(testKey)
	assert.Error(t, err)
}

func TestCacheService_Clear(t *testing.T) {
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	require.NotNil(t, cache)

	// Set multiple keys
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := cache.Set(key, []byte("data for "+key), 5*time.Minute)
		assert.NoError(t, err)
	}

	// Verify all keys exist
	for _, key := range keys {
		_, err := cache.Get(key)
		assert.NoError(t, err)
	}

	// Clear cache
	err := cache.Clear()
	assert.NoError(t, err)

	// Verify all keys are gone
	for _, key := range keys {
		_, err := cache.Get(key)
		assert.Error(t, err)
	}
}

func TestCacheService_PatternDelete(t *testing.T) {
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	require.NotNil(t, cache)

	// Set keys with different prefixes
	courseKeys := []string{"course:1", "course:2", "course:3"}
	userKeys := []string{"user:1", "user:2"}
	
	for _, key := range courseKeys {
		err := cache.Set(key, []byte("course data"), 5*time.Minute)
		assert.NoError(t, err)
	}
	
	for _, key := range userKeys {
		err := cache.Set(key, []byte("user data"), 5*time.Minute)
		assert.NoError(t, err)
	}

	// Delete course keys with pattern
	err := cache.DeletePattern("course:*")
	assert.NoError(t, err)

	// Verify course keys are deleted
	for _, key := range courseKeys {
		_, err := cache.Get(key)
		assert.Error(t, err)
	}

	// Verify user keys still exist
	for _, key := range userKeys {
		_, err := cache.Get(key)
		assert.NoError(t, err)
	}
}

func TestCacheService_Stats(t *testing.T) {
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	require.NotNil(t, cache)

	// Add some data
	cache.Set("key1", []byte("data1"), 5*time.Minute)
	cache.Set("key2", []byte("data2"), 5*time.Minute)

	stats := cache.GetStats()
	assert.NotNil(t, stats)
	// Memory items count may vary due to test execution order
	assert.GreaterOrEqual(t, stats["memory_items"], 2)
	// Check fallback mode exists in stats
	assert.Contains(t, stats, "fallback_mode")
}

func TestCachedCourseService_LoadCourses(t *testing.T) {
	// Initialize database for testing
	if err := InitDatabase(); err != nil {
		t.Skipf("Database not available: %v", err)
	}

	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	InitCacheService(config)

	service := NewCachedCourseService()
	require.NotNil(t, service)

	// First call should hit database
	courses1, err := service.LoadCourses()
	assert.NoError(t, err)
	assert.NotNil(t, courses1)

	// Second call should hit cache
	courses2, err := service.LoadCourses()
	assert.NoError(t, err)
	assert.NotNil(t, courses2)
	assert.Equal(t, len(courses1), len(courses2))
}

func TestCachedCourseService_GetCoursesWithCoordinates(t *testing.T) {
	// Initialize database for testing
	if err := InitDatabase(); err != nil {
		t.Skipf("Database not available: %v", err)
	}

	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	InitCacheService(config)

	service := NewCachedCourseService()
	require.NotNil(t, service)

	// First call should hit database
	courses, err := service.GetCoursesWithCoordinates()
	assert.NoError(t, err)
	assert.NotNil(t, courses)

	// Verify all courses have coordinates
	for _, course := range courses {
		assert.NotNil(t, course.Latitude)
		assert.NotNil(t, course.Longitude)
	}
}

func TestCachedCourseService_CacheInvalidation(t *testing.T) {
	// Initialize database for testing
	if err := InitDatabase(); err != nil {
		t.Skipf("Database not available: %v", err)
	}

	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	cache := InitCacheService(config)
	service := NewCachedCourseService()
	require.NotNil(t, service)

	// Load courses to populate cache
	courses, err := service.LoadCourses()
	assert.NoError(t, err)
	assert.NotEmpty(t, courses)

	// Verify cache is populated
	_, err = cache.Get("courses:all")
	assert.NoError(t, err)

	// Save a course (should invalidate cache)
	if len(courses) > 0 {
		err = service.SaveCourse(courses[0])
		assert.NoError(t, err)

		// Cache should be invalidated
		_, err = cache.Get("courses:all")
		assert.Error(t, err)
	}
}

func TestCachedCourseService_CacheStats(t *testing.T) {
	config := &CacheConfig{
		RedisURL:     "redis://localhost:6379",
		EnableRedis:  false,
		EnableMemory: true,
		DefaultTTL:   30 * time.Minute,
		MaxMemoryMB:  100,
	}

	InitCacheService(config)

	service := NewCachedCourseService()
	require.NotNil(t, service)

	stats := service.GetCacheStats()
	assert.NotNil(t, stats)
	assert.Equal(t, true, stats["cache_enabled"])
}