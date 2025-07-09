// Package services provides a clean service layer architecture for the golf course management system
package services

// This file serves as the main export point for the services package
// It re-exports all the key interfaces and types for easy importing

// Re-export key interfaces
type (
	// Service interfaces
	CourseServiceInterface  = CourseService
	AuthServiceInterface    = AuthService
	SessionServiceInterface = SessionService
	ReviewServiceInterface  = ReviewService
	
	// Repository interfaces
	CourseRepositoryInterface = CourseRepository
	UserRepositoryInterface   = UserRepository
	ReviewRepositoryInterface = ReviewRepository
	
	// Container interface
	ServiceContainerInterface = ServiceContainer
)

// Re-export key types
type (
	// Domain models
	CourseModel      = Course
	HoleModel        = Hole
	ScoreModel       = Score
	RankingModel     = Ranking
	GoogleUserModel  = GoogleUser
	CourseReviewModel = CourseReview
	
	// Service types
	UserCourseScoreModel = UserCourseScore
	UserCourseHoleModel  = UserCourseHole
	AuthConfigModel      = AuthConfig
	ServiceConfigModel   = ServiceConfig
)

// Package constants
const (
	// Version of the service layer
	ServiceLayerVersion = "1.0.0"
	
	// Default pagination limits
	DefaultPageSize = 20
	MaxPageSize     = 100
	
	// Validation constants
	MinCourseName        = 3
	MaxCourseName        = 100
	MinCourseAddress     = 10
	MaxCourseAddress     = 200
	MinReviewLength      = 10
	MaxReviewLength      = 2000
	MinRating            = 1
	MaxRating            = 5
	MinScore             = 1
	MaxScore             = 20
	MinHandicap          = -5
	MaxHandicap          = 40
	MinPar               = 3
	MaxPar               = 6
	MinHole              = 1
	MaxHole              = 18
	MaxYardage           = 800
)

// Package-level utility functions
func init() {
	// Initialize any package-level resources if needed
}

// ValidatePageSize validates pagination parameters
func ValidatePageSize(pageSize int) int {
	if pageSize <= 0 {
		return DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return MaxPageSize
	}
	return pageSize
}

// ValidateOffset validates pagination offset
func ValidateOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

// Package documentation
/*
Package services provides a comprehensive service layer architecture for the golf course management system.

## Overview

This package implements a clean architecture pattern with the following layers:

1. **Domain Models**: Core business entities (Course, User, Review, etc.)
2. **Repository Layer**: Data access interfaces and implementations
3. **Service Layer**: Business logic and use cases
4. **Container**: Dependency injection and service lifecycle management

## Key Components

### Services
- CourseService: Manages course CRUD operations, validation, and permissions
- AuthService: Handles Google OAuth authentication and user management
- SessionService: Manages user sessions and authentication state
- ReviewService: Manages course reviews, ratings, and scoring

### Repositories
- CourseRepository: Data access for course entities
- UserRepository: Data access for user entities
- ReviewRepository: Data access for review and scoring entities

### Container
- ServiceContainer: Dependency injection container with singleton service management

## Usage

### Basic Setup

```go
// Initialize the service container
db := // your GORM database connection
config := CreateServiceConfig()
container := NewServiceContainer(db, config)

// Use in Echo middleware
e.Use(ServiceMiddleware(container))
```

### Using Services in Handlers

```go
func (h *Handlers) SomeHandler(c echo.Context) error {
    ctx := c.Request().Context()
    
    // Get service from context
    courseService := services.GetCourseService(c)
    
    // Use service methods
    courses, err := courseService.GetAllCourses(ctx)
    if err != nil {
        return err
    }
    
    // ... rest of handler logic
}
```

### Testing

Services are designed to be easily testable with interface mocking:

```go
func TestCourseService_CreateCourse(t *testing.T) {
    // Mock repository
    mockRepo := &MockCourseRepository{}
    
    // Create service with mock
    service := NewCourseService(mockRepo, mockUserRepo)
    
    // Test the service
    err := service.CreateCourse(ctx, course, userID)
    assert.NoError(t, err)
}
```

## Benefits

1. **Clean Architecture**: Clear separation of concerns between layers
2. **Testability**: Interface-based design enables easy mocking
3. **Dependency Injection**: Centralized service management
4. **Validation**: Built-in business logic validation
5. **Error Handling**: Consistent error handling patterns
6. **Performance**: Singleton services reduce initialization overhead

## Migration Guide

To migrate existing handlers to use the service layer:

1. Remove direct database service creation
2. Use GetXXXService(c) to get services from context
3. Use context.Context for all service calls
4. Update main.go to use ServiceMiddleware
5. Initialize the service container in main.go

See handlers_service_layer.go for complete examples.
*/