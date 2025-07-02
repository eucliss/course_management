# Mobile App Development Guide for Golf Course Management

## Overview

This document outlines the strategy for adding mobile apps (iOS/Android) to the existing Go-based golf course management application while maintaining the current HTMX web interface.

## Current Application Architecture

- **Backend**: Go with Echo framework
- **Frontend**: HTMX with server-side HTML rendering
- **Data Storage**: JSON files (migrating to PostgreSQL recommended)
- **Authentication**: Google OAuth integration

## Database Strategy for Mobile

### Recommended Database: PostgreSQL

**Why PostgreSQL is perfect for mobile + web:**
- **Hybrid Data Handling**: JSONB support for complex course data + relational capabilities
- **Concurrent Access**: Handles both web and mobile clients efficiently
- **Advanced Queries**: Complex filtering, social features, aggregated ratings
- **Mobile-Optimized Indexes**: Fast API responses for mobile apps

### Schema Design
```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    handicap DECIMAL(4,2),
    profile_data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Courses with JSONB for existing structure
CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    course_data JSONB NOT NULL, -- Store existing JSON structure
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Social features
CREATE TABLE user_follows (
    id SERIAL PRIMARY KEY,
    follower_id INTEGER REFERENCES users(id),
    following_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(follower_id, following_id)
);

-- Course reviews
CREATE TABLE course_reviews (
    id SERIAL PRIMARY KEY,
    course_id INTEGER REFERENCES courses(id),
    user_id INTEGER REFERENCES users(id),
    overall_rating VARCHAR(1) CHECK (overall_rating IN ('S', 'A', 'B', 'C', 'D', 'F')),
    individual_ratings JSONB,
    review_text TEXT,
    photos TEXT[], -- Array of photo URLs
    created_at TIMESTAMP DEFAULT NOW()
);

-- Activity feed
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    activity_type VARCHAR(50) NOT NULL,
    course_id INTEGER REFERENCES courses(id),
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Photo Storage Strategy

### ❌ Do NOT Store Photos in Database
- Performance issues (slower queries, massive backups)
- Expensive storage costs
- Scalability problems

### ✅ Recommended: File System + Database References
```go
// Photo storage structure
/uploads/course-photos/{course_id}/{user_id}/{timestamp}_{random}.jpg

// Database stores URLs/paths only
photos TEXT[] -- Array of photo file paths/URLs
```

## Backend Architecture: Dual-Mode Approach

### Strategy: Keep HTMX + Add API Routes

**NO BREAKING CHANGES** to existing HTMX setup - add parallel API routes:

```go
func main() {
    e := echo.New()
    
    // EXISTING HTMX ROUTES (unchanged)
    e.GET("/", handlers.Home)                    // Returns HTML
    e.GET("/courses", handlers.GetCoursesHTML)   // Returns HTML for HTMX
    e.GET("/course/:id", handlers.GetCourseHTML) // Returns HTML for HTMX
    e.POST("/create-course", handlers.CreateCourseHTML) // HTMX form handling
    
    // NEW: API routes for mobile apps
    api := e.Group("/api/v1")
    api.Use(middleware.CORS())
    api.GET("/courses", handlers.GetCoursesAPI)     // Returns JSON
    api.GET("/courses/:id", handlers.GetCourseAPI)  // Returns JSON
    api.POST("/courses", handlers.CreateCourseAPI)  // Returns JSON
    api.GET("/users/:id/feed", handlers.GetActivityFeedAPI)
    api.POST("/auth/mobile", handlers.MobileAuthAPI)
}
```

### File Structure
```
course_management/
├── handlers.go          # Existing HTMX handlers (unchanged)
├── api_handlers.go      # New mobile API handlers
├── models.go           # Shared models (unchanged)
├── course_service.go   # Business logic (unchanged)
├── main.go             # Route setup (add API routes)
└── views/              # HTMX templates (unchanged)
```

### Example API Handler
```go
// api_handlers.go
func (h *Handlers) GetCoursesAPI(c echo.Context) error {
    courses, err := h.courseService.LoadCourses()
    if err != nil {
        return c.JSON(500, map[string]string{"error": "Failed to load courses"})
    }
    
    return c.JSON(200, map[string]interface{}{
        "courses": courses,
        "total":   len(courses),
    })
}

func (h *Handlers) CreateCourseAPI(c echo.Context) error {
    var course Course
    if err := c.Bind(&course); err != nil {
        return c.JSON(400, map[string]string{"error": "Invalid course data"})
    }
    
    err := h.courseService.SaveCourse(course)
    if err != nil {
        return c.JSON(500, map[string]string{"error": "Failed to save course"})
    }
    
    return c.JSON(201, course)
}
```

## Mobile App Technology Recommendations

### Option 1: React Native (Recommended)

**Why React Native:**
- ✅ Single codebase for iOS and Android
- ✅ Native UI components (truly native iOS/Android look)
- ✅ Large ecosystem for sports apps
- ✅ Easy API integration
- ✅ Expo for rapid development

```javascript
// React Native example
import React, { useEffect, useState } from 'react';
import { View, Text, FlatList } from 'react-native';

const CourseList = () => {
  const [courses, setCourses] = useState([]);
  
  useEffect(() => {
    fetch('https://your-api.com/api/v1/courses')
      .then(response => response.json())
      .then(data => setCourses(data.courses));
  }, []);
  
  return (
    <FlatList
      data={courses}
      renderItem={({ item }) => (
        <View>
          <Text>{item.name}</Text>
          <Text>{item.overallRating}</Text>
        </View>
      )}
    />
  );
};
```

### Option 2: Flutter
- ✅ Single codebase
- ✅ Excellent performance
- ✅ Beautiful, customizable UI
- ✅ Great for complex layouts

### Option 3: Native (Swift/Kotlin)
- Only if maximum performance needed
- More development time
- Platform-specific expertise required

## Deployment Recommendations

### Option 1: DigitalOcean (Recommended for Mobile)

**Why DigitalOcean excels for mobile:**
- **Global CDN**: Built-in Spaces CDN for fast photo delivery
- **Predictable Costs**: $12-25/month vs complex pricing elsewhere
- **S3-Compatible Storage**: Perfect for mobile photo uploads
- **Auto-scaling**: Handles mobile traffic spikes

```yaml
# .do/app.yaml
name: golf-course-api
services:
- name: api
  source_dir: /
  run_command: ./main
  environment_slug: go
  instance_count: 2
  instance_size_slug: basic-s

databases:
- engine: PG
  name: golf-db
  version: "15"
```

### Option 2: Railway
- **Simplest setup**: Built-in PostgreSQL
- **Git-based deployments**
- **Cost**: ~$15-30/month
- **Limited mobile features** (no CDN)

### Option 3: Fly.io
- **Best performance**: Global edge locations
- **Complex pricing**: Can get expensive
- **Excellent scaling**

## Photo Handling for Mobile

### DigitalOcean Spaces Integration
```go
func (h *Handlers) UploadPhotoAPI(c echo.Context) error {
    file, err := c.FormFile("photo")
    if err != nil {
        return c.JSON(400, map[string]string{"error": "No file uploaded"})
    }
    
    // Upload to Spaces
    spacesURL := "https://golf-photos.nyc3.digitaloceanspaces.com/"
    cdnURL := "https://golf-photos.nyc3.cdn.digitaloceanspaces.com/"
    
    // Save file to Spaces
    filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    // ... upload logic ...
    
    // Return CDN URL for mobile caching
    return c.JSON(200, map[string]string{
        "url": cdnURL + filename,
        "thumbnail": cdnURL + filename + "?w=300&h=200",
    })
}
```

## Development Roadmap

### Phase 1: Backend API Setup (Week 1-2)
1. **Add API routes** alongside existing HTMX routes
2. **Set up PostgreSQL** database
3. **Migrate course data** from JSON to database
4. **Test API endpoints** with Postman

### Phase 2: Mobile App Foundation (Week 3-4)
1. **Create React Native app** with Expo
2. **Implement authentication** (Google OAuth for mobile)
3. **Build core screens**: Course list, course detail, profile
4. **Connect to API** endpoints

### Phase 3: Core Features (Week 5-6)
1. **Course creation** and editing from mobile
2. **Photo upload** functionality
3. **User profiles** and settings
4. **Basic search** and filtering

### Phase 4: Social Features (Week 7-8)
1. **User following** system
2. **Activity feed** implementation
3. **Course reviews** and ratings
4. **Push notifications** setup

### Phase 5: Advanced Features (Week 9-10)
1. **GPS integration** for nearby courses
2. **Advanced filtering** and search
3. **Offline capabilities**
4. **Performance optimization**

## Mobile-Specific Features

### React Native Capabilities
```javascript
// GPS for course discovery
import * as Location from 'expo-location';

// Camera for course photos
import { Camera } from 'expo-camera';

// Maps for course locations
import MapView, { Marker } from 'react-native-maps';

// Push notifications
import * as Notifications from 'expo-notifications';

// Social sharing
import * as Sharing from 'expo-sharing';
```

### Native iOS UI Components
```javascript
import {
  SafeAreaView,      // iOS safe area
  ScrollView,        // Native scrolling
  RefreshControl,    // Pull-to-refresh
  ActionSheetIOS,    // iOS action sheets
  Alert,             // Native alerts
} from 'react-native';
```

## Final Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web App       │    │   iOS App       │    │  Android App    │
│   (HTMX)        │    │ (React Native)  │    │ (React Native)  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   Go API Backend        │
                    │   (HTMX + JSON APIs)    │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │    PostgreSQL DB        │
                    │   (DigitalOcean)        │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   Spaces CDN            │
                    │   (Photo Storage)       │
                    └─────────────────────────┘
```

## Cost Estimates

**Total Monthly Costs:**
- **DigitalOcean App Platform**: $12-25/month
- **PostgreSQL Database**: $15/month
- **Spaces + CDN**: $5-15/month
- **Push Notifications**: $0-10/month (Firebase)
- **Apple Developer**: $99/year
- **Google Play**: $25 one-time

**Total: $32-65/month + $124/year for app stores**

## Key Benefits

1. **✅ No Disruption**: Existing HTMX web app continues working
2. **✅ Shared Backend**: Same Go API serves web and mobile
3. **✅ Native Experience**: True iOS/Android UI components
4. **✅ Scalable Architecture**: PostgreSQL + CDN handles growth
5. **✅ Cost-Effective**: Predictable monthly costs
6. **✅ Fast Development**: React Native + Expo rapid iteration

## Next Steps

1. **Choose deployment platform** (recommend DigitalOcean)
2. **Set up PostgreSQL** database
3. **Add API routes** to existing Go backend
4. **Create React Native app** with Expo
5. **Implement core features** following the roadmap

This approach ensures your existing HTMX web application continues to work perfectly while you build a modern mobile experience that shares the same robust backend infrastructure.
