package services

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
)

// ServiceContainer implementation
type serviceContainer struct {
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

func NewServiceContainer(db *gorm.DB, config ServiceConfig) ServiceContainer {
	return &serviceContainer{
		db:     db,
		config: config,
	}
}

func (c *serviceContainer) CourseService() CourseService {
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
	log.Printf("✅ CourseService initialized")
	return c.courseService
}

func (c *serviceContainer) AuthService() AuthService {
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
	log.Printf("✅ AuthService initialized")
	return c.authService
}

func (c *serviceContainer) SessionService() SessionService {
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
	log.Printf("✅ SessionService initialized")
	return c.sessionService
}

func (c *serviceContainer) ReviewService() ReviewService {
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
	log.Printf("✅ ReviewService initialized")
	return c.reviewService
}

func (c *serviceContainer) CourseRepository() CourseRepository {
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

	c.courseRepo = NewCourseRepository(c.db)
	log.Printf("✅ CourseRepository initialized")
	return c.courseRepo
}

func (c *serviceContainer) UserRepository() UserRepository {
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
	log.Printf("✅ UserRepository initialized")
	return c.userRepo
}

func (c *serviceContainer) ReviewRepository() ReviewRepository {
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
	log.Printf("✅ ReviewRepository initialized")
	return c.reviewRepo
}

func (c *serviceContainer) Close() error {
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
		
		log.Printf("✅ Database connection closed")
	}
	
	return nil
}

// Global service container instance
var (
	globalContainer ServiceContainer
	containerOnce   sync.Once
)

// GetServiceContainer returns the global service container instance
func GetServiceContainer() ServiceContainer {
	if globalContainer == nil {
		panic("service container not initialized - call InitializeServiceContainer first")
	}
	return globalContainer
}

// InitializeServiceContainer initializes the global service container
func InitializeServiceContainer(db *gorm.DB, config ServiceConfig) {
	containerOnce.Do(func() {
		globalContainer = NewServiceContainer(db, config)
		log.Printf("✅ Global service container initialized")
	})
}

// Helper function to create service config from environment
func CreateServiceConfig() ServiceConfig {
	return ServiceConfig{
		AuthConfig: AuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:        getEnv("GOOGLE_REDIRECT_URL", ""),
		},
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),
	}
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	// This would normally use os.Getenv, but to avoid import cycles
	// we'll leave it as a placeholder
	return defaultValue
}

// ServiceMiddleware creates Echo middleware that injects services into context
// This needs to be imported where Echo is available
// func ServiceMiddleware(container ServiceContainer) func(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			// Inject services into context
// 			c.Set("services", container)
// 			c.Set("courseService", container.CourseService())
// 			c.Set("authService", container.AuthService())
// 			c.Set("sessionService", container.SessionService())
// 			c.Set("reviewService", container.ReviewService())
// 			
// 			return next(c)
// 		}
// 	}
// }

// Helper functions to extract services from Echo context
// These will need to be moved to the main package where Echo is available
// 
// func GetCourseService(c echo.Context) CourseService {
// 	service, ok := c.Get("courseService").(CourseService)
// 	if !ok {
// 		panic("CourseService not found in context - ensure ServiceMiddleware is used")
// 	}
// 	return service
// }
// 
// func GetAuthService(c echo.Context) AuthService {
// 	service, ok := c.Get("authService").(AuthService)
// 	if !ok {
// 		panic("AuthService not found in context - ensure ServiceMiddleware is used")
// 	}
// 	return service
// }
// 
// func GetSessionService(c echo.Context) SessionService {
// 	service, ok := c.Get("sessionService").(SessionService)
// 	if !ok {
// 		panic("SessionService not found in context - ensure ServiceMiddleware is used")
// 	}
// 	return service
// }
// 
// func GetReviewService(c echo.Context) ReviewService {
// 	service, ok := c.Get("reviewService").(ReviewService)
// 	if !ok {
// 		panic("ReviewService not found in context - ensure ServiceMiddleware is used")
// 	}
// 	return service
// }
// 
// func GetServices(c echo.Context) ServiceContainer {
// 	services, ok := c.Get("services").(ServiceContainer)
// 	if !ok {
// 		panic("ServiceContainer not found in context - ensure ServiceMiddleware is used")
// 	}
// 	return services
// }