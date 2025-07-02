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

		// Composite indexes for common queries
		"CREATE INDEX IF NOT EXISTS idx_course_ownership ON course_dbs(created_by, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_course_search ON course_dbs(name, address)",

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
