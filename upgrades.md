# Course Management System Upgrades

This document outlines the optimization and restructuring tasks needed to improve the project's architecture, performance, and maintainability.

## 1. Database Schema & Data Architecture

### Issues:
- [ ] Mixing JSON files with PostgreSQL creates complexity. Remove JSON file dependencies.
- [ ] Course data stored as JSON strings in database
- [ ] Inconsistent data access patterns
- [ ] No proper relational structure for course holes

### Tasks:
- [ ] Create proper Course model with individual fields instead of JSON storage
- [ ] Add separate Hole model with foreign key relationship
- [ ] Migrate existing JSON course data to relational schema
- [ ] Add database constraints and indexes
- [ ] Remove JSON file dependencies
- [ ] Add proper course location fields (city, state, zipcode)

### Implementation:
```go
// Replace existing CourseDB with:
type Course struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"not null;index"`
    Address     string    `gorm:"not null"`
    City        string    `gorm:"index"`
    State       string    `gorm:"index"`
    ZipCode     string    
    Latitude    *float64
    Longitude   *float64
    Phone       string
    Website     string
    Holes       []Hole    `gorm:"foreignKey:CourseID"`
    CreatedBy   *uint
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Hole struct {
    ID       uint `gorm:"primaryKey"`
    CourseID uint `gorm:"not null;index"`
    Number   int  `gorm:"not null"`
    Par      int
    Yardage  int
    Description string
}
```

## 2. Service Layer Architecture

### Issues:
- [ ] Business logic mixed in handlers
- [ ] Direct database calls from handlers
- [ ] No clear separation of concerns
- [ ] No dependency injection

### Tasks:
- [ ] Create service layer interfaces
- [ ] Implement repository pattern
- [ ] Extract business logic from handlers
- [ ] Add dependency injection container
- [ ] Create service structs for each domain (Course, Review, User)
- [ ] Add service tests

### Implementation:
```go
// Create service layer
type CourseService struct {
    repo CourseRepository
}

type CourseRepository interface {
    Create(course *Course) error
    GetByID(id uint) (*Course, error)
    GetByNameAndAddress(name, address string) (*Course, error)
    Search(filters CourseFilters) ([]Course, error)
}

type Services struct {
    Course CourseService
    Review ReviewService
    User   UserService
}
```

## 3. API Design & Response Optimization

### Issues:
- [ ] Server-side rendering limits scalability
- [ ] No API versioning
- [ ] Inconsistent response formats
- [ ] No standardized error handling

### Tasks:
- [ ] Add REST API endpoints alongside existing routes
- [ ] Implement API versioning (v1)
- [ ] Create standardized response format
- [ ] Add API documentation (OpenAPI/Swagger)
- [ ] Implement proper HTTP status codes
- [ ] Add content negotiation (JSON/HTML)

### Implementation:
```go
// Add REST API endpoints
func (h *Handlers) SetupAPIRoutes(e *echo.Echo) {
    api := e.Group("/api/v1")
    
    // Courses
    api.GET("/courses", h.GetCoursesAPI)
    api.GET("/courses/:id", h.GetCourseAPI)
    api.POST("/courses", h.CreateCourseAPI)
    
    // Reviews
    api.GET("/courses/:id/reviews", h.GetCourseReviewsAPI)
    api.POST("/courses/:id/reviews", h.CreateReviewAPI)
}

// Standardized response format
type APIResponse struct {
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Message string      `json:"message,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}
```

## 4. Performance Optimizations

### Issues:
- [ ] No caching layer
- [ ] Inefficient course lookups
- [ ] No pagination on large datasets
- [ ] Missing database indexes
- [ ] No query optimization

### Tasks:
- [ ] Add Redis caching layer
- [ ] Implement proper pagination
- [ ] Add database indexes for common queries
- [ ] Optimize N+1 query problems
- [ ] Add query result caching
- [ ] Implement lazy loading for related data
- [ ] Add connection pooling configuration

### Implementation:
```go
// Add Redis caching
type CacheService struct {
    client *redis.Client
}

// Implement proper pagination
type PaginationParams struct {
    Page   int `query:"page" validate:"min=1"`
    Limit  int `query:"limit" validate:"min=1,max=100"`
    Sort   string `query:"sort"`
    Order  string `query:"order" validate:"oneof=asc desc"`
}

// Add database indexes
func CreateIndexes(db *gorm.DB) error {
    indexes := []string{
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courses_location ON courses(city, state)",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courses_name_address ON courses(name, address)",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_course_user ON course_reviews(course_id, user_id)",
    }
}
```

## 5. Security Improvements

### Issues:
- [ ] No input validation
- [ ] Missing rate limiting
- [ ] No request logging
- [ ] Insufficient authorization checks
- [ ] No CSRF protection

### Tasks:
- [ ] Add input validation middleware
- [ ] Implement rate limiting
- [ ] Add request/response logging
- [ ] Enhance authorization checks
- [ ] Add CSRF protection
- [ ] Implement API key authentication
- [ ] Add request sanitization

### Implementation:
```go
// Add validation middleware
func ValidateMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            if err := c.Validate(req); err != nil {
                return c.JSON(400, APIResponse{Error: "Invalid input"})
            }
            return next(c)
        }
    }
}

// Input validation structs
type CreateCourseRequest struct {
    Name    string `json:"name" validate:"required,min=3,max=100"`
    Address string `json:"address" validate:"required,min=10,max=200"`
    City    string `json:"city" validate:"required,min=2,max=50"`
    State   string `json:"state" validate:"required,len=2"`
}
```

## 6. Code Organization & Structure

### Issues:
- [ ] Large handler files (1500+ lines)
- [ ] Mixed concerns in single files
- [ ] No clear module boundaries
- [ ] Scripts scattered in single directory

### Tasks:
- [ ] Restructure project into clean architecture
- [ ] Split large files into smaller modules
- [ ] Organize by domain, not by type
- [ ] Create proper package structure
- [ ] Move scripts to appropriate locations
- [ ] Add internal and pkg directories

### Implementation:
```
project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes/
│   ├── services/
│   ├── repositories/
│   ├── models/
│   └── config/
├── pkg/
│   ├── auth/
│   ├── cache/
│   └── utils/
├── migrations/
├── tests/
└── docs/
```

## 7. Testing Infrastructure

### Issues:
- [ ] No automated tests
- [ ] No test database setup
- [ ] No CI/CD pipeline
- [ ] No mocking framework

### Tasks:
- [ ] Add unit tests for services
- [ ] Add integration tests for API endpoints
- [ ] Create test database setup
- [ ] Add mocking for external dependencies
- [ ] Implement test coverage reporting
- [ ] Add GitHub Actions CI/CD
- [ ] Create test fixtures and factories

### Implementation:
```go
// Add comprehensive tests
func TestCourseService_Create(t *testing.T) {
    db := setupTestDB()
    defer db.Close()
    
    service := NewCourseService(db)
    
    tests := []struct {
        name    string
        course  Course
        wantErr bool
    }{
        // Test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.Create(&tt.course)
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 8. Configuration Management

### Issues:
- [ ] Environment variables scattered
- [ ] No configuration validation
- [ ] No environment-specific configs
- [ ] Hardcoded values in code

### Tasks:
- [ ] Centralize configuration management
- [ ] Add configuration validation
- [ ] Create environment-specific config files
- [ ] Add configuration hot-reloading
- [ ] Implement configuration encryption for secrets
- [ ] Add configuration documentation

### Implementation:
```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Google   GoogleConfig   `mapstructure:"google"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## 9. Monitoring & Observability

### Issues:
- [ ] No structured logging
- [ ] No metrics collection
- [ ] No health check endpoints
- [ ] No error tracking

### Tasks:
- [ ] Add structured logging with logrus/zap
- [ ] Implement metrics collection (Prometheus)
- [ ] Add health check endpoints
- [ ] Integrate error tracking (Sentry)
- [ ] Add distributed tracing
- [ ] Create monitoring dashboards
- [ ] Add alerting rules

### Implementation:
```go
// Health check endpoint
func (h *Handlers) HealthCheck(c echo.Context) error {
    status := map[string]string{
        "status":   "healthy",
        "database": h.checkDatabase(),
        "redis":    h.checkRedis(),
    }
    return c.JSON(200, status)
}

// Structured logging
func NewLogger() *logrus.Logger {
    log := logrus.New()
    log.SetFormatter(&logrus.JSONFormatter{})
    log.SetLevel(logrus.InfoLevel)
    return log
}
```

## 10. Frontend Modernization

### Issues:
- [ ] Heavy server-side rendering
- [ ] Limited interactivity
- [ ] No build process
- [ ] No TypeScript

### Tasks:
- [ ] Add TypeScript for better type safety
- [ ] Implement proper bundling (Vite/Webpack)
- [ ] Add state management for complex interactions
- [ ] Implement Progressive Web App features
- [ ] Add client-side routing
- [ ] Optimize bundle size
- [ ] Add modern CSS framework

### Implementation:
```javascript
// Add TypeScript interfaces
interface Course {
  id: number;
  name: string;
  address: string;
  city: string;
  state: string;
  latitude?: number;
  longitude?: number;
}

// Add proper build process
// package.json
{
  "scripts": {
    "build": "vite build",
    "dev": "vite",
    "type-check": "tsc --noEmit"
  }
}
```

## Implementation Priority

### Phase 1 (High Priority - Critical)
- [ ] Service layer implementation
- [ ] Database schema migration
- [ ] Input validation
- [ ] Basic testing infrastructure
- [ ] Configuration management

### Phase 2 (Medium Priority - Important)
- [ ] API endpoints
- [ ] Caching layer
- [ ] Performance optimizations
- [ ] Security enhancements
- [ ] Code restructuring

### Phase 3 (Low Priority - Nice to Have)
- [ ] Frontend modernization
- [ ] Advanced monitoring
- [ ] CI/CD pipeline
- [ ] Documentation improvements
- [ ] Progressive Web App features

## Estimated Timeline

- **Phase 1**: 3-4 weeks
- **Phase 2**: 4-6 weeks  
- **Phase 3**: 2-3 weeks

## Success Metrics

- [ ] Code coverage > 80%
- [ ] API response time < 200ms
- [ ] Database query optimization (reduce N+1 queries)
- [ ] Bundle size reduction > 30%
- [ ] Zero security vulnerabilities
- [ ] 100% uptime with health checks

## Notes

- Implement changes incrementally to avoid breaking existing functionality
- Maintain backward compatibility during transition
- Create migration scripts for database changes
- Document all API changes
- Test thoroughly in staging environment before production deployment