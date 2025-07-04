package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Copy the models from main package for this script
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
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Address    string `json:"address"`
	Hash       string `gorm:"uniqueIndex;not null" json:"hash"`
	CourseData string `gorm:"type:jsonb" json:"course_data"`
	CreatedBy  *uint  `json:"created_by"`
	UpdatedBy  *uint  `json:"updated_by"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at"`

	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

type CourseReview struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	OverallRating      *string `gorm:"type:varchar(1);check:overall_rating IN ('S','A','B','C','D','F')" json:"overall_rating"`
	Price              *string `gorm:"type:varchar(10)" json:"price"`
	HandicapDifficulty *int    `json:"handicap_difficulty"`
	HazardDifficulty   *int    `json:"hazard_difficulty"`
	Merch              *string `gorm:"type:varchar(1);check:merch IN ('S','A','B','C','D','F')" json:"merch"`
	Condition          *string `gorm:"type:varchar(1);check:condition IN ('S','A','B','C','D','F')" json:"condition"`
	EnjoymentRating    *string `gorm:"type:varchar(1);check:enjoyment_rating IN ('S','A','B','C','D','F')" json:"enjoyment_rating"`
	Vibe               *string `gorm:"type:varchar(1);check:vibe IN ('S','A','B','C','D','F')" json:"vibe"`
	RangeRating        *string `gorm:"type:varchar(1);check:range_rating IN ('S','A','B','C','D','F')" json:"range_rating"`
	Amenities          *string `gorm:"type:varchar(1);check:amenities IN ('S','A','B','C','D','F')" json:"amenities"`
	Glizzies           *string `gorm:"type:varchar(1);check:glizzies IN ('S','A','B','C','D','F')" json:"glizzies"`
	ReviewText         *string `gorm:"type:text" json:"review_text"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updated_at"`

	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type UserCourseScore struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CourseID uint `gorm:"not null" json:"course_id"`
	UserID   uint `gorm:"not null" json:"user_id"`

	Score      int      `gorm:"not null" json:"score"`
	Handicap   *float64 `gorm:"type:decimal(4,2)" json:"handicap"`
	DatePlayed *string  `gorm:"type:date;default:current_date" json:"date_played"`
	OutScore   *int     `json:"out_score"`
	InScore    *int     `json:"in_score"`
	Notes      *string  `gorm:"type:text" json:"notes"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`

	Course *CourseDB `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectDatabase() (*gorm.DB, error) {
	// Load environment variables - try current directory first, then parent
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found in current or parent directory")
		}
	}

	config := struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:   getEnvOrDefault("DB_NAME", "course_management"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	if config.Password == "" {
		return nil, fmt.Errorf("database password not set in environment variables")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

func formatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("Jan 2, 2006 at 3:04 PM")
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

func main() {
	fmt.Println("🔍 COURSE REVIEWS REPORT")
	fmt.Println("========================")
	fmt.Println()

	// Connect to database
	fmt.Println("🔌 Connecting to database...")
	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	fmt.Println("✅ Connected successfully!")
	fmt.Println()

	// Get first 10 reviews with user and course data
	var reviews []CourseReview
	result := db.Preload("User").Preload("Course").
		Order("created_at DESC").
		Limit(10).
		Find(&reviews)

	if result.Error != nil {
		log.Fatalf("❌ Failed to fetch reviews: %v", result.Error)
	}

	if len(reviews) == 0 {
		fmt.Println("📭 No reviews found in the database")
		fmt.Println()
		fmt.Println("💡 To create reviews:")
		fmt.Println("   1. Start the application: go run *.go")
		fmt.Println("   2. Login with Google")
		fmt.Println("   3. Click 'Review Course' and submit a review")
		return
	}

	fmt.Printf("📊 Found %d reviews (showing first 10):\n", len(reviews))
	fmt.Println()

	// Display reviews
	for i, review := range reviews {
		fmt.Printf("🎯 REVIEW #%d\n", i+1)
		fmt.Println("─────────────")

		// Course info
		courseName := "Unknown Course"
		courseAddress := "Unknown Address"
		if review.Course != nil {
			courseName = review.Course.Name
			courseAddress = review.Course.Address
		}
		fmt.Printf("🏌️  Course: %s\n", courseName)
		fmt.Printf("📍 Address: %s\n", courseAddress)

		// User info
		reviewerName := "Unknown User"
		if review.User != nil {
			reviewerName = review.User.Name
			if review.User.DisplayName != nil && *review.User.DisplayName != "" {
				reviewerName = *review.User.DisplayName
			}
		}
		fmt.Printf("👤 Reviewer: %s\n", reviewerName)

		// Review date
		fmt.Printf("📅 Date: %s\n", formatTimestamp(review.CreatedAt))

		// Ratings
		fmt.Printf("⭐ Overall Rating: %s\n", safeString(review.OverallRating))
		fmt.Printf("💰 Price: %s\n", safeString(review.Price))
		fmt.Printf("🎯 Handicap Difficulty: %s\n", safeInt(review.HandicapDifficulty))
		fmt.Printf("⚠️  Hazard Difficulty: %s\n", safeInt(review.HazardDifficulty))
		fmt.Printf("🛍️  Merch: %s\n", safeString(review.Merch))
		fmt.Printf("🌱 Condition: %s\n", safeString(review.Condition))
		fmt.Printf("😊 Enjoyment: %s\n", safeString(review.EnjoymentRating))
		fmt.Printf("✨ Vibe: %s\n", safeString(review.Vibe))
		fmt.Printf("🏹 Range: %s\n", safeString(review.RangeRating))
		fmt.Printf("🏨 Amenities: %s\n", safeString(review.Amenities))
		fmt.Printf("🌭 Turn Dog: %s\n", safeString(review.Glizzies))

		// Review text
		if review.ReviewText != nil && *review.ReviewText != "" {
			fmt.Printf("📝 Review:\n")
			reviewText := *review.ReviewText
			if len(reviewText) > 200 {
				reviewText = reviewText[:200] + "..."
			}
			fmt.Printf("   \"%s\"\n", reviewText)
		}

		fmt.Println()
	}

	// Get review stats
	var totalReviews int64
	db.Model(&CourseReview{}).Count(&totalReviews)

	var totalCourses int64
	db.Model(&CourseDB{}).Count(&totalCourses)

	var totalUsers int64
	db.Model(&User{}).Count(&totalUsers)

	// Get reviews by course
	var coursesWithReviews []struct {
		CourseID    uint
		CourseName  string
		ReviewCount int64
	}

	db.Table("course_reviews").
		Select("course_reviews.course_id, course_dbs.name as course_name, COUNT(*) as review_count").
		Joins("JOIN course_dbs ON course_reviews.course_id = course_dbs.id").
		Group("course_reviews.course_id, course_dbs.name").
		Order("review_count DESC").
		Limit(5).
		Scan(&coursesWithReviews)

	// Get most active reviewers
	var activeReviewers []struct {
		UserID      uint
		UserName    string
		ReviewCount int64
	}

	db.Table("course_reviews").
		Select("course_reviews.user_id, users.name as user_name, COUNT(*) as review_count").
		Joins("JOIN users ON course_reviews.user_id = users.id").
		Group("course_reviews.user_id, users.name").
		Order("review_count DESC").
		Limit(5).
		Scan(&activeReviewers)

	// Summary stats
	fmt.Println("📈 REVIEW STATISTICS")
	fmt.Println("====================")
	fmt.Printf("📊 Total Reviews: %d\n", totalReviews)
	fmt.Printf("🏌️  Total Courses: %d\n", totalCourses)
	fmt.Printf("👥 Total Users: %d\n", totalUsers)
	fmt.Println()

	if len(coursesWithReviews) > 0 {
		fmt.Println("🏆 TOP REVIEWED COURSES:")
		for i, course := range coursesWithReviews {
			fmt.Printf("   %d. %s (%d reviews)\n", i+1, course.CourseName, course.ReviewCount)
		}
		fmt.Println()
	}

	if len(activeReviewers) > 0 {
		fmt.Println("🌟 MOST ACTIVE REVIEWERS:")
		for i, reviewer := range activeReviewers {
			fmt.Printf("   %d. %s (%d reviews)\n", i+1, reviewer.UserName, reviewer.ReviewCount)
		}
		fmt.Println()
	}

	// Get scores if any exist
	var totalScores int64
	db.Model(&UserCourseScore{}).Count(&totalScores)

	if totalScores > 0 {
		fmt.Printf("⛳ Total Scores Posted: %d\n", totalScores)

		// Get recent scores
		var recentScores []UserCourseScore
		db.Preload("User").Preload("Course").
			Order("created_at DESC").
			Limit(5).
			Find(&recentScores)

		if len(recentScores) > 0 {
			fmt.Println("\n🎯 RECENT SCORES:")
			for _, score := range recentScores {
				userName := "Unknown User"
				if score.User != nil {
					userName = score.User.Name
					if score.User.DisplayName != nil && *score.User.DisplayName != "" {
						userName = *score.User.DisplayName
					}
				}

				courseName := "Unknown Course"
				if score.Course != nil {
					courseName = score.Course.Name
				}

				handicapStr := ""
				if score.Handicap != nil {
					handicapStr = fmt.Sprintf(" (%.1f handicap)", *score.Handicap)
				}

				fmt.Printf("   %s shot %d at %s%s\n", userName, score.Score, courseName, handicapStr)
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ Report completed!")
}
