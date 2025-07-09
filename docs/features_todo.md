# Golf Course Management App - Features To-Do List

## 🗄️ Database Migration
- [x] **Migrate from JSON files to PostgreSQL**
  - Set up PostgreSQL database with proper schema
  - Import existing course JSON data into JSONB columns
  - Update Go models to work with database

## 👥 User Management System
- [x] **User Registration & Authentication**
  - Extend current Google OAuth to full user profiles
  - Create user table with handicap tracking
  - Add user profile pages and settings

- [x] **User Course Ownership**
  - Allow users to add new courses to the platform
  - Track which user created each course
  - Enable users to edit their own course submissions

1. Database Schema Updates
Add CreatedBy field to Course model (foreign key to User)
Add UpdatedBy field to Course model (optional, for tracking edits)
Run database migration to add these columns
2. Course Creation Updates
Modify CreateCourse handler to capture authenticated user ID
Update course creation form to associate new courses with current user
Add validation to ensure only authenticated users can create courses
3. Course Edit Authorization
Add CanEditCourse() method to check if user owns the course
Modify EditCourse handler to verify ownership before allowing edits
Update edit buttons in UI to only show for course owners
4. Database Service Methods
Add GetCoursesByUser(userID) method for user's course listings
Add GetCourseOwner(courseID) method for ownership verification
Update existing course methods to handle ownership
5. UI/Template Updates
Add "My Courses" vs "All Courses" sections on profile page
Show ownership indicators (e.g., "Created by you" badge)
Hide/show edit buttons based on ownership
Add user attribution on course detail pages
6. Route Protection
Add middleware to protect edit routes with ownership checks
Update existing course routes to pass ownership context

## 🤝 Social Features
- [ ] **User Following System**
  - Allow users to follow other golfers
  - Create friends/following lists
  - Add follow/unfollow functionality

- [ ] **Activity Feed**
  - Show recent activity from followed users
  - Include course reviews, new scores, course additions
  - Real-time or near-real-time updates

## ⭐ Course Review & Rating System
- [x] **Multi-User Course Reviews**
  - Allow multiple users to review each course
  - Store individual ratings for all course aspects
  - Add photo upload capability for course reviews

- [ ] **Aggregated Rating System**
  - Calculate average ratings across all users
  - Segment ratings by handicap level ranges
  - Display consolidated scores on course pages

## 🔍 Advanced Filtering & Search
- [ ] **Course Filtering**
  - Filter by price range, difficulty, ratings
  - Location-based filtering
  - Filter by specific amenities (range, merch, etc.)

- [ ] **User & Activity Filtering**
  - Filter activity feed by activity type
  - Search for users by name or handicap
  - Filter reviews by handicap level

## 📸 Photo Management
- [ ] **Course Photo System**
  - File upload handling for course images
  - Photo storage strategy (local or cloud)
  - Display photos in course reviews and galleries

## 🎯 Enhanced Course Data
- [ ] **Long-Form Content Support**
  - Expand course descriptions with rich text
  - Support for detailed hole-by-hole descriptions
  - Course history and additional metadata

## 📊 Analytics & Insights
- [ ] **Rating Analytics**
  - Trending courses by rating improvements
  - Handicap-based course difficulty insights
  - User engagement metrics

---

**Priority Order:**
1. Database Migration (foundation for everything else)
2. User Management System (enables social features)
3. Social Features (core differentiator)
4. Review & Rating System (main value proposition)
5. Advanced Filtering (user experience enhancement)
6. Photo Management (content richness)
7. Analytics & Insights (future growth features) 