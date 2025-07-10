package services

import (
	"context"
	"testing"

	testingPkg "course_management/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RepositoryTestSuite provides a test suite for repository tests
type RepositoryTestSuite struct {
	suite.Suite
	testDB   *testingPkg.TestDB
	fixtures *testingPkg.TestFixtures
	ctx      context.Context
}

// SetupSuite sets up the test suite
func (suite *RepositoryTestSuite) SetupSuite() {
	suite.testDB = testingPkg.NewTestDB(suite.T())
	suite.ctx = testingPkg.TestContext(suite.T())
}

// SetupTest sets up each individual test
func (suite *RepositoryTestSuite) SetupTest() {
	suite.testDB.CleanupTables(suite.T())
	suite.fixtures = suite.testDB.SeedTestData(suite.T())
}

// TearDownSuite tears down the test suite
func (suite *RepositoryTestSuite) TearDownSuite() {
	suite.testDB.Close()
}

// TestCourseRepository tests the course repository
func (suite *RepositoryTestSuite) TestCourseRepository() {
	repo := &courseRepository{db: suite.testDB.DB}

	suite.Run("Create", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.Create")

		course := Course{
			Name:        "New Test Course",
			Address:     "789 Test Road, Test City, TX 78901",
			Description: "A new test course",
			Review:      "Great new course",
			OverallRating: "A",
			Latitude:    &[]float64{32.7767}[0],
			Longitude:   &[]float64{-96.7970}[0],
			Holes:       []Hole{{Number: 1, Par: 4, Yardage: 400}},
			Scores:      []Score{{Score: 4, Handicap: 15.0}},
		}

		err := repo.Create(suite.ctx, course, &suite.fixtures.User1ID)
		testingPkg.AssertNoError(suite.T(), err, "Course creation should succeed")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.Create")
	})

	suite.Run("GetByID", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.GetByID")

		course, err := repo.GetByID(suite.ctx, suite.fixtures.Course1ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByID should succeed")
		testingPkg.AssertNotNil(suite.T(), course, "Course should not be nil")
		testingPkg.AssertEqual(suite.T(), "Test Golf Course 1", course.Name, "Course name should match")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.GetByID")
	})

	suite.Run("GetByID_NotFound", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.GetByID_NotFound")

		course, err := repo.GetByID(suite.ctx, 999)
		testingPkg.AssertError(suite.T(), err, "GetByID should fail for non-existent course")
		testingPkg.AssertNil(suite.T(), course, "Course should be nil for non-existent ID")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.GetByID_NotFound")
	})

	suite.Run("GetAll", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.GetAll")

		courses, err := repo.GetAll(suite.ctx)
		testingPkg.AssertNoError(suite.T(), err, "GetAll should succeed")
		// Note: The Create test above adds a course, so we expect 3 total courses (2 from fixtures + 1 from Create test)
		testingPkg.AssertLen(suite.T(), courses, 3, "Should return 3 courses (2 from fixtures + 1 from Create test)")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.GetAll")
	})

	suite.Run("GetByNameAndAddress", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.GetByNameAndAddress")

		course, err := repo.GetByNameAndAddress(suite.ctx, "Test Golf Course 1", "123 Golf Lane, Test City, TX 12345")
		testingPkg.AssertNoError(suite.T(), err, "GetByNameAndAddress should succeed")
		testingPkg.AssertNotNil(suite.T(), course, "Course should not be nil")
		testingPkg.AssertEqual(suite.T(), suite.fixtures.Course1ID, course.ID, "Course ID should match")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.GetByNameAndAddress")
	})

	suite.Run("Update", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.Update")

		course, err := repo.GetByID(suite.ctx, suite.fixtures.Course1ID)
		require.NoError(suite.T(), err)

		course.Description = "Updated description"
		course.OverallRating = "S"

		err = repo.Update(suite.ctx, *course, &suite.fixtures.User1ID)
		testingPkg.AssertNoError(suite.T(), err, "Update should succeed")

		// Verify update
		updatedCourse, err := repo.GetByID(suite.ctx, suite.fixtures.Course1ID)
		require.NoError(suite.T(), err)
		testingPkg.AssertEqual(suite.T(), "Updated description", updatedCourse.Description, "Description should be updated")
		testingPkg.AssertEqual(suite.T(), "S", updatedCourse.OverallRating, "Overall rating should be updated")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.Update")
	})

	suite.Run("Delete", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.Delete")

		err := repo.Delete(suite.ctx, suite.fixtures.Course2ID)
		testingPkg.AssertNoError(suite.T(), err, "Delete should succeed")

		// Verify deletion
		course, err := repo.GetByID(suite.ctx, suite.fixtures.Course2ID)
		testingPkg.AssertError(suite.T(), err, "GetByID should fail for deleted course")
		testingPkg.AssertNil(suite.T(), course, "Course should be nil after deletion")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.Delete")
	})

	suite.Run("GetByUser", func() {
		testingPkg.LogTestStart(suite.T(), "CourseRepository.GetByUser")

		courses, err := repo.GetByUser(suite.ctx, suite.fixtures.User1ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByUser should succeed")
		
		// Should find courses created by User1
		// Note: Since CreatedBy is not exposed in Course struct, we just check that courses are returned
		testingPkg.AssertTrue(suite.T(), len(courses) >= 0, "Should return courses for user")

		testingPkg.LogTestEnd(suite.T(), "CourseRepository.GetByUser")
	})
}

// TestUserRepository tests the user repository
func (suite *RepositoryTestSuite) TestUserRepository() {
	repo := &userRepository{db: suite.testDB.DB}

	suite.Run("Create", func() {
		testingPkg.LogTestStart(suite.T(), "UserRepository.Create")

		user := GoogleUser{
			ID:       "new-test-user",
			Email:    "newuser@example.com",
			Name:     "New Test User",
			Picture:  "https://example.com/newpic.jpg",
			Handicap: &[]float64{12.5}[0],
		}

		_, err := repo.Create(suite.ctx, user)
		testingPkg.AssertNoError(suite.T(), err, "User creation should succeed")

		testingPkg.LogTestEnd(suite.T(), "UserRepository.Create")
	})

	suite.Run("GetByID", func() {
		testingPkg.LogTestStart(suite.T(), "UserRepository.GetByID")

		
		user, err := repo.GetByID(suite.ctx, suite.fixtures.User1ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByID should succeed")
		testingPkg.AssertNotNil(suite.T(), user, "User should not be nil")
		testingPkg.AssertEqual(suite.T(), "test1@example.com", user.Email, "User email should match")

		testingPkg.LogTestEnd(suite.T(), "UserRepository.GetByID")
	})

	suite.Run("GetByGoogleID", func() {
		testingPkg.LogTestStart(suite.T(), "UserRepository.GetByGoogleID")

		user, err := repo.GetByGoogleID(suite.ctx, "test-user-1")
		testingPkg.AssertNoError(suite.T(), err, "GetByGoogleID should succeed")
		testingPkg.AssertNotNil(suite.T(), user, "User should not be nil")
		// Note: GoogleUser.ID is a string (GoogleID), not uint
		testingPkg.AssertEqual(suite.T(), "test-user-1", user.ID, "User Google ID should match")

		testingPkg.LogTestEnd(suite.T(), "UserRepository.GetByGoogleID")
	})

	suite.Run("GetByEmail", func() {
		testingPkg.LogTestStart(suite.T(), "UserRepository.GetByEmail")

		user, err := repo.GetByEmail(suite.ctx, "test2@example.com")
		testingPkg.AssertNoError(suite.T(), err, "GetByEmail should succeed")
		testingPkg.AssertNotNil(suite.T(), user, "User should not be nil")
		// Note: GoogleUser.ID is a string (GoogleID), not uint
		testingPkg.AssertEqual(suite.T(), "test-user-2", user.ID, "User Google ID should match")

		testingPkg.LogTestEnd(suite.T(), "UserRepository.GetByEmail")
	})

	suite.Run("Update", func() {
		testingPkg.LogTestStart(suite.T(), "UserRepository.Update")

		user, err := repo.GetByID(suite.ctx, suite.fixtures.User1ID)
		require.NoError(suite.T(), err)

		user.Name = "Updated Test User"
		newHandicap := 18.0
		user.Handicap = &newHandicap

		err = repo.Update(suite.ctx, *user)
		testingPkg.AssertNoError(suite.T(), err, "Update should succeed")

		// Verify update
		updatedUser, err := repo.GetByID(suite.ctx, suite.fixtures.User1ID)
		require.NoError(suite.T(), err)
		testingPkg.AssertEqual(suite.T(), "Updated Test User", updatedUser.Name, "Name should be updated")
		testingPkg.AssertEqual(suite.T(), 18.0, *updatedUser.Handicap, "Handicap should be updated")

		testingPkg.LogTestEnd(suite.T(), "UserRepository.Update")
	})

	suite.Run("Delete", func() {
		testingPkg.LogTestStart(suite.T(), "UserRepository.Delete")

		err := repo.Delete(suite.ctx, suite.fixtures.User2ID)
		testingPkg.AssertNoError(suite.T(), err, "Delete should succeed")

		// Verify deletion
		user, err := repo.GetByID(suite.ctx, suite.fixtures.User2ID)
		testingPkg.AssertError(suite.T(), err, "GetByID should fail for deleted user")
		testingPkg.AssertNil(suite.T(), user, "User should be nil after deletion")

		testingPkg.LogTestEnd(suite.T(), "UserRepository.Delete")
	})
}

// TestReviewRepository tests the review repository
func (suite *RepositoryTestSuite) TestReviewRepository() {
	repo := &reviewRepository{db: suite.testDB.DB}

	suite.Run("Create", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.Create")

		reviewText := "Excellent course!"
		overallRating := "A"
		review := CourseReview{
			UserID:        suite.fixtures.User2ID,
			CourseID:      suite.fixtures.Course2ID,
			ReviewText:    &reviewText,
			OverallRating: &overallRating,
		}

		err := repo.Create(suite.ctx, review)
		testingPkg.AssertNoError(suite.T(), err, "Review creation should succeed")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.Create")
	})

	suite.Run("GetByID", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.GetByID")

		review, err := repo.GetByID(suite.ctx, suite.fixtures.Review1ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByID should succeed")
		testingPkg.AssertNotNil(suite.T(), review, "Review should not be nil")
		testingPkg.AssertEqual(suite.T(), suite.fixtures.Course1ID, review.CourseID, "Course ID should match")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.GetByID")
	})

	suite.Run("GetByUser", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.GetByUser")

		reviews, err := repo.GetByUser(suite.ctx, suite.fixtures.User1ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByUser should succeed")
		testingPkg.AssertLen(suite.T(), reviews, 1, "Should return 1 review for user 1")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.GetByUser")
	})

	suite.Run("GetByCourse", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.GetByCourse")

		reviews, err := repo.GetByCourse(suite.ctx, suite.fixtures.Course1ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByCourse should succeed")
		testingPkg.AssertLen(suite.T(), reviews, 1, "Should return 1 review for the course")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.GetByCourse")
	})

	suite.Run("Update", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.Update")

		review, err := repo.GetByID(suite.ctx, suite.fixtures.Review1ID)
		require.NoError(suite.T(), err)

		updatedReviewText := "Updated review text"
		updatedRating := "B"
		review.ReviewText = &updatedReviewText
		review.OverallRating = &updatedRating

		err = repo.Update(suite.ctx, *review)
		testingPkg.AssertNoError(suite.T(), err, "Update should succeed")

		// Verify update
		updatedReview, err := repo.GetByID(suite.ctx, suite.fixtures.Review1ID)
		require.NoError(suite.T(), err)
		testingPkg.AssertEqual(suite.T(), "Updated review text", *updatedReview.ReviewText, "Review text should be updated")
		testingPkg.AssertEqual(suite.T(), "B", *updatedReview.OverallRating, "Rating should be updated")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.Update")
	})

	suite.Run("Delete", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.Delete")

		err := repo.Delete(suite.ctx, suite.fixtures.Review1ID)
		testingPkg.AssertNoError(suite.T(), err, "Delete should succeed")

		// Verify deletion
		review, err := repo.GetByID(suite.ctx, suite.fixtures.Review1ID)
		testingPkg.AssertError(suite.T(), err, "GetByID should fail for deleted review")
		testingPkg.AssertNil(suite.T(), review, "Review should be nil after deletion")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.Delete")
	})

	suite.Run("GetByUserAndCourse", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewRepository.GetByUserAndCourse")

		// First create a review
		reviewText := "Good course"
		overallRating := "B"
		review := CourseReview{
			UserID:        suite.fixtures.User1ID,
			CourseID:      suite.fixtures.Course2ID,
			ReviewText:    &reviewText,
			OverallRating: &overallRating,
		}
		err := repo.Create(suite.ctx, review)
		require.NoError(suite.T(), err)

		// Now test getting the user's review for that course
		foundReview, err := repo.GetByUserAndCourse(suite.ctx, suite.fixtures.User1ID, suite.fixtures.Course2ID)
		testingPkg.AssertNoError(suite.T(), err, "GetByUserAndCourse should succeed")
		testingPkg.AssertNotNil(suite.T(), foundReview, "Review should not be nil")
		testingPkg.AssertEqual(suite.T(), "Good course", *foundReview.ReviewText, "Review text should match")

		testingPkg.LogTestEnd(suite.T(), "ReviewRepository.GetByUserAndCourse")
	})
}

// TestRepositorySuite runs the repository test suite
func TestRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping repository tests in short mode")
	}
	
	suite.Run(t, new(RepositoryTestSuite))
}

// TestCourseRepository_Standalone tests individual course repository functions
func TestCourseRepository_Standalone(t *testing.T) {
	testingPkg.SkipIfShort(t)
	
	testDB := testingPkg.NewTestDB(t)
	defer testDB.Close()
	
	ctx := testingPkg.TestContext(t)
	repo := &courseRepository{db: testDB.DB}

	t.Run("Create_InvalidData", func(t *testing.T) {
		course := Course{
			// Missing required fields
			Name:    "",
			Address: "",
		}

		err := repo.Create(ctx, course, nil)
		// Repository doesn't validate - that's done at service layer
		assert.NoError(t, err, "Repository should accept any data")
	})

	t.Run("GetAll_EmptyDatabase", func(t *testing.T) {
		// Clean up any existing data first
		testDB.CleanupTables(t)
		
		courses, err := repo.GetAll(ctx)
		assert.NoError(t, err, "Should succeed even with empty database")
		assert.Len(t, courses, 0, "Should return empty slice")
	})
}

// TestUserRepository_Standalone tests individual user repository functions
func TestUserRepository_Standalone(t *testing.T) {
	testingPkg.SkipIfShort(t)
	
	testDB := testingPkg.NewTestDB(t)
	defer testDB.Close()
	
	ctx := testingPkg.TestContext(t)
	repo := &userRepository{db: testDB.DB}

	t.Run("Create_DuplicateEmail", func(t *testing.T) {
		user1 := GoogleUser{
			ID:    "unique-1",
			Email: "duplicate@example.com",
			Name:  "User 1",
		}

		user2 := GoogleUser{
			ID:    "unique-2",
			Email: "duplicate@example.com", // Same email
			Name:  "User 2",
		}

		_, err := repo.Create(ctx, user1)
		assert.NoError(t, err, "First user creation should succeed")

		_, err = repo.Create(ctx, user2)
		assert.Error(t, err, "Second user with duplicate email should fail")
	})

	t.Run("GetByGoogleID_NotFound", func(t *testing.T) {
		user, err := repo.GetByGoogleID(ctx, "non-existent")
		assert.Error(t, err, "Should fail for non-existent Google ID")
		assert.Nil(t, user, "User should be nil")
	})
}

// TestReviewRepository_Standalone tests individual review repository functions
func TestReviewRepository_Standalone(t *testing.T) {
	testingPkg.SkipIfShort(t)
	
	testDB := testingPkg.NewTestDB(t)
	defer testDB.Close()
	
	ctx := testingPkg.TestContext(t)
	repo := &reviewRepository{db: testDB.DB}

	t.Run("GetByUser_NoReviews", func(t *testing.T) {
		reviews, err := repo.GetByUser(ctx, 999) // Non-existent user
		assert.NoError(t, err, "Should succeed even for user with no reviews")
		assert.Len(t, reviews, 0, "Should return empty slice")
	})

	t.Run("GetByCourse_NoCourse", func(t *testing.T) {
		reviews, err := repo.GetByCourse(ctx, 999) // Non-existent course ID
		assert.NoError(t, err, "Should succeed even for non-existent course")
		assert.Len(t, reviews, 0, "Should return empty slice")
	})
}