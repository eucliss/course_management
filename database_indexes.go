package main

import (
	"fmt"
	"log"
)

// CreatePerformanceIndexes adds critical database indexes for performance
func CreatePerformanceIndexes() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	log.Printf("🚀 Creating performance indexes...")

	indexes := []string{
		// Course indexes
		"CREATE INDEX IF NOT EXISTS idx_course_dbs_created_by ON course_dbs(created_by)",
		"CREATE INDEX IF NOT EXISTS idx_course_dbs_name ON course_dbs(name)",
		"CREATE INDEX IF NOT EXISTS idx_course_dbs_created_at ON course_dbs(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_course_dbs_hash ON course_dbs(hash)", // Already unique, but explicit

		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)", // Already unique, but explicit
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",         // Already unique, but explicit

		// Review system indexes
		"CREATE INDEX IF NOT EXISTS idx_course_reviews_course_id ON course_reviews(course_id)",
		"CREATE INDEX IF NOT EXISTS idx_course_reviews_user_id ON course_reviews(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_course_reviews_course_user ON course_reviews(course_id, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_course_reviews_created_at ON course_reviews(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_course_reviews_rating ON course_reviews(overall_rating)",

		// Activity system indexes (if UserActivity table exists)
		"CREATE INDEX IF NOT EXISTS idx_user_activities_user_id ON user_activities(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_activities_user_created ON user_activities(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_user_activities_type ON user_activities(activity_type)",

		// Composite indexes for common queries
		"CREATE INDEX IF NOT EXISTS idx_course_ownership ON course_dbs(created_by, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_course_search ON course_dbs(name, address)",

		// Geospatial indexes for location-based queries
		"CREATE INDEX IF NOT EXISTS idx_course_dbs_location ON course_dbs(latitude, longitude)",

		// JSONB indexes for course data queries (if needed)
		"CREATE INDEX IF NOT EXISTS idx_course_data_gin ON course_dbs USING GIN(course_data)",
	}

	for _, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️ Warning: Failed to create index: %v", err)
			// Continue with other indexes even if one fails
		}
	}

	log.Printf("✅ Performance indexes created successfully")
	return nil
}
