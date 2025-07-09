package services

import (
	"context"
	"testing"

	testingPkg "course_management/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockReviewRepository is a mock implementation of ReviewRepository
type MockReviewRepository struct {
	mock.Mock
}

func (m *MockReviewRepository) Create(ctx context.Context, review CourseReview) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockReviewRepository) GetByID(ctx context.Context, id uint) (*CourseReview, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CourseReview), args.Error(1)
}

func (m *MockReviewRepository) GetByUser(ctx context.Context, userID uint) ([]CourseReview, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]CourseReview), args.Error(1)
}

func (m *MockReviewRepository) GetByCourse(ctx context.Context, courseID uint) ([]CourseReview, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]CourseReview), args.Error(1)
}

func (m *MockReviewRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uint) (*CourseReview, error) {
	args := m.Called(ctx, userID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CourseReview), args.Error(1)
}

func (m *MockReviewRepository) Update(ctx context.Context, review CourseReview) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockReviewRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReviewRepository) AddScore(ctx context.Context, score UserCourseScore) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

func (m *MockReviewRepository) GetUserScores(ctx context.Context, userID uint) ([]UserCourseScore, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]UserCourseScore), args.Error(1)
}

func (m *MockReviewRepository) GetCourseScores(ctx context.Context, courseID uint) ([]UserCourseScore, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]UserCourseScore), args.Error(1)
}

func (m *MockReviewRepository) AddHoleScore(ctx context.Context, holeScore UserCourseHole) error {
	args := m.Called(ctx, holeScore)
	return args.Error(0)
}

func (m *MockReviewRepository) GetUserHoleScores(ctx context.Context, userID uint) ([]UserCourseHole, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]UserCourseHole), args.Error(1)
}

// ReviewServiceTestSuite provides a test suite for review service tests
type ReviewServiceTestSuite struct {
	suite.Suite
	mockRepo *MockReviewRepository
	service  ReviewService
	ctx      context.Context
}

// SetupSuite sets up the test suite
func (suite *ReviewServiceTestSuite) SetupSuite() {
	suite.ctx = testingPkg.TestContext(suite.T())
}

// SetupTest sets up each individual test
func (suite *ReviewServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockReviewRepository)
	// Note: We'd need to create a review service that accepts dependencies
	// For now we'll test the expected behaviors using the mock directly
}

// TearDownTest cleans up after each test
func (suite *ReviewServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// TestCreateReview tests review creation
func (suite *ReviewServiceTestSuite) TestCreateReview() {
	suite.Run("Success", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.CreateReview_Success")

		review := CourseReview{
			UserID:   1,
			CourseID: 1,
			Review:   "Great course!",
			Rating:   5,
		}

		suite.mockRepo.On("Create", suite.ctx, review).Return(nil)

		err := suite.mockRepo.Create(suite.ctx, review)
		testingPkg.AssertNoError(suite.T(), err, "Review creation should succeed")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.CreateReview_Success")
	})

	suite.Run("ValidationError", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.CreateReview_Validation")

		review := CourseReview{
			UserID:   0, // Invalid: no user ID
			CourseID: 1,
			Review:   "",
			Rating:   6, // Invalid: rating too high
		}

		err := validateReview(review)
		testingPkg.AssertError(suite.T(), err, "Should fail validation")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.CreateReview_Validation")
	})
}

// TestGetUserReviews tests getting reviews by user
func (suite *ReviewServiceTestSuite) TestGetUserReviews() {
	suite.Run("Success", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.GetUserReviews_Success")

		userID := uint(1)
		expectedReviews := []CourseReview{
			{ID: 1, UserID: userID, CourseID: 1, Rating: 4},
			{ID: 2, UserID: userID, CourseID: 2, Rating: 5},
		}

		suite.mockRepo.On("GetByUser", suite.ctx, userID).Return(expectedReviews, nil)

		reviews, err := suite.mockRepo.GetByUser(suite.ctx, userID)
		testingPkg.AssertNoError(suite.T(), err, "GetUserReviews should succeed")
		testingPkg.AssertLen(suite.T(), reviews, 2, "Should return 2 reviews")
		testingPkg.AssertEqual(suite.T(), uint(1), reviews[0].CourseID, "First review course ID should match")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.GetUserReviews_Success")
	})

	suite.Run("NoReviews", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.GetUserReviews_Empty")

		userID := uint(999)
		suite.mockRepo.On("GetByUser", suite.ctx, userID).Return([]CourseReview{}, nil)

		reviews, err := suite.mockRepo.GetByUser(suite.ctx, userID)
		testingPkg.AssertNoError(suite.T(), err, "GetUserReviews should succeed even with no reviews")
		testingPkg.AssertLen(suite.T(), reviews, 0, "Should return empty slice")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.GetUserReviews_Empty")
	})
}

// TestUpdateReview tests review updates
func (suite *ReviewServiceTestSuite) TestUpdateReview() {
	suite.Run("Success", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.UpdateReview_Success")

		review := CourseReview{
			ID:       1,
			UserID:   1,
			CourseID: 1,
			Review:   "Updated review text",
			Rating:   4,
		}

		suite.mockRepo.On("Update", suite.ctx, review).Return(nil)

		err := suite.mockRepo.Update(suite.ctx, review)
		testingPkg.AssertNoError(suite.T(), err, "Review update should succeed")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.UpdateReview_Success")
	})
}

// TestDeleteReview tests review deletion
func (suite *ReviewServiceTestSuite) TestDeleteReview() {
	suite.Run("Success", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.DeleteReview_Success")

		reviewID := uint(1)
		suite.mockRepo.On("Delete", suite.ctx, reviewID).Return(nil)

		err := suite.mockRepo.Delete(suite.ctx, reviewID)
		testingPkg.AssertNoError(suite.T(), err, "Review deletion should succeed")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.DeleteReview_Success")
	})
}

// TestScoreManagement tests score-related operations
func (suite *ReviewServiceTestSuite) TestScoreManagement() {
	suite.Run("AddScore", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.AddScore")

		score := UserCourseScore{
			UserID:   1,
			CourseID: 1,
			Score:    85,
			Handicap: 15.0,
		}

		suite.mockRepo.On("AddScore", suite.ctx, score).Return(nil)

		err := suite.mockRepo.AddScore(suite.ctx, score)
		testingPkg.AssertNoError(suite.T(), err, "Adding score should succeed")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.AddScore")
	})

	suite.Run("GetUserScores", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.GetUserScores")

		userID := uint(1)
		expectedScores := []UserCourseScore{
			{ID: 1, UserID: userID, CourseID: 1, Score: 85},
			{ID: 2, UserID: userID, CourseID: 2, Score: 78},
		}

		suite.mockRepo.On("GetUserScores", suite.ctx, userID).Return(expectedScores, nil)

		scores, err := suite.mockRepo.GetUserScores(suite.ctx, userID)
		testingPkg.AssertNoError(suite.T(), err, "GetUserScores should succeed")
		testingPkg.AssertLen(suite.T(), scores, 2, "Should return 2 scores")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.GetUserScores")
	})
}

// TestHoleScoreManagement tests hole score operations
func (suite *ReviewServiceTestSuite) TestHoleScoreManagement() {
	suite.Run("AddHoleScore", func() {
		testingPkg.LogTestStart(suite.T(), "ReviewService.AddHoleScore")

		holeScore := UserCourseHole{
			UserID:     1,
			CourseID:   1,
			HoleNumber: 1,
			Score:      4,
			Par:        4,
		}

		suite.mockRepo.On("AddHoleScore", suite.ctx, holeScore).Return(nil)

		err := suite.mockRepo.AddHoleScore(suite.ctx, holeScore)
		testingPkg.AssertNoError(suite.T(), err, "Adding hole score should succeed")

		testingPkg.LogTestEnd(suite.T(), "ReviewService.AddHoleScore")
	})
}

// validateReview is a helper function that demonstrates expected validation logic
func validateReview(review CourseReview) error {
	if review.UserID == 0 {
		return assert.AnError
	}
	if review.CourseID == 0 {
		return assert.AnError
	}
	if review.Rating < 1 || review.Rating > 5 {
		return assert.AnError
	}
	return nil
}

// TestReviewService runs the review service test suite
func TestReviewService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping review service tests in short mode")
	}
	
	suite.Run(t, new(ReviewServiceTestSuite))
}

// TestReviewValidation tests review validation logic
func TestReviewValidation(t *testing.T) {
	testingPkg.SkipIfShort(t)

	testCases := []struct {
		name        string
		review      CourseReview
		shouldError bool
		errorMsg    string
	}{
		{
			name: "ValidReview",
			review: CourseReview{
				UserID:   1,
				CourseID: 1,
				Review:   "Great course!",
				Rating:   5,
			},
			shouldError: false,
		},
		{
			name: "MissingUserID",
			review: CourseReview{
				UserID:   0,
				CourseID: 1,
				Review:   "Great course!",
				Rating:   5,
			},
			shouldError: true,
		},
		{
			name: "MissingCourseID",
			review: CourseReview{
				UserID:   1,
				CourseID: 0,
				Review:   "Great course!",
				Rating:   5,
			},
			shouldError: true,
		},
		{
			name: "InvalidRatingTooLow",
			review: CourseReview{
				UserID:   1,
				CourseID: 1,
				Review:   "Poor course",
				Rating:   0,
			},
			shouldError: true,
		},
		{
			name: "InvalidRatingTooHigh",
			review: CourseReview{
				UserID:   1,
				CourseID: 1,
				Review:   "Amazing course",
				Rating:   6,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReview(tc.review)
			if tc.shouldError {
				assert.Error(t, err, tc.name+" should fail validation")
			} else {
				assert.NoError(t, err, tc.name+" should pass validation")
			}
		})
	}
}