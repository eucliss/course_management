package services

// New relational database models for proper schema
// These replace the JSON-based storage with proper relational structure

// CourseNewDB represents the new relational course model
type CourseNewDB struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"size:100;not null;index" json:"name"`
	Address       string    `gorm:"type:text;not null" json:"address"`
	Description   string    `gorm:"type:text" json:"description"`
	City          string    `gorm:"size:50;index" json:"city"`
	State         string    `gorm:"size:2" json:"state"`
	ZipCode       string    `gorm:"size:10" json:"zip_code"`
	Phone         string    `gorm:"size:20" json:"phone"`
	Website       string    `gorm:"size:255" json:"website"`
	OverallRating string    `gorm:"size:1;check:overall_rating IN ('','S','A','B','C','D','F')" json:"overall_rating"`
	Review        string    `gorm:"type:text" json:"review"`
	Hash          string    `gorm:"uniqueIndex;not null" json:"hash"`
	Latitude      *float64  `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude     *float64  `gorm:"type:decimal(11,8)" json:"longitude"`
	CreatedBy     *uint     `gorm:"index" json:"created_by"`
	UpdatedBy     *uint     `json:"updated_by"`
	CreatedAt     int64     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     int64     `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	Holes    []CourseHoleNewDB    `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"holes"`
	Rankings *CourseRankingNewDB  `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"rankings"`
	Scores   []UserCourseScoreNewDB `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"scores"`
}

// CourseHoleNewDB represents individual holes for a course
type CourseHoleNewDB struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID    uint   `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"course_id"`
	HoleNumber  int    `gorm:"not null;check:hole_number BETWEEN 1 AND 18" json:"hole_number"`
	Par         int    `gorm:"check:par BETWEEN 3 AND 6" json:"par"`
	Yardage     int    `gorm:"check:yardage BETWEEN 0 AND 800" json:"yardage"`
	Description string `gorm:"type:text" json:"description"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
}

// CourseRankingNewDB represents the ranking/rating information for a course
type CourseRankingNewDB struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID           uint   `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE" json:"course_id"`
	Price              string `gorm:"size:10" json:"price"`
	HandicapDifficulty int    `gorm:"check:handicap_difficulty BETWEEN 1 AND 10" json:"handicap_difficulty"`
	HazardDifficulty   int    `gorm:"check:hazard_difficulty BETWEEN 1 AND 10" json:"hazard_difficulty"`
	Merch              string `gorm:"size:1;check:merch IN ('','S','A','B','C','D','F')" json:"merch"`
	Condition          string `gorm:"size:1;check:condition IN ('','S','A','B','C','D','F')" json:"condition"`
	EnjoymentRating    string `gorm:"size:1;check:enjoyment_rating IN ('','S','A','B','C','D','F')" json:"enjoyment_rating"`
	Vibe               string `gorm:"size:1;check:vibe IN ('','S','A','B','C','D','F')" json:"vibe"`
	RangeRating        string `gorm:"size:1;check:range_rating IN ('','S','A','B','C','D','F')" json:"range_rating"`
	Amenities          string `gorm:"size:1;check:amenities IN ('','S','A','B','C','D','F')" json:"amenities"`
	Glizzies           string `gorm:"size:1;check:glizzies IN ('','S','A','B','C','D','F')" json:"glizzies"`
	CreatedAt          int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// UserCourseScoreNewDB represents user scores for courses
type UserCourseScoreNewDB struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint    `gorm:"not null;index" json:"user_id"`
	CourseID  uint    `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"course_id"`
	Score     int     `gorm:"not null;check:score BETWEEN 1 AND 200" json:"score"`
	Handicap  float64 `gorm:"type:decimal(4,1);check:handicap BETWEEN -5 AND 40" json:"handicap"`
	CreatedAt int64   `gorm:"autoCreateTime" json:"created_at"`
}

// Table name mappings
func (CourseNewDB) TableName() string {
	return "courses"
}

func (CourseHoleNewDB) TableName() string {
	return "course_holes"
}

func (CourseRankingNewDB) TableName() string {
	return "course_rankings"
}

func (UserCourseScoreNewDB) TableName() string {
	return "user_course_scores"
}

// Conversion methods from database models to service models

// ToCourse converts CourseNewDB to Course model for service layer
func (cdb *CourseNewDB) ToCourse() Course {
	course := Course{
		ID:            uint(cdb.ID),
		Name:          cdb.Name,
		Address:       cdb.Address,
		Description:   cdb.Description,
		OverallRating: cdb.OverallRating,
		Review:        cdb.Review,
		Latitude:      cdb.Latitude,
		Longitude:     cdb.Longitude,
		Holes:         make([]Hole, len(cdb.Holes)),
		Scores:        make([]Score, len(cdb.Scores)),
	}

	// Convert holes
	for i, hole := range cdb.Holes {
		course.Holes[i] = Hole{
			Number:      hole.HoleNumber,
			Par:         hole.Par,
			Yardage:     hole.Yardage,
			Description: hole.Description,
		}
	}

	// Convert rankings
	if cdb.Rankings != nil {
		course.Ranks = Ranking{
			Price:              cdb.Rankings.Price,
			HandicapDifficulty: cdb.Rankings.HandicapDifficulty,
			HazardDifficulty:   cdb.Rankings.HazardDifficulty,
			Merch:              cdb.Rankings.Merch,
			Condition:          cdb.Rankings.Condition,
			EnjoymentRating:    cdb.Rankings.EnjoymentRating,
			Vibe:               cdb.Rankings.Vibe,
			Range:              cdb.Rankings.RangeRating,
			Amenities:          cdb.Rankings.Amenities,
			Glizzies:           cdb.Rankings.Glizzies,
		}
	}

	// Convert scores
	for i, score := range cdb.Scores {
		course.Scores[i] = Score{
			Score:    score.Score,
			Handicap: score.Handicap,
		}
	}

	return course
}

// FromCourse converts Course model to CourseNewDB for database storage
func (cdb *CourseNewDB) FromCourse(course Course) {
	cdb.ID = course.ID
	cdb.Name = course.Name
	cdb.Address = course.Address
	cdb.Description = course.Description
	cdb.OverallRating = course.OverallRating
	cdb.Review = course.Review
	cdb.Latitude = course.Latitude
	cdb.Longitude = course.Longitude

	// Generate hash for uniqueness
	cdb.Hash = course.Name + "|" + course.Address

	// Convert holes
	cdb.Holes = make([]CourseHoleNewDB, len(course.Holes))
	for i, hole := range course.Holes {
		cdb.Holes[i] = CourseHoleNewDB{
			CourseID:    cdb.ID,
			HoleNumber:  hole.Number,
			Par:         hole.Par,
			Yardage:     hole.Yardage,
			Description: hole.Description,
		}
	}

	// Convert rankings
	if course.Ranks != (Ranking{}) {
		cdb.Rankings = &CourseRankingNewDB{
			CourseID:           cdb.ID,
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
	}

	// Convert scores
	cdb.Scores = make([]UserCourseScoreNewDB, len(course.Scores))
	for i, score := range course.Scores {
		cdb.Scores[i] = UserCourseScoreNewDB{
			CourseID: cdb.ID,
			Score:    score.Score,
			Handicap: score.Handicap,
			// Note: UserID will need to be set separately
		}
	}
}