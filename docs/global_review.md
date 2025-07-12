# Global Review Implementation Plan

## Overview
This document outlines the implementation of a global review feature that allows users to toggle between their personal course reviews and aggregated global reviews from all users.

## Current State Analysis

### Existing Architecture
- **CourseReview Model**: Individual user reviews with comprehensive rating categories
- **Rating System**: Letter grades (S, A, B, C, D, F) and numeric scales (1-10)
- **API Layer**: Modern REST API with pagination and filtering
- **Frontend**: HTMX-powered dynamic interactions
- **Database**: PostgreSQL with GORM ORM

### Rating Categories
- Overall Rating (S-F)
- Price ($-$$$$)
- Handicap Difficulty (1-10)
- Hazard Difficulty (1-10)
- Categorical Ratings (S-F): Merch, Condition, Enjoyment, Vibe, Range, Amenities, Glizzies, Walkability

## Implementation Plan

### Phase 1: Database Schema Extensions

#### 1.1 Add Global Review Cache Table
```sql
-- New table to cache aggregated global reviews
CREATE TABLE global_course_reviews (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    course_hash VARCHAR(255) NOT NULL,
    
    -- Aggregated ratings
    overall_rating VARCHAR(1), -- Most common S-F rating
    price VARCHAR(10), -- Most common price tier
    handicap_difficulty DECIMAL(3,2), -- Average of all ratings
    hazard_difficulty DECIMAL(3,2), -- Average of all ratings
    
    -- Categorical ratings (most common)
    merch VARCHAR(1),
    condition VARCHAR(1),
    enjoyment VARCHAR(1),
    vibe VARCHAR(1),
    range VARCHAR(1),
    amenities VARCHAR(1),
    glizzies VARCHAR(1),
    walkability VARCHAR(1),
    
    -- Metadata
    total_reviews INTEGER DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    UNIQUE(course_id, course_hash),
    INDEX idx_course_hash (course_hash),
    INDEX idx_course_id (course_id)
);
```

#### 1.2 Add Rating Distribution Table
```sql
-- Track distribution of ratings for better aggregation
CREATE TABLE rating_distributions (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    course_hash VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL, -- 'overall', 'condition', etc.
    rating_value VARCHAR(10) NOT NULL, -- 'S', 'A', '$$$', etc.
    count INTEGER DEFAULT 0,
    
    UNIQUE(course_id, course_hash, category, rating_value),
    INDEX idx_course_category (course_id, category)
);
```

### Phase 2: Service Layer Extensions

#### 2.1 Global Review Service Interface
```go
// services/global_review_service.go
type GlobalReviewService interface {
    GetGlobalReview(courseID int, courseHash string) (*GlobalReview, error)
    RefreshGlobalReview(courseID int, courseHash string) error
    GetRatingDistribution(courseID int, courseHash string) (*RatingDistribution, error)
    ToggleUserReviewMode(userID int, mode string) error
    GetUserReviewMode(userID int) (string, error)
}

type GlobalReview struct {
    CourseID           int                `json:"course_id"`
    CourseHash         string             `json:"course_hash"`
    OverallRating      string             `json:"overall_rating"`
    Price              string             `json:"price"`
    HandicapDifficulty float64            `json:"handicap_difficulty"`
    HazardDifficulty   float64            `json:"hazard_difficulty"`
    CategoricalRatings map[string]string  `json:"categorical_ratings"`
    TotalReviews       int                `json:"total_reviews"`
    LastUpdated        time.Time          `json:"last_updated"`
    Distribution       *RatingDistribution `json:"distribution,omitempty"`
}

type RatingDistribution struct {
    CourseID   int                           `json:"course_id"`
    CourseHash string                        `json:"course_hash"`
    Categories map[string]map[string]int     `json:"categories"`
}
```

#### 2.2 Aggregation Logic
```go
// services/aggregation_service.go
type AggregationService interface {
    CalculateGlobalReview(courseID int, courseHash string) (*GlobalReview, error)
    GetMostCommonRating(ratings []string) string
    GetAverageNumericRating(ratings []float64) float64
    UpdateRatingDistribution(courseID int, courseHash string) error
}

// Implementation functions:
// - CalculateGlobalReview: Aggregates all user reviews for a course
// - GetMostCommonRating: Finds mode of categorical ratings (S, A, B, C, D, F)
// - GetAverageNumericRating: Calculates mean of numeric ratings (1-10)
// - UpdateRatingDistribution: Updates distribution table for analytics
```

### Phase 3: API Layer Extensions

#### 3.1 New API Endpoints
```go
// api/global_review_handlers.go

// GET /api/courses/:courseId/global-review
func GetGlobalReview(c echo.Context) error

// GET /api/courses/:courseId/review-distribution  
func GetReviewDistribution(c echo.Context) error

// POST /api/user/review-mode
func SetUserReviewMode(c echo.Context) error

// GET /api/user/review-mode
func GetUserReviewMode(c echo.Context) error

// POST /api/courses/:courseId/refresh-global-review
func RefreshGlobalReview(c echo.Context) error
```

#### 3.2 Enhanced Course Review Endpoint
```go
// Modify existing GET /api/courses/:courseId/reviews
// Add query parameter: ?view_mode=personal|global
func GetCourseReviews(c echo.Context) error {
    viewMode := c.QueryParam("view_mode") // "personal" or "global"
    
    if viewMode == "global" {
        // Return global aggregated review
        return getGlobalReview(c)
    }
    
    // Return personal review (existing logic)
    return getPersonalReview(c)
}
```

### Phase 4: Frontend Implementation

#### 4.1 Toggle Component
```html
<!-- views/partials/review-toggle.html -->
<div class="review-toggle" id="review-toggle">
    <div class="toggle-buttons">
        <button class="toggle-btn active" 
                hx-get="/api/courses/{{.CourseID}}/reviews?view_mode=personal"
                hx-target="#review-content"
                hx-swap="innerHTML"
                onclick="setActiveToggle(this)">
            My Review
        </button>
        <button class="toggle-btn" 
                hx-get="/api/courses/{{.CourseID}}/reviews?view_mode=global"
                hx-target="#review-content"
                hx-swap="innerHTML"
                onclick="setActiveToggle(this)">
            Global Review
        </button>
    </div>
</div>

<div id="review-content">
    <!-- Review content will be loaded here -->
</div>
```

#### 4.2 Global Review Template
```html
<!-- views/global-review.html -->
<div class="global-review-container">
    <div class="global-review-header">
        <h3>Global Course Review</h3>
        <p class="review-count">Based on {{.TotalReviews}} reviews</p>
    </div>
    
    <div class="rating-grid">
        <div class="rating-item">
            <label>Overall Rating</label>
            <div class="letter-grade {{.OverallRating}}">{{.OverallRating}}</div>
        </div>
        
        <div class="rating-item">
            <label>Price</label>
            <div class="price-tier">{{.Price}}</div>
        </div>
        
        <div class="rating-item">
            <label>Handicap Difficulty</label>
            <div class="numeric-rating">{{printf "%.1f" .HandicapDifficulty}}</div>
        </div>
        
        <div class="rating-item">
            <label>Hazard Difficulty</label>
            <div class="numeric-rating">{{printf "%.1f" .HazardDifficulty}}</div>
        </div>
        
        <!-- Categorical ratings -->
        {{range $category, $rating := .CategoricalRatings}}
        <div class="rating-item">
            <label>{{$category | title}}</label>
            <div class="letter-grade {{$rating}}">{{$rating}}</div>
        </div>
        {{end}}
    </div>
    
    <div class="global-review-note">
        <p><em>Note: Scores and hole-by-hole data are not included in global reviews.</em></p>
    </div>
</div>
```

#### 4.3 JavaScript Toggle Logic
```javascript
// static/js/review-toggle.js
function setActiveToggle(button) {
    // Remove active class from all buttons
    document.querySelectorAll('.toggle-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Add active class to clicked button
    button.classList.add('active');
    
    // Store user preference
    const viewMode = button.textContent.includes('My') ? 'personal' : 'global';
    localStorage.setItem('preferredViewMode', viewMode);
}

// Initialize toggle state on page load
document.addEventListener('DOMContentLoaded', function() {
    const preferredMode = localStorage.getItem('preferredViewMode') || 'personal';
    const button = document.querySelector(`[onclick*="${preferredMode}"]`);
    if (button) {
        button.click();
    }
});
```

### Phase 5: Database Migration and Seeding

#### 5.1 Migration Script
```sql
-- scripts/add_global_reviews.sql
BEGIN;

-- Add global review cache table
CREATE TABLE global_course_reviews (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    course_hash VARCHAR(255) NOT NULL,
    overall_rating VARCHAR(1),
    price VARCHAR(10),
    handicap_difficulty DECIMAL(3,2),
    hazard_difficulty DECIMAL(3,2),
    merch VARCHAR(1),
    condition VARCHAR(1),
    enjoyment VARCHAR(1),
    vibe VARCHAR(1),
    range VARCHAR(1),
    amenities VARCHAR(1),
    glizzies VARCHAR(1),
    walkability VARCHAR(1),
    total_reviews INTEGER DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, course_hash)
);

-- Add rating distribution table
CREATE TABLE rating_distributions (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    course_hash VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL,
    rating_value VARCHAR(10) NOT NULL,
    count INTEGER DEFAULT 0,
    UNIQUE(course_id, course_hash, category, rating_value)
);

-- Add indexes
CREATE INDEX idx_global_course_hash ON global_course_reviews(course_hash);
CREATE INDEX idx_global_course_id ON global_course_reviews(course_id);
CREATE INDEX idx_dist_course_category ON rating_distributions(course_id, category);

-- Add user preference column
ALTER TABLE users ADD COLUMN preferred_review_mode VARCHAR(20) DEFAULT 'personal';

COMMIT;
```

#### 5.2 Seeding Script
```go
// scripts/seed_global_reviews.go
func SeedGlobalReviews(db *gorm.DB) error {
    // Get all courses that have reviews
    var courses []database.CourseDB
    if err := db.Find(&courses).Error; err != nil {
        return err
    }
    
    for _, course := range courses {
        // Calculate and store global review for each course
        if err := aggregationService.CalculateGlobalReview(course.ID, course.Hash); err != nil {
            log.Printf("Error calculating global review for course %d: %v", course.ID, err)
        }
    }
    
    return nil
}
```

### Phase 6: Performance Optimizations

#### 6.1 Caching Strategy
```go
// services/cache_service.go additions
type GlobalReviewCache interface {
    GetGlobalReview(courseHash string) (*GlobalReview, error)
    SetGlobalReview(courseHash string, review *GlobalReview) error
    InvalidateGlobalReview(courseHash string) error
}

// Cache keys
const (
    GlobalReviewCacheKey = "global_review:%s"
    GlobalReviewCacheTTL = 1 * time.Hour
)
```

#### 6.2 Background Jobs
```go
// services/background_jobs.go
type BackgroundJobs interface {
    ScheduleGlobalReviewRefresh(courseID int, courseHash string)
    RefreshAllGlobalReviews()
}

// Refresh global reviews when new reviews are added
func (s *reviewService) CreateReview(review *CourseReview) error {
    // ... existing logic
    
    // Schedule background refresh of global review
    s.backgroundJobs.ScheduleGlobalReviewRefresh(review.CourseID, review.CourseHash)
    
    return nil
}
```

### Phase 7: Testing Strategy

#### 7.1 Unit Tests
```go
// services/global_review_service_test.go
func TestCalculateGlobalReview(t *testing.T) {
    // Test aggregation logic with various rating distributions
}

func TestGetMostCommonRating(t *testing.T) {
    // Test mode calculation for categorical ratings
}

func TestGetAverageNumericRating(t *testing.T) {
    // Test mean calculation for numeric ratings
}
```

#### 7.2 Integration Tests
```go
// api/global_review_handlers_test.go
func TestGetGlobalReview(t *testing.T) {
    // Test API endpoint with various scenarios
}

func TestToggleReviewMode(t *testing.T) {
    // Test user preference persistence
}
```

## Implementation Steps

### Step 1: Database Schema
1. Run migration script to add new tables
2. Add user preference column
3. Create necessary indexes

### Step 2: Service Layer
1. Implement GlobalReviewService interface
2. Create aggregation logic
3. Add caching layer

### Step 3: API Layer
1. Add new endpoints for global reviews
2. Modify existing endpoints to support view modes
3. Add user preference endpoints

### Step 4: Frontend
1. Create toggle component
2. Build global review template
3. Add JavaScript for toggle functionality

### Step 5: Data Migration
1. Run seeding script to populate global reviews
2. Calculate initial rating distributions
3. Test data integrity

### Step 6: Testing
1. Write comprehensive unit tests
2. Add integration tests
3. Test performance with large datasets

### Step 7: Performance Optimization
1. Implement caching strategy
2. Add background job processing
3. Monitor and optimize database queries

## Success Criteria

1. **Functionality**: Users can toggle between personal and global reviews
2. **Performance**: Global reviews load in under 500ms
3. **Accuracy**: Aggregation logic correctly calculates ratings
4. **Persistence**: User preferences are saved and restored
5. **Scalability**: System handles courses with 100+ reviews
6. **Compatibility**: Feature works with existing HTMX frontend
7. **Testing**: 90%+ test coverage for new code

## Rollback Plan

1. **Database**: Keep migration scripts reversible
2. **API**: Maintain backward compatibility
3. **Frontend**: Feature flag for toggle functionality
4. **Data**: Backup existing review data before migration

## Post-Implementation Enhancements

1. **Analytics**: Track usage of personal vs global views
2. **Filtering**: Add filters for global reviews (date range, user count)
3. **Visualizations**: Add charts for rating distributions
4. **Notifications**: Alert users when global ratings change significantly
5. **Export**: Allow users to export global review data