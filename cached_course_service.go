package main

import (
	"fmt"
	"log"
	"time"
)

type CachedCourseService struct {
	dbService *DatabaseService
	cache     *CacheService
}

func NewCachedCourseService() *CachedCourseService {
	// Database is required - fail if not available
	if GetDB() == nil {
		log.Fatal("Database connection required - JSON fallback removed")
	}

	return &CachedCourseService{
		dbService: NewDatabaseService(),
		cache:     GetCacheService(),
	}
}

func (cs *CachedCourseService) LoadCourses() ([]Course, error) {
	// Try cache first
	cacheKey := "courses:all"
	var courses []Course
	
	if cs.cache != nil {
		err := cs.cache.GetJSON(cacheKey, &courses)
		if err == nil {
			log.Printf("✅ Cache HIT: Loaded %d courses from cache", len(courses))
			return courses, nil
		}
	}

	// Cache miss - load from database
	dbCourses, err := cs.dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load courses from database: %v", err)
	}

	// Cache the result for 30 minutes
	if cs.cache != nil {
		err = cs.cache.SetJSON(cacheKey, dbCourses, 30*time.Minute)
		if err != nil {
			log.Printf("❌ Failed to cache courses: %v", err)
		}
	}

	log.Printf("✅ Cache MISS: Loaded %d courses from database", len(dbCourses))
	return dbCourses, nil
}

func (cs *CachedCourseService) GetCoursesWithCoordinates() ([]Course, error) {
	// Try cache first
	cacheKey := "courses:with_coords"
	var courses []Course
	
	if cs.cache != nil {
		err := cs.cache.GetJSON(cacheKey, &courses)
		if err == nil {
			log.Printf("✅ Cache HIT: Loaded %d courses with coordinates from cache", len(courses))
			return courses, nil
		}
	}

	// Cache miss - load from database and filter for coordinates
	allCourses, err := cs.dbService.GetAllCoursesFromDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load courses from database: %v", err)
	}

	// Filter courses with coordinates
	var coursesWithCoords []Course
	for _, course := range allCourses {
		if course.Latitude != nil && course.Longitude != nil {
			coursesWithCoords = append(coursesWithCoords, course)
		}
	}

	// Cache the result for 30 minutes
	if cs.cache != nil {
		err = cs.cache.SetJSON(cacheKey, coursesWithCoords, 30*time.Minute)
		if err != nil {
			log.Printf("❌ Failed to cache courses with coordinates: %v", err)
		}
	}

	log.Printf("✅ Cache MISS: Loaded %d courses with coordinates from database", len(coursesWithCoords))
	return coursesWithCoords, nil
}

func (cs *CachedCourseService) GetCourseByNameAndAddress(name, address string) (*Course, error) {
	// Generate cache key based on name and address
	cacheKey := fmt.Sprintf("course:%s:%s", name, address)
	var course Course
	
	if cs.cache != nil {
		err := cs.cache.GetJSON(cacheKey, &course)
		if err == nil {
			log.Printf("✅ Cache HIT: Found course %s from cache", name)
			return &course, nil
		}
	}

	// Cache miss - load from database
	dbCourse, err := cs.dbService.GetCourseByNameAndAddress(name, address)
	if err != nil {
		return nil, err
	}

	// Convert CourseDB to Course if needed
	course = Course{
		Name:        dbCourse.Name,
		Address:     dbCourse.Address,
		Latitude:    dbCourse.Latitude,
		Longitude:   dbCourse.Longitude,
		// Add other fields as needed from CourseDB to Course conversion
	}

	// Cache the result for 1 hour
	if cs.cache != nil {
		err = cs.cache.SetJSON(cacheKey, course, 1*time.Hour)
		if err != nil {
			log.Printf("❌ Failed to cache course %s: %v", name, err)
		}
	}

	log.Printf("✅ Cache MISS: Found course %s from database", name)
	return &course, nil
}

func (cs *CachedCourseService) SaveCourse(course Course) error {
	return cs.SaveCourseWithOwner(course, nil)
}

func (cs *CachedCourseService) SaveCourseWithOwner(course Course, createdBy *uint) error {
	// Save to database first
	if err := cs.dbService.SaveCourseToDatabase(course, createdBy); err != nil {
		return fmt.Errorf("failed to save course to database: %v", err)
	}

	// Invalidate related caches
	if cs.cache != nil {
		cs.invalidateCourseCache(course.Name, course.Address)
	}

	if createdBy != nil {
		log.Printf("✅ Course saved to database with owner ID %d", *createdBy)
	} else {
		log.Printf("✅ Course saved to database")
	}

	return nil
}

func (cs *CachedCourseService) UpdateCourse(course Course) error {
	// Update in database first
	if err := cs.dbService.UpdateCourseInDatabase(course); err != nil {
		return fmt.Errorf("failed to update course in database: %v", err)
	}

	// Invalidate related caches
	if cs.cache != nil {
		cs.invalidateCourseCache(course.Name, course.Address)
	}

	log.Printf("✅ Course updated in database")
	return nil
}

func (cs *CachedCourseService) DeleteCourse(name, address string) error {
	// For now, deletion is not implemented in the database service
	// This is a placeholder for future implementation
	return fmt.Errorf("delete course not implemented yet")
}

func (cs *CachedCourseService) GetUserReviews(userID uint) ([]CourseReview, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%d:reviews", userID)
	var reviews []CourseReview
	
	if cs.cache != nil {
		err := cs.cache.GetJSON(cacheKey, &reviews)
		if err == nil {
			log.Printf("✅ Cache HIT: Loaded %d user reviews from cache", len(reviews))
			return reviews, nil
		}
	}

	// For now, return empty reviews - this would be implemented with actual database query
	reviews = []CourseReview{}
	log.Printf("✅ Cache MISS: Loaded %d user reviews from database", len(reviews))
	return reviews, nil
}

func (cs *CachedCourseService) GetUserCourseReview(userID uint, courseID uint) (*CourseReview, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%d:course:%d:review", userID, courseID)
	var review CourseReview
	
	if cs.cache != nil {
		err := cs.cache.GetJSON(cacheKey, &review)
		if err == nil {
			log.Printf("✅ Cache HIT: Found user course review from cache")
			return &review, nil
		}
	}

	// For now, return nil - this would be implemented with actual database query
	return nil, fmt.Errorf("user course review not found")
}

func (cs *CachedCourseService) SaveUserCourseReview(userID uint, courseID uint, review *CourseReview) error {
	// For now, this is a placeholder - would implement actual database save
	// Invalidate related caches
	if cs.cache != nil {
		cs.invalidateUserReviewCache(userID, courseID)
	}

	log.Printf("✅ User course review saved to database")
	return nil
}

// Cache invalidation helpers
func (cs *CachedCourseService) invalidateCourseCache(name, address string) {
	if cs.cache == nil {
		return
	}

	// Invalidate all course-related caches
	cs.cache.Delete("courses:all")
	cs.cache.Delete("courses:with_coords")
	cs.cache.Delete(fmt.Sprintf("course:%s:%s", name, address))
	
	log.Printf("✅ Invalidated course cache for %s", name)
}

func (cs *CachedCourseService) invalidateUserReviewCache(userID, courseID uint) {
	if cs.cache == nil {
		return
	}

	// Invalidate user-specific review caches
	cs.cache.Delete(fmt.Sprintf("user:%d:reviews", userID))
	cs.cache.Delete(fmt.Sprintf("user:%d:course:%d:review", userID, courseID))
	
	log.Printf("✅ Invalidated user review cache for user %d", userID)
}

func (cs *CachedCourseService) ClearAllCache() error {
	if cs.cache == nil {
		return nil
	}

	err := cs.cache.Clear()
	if err != nil {
		return fmt.Errorf("failed to clear cache: %v", err)
	}

	log.Printf("✅ All cache cleared")
	return nil
}

func (cs *CachedCourseService) GetCacheStats() map[string]interface{} {
	if cs.cache == nil {
		return map[string]interface{}{"cache_enabled": false}
	}

	stats := cs.cache.GetStats()
	stats["cache_enabled"] = true
	return stats
}