# Course Management

A golf course management system that tracks course ratings, reviews, and scores.

## Local Development

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your environment variables
3. Run `go mod download` to install dependencies
4. **Optional:** Set up PostgreSQL database (see [DATABASE_SETUP.md](DATABASE_SETUP.md))
5. Run `go run .` to start the server

### Database Setup (Optional)

The application supports both JSON files and PostgreSQL:
- **Without database:** Uses JSON files (current behavior)
- **With database:** Automatically migrates JSON data and uses PostgreSQL

For PostgreSQL setup, see detailed instructions in [DATABASE_SETUP.md](DATABASE_SETUP.md)

## Deployment

This application can be deployed using Railway:

1. Create a new project on [Railway](https://railway.app/)
2. Connect your GitHub repository
3. Railway will automatically detect the Dockerfile and deploy your application
4. Set your environment variables in Railway's dashboard
5. Railway will provide you with a public URL for your application

## Environment Variables

- `PORT`: The port number for the server (default: 42069)
- Add any additional environment variables here

## Tech Stack

- Go
- Echo Framework
- HTML Templates
- Docker 
```
air
```


# Golf Course Management App - Features To-Do List

## 🗄️ Database Migration
- [ ] **Migrate from JSON files to PostgreSQL**
  - Set up PostgreSQL database with proper schema
  - Import existing course JSON data into JSONB columns
  - Update Go models to work with database

## 👥 User Management System
- [ ] **User Registration & Authentication**
  - Extend current Google OAuth to full user profiles
  - Create user table with handicap tracking
  - Add user profile pages and settings

- [ ] **User Course Ownership**
  - Allow users to add new courses to the platform
  - Track which user created each course
  - Enable users to edit their own course submissions

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
- [ ] **Multi-User Course Reviews**
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

-- Users table
users (id, username, email, handicap, created_at, updated_at)

-- Courses table with JSONB for your complex course data
courses (id, name, address, course_data JSONB, created_by, created_at)

-- User follows/friendships
user_follows (follower_id, following_id, created_at)

-- Course reviews (separate from course data for aggregation)
course_reviews (id, course_id, user_id, overall_rating, individual_ratings JSONB, review_text, created_at)

-- User course scores
user_scores (id, course_id, user_id, score, handicap, date_played, created_at)

-- Activity feed items
activities (id, user_id, activity_type, course_id, data JSONB, created_at)