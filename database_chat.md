# Database and Deployment Strategy for Golf Course Management App

## Application Overview

This is a Go-based golf course management application with the following current structure:
- **Backend**: Go with Echo framework
- **Current Data Storage**: JSON files for course data
- **Authentication**: Google OAuth integration
- **Frontend**: HTML templates with server-side rendering

### Current Data Models
```go
type Course struct {
    Name          string  `json:"name"`
    ID            int     `json:"ID"`
    Description   string  `json:"description"`
    Ranks         Ranking `json:"ranks"`
    OverallRating string  `json:"overallRating"`
    Review        string  `json:"review"`
    Holes         []Hole  `json:"holes"`
    Scores        []Score `json:"scores"`
    Address       string  `json:"address"`
}

type Ranking struct {
    Price              string `json:"price"`
    HandicapDifficulty int    `json:"handicapDifficulty"`
    HazardDifficulty   int    `json:"hazardDifficulty"`
    Merch              string `json:"merch"`
    Condition          string `json:"condition"`
    EnjoymentRating    string `json:"enjoymentRating"`
    Vibe               string `json:"vibe"`
    Range              string `json:"range"`
    Amenities          string `json:"amenities"`
    Glizzies           string `json:"glizzies"`
}
```

## Requirements Analysis

The application needs to support:
1. **Complex Course Data**: Long text descriptions, detailed rankings, hole-by-hole information
2. **User Management**: User accounts, profiles, authentication
3. **Social Features**: Users can follow other golfers, see friend activity feeds
4. **Course Reviews**: Multiple users can review courses, aggregate ratings by handicap level
5. **Photo Storage**: Users can upload photos for course reviews
6. **Advanced Filtering**: Complex queries across all data types
7. **Activity Feeds**: Real-time updates based on friend networks

## Database Recommendation: PostgreSQL

### Why PostgreSQL is the Best Choice

1. **Hybrid Data Handling**
   - Native JSONB support for complex course data structure
   - Relational capabilities for user management and social features
   - Can store existing JSON course files directly in JSONB columns

2. **Advanced Query Capabilities**
   - Complex JOIN operations for social features
   - Window functions for activity feeds and rankings
   - Full-text search for course descriptions and reviews
   - GIN indexes on JSONB for fast JSON queries

3. **Aggregation & Analytics**
   - Excellent for calculating average ratings across handicap levels
   - Window functions for complex feed queries
   - Materialized views for performance optimization

4. **Scalability & Performance**
   - Handles concurrent users effectively
   - Efficient indexing strategies
   - Strong consistency guarantees

### Recommended Database Schema

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    handicap DECIMAL(4,2),
    profile_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Courses table with JSONB for complex course data
CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    course_data JSONB NOT NULL, -- Store existing JSON structure
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- User follows/friendships
CREATE TABLE user_follows (
    id SERIAL PRIMARY KEY,
    follower_id INTEGER REFERENCES users(id),
    following_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(follower_id, following_id)
);

-- Course reviews (separate from course data for aggregation)
CREATE TABLE course_reviews (
    id SERIAL PRIMARY KEY,
    course_id INTEGER REFERENCES courses(id),
    user_id INTEGER REFERENCES users(id),
    overall_rating VARCHAR(1) CHECK (overall_rating IN ('S', 'A', 'B', 'C', 'D', 'F')),
    individual_ratings JSONB, -- Store detailed rankings
    review_text TEXT,
    photos TEXT[], -- Array of photo file paths/URLs
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(course_id, user_id) -- One review per user per course
);

-- User course scores
CREATE TABLE user_scores (
    id SERIAL PRIMARY KEY,
    course_id INTEGER REFERENCES courses(id),
    user_id INTEGER REFERENCES users(id),
    score INTEGER NOT NULL,
    handicap DECIMAL(4,2),
    date_played DATE,
    score_data JSONB, -- Hole-by-hole scores if needed
    created_at TIMESTAMP DEFAULT NOW()
);

-- Activity feed items
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    activity_type VARCHAR(50) NOT NULL, -- 'course_review', 'score_posted', 'course_added'
    course_id INTEGER REFERENCES courses(id),
    data JSONB, -- Additional activity-specific data
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_course_reviews_course_id ON course_reviews(course_id);
CREATE INDEX idx_course_reviews_user_id ON course_reviews(user_id);
CREATE INDEX idx_user_follows_follower ON user_follows(follower_id);
CREATE INDEX idx_user_follows_following ON user_follows(following_id);
CREATE INDEX idx_activities_user_id_created ON activities(user_id, created_at DESC);
CREATE INDEX idx_courses_course_data ON courses USING GIN(course_data);
```

### Migration Strategy
1. Import existing JSON course files into the `course_data` JSONB column
2. Maintain backward compatibility during transition
3. Gradually move specific fields to relational columns if needed for performance

## Photo Storage Strategy

### ❌ Do NOT Store Photos in Database
**Reasons to avoid database photo storage:**
- Performance degradation (slower queries, massive backups)
- Expensive storage costs
- Scalability issues
- Connection timeouts for large files

### ✅ Recommended Approach: File System + Database References

**Schema Addition:**
```sql
-- Photos stored as file paths/URLs in course_reviews table
photos TEXT[] -- Array of photo file paths/URLs
```

**File Storage Options (Best to Worst):**

1. **Cloud Storage (Recommended)**
   - AWS S3, Google Cloud Storage, Cloudflare R2
   - Built-in CDN for fast global delivery
   - Automatic backups and redundancy
   - Cost-effective scaling

2. **Local File System**
   - Store in `/uploads/course-photos/` directory
   - Simple for development
   - Requires backup strategy

**File Naming Convention:**
```
/uploads/course-photos/{course_id}/{user_id}/{timestamp}_{random}.jpg
```

**Go Implementation Example:**
```go
type CourseReview struct {
    ID               int      `json:"id"`
    CourseID         int      `json:"course_id"`
    UserID           int      `json:"user_id"`
    OverallRating    string   `json:"overall_rating"`
    ReviewText       string   `json:"review_text"`
    PhotoURLs        []string `json:"photo_urls"` // Store URLs/paths
    CreatedAt        time.Time `json:"created_at"`
}
```

## Deployment Recommendations

### Option 1: Railway (Recommended for Simplicity)
**Perfect for getting started:**
- Built-in PostgreSQL database (no setup required)
- Automatic HTTPS and custom domains
- Git-based deployments
- File storage works out of the box
- Cost: ~$5-15/month
- Zero DevOps overhead

**Configuration:**
```yaml
# railway.toml
[build]
  builder = "NIXPACKS"

[deploy]
  healthcheckPath = "/"
  healthcheckTimeout = 100
  restartPolicyType = "ON_FAILURE"
```

### Option 2: Fly.io (Best Performance)
**For production-ready applications:**
- Excellent global performance
- Built-in PostgreSQL via Fly Postgres
- Persistent volumes for photo storage
- Auto-scaling capabilities
- Cost: ~$10-20/month

**Configuration:**
```toml
# fly.toml
app = "golf-course-app"

[build]
  builder = "paketobuildpacks/builder:base"

[[services]]
  http_checks = []
  internal_port = 8080
  processes = ["app"]
  protocol = "tcp"

[mounts]
  source = "photos_volume"
  destination = "/app/uploads"
```

### Option 3: DigitalOcean App Platform (Balanced)
**Good middle ground:**
- Managed PostgreSQL database
- Spaces (S3-compatible) for photo storage
- Simple configuration
- Cost: ~$12-25/month

## Complete Deployment Setup

### Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Create uploads directory
RUN mkdir -p /app/uploads/course-photos

COPY --from=builder /app/main .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static
COPY --from=builder /app/courses ./courses

EXPOSE 8080
CMD ["./main"]
```

### Required Environment Variables
```bash
DATABASE_URL=postgresql://user:pass@host:5432/dbname
SESSION_SECRET=your-32-char-secret
PORT=8080
UPLOAD_PATH=/app/uploads/course-photos
```

### Photo Upload Handler Example
```go
func (h *Handlers) UploadCoursePhoto(c echo.Context) error {
    file, err := c.FormFile("photo")
    if err != nil {
        return err
    }
    
    // Create directory structure
    uploadPath := os.Getenv("UPLOAD_PATH")
    courseID := c.Param("courseId")
    userID := c.Get("user_id").(string)
    
    dir := filepath.Join(uploadPath, courseID, userID)
    os.MkdirAll(dir, 0755)
    
    // Save file
    filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    dst := filepath.Join(dir, filename)
    
    // Save and return URL
    photoURL := fmt.Sprintf("/uploads/course-photos/%s/%s/%s", courseID, userID, filename)
    return c.JSON(200, map[string]string{"url": photoURL})
}
```

## Implementation Roadmap

### Phase 1: Database Migration
1. Set up PostgreSQL database
2. Create schema with tables above
3. Migrate existing JSON course data to JSONB columns
4. Update Go models and database layer

### Phase 2: User Management
1. Implement user registration/authentication
2. Add user profiles and preferences
3. Create user-course relationships

### Phase 3: Social Features
1. Implement user following system
2. Create activity feed functionality
3. Add course review system with aggregated ratings

### Phase 4: Photo Management
1. Implement photo upload functionality
2. Add photo storage and serving
3. Integrate photos with course reviews

### Phase 5: Advanced Features
1. Implement advanced filtering and search
2. Add analytics and reporting
3. Optimize performance with proper indexing

## Key Benefits of This Architecture

1. **Flexibility**: JSONB allows storing complex course data while maintaining relational integrity
2. **Scalability**: PostgreSQL handles growth in users, courses, and social interactions
3. **Performance**: Proper indexing and query optimization for fast responses
4. **Cost-Effective**: Efficient photo storage strategy keeps costs manageable
5. **Developer-Friendly**: Familiar SQL with modern JSON capabilities
6. **Production-Ready**: Proven technology stack with excellent tooling

This architecture provides a solid foundation for your golf course management application that can scale from a small user base to thousands of golfers sharing courses and building a community around the sport. 