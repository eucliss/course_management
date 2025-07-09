package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// CourseRepository implementation using new relational schema
type courseRepositoryNew struct {
	db *gorm.DB
}

func NewCourseRepositoryNew(db *gorm.DB) CourseRepository {
	return &courseRepositoryNew{db: db}
}

func (r *courseRepositoryNew) Create(ctx context.Context, course Course, createdBy *uint) error {
	var courseDB CourseNewDB
	courseDB.FromCourse(course)
	courseDB.CreatedBy = createdBy

	// Use transaction to ensure all related data is created atomically
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create course
		if err := tx.Create(&courseDB).Error; err != nil {
			return fmt.Errorf("failed to create course: %w", err)
		}

		// Update foreign keys for related entities
		for i := range courseDB.Holes {
			courseDB.Holes[i].CourseID = courseDB.ID
		}
		if courseDB.Rankings != nil {
			courseDB.Rankings.CourseID = courseDB.ID
		}
		for i := range courseDB.Scores {
			courseDB.Scores[i].CourseID = courseDB.ID
		}

		return nil
	})
}

func (r *courseRepositoryNew) GetByID(ctx context.Context, id uint) (*Course, error) {
	var courseDB CourseNewDB
	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		First(&courseDB, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("course not found")
		}
		return nil, err
	}

	course := courseDB.ToCourse()
	return &course, nil
}

func (r *courseRepositoryNew) GetByIndex(ctx context.Context, index int) (*Course, error) {
	var courseIDs []uint
	if err := r.db.WithContext(ctx).Model(&CourseNewDB{}).
		Select("id").
		Order("created_at ASC").
		Find(&courseIDs).Error; err != nil {
		return nil, err
	}

	if index < 0 || index >= len(courseIDs) {
		return nil, fmt.Errorf("course index out of range")
	}

	return r.GetByID(ctx, courseIDs[index])
}

func (r *courseRepositoryNew) GetByName(ctx context.Context, name string) (*Course, error) {
	var courseDB CourseNewDB
	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Where("name = ?", name).
		First(&courseDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("course not found")
		}
		return nil, err
	}

	course := courseDB.ToCourse()
	return &course, nil
}

func (r *courseRepositoryNew) GetByNameAndAddress(ctx context.Context, name, address string) (*Course, error) {
	var courseDB CourseNewDB
	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Where("name = ? AND address = ?", name, address).
		First(&courseDB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("course not found")
		}
		return nil, err
	}

	course := courseDB.ToCourse()
	return &course, nil
}

func (r *courseRepositoryNew) GetAll(ctx context.Context) ([]Course, error) {
	var coursesDB []CourseNewDB
	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Order("created_at ASC").
		Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, len(coursesDB))
	for i, courseDB := range coursesDB {
		courses[i] = courseDB.ToCourse()
	}

	return courses, nil
}

func (r *courseRepositoryNew) Update(ctx context.Context, course Course, updatedBy *uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update basic course information
		updates := map[string]interface{}{
			"name":           course.Name,
			"address":        course.Address,
			"description":    course.Description,
			"overall_rating": course.OverallRating,
			"review":         course.Review,
			"latitude":       course.Latitude,
			"longitude":      course.Longitude,
			"updated_by":     updatedBy,
		}

		result := tx.Model(&CourseNewDB{}).Where("id = ?", course.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("course not found")
		}

		// Update holes - delete and recreate for simplicity
		if err := tx.Where("course_id = ?", course.ID).Delete(&CourseHoleNewDB{}).Error; err != nil {
			return fmt.Errorf("failed to delete old holes: %w", err)
		}

		for _, hole := range course.Holes {
			holeDB := CourseHoleNewDB{
				CourseID:    course.ID,
				HoleNumber:  hole.Number,
				Par:         hole.Par,
				Yardage:     hole.Yardage,
				Description: hole.Description,
			}
			if err := tx.Create(&holeDB).Error; err != nil {
				return fmt.Errorf("failed to create hole %d: %w", hole.Number, err)
			}
		}

		// Update rankings
		if course.Ranks != (Ranking{}) {
			rankingDB := CourseRankingNewDB{
				CourseID:           course.ID,
				Price:              course.Ranks.Price,
				HandicapDifficulty: course.Ranks.HandicapDifficulty,
				HazardDifficulty:   course.Ranks.HazardDifficulty,
				Merch:              course.Ranks.Merch,
				Condition:          course.Ranks.Condition,
				EnjoymentRating:    course.Ranks.EnjoymentRating,
				Vibe:               course.Ranks.Vibe,
				RangeRating:        course.Ranks.Range,
				Amenities:          course.Ranks.Amenities,
				Glizzies:           course.Ranks.Glizzies,
			}

			// Upsert rankings
			if err := tx.Where("course_id = ?", course.ID).
				Assign(rankingDB).
				FirstOrCreate(&rankingDB).Error; err != nil {
				return fmt.Errorf("failed to update rankings: %w", err)
			}
		}

		return nil
	})
}

func (r *courseRepositoryNew) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&CourseNewDB{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("course not found")
	}

	// Related data will be deleted automatically due to CASCADE constraints
	return nil
}

func (r *courseRepositoryNew) GetByUser(ctx context.Context, userID uint) ([]Course, error) {
	var coursesDB []CourseNewDB
	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Where("created_by = ?", userID).
		Order("created_at ASC").
		Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, len(coursesDB))
	for i, courseDB := range coursesDB {
		courses[i] = courseDB.ToCourse()
	}

	return courses, nil
}

func (r *courseRepositoryNew) CanEdit(ctx context.Context, courseID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseNewDB{}).
		Where("id = ? AND created_by = ?", courseID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *courseRepositoryNew) CanEditByIndex(ctx context.Context, index int, userID uint) (bool, error) {
	course, err := r.GetByIndex(ctx, index)
	if err != nil {
		return false, err
	}
	return r.CanEdit(ctx, course.ID, userID)
}

func (r *courseRepositoryNew) IsOwner(ctx context.Context, userID uint, courseName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseNewDB{}).
		Where("name = ? AND created_by = ?", courseName, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *courseRepositoryNew) GetWithPagination(ctx context.Context, offset, limit int) ([]Course, int64, error) {
	var coursesDB []CourseNewDB
	var totalCount int64

	if err := r.db.WithContext(ctx).Model(&CourseNewDB{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&coursesDB).Error; err != nil {
		return nil, 0, err
	}

	courses := make([]Course, len(coursesDB))
	for i, courseDB := range coursesDB {
		courses[i] = courseDB.ToCourse()
	}

	return courses, totalCount, nil
}

func (r *courseRepositoryNew) GetByUserWithPagination(ctx context.Context, userID uint, offset, limit int) ([]Course, int64, error) {
	var coursesDB []CourseNewDB
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&CourseNewDB{}).Where("created_by = ?", userID)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&coursesDB).Error; err != nil {
		return nil, 0, err
	}

	courses := make([]Course, len(coursesDB))
	for i, courseDB := range coursesDB {
		courses[i] = courseDB.ToCourse()
	}

	return courses, totalCount, nil
}

func (r *courseRepositoryNew) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseNewDB{}).Count(&count).Error
	return count, err
}

func (r *courseRepositoryNew) Exists(ctx context.Context, name, address string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CourseNewDB{}).
		Where("name = ? AND address = ?", name, address).
		Count(&count).Error
	return count > 0, err
}

func (r *courseRepositoryNew) GetAvailableForReview(ctx context.Context, userID uint) ([]Course, error) {
	var coursesDB []CourseNewDB
	
	// Get courses that the user hasn't reviewed yet
	subQuery := r.db.Select("course_id").
		Where("user_id = ?", userID).
		Table("course_reviews")
	
	if err := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores").
		Where("id NOT IN (?)", subQuery).
		Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, len(coursesDB))
	for i, courseDB := range coursesDB {
		courses[i] = courseDB.ToCourse()
	}

	return courses, nil
}

// Search method for enhanced querying capabilities
func (r *courseRepositoryNew) Search(ctx context.Context, name, address string) ([]Course, error) {
	var coursesDB []CourseNewDB
	
	query := r.db.WithContext(ctx).
		Preload("Holes").
		Preload("Rankings").
		Preload("Scores")

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if address != "" {
		query = query.Where("address ILIKE ?", "%"+address+"%")
	}

	if err := query.Order("created_at ASC").Find(&coursesDB).Error; err != nil {
		return nil, err
	}

	courses := make([]Course, len(coursesDB))
	for i, courseDB := range coursesDB {
		courses[i] = courseDB.ToCourse()
	}

	return courses, nil
}

// Helper method to migrate existing table
func (r *courseRepositoryNew) MigrateSchema(ctx context.Context) error {
	log.Println("🔄 Running database schema migration...")
	
	// Auto-migrate new tables
	if err := r.db.WithContext(ctx).AutoMigrate(
		&CourseNewDB{},
		&CourseHoleNewDB{},
		&CourseRankingNewDB{},
		&UserCourseScoreNewDB{},
	); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	log.Println("✅ Database schema migration completed")
	return nil
}