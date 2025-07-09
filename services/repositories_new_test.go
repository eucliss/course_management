package services

import (
	"fmt"
	"testing"

	testingPkg "course_management/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepositoryBasic(t *testing.T) {
	testingPkg.SkipIfShort(t)

	testDB := testingPkg.NewTestDB(t)
	defer testDB.Close()

	ctx := testingPkg.TestContext(t)

	// Create new repository
	repo := NewCourseRepositoryNew(testDB.DB)

	// Test migration
	courseRepoNew, ok := repo.(*courseRepositoryNew)
	require.True(t, ok, "Repository should be of type courseRepositoryNew")

	// Run schema migration
	err := courseRepoNew.MigrateSchema(ctx)
	require.NoError(t, err, "Schema migration should succeed")

	// Test basic operations
	t.Run("CreateAndRetrieve", func(t *testing.T) {
		course := Course{
			Name:          "Test Course New Schema",
			Address:       "123 New Schema St, Test City, TX 12345",
			Description:   "A test course for new schema",
			OverallRating: "A",
			Holes: []Hole{
				{Number: 1, Par: 4, Yardage: 400, Description: "First hole"},
				{Number: 2, Par: 3, Yardage: 150, Description: "Second hole"},
			},
			Ranks: Ranking{
				Price:              "$50",
				HandicapDifficulty: 6,
				HazardDifficulty:   5,
				Merch:              "A",
				Condition:          "A",
				EnjoymentRating:    "A",
				Vibe:               "A",
				Range:              "A",
				Amenities:          "A",
				Glizzies:           "A",
			},
			Scores: []Score{
				{Score: 80, Handicap: 15.0},
			},
		}

		userID := uint(1)

		// Create course
		err := repo.Create(ctx, course, &userID)
		require.NoError(t, err, "Course creation should succeed")

		// Retrieve by name
		retrievedCourse, err := repo.GetByName(ctx, course.Name)
		require.NoError(t, err, "Course retrieval should succeed")
		assert.Equal(t, course.Name, retrievedCourse.Name, "Course name should match")
		assert.Equal(t, course.Address, retrievedCourse.Address, "Course address should match")
		assert.Equal(t, course.Description, retrievedCourse.Description, "Course description should match")
		assert.Equal(t, course.OverallRating, retrievedCourse.OverallRating, "Course rating should match")

		// Verify holes
		assert.Len(t, retrievedCourse.Holes, 2, "Should have 2 holes")
		assert.Equal(t, 1, retrievedCourse.Holes[0].Number, "First hole number should be 1")
		assert.Equal(t, 4, retrievedCourse.Holes[0].Par, "First hole par should be 4")

		// Verify rankings
		assert.Equal(t, "$50", retrievedCourse.Ranks.Price, "Price should match")
		assert.Equal(t, 6, retrievedCourse.Ranks.HandicapDifficulty, "Handicap difficulty should match")

		// Verify scores
		assert.Len(t, retrievedCourse.Scores, 1, "Should have 1 score")
		assert.Equal(t, 80, retrievedCourse.Scores[0].Score, "Score should match")
	})

	t.Run("GetAll", func(t *testing.T) {
		courses, err := repo.GetAll(ctx)
		require.NoError(t, err, "GetAll should succeed")
		assert.GreaterOrEqual(t, len(courses), 1, "Should have at least 1 course")
	})
}

func TestNewRepositoryPerformance(t *testing.T) {
	testingPkg.SkipIfShort(t)

	testDB := testingPkg.NewTestDB(t)
	defer testDB.Close()

	ctx := testingPkg.TestContext(t)

	// Create new repository
	repo := NewCourseRepositoryNew(testDB.DB)

	// Test migration
	courseRepoNew, ok := repo.(*courseRepositoryNew)
	require.True(t, ok, "Repository should be of type courseRepositoryNew")

	// Run schema migration
	err := courseRepoNew.MigrateSchema(ctx)
	require.NoError(t, err, "Schema migration should succeed")

	// Create multiple courses for performance testing
	numCourses := 10
	for i := 0; i < numCourses; i++ {
		course := Course{
			Name:          fmt.Sprintf("Performance Test Course %d", i),
			Address:       fmt.Sprintf("123 Performance St %d, Test City, TX 1234%d", i, i),
			Description:   fmt.Sprintf("Performance test course %d", i),
			OverallRating: "B",
			Holes: []Hole{
				{Number: 1, Par: 4, Yardage: 400, Description: "Test hole 1"},
				{Number: 2, Par: 3, Yardage: 150, Description: "Test hole 2"},
			},
		}

		userID := uint(1)
		err := repo.Create(ctx, course, &userID)
		require.NoError(t, err, "Course creation should succeed")
	}

	// Test bulk retrieval performance
	t.Run("BulkRetrievalPerformance", func(t *testing.T) {
		courses, err := repo.GetAll(ctx)
		require.NoError(t, err, "GetAll should succeed")
		assert.GreaterOrEqual(t, len(courses), numCourses, "Should have at least the created courses")

		// Verify that all courses have their related data loaded
		for _, course := range courses {
			// Each course should have its holes and rankings loaded
			if len(course.Holes) > 0 {
				assert.Greater(t, course.Holes[0].Number, 0, "Hole number should be set")
			}
		}
	})
}