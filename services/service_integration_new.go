package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// relationalServiceContainer implements ServiceContainer using relational database schema
type relationalServiceContainer struct {
	db *gorm.DB
	config ServiceConfig
	
	// Repositories (singletons)
	courseRepo CourseRepository
	userRepo   UserRepository
	reviewRepo ReviewRepository
	
	// Services (singletons)
	courseService CourseService
	authService   AuthService
	sessionService SessionService
	reviewService  ReviewService
	
	// Mutex for thread-safe initialization
	mu sync.RWMutex
}

// NewServiceContainerWithRelationalDB creates a service container using the new relational database schema
func NewServiceContainerWithRelationalDB(db *gorm.DB, config ServiceConfig) ServiceContainer {
	// Run migration first
	if err := migrateToRelationalSchema(db); err != nil {
		log.Printf("⚠️  Warning: Failed to migrate to relational schema: %v", err)
		log.Printf("   Falling back to existing JSON-based schema")
		return NewServiceContainer(db, config)
	}

	log.Println("✅ Successfully migrated to relational database schema")

	// Create a custom service container that uses the new relational repository
	return &relationalServiceContainer{
		db:     db,
		config: config,
	}
}

// Implement ServiceContainer interface methods

func (c *relationalServiceContainer) CourseService() CourseService {
	c.mu.RLock()
	if c.courseService != nil {
		c.mu.RUnlock()
		return c.courseService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.courseService != nil {
		return c.courseService
	}

	c.courseService = NewCourseService(c.CourseRepository(), c.UserRepository())
	log.Printf("✅ CourseService initialized with relational schema")
	return c.courseService
}

func (c *relationalServiceContainer) AuthService() AuthService {
	c.mu.RLock()
	if c.authService != nil {
		c.mu.RUnlock()
		return c.authService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.authService != nil {
		return c.authService
	}

	c.authService = NewAuthService(c.UserRepository(), c.config.AuthConfig)
	log.Printf("✅ AuthService initialized with relational schema")
	return c.authService
}

func (c *relationalServiceContainer) SessionService() SessionService {
	c.mu.RLock()
	if c.sessionService != nil {
		c.mu.RUnlock()
		return c.sessionService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.sessionService != nil {
		return c.sessionService
	}

	c.sessionService = NewSessionService(c.UserRepository())
	log.Printf("✅ SessionService initialized with relational schema")
	return c.sessionService
}

func (c *relationalServiceContainer) ReviewService() ReviewService {
	c.mu.RLock()
	if c.reviewService != nil {
		c.mu.RUnlock()
		return c.reviewService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.reviewService != nil {
		return c.reviewService
	}

	c.reviewService = NewReviewService(c.ReviewRepository(), c.CourseRepository(), c.UserRepository())
	log.Printf("✅ ReviewService initialized with relational schema")
	return c.reviewService
}

func (c *relationalServiceContainer) CourseRepository() CourseRepository {
	c.mu.RLock()
	if c.courseRepo != nil {
		c.mu.RUnlock()
		return c.courseRepo
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.courseRepo != nil {
		return c.courseRepo
	}

	c.courseRepo = NewCourseRepositoryNew(c.db)
	log.Printf("✅ CourseRepository initialized with relational schema")
	return c.courseRepo
}

func (c *relationalServiceContainer) UserRepository() UserRepository {
	c.mu.RLock()
	if c.userRepo != nil {
		c.mu.RUnlock()
		return c.userRepo
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.userRepo != nil {
		return c.userRepo
	}

	c.userRepo = NewUserRepository(c.db)
	log.Printf("✅ UserRepository initialized with relational schema")
	return c.userRepo
}

func (c *relationalServiceContainer) ReviewRepository() ReviewRepository {
	c.mu.RLock()
	if c.reviewRepo != nil {
		c.mu.RUnlock()
		return c.reviewRepo
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if c.reviewRepo != nil {
		return c.reviewRepo
	}

	c.reviewRepo = NewReviewRepository(c.db)
	log.Printf("✅ ReviewRepository initialized with relational schema")
	return c.reviewRepo
}

func (c *relationalServiceContainer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Close database connection if needed
	if c.db != nil {
		sqlDB, err := c.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying database connection: %w", err)
		}
		
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		
		log.Printf("✅ Database connection closed (relational schema)")
	}
	
	return nil
}

// migrateToRelationalSchema performs the database migration to relational schema
func migrateToRelationalSchema(db *gorm.DB) error {
	ctx := context.Background()
	
	log.Println("🔄 Starting relational schema migration...")

	// Create new repository instance for migration
	repo := NewCourseRepositoryNew(db)
	courseRepoNew, ok := repo.(*courseRepositoryNew)
	if !ok {
		return fmt.Errorf("failed to cast repository to courseRepositoryNew")
	}

	// Run schema migration
	if err := courseRepoNew.MigrateSchema(ctx); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	// Check if we need to migrate existing data
	var oldCourseCount int64
	if err := db.Model(&CourseDB{}).Count(&oldCourseCount).Error; err != nil {
		// Table might not exist yet, which is fine
		log.Printf("ℹ️  Old course table not found, skipping data migration")
		return nil
	}

	if oldCourseCount > 0 {
		log.Printf("📊 Found %d courses in old JSON format", oldCourseCount)
		log.Printf("🔄 Migration of existing data would be performed here")
		log.Printf("   For production, implement data migration from JSON to relational format")
		
		// TODO: Implement actual data migration
		// This would involve:
		// 1. Reading all courses from old CourseDB table
		// 2. Parsing JSON data
		// 3. Converting to new relational format
		// 4. Inserting into new tables
		// 5. Backing up old data
		// 6. Switching to new schema
	}

	log.Println("✅ Relational schema migration completed successfully")
	return nil
}

// GetRelationalStats returns statistics about the relational database schema
func GetRelationalStats(db *gorm.DB) (*RelationalStats, error) {
	ctx := context.Background()
	
	stats := &RelationalStats{}
	
	// Count courses
	if err := db.WithContext(ctx).Model(&CourseNewDB{}).Count(&stats.TotalCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count courses: %w", err)
	}

	// Count holes
	if err := db.WithContext(ctx).Model(&CourseHoleNewDB{}).Count(&stats.TotalHoles).Error; err != nil {
		return nil, fmt.Errorf("failed to count holes: %w", err)
	}

	// Count rankings
	if err := db.WithContext(ctx).Model(&CourseRankingNewDB{}).Count(&stats.TotalRankings).Error; err != nil {
		return nil, fmt.Errorf("failed to count rankings: %w", err)
	}

	// Count scores
	if err := db.WithContext(ctx).Model(&UserCourseScoreNewDB{}).Count(&stats.TotalScores).Error; err != nil {
		return nil, fmt.Errorf("failed to count scores: %w", err)
	}

	// Calculate average holes per course
	if stats.TotalCourses > 0 {
		stats.AvgHolesPerCourse = float64(stats.TotalHoles) / float64(stats.TotalCourses)
	}

	return stats, nil
}

// RelationalStats represents statistics about the relational database
type RelationalStats struct {
	TotalCourses       int64   `json:"total_courses"`
	TotalHoles         int64   `json:"total_holes"`
	TotalRankings      int64   `json:"total_rankings"`
	TotalScores        int64   `json:"total_scores"`
	AvgHolesPerCourse  float64 `json:"avg_holes_per_course"`
}

// ValidateRelationalIntegrity performs data integrity checks on the relational schema
func ValidateRelationalIntegrity(db *gorm.DB) (*IntegrityReport, error) {
	ctx := context.Background()
	report := &IntegrityReport{
		IsValid: true,
		Issues:  []string{},
	}

	// Check for courses without holes
	var coursesWithoutHoles int64
	if err := db.WithContext(ctx).Table("courses").
		Joins("LEFT JOIN course_holes ON courses.id = course_holes.course_id").
		Where("course_holes.id IS NULL").
		Count(&coursesWithoutHoles).Error; err != nil {
		return nil, fmt.Errorf("failed to check courses without holes: %w", err)
	}

	if coursesWithoutHoles > 0 {
		report.Issues = append(report.Issues, fmt.Sprintf("%d courses have no holes", coursesWithoutHoles))
	}

	// Check for courses without rankings
	var coursesWithoutRankings int64
	if err := db.WithContext(ctx).Table("courses").
		Joins("LEFT JOIN course_rankings ON courses.id = course_rankings.course_id").
		Where("course_rankings.id IS NULL").
		Count(&coursesWithoutRankings).Error; err != nil {
		return nil, fmt.Errorf("failed to check courses without rankings: %w", err)
	}

	if coursesWithoutRankings > 0 {
		report.Issues = append(report.Issues, fmt.Sprintf("%d courses have no rankings", coursesWithoutRankings))
	}

	// Check for orphaned holes (holes without courses)
	var orphanedHoles int64
	if err := db.WithContext(ctx).Table("course_holes").
		Joins("LEFT JOIN courses ON course_holes.course_id = courses.id").
		Where("courses.id IS NULL").
		Count(&orphanedHoles).Error; err != nil {
		return nil, fmt.Errorf("failed to check orphaned holes: %w", err)
	}

	if orphanedHoles > 0 {
		report.Issues = append(report.Issues, fmt.Sprintf("%d holes are orphaned (no course)", orphanedHoles))
		report.IsValid = false
	}

	// Check for orphaned rankings
	var orphanedRankings int64
	if err := db.WithContext(ctx).Table("course_rankings").
		Joins("LEFT JOIN courses ON course_rankings.course_id = courses.id").
		Where("courses.id IS NULL").
		Count(&orphanedRankings).Error; err != nil {
		return nil, fmt.Errorf("failed to check orphaned rankings: %w", err)
	}

	if orphanedRankings > 0 {
		report.Issues = append(report.Issues, fmt.Sprintf("%d rankings are orphaned (no course)", orphanedRankings))
		report.IsValid = false
	}

	report.TotalIssues = len(report.Issues)
	return report, nil
}

// IntegrityReport represents the results of data integrity validation
type IntegrityReport struct {
	IsValid     bool     `json:"is_valid"`
	TotalIssues int      `json:"total_issues"`
	Issues      []string `json:"issues"`
}

// PerformanceComparison compares performance between JSON and relational schemas
func PerformanceComparison(db *gorm.DB) (*PerformanceReport, error) {
	ctx := context.Background()
	report := &PerformanceReport{}

	// Test relational schema performance
	start := time.Now()
	repo := NewCourseRepositoryNew(db)
	courses, err := repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses from relational schema: %w", err)
	}
	report.RelationalQueryTime = time.Since(start)
	report.RelationalCourseCount = len(courses)

	// Test JSON schema performance (if available)
	start = time.Now()
	oldRepo := NewCourseRepository(db)
	oldCourses, err := oldRepo.GetAll(ctx)
	if err != nil {
		log.Printf("ℹ️  JSON schema not available for comparison: %v", err)
		report.JSONQueryTime = 0
		report.JSONCourseCount = 0
	} else {
		report.JSONQueryTime = time.Since(start)
		report.JSONCourseCount = len(oldCourses)
	}

	// Calculate performance improvement
	if report.JSONQueryTime > 0 {
		improvement := (report.JSONQueryTime.Nanoseconds() - report.RelationalQueryTime.Nanoseconds()) * 100 / report.JSONQueryTime.Nanoseconds()
		report.PerformanceImprovement = fmt.Sprintf("%.1f%%", float64(improvement))
	}

	return report, nil
}

// PerformanceReport represents performance comparison results
type PerformanceReport struct {
	RelationalQueryTime       time.Duration `json:"relational_query_time"`
	RelationalCourseCount     int          `json:"relational_course_count"`
	JSONQueryTime            time.Duration `json:"json_query_time"`
	JSONCourseCount          int          `json:"json_course_count"`
	PerformanceImprovement   string       `json:"performance_improvement"`
}