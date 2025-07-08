package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models
type User struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	GoogleID    string   `gorm:"uniqueIndex" json:"google_id"`
	Email       string   `gorm:"uniqueIndex" json:"email"`
	Name        string   `json:"name"`
	DisplayName *string  `json:"display_name"`
	Picture     string   `json:"picture"`
	Handicap    *float64 `json:"handicap,omitempty"`
	CreatedAt   int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64    `gorm:"autoUpdateTime" json:"updated_at"`
}

type CourseDB struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	Name       string   `gorm:"not null" json:"name"`
	Address    string   `json:"address"`
	Hash       string   `gorm:"uniqueIndex;not null" json:"hash"`
	CourseData string   `gorm:"type:jsonb" json:"course_data"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	CreatedBy  *uint    `json:"created_by"`
	UpdatedBy  *uint    `json:"updated_by"`
	CreatedAt  int64    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64    `gorm:"autoUpdateTime" json:"updated_at"`
	Creator    *User    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater    *User    `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

type CourseReview struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	CourseID           uint      `gorm:"not null" json:"course_id"`
	UserID             uint      `gorm:"not null" json:"user_id"`
	OverallRating      *string   `json:"overall_rating"`
	Price              *string   `json:"price"`
	HandicapDifficulty *int      `json:"handicap_difficulty"`
	HazardDifficulty   *int      `json:"hazard_difficulty"`
	Merch              *string   `json:"merch"`
	Condition          *string   `json:"condition"`
	EnjoymentRating    *string   `json:"enjoyment_rating"`
	Vibe               *string   `json:"vibe"`
	RangeRating        *string   `json:"range_rating"`
	Amenities          *string   `json:"amenities"`
	Glizzies           *string   `json:"glizzies"`
	ReviewText         *string   `json:"review_text"`
	CreatedAt          int64     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          int64     `gorm:"autoUpdateTime" json:"updated_at"`
	Course             *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User               *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type UserCourseScore struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CourseID   uint      `gorm:"not null" json:"course_id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Score      int       `gorm:"not null" json:"score"`
	Handicap   *float64  `json:"handicap"`
	DatePlayed *string   `json:"date_played"`
	OutScore   *int      `json:"out_score"`
	InScore    *int      `json:"in_score"`
	Notes      *string   `json:"notes"`
	CreatedAt  int64     `gorm:"autoCreateTime" json:"created_at"`
	Course     *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type UserActivity struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	ActivityType string    `gorm:"type:varchar(50);not null" json:"activity_type"`
	CourseID     *uint     `json:"course_id"`
	TargetUserID *uint     `json:"target_user_id"`
	Data         string    `gorm:"type:jsonb" json:"data"`
	CreatedAt    int64     `gorm:"autoCreateTime" json:"created_at"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course       *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	TargetUser   *User     `gorm:"foreignKey:TargetUserID" json:"target_user,omitempty"`
}

type Course struct {
	Name          string `json:"name"`
	ID            int    `json:"ID"`
	Description   string `json:"description"`
	OverallRating string `json:"overallRating"`
	Review        string `json:"review"`
	Address       string `json:"address"`
}

type TableInfo struct {
	TableName string
	RowCount  int64
}

type IndexInfo struct {
	IndexName string
	TableName string
	Columns   string
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToDatabase() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found")
		}
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "course_management")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	if password == "" {
		return nil, fmt.Errorf("database password not set")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "Never"
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

func safeString(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func safeInt(i *int) string {
	if i == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *i)
}

func safeFloat(f *float64) string {
	if f == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f", *f)
}

func main() {
	fmt.Println("🏌️ GOLF COURSE MANAGEMENT DATABASE ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	db, err := connectToDatabase()
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		fmt.Println("\n💡 Setup instructions:")
		fmt.Println("   1. Create .env file with database credentials")
		fmt.Println("   2. Ensure PostgreSQL is running")
		return
	}

	fmt.Println("✅ Connected to database successfully!")
	fmt.Println()

	// Database Overview
	fmt.Println("📊 DATABASE OVERVIEW:")
	fmt.Println("────────────────────")

	var userCount, courseCount, reviewCount, scoreCount, activityCount int64
	db.Model(&User{}).Count(&userCount)
	db.Model(&CourseDB{}).Count(&courseCount)
	db.Model(&CourseReview{}).Count(&reviewCount)
	db.Model(&UserCourseScore{}).Count(&scoreCount)
	db.Model(&UserActivity{}).Count(&activityCount)

	fmt.Printf("   %-20s %d records\n", "users:", userCount)
	fmt.Printf("   %-20s %d records\n", "course_dbs:", courseCount)
	fmt.Printf("   %-20s %d records\n", "course_reviews:", reviewCount)
	fmt.Printf("   %-20s %d records\n", "user_course_scores:", scoreCount)
	fmt.Printf("   %-20s %d records\n", "user_activities:", activityCount)
	fmt.Println()

	// Users Section
	fmt.Println("👥 USERS:")
	fmt.Println("─────────")
	var users []User
	if db.Find(&users).Error == nil && len(users) > 0 {
		for i, user := range users {
			displayName := user.Name
			if user.DisplayName != nil && *user.DisplayName != "" {
				displayName = fmt.Sprintf("%s (%s)", *user.DisplayName, user.Name)
			}
			fmt.Printf("   %d. %s\n", i+1, displayName)
			fmt.Printf("      📧 %s | 🆔 ID:%d | ⛳ Handicap:%s\n",
				user.Email, user.ID, safeFloat(user.Handicap))
			fmt.Printf("      📅 Joined: %s\n", formatTime(user.CreatedAt))
			fmt.Println()
		}
	} else {
		fmt.Println("   No users found")
		fmt.Println()
	}

	// Courses Section
	fmt.Println("⛳ COURSES:")
	fmt.Println("──────────")
	var courses []CourseDB
	if db.Preload("Creator").Find(&courses).Error == nil && len(courses) > 0 {
		fmt.Printf("   Total: %d courses (showing first 5)\n", len(courses))
		fmt.Println()

		// Show only first 5 courses
		maxCourses := 5
		if len(courses) < maxCourses {
			maxCourses = len(courses)
		}

		for i := 0; i < maxCourses; i++ {
			courseDB := courses[i]
			var course Course
			json.Unmarshal([]byte(courseDB.CourseData), &course)

			createdBy := "System"
			if courseDB.Creator != nil {
				createdBy = courseDB.Creator.Name
				if courseDB.Creator.DisplayName != nil && *courseDB.Creator.DisplayName != "" {
					createdBy = *courseDB.Creator.DisplayName
				}
			}

			fmt.Printf("   %d. %s\n", i+1, course.Name)
			fmt.Printf("      📍 %s\n", course.Address)
			fmt.Printf("      🏆 Rating:%s | 👤 Creator:%s\n",
				func() string {
					if course.OverallRating == "" {
						return "-"
					}
					return course.OverallRating
				}(),
				createdBy)
			fmt.Printf("      📅 Created: %s\n", formatTime(courseDB.CreatedAt))
			fmt.Println()
		}

		if len(courses) > 5 {
			fmt.Printf("   ... and %d more courses\n", len(courses)-5)
			fmt.Println()
		}
	} else {
		fmt.Println("   No courses found")
		fmt.Println()
	}

	// Courses with Coordinates Section
	fmt.Println("📍 COURSES WITH COORDINATES:")
	fmt.Println("────────────────────────────")
	var coursesWithCoords []CourseDB
	if db.Where("latitude IS NOT NULL AND longitude IS NOT NULL").Find(&coursesWithCoords).Error == nil && len(coursesWithCoords) > 0 {
		fmt.Printf("   Total courses with coordinates: %d\n", len(coursesWithCoords))
		fmt.Println()

		// Show first course with coordinates as example
		if len(coursesWithCoords) > 0 {
			courseDB := coursesWithCoords[0]
			var course Course
			json.Unmarshal([]byte(courseDB.CourseData), &course)

			fmt.Printf("   📍 EXAMPLE COURSE WITH COORDINATES:\n")
			fmt.Printf("   ═══════════════════════════════════\n")
			fmt.Printf("   🏌️ Name: %s\n", course.Name)
			fmt.Printf("   📍 Address: %s\n", course.Address)
			fmt.Printf("   🌐 Latitude: %.6f\n", *courseDB.Latitude)
			fmt.Printf("   🌐 Longitude: %.6f\n", *courseDB.Longitude)
			fmt.Printf("   🏆 Rating: %s\n", func() string {
				if course.OverallRating == "" {
					return "-"
				}
				return course.OverallRating
			}())
			fmt.Printf("   📅 Created: %s\n", formatTime(courseDB.CreatedAt))
			fmt.Println()
		}
	} else {
		fmt.Println("   No courses with coordinates found")
		fmt.Println()
	}

	// Reviews Section
	fmt.Println("⭐ COURSE REVIEWS:")
	fmt.Println("──────────────────")
	var reviews []CourseReview
	if db.Preload("Course").Preload("User").Find(&reviews).Error == nil && len(reviews) > 0 {
		for i, review := range reviews {
			courseName := "Unknown Course"
			if review.Course != nil {
				courseName = review.Course.Name
			}
			userName := "Unknown User"
			if review.User != nil {
				userName = review.User.Name
				if review.User.DisplayName != nil && *review.User.DisplayName != "" {
					userName = *review.User.DisplayName
				}
			}

			fmt.Printf("   %d. %s reviewed %s\n", i+1, userName, courseName)
			fmt.Printf("      🏆 Overall:%s | 💰 Price:%s | 🏌️ Handicap:%s | ⚠️ Hazard:%s\n",
				safeString(review.OverallRating), safeString(review.Price),
				safeInt(review.HandicapDifficulty), safeInt(review.HazardDifficulty))
			fmt.Printf("      🛍️ Merch:%s | 🌿 Condition:%s | 😊 Enjoyment:%s | 🎯 Vibe:%s\n",
				safeString(review.Merch), safeString(review.Condition),
				safeString(review.EnjoymentRating), safeString(review.Vibe))
			fmt.Printf("      📅 %s\n", formatTime(review.CreatedAt))
			fmt.Println()
		}
	} else {
		fmt.Println("   No reviews found")
		fmt.Println()
	}

	// Scores Section
	fmt.Println("🏌️ GOLF SCORES:")
	fmt.Println("───────────────")
	var scores []UserCourseScore
	if db.Preload("Course").Preload("User").Find(&scores).Error == nil && len(scores) > 0 {
		for i, score := range scores {
			courseName := "Unknown Course"
			if score.Course != nil {
				courseName = score.Course.Name
			}
			userName := "Unknown User"
			if score.User != nil {
				userName = score.User.Name
				if score.User.DisplayName != nil && *score.User.DisplayName != "" {
					userName = *score.User.DisplayName
				}
			}

			fmt.Printf("   %d. %s at %s\n", i+1, userName, courseName)
			fmt.Printf("      🏆 Score:%d | ⛳ Handicap:%s", score.Score, safeFloat(score.Handicap))
			if score.DatePlayed != nil {
				fmt.Printf(" | 📅 %s", *score.DatePlayed)
			}
			fmt.Printf("\n      📅 Recorded: %s\n", formatTime(score.CreatedAt))
			fmt.Println()
		}
	} else {
		fmt.Println("   No scores found")
		fmt.Println()
	}

	// Activities Section
	fmt.Println("📱 RECENT ACTIVITIES:")
	fmt.Println("────────────────────")
	var activities []UserActivity
	if db.Preload("User").Preload("Course").Order("created_at DESC").Limit(10).Find(&activities).Error == nil && len(activities) > 0 {
		for i, activity := range activities {
			userName := "Unknown User"
			if activity.User != nil {
				userName = activity.User.Name
				if activity.User.DisplayName != nil && *activity.User.DisplayName != "" {
					userName = *activity.User.DisplayName
				}
			}
			courseName := ""
			if activity.Course != nil {
				courseName = fmt.Sprintf(" | 🏌️ %s", activity.Course.Name)
			}

			fmt.Printf("   %d. %s - %s%s\n", i+1, userName, activity.ActivityType, courseName)
			fmt.Printf("      📅 %s\n", formatTime(activity.CreatedAt))
			fmt.Println()
		}
	} else {
		fmt.Println("   No activities found")
		fmt.Println()
	}

	// Database Schema Info
	fmt.Println("🗄️ DATABASE FEATURES:")
	fmt.Println("─────────────────────")
	fmt.Println("   ✨ Multi-user review system with individual ratings")
	fmt.Println("   ✨ JSONB storage for complex course data")
	fmt.Println("   ✨ Golf score tracking with handicap support")
	fmt.Println("   ✨ Activity feed for social features")
	fmt.Println("   ✨ Course ownership and edit permissions")
	fmt.Println("   ✨ Unique course hashing to prevent duplicates")
	fmt.Println("   ✨ Performance indexes for fast queries")
	fmt.Println("   ✨ PostgreSQL constraints for data integrity")
	fmt.Println()

	// Summary Statistics
	fmt.Println("📈 SUMMARY STATISTICS:")
	fmt.Println("──────────────────────")
	fmt.Printf("   👥 Total Users: %d\n", userCount)
	fmt.Printf("   ⛳ Total Courses: %d\n", courseCount)
	fmt.Printf("   ⭐ Total Reviews: %d\n", reviewCount)
	fmt.Printf("   🏌️ Total Scores: %d\n", scoreCount)
	fmt.Printf("   📱 Total Activities: %d\n", activityCount)

	if courseCount > 0 {
		avgReviews := float64(reviewCount) / float64(courseCount)
		fmt.Printf("   📊 Average Reviews per Course: %.1f\n", avgReviews)
	}

	// Most active reviewer
	var mostActive struct {
		Name        string
		DisplayName *string
		ReviewCount int64
	}
	db.Table("course_reviews").
		Select("users.name, users.display_name, COUNT(*) as review_count").
		Joins("JOIN users ON course_reviews.user_id = users.id").
		Group("users.name, users.display_name").
		Order("review_count DESC").
		Limit(1).
		Scan(&mostActive)

	if mostActive.ReviewCount > 0 {
		displayName := mostActive.Name
		if mostActive.DisplayName != nil && *mostActive.DisplayName != "" {
			displayName = *mostActive.DisplayName
		}
		fmt.Printf("   🏆 Most Active Reviewer: %s (%d reviews)\n", displayName, mostActive.ReviewCount)
	}

	fmt.Println()
	fmt.Println("✅ Database analysis complete!")
}
