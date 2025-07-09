# Testing Guide

This document provides comprehensive information about running and maintaining tests in the Course Management System.

## Table of Contents
- [Quick Start](#quick-start)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Writing Tests](#writing-tests)
- [CI/CD Pipeline](#cicd-pipeline)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites
- Go 1.24 or higher
- SQLite (automatically handled by Go modules)

### Run All Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run only fast tests (skip integration tests)
go test -short ./...
```

### Run Specific Test Categories
```bash
# Service layer tests only
go test -v ./services

# Handler integration tests only
go test -v -run TestHandlers

# Repository tests only
go test -v ./services -run TestRepository
```

## Test Structure

### Directory Layout
```
course_management/
├── testing/                    # Test utilities and helpers
│   ├── test_database.go       # Test database setup
│   └── test_helpers.go        # Test helper functions
├── services/                   # Service layer tests
│   ├── course_service_test.go # Course service unit tests
│   ├── auth_service_test.go   # Auth service unit tests
│   ├── review_service_test.go # Review service unit tests
│   └── repositories_test.go   # Repository integration tests
├── handlers_test.go           # HTTP handler tests
└── .github/workflows/test.yml # CI/CD pipeline
```

### Test Types

#### 1. Unit Tests (Service Layer)
- **Location**: `services/*_test.go`
- **Purpose**: Test business logic with mocked dependencies
- **Speed**: Fast (< 1ms per test)
- **Coverage**: Service interfaces, validation, error handling

#### 2. Integration Tests (Repository Layer)
- **Location**: `services/repositories_test.go`
- **Purpose**: Test database interactions with real SQLite
- **Speed**: Medium (10-100ms per test)
- **Coverage**: CRUD operations, data persistence

#### 3. Handler Tests
- **Location**: `handlers_test.go`
- **Purpose**: Test HTTP endpoints and middleware
- **Speed**: Fast (1-10ms per test)
- **Coverage**: Request/response handling, validation, security

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run tests with detailed output
go test -v ./...

# Run tests with coverage report
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...
```

### Test Filtering

```bash
# Run specific test function
go test -run TestCourseService ./services

# Run tests matching pattern
go test -run "TestCourse.*Create" ./services

# Run only short tests (skip slow integration tests)
go test -short ./...

# Run only integration tests
go test -run Integration ./...
```

### Service Layer Tests

```bash
# All service tests
go test -v ./services

# Course service tests only
go test -v ./services -run TestCourseService

# Auth service tests only
go test -v ./services -run TestAuthService

# Review service tests only
go test -v ./services -run TestReviewService
```

### Repository Tests

```bash
# All repository tests
go test -v ./services -run TestRepository

# Specific repository suite
go test -v ./services -run TestRepositorySuite

# Individual repository functions
go test -v ./services -run TestCourseRepository_Standalone
```

### Handler Tests

```bash
# All handler tests
go test -v -run TestHandlers

# Specific handler test categories
go test -v -run TestHandlers_Integration
go test -v -run TestHandlers_Validation
go test -v -run TestHandlers_Security
```

### Performance Tests

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkHandlers_Performance

# Benchmark with memory profiling
go test -bench=. -benchmem ./...
```

## Test Coverage

### Generate Coverage Report

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open coverage report in browser
open coverage.html
```

### Coverage by Package

```bash
# Services package coverage
go test -cover ./services

# Individual service coverage
go test -cover ./services -run TestCourseService

# Handler coverage
go test -cover -run TestHandlers
```

### Coverage Targets

- **Service Layer**: 95%+ coverage target
- **Repository Layer**: 90%+ coverage target
- **Handler Layer**: 85%+ coverage target
- **Overall Project**: 90%+ coverage target

## Writing Tests

### Service Layer Tests

```go
func TestCourseService_CreateCourse(t *testing.T) {
    // Setup
    mockRepo := new(MockCourseRepository)
    service := &courseService{courseRepo: mockRepo}
    ctx := context.Background()
    
    // Test data
    course := Course{
        Name:    "Test Course",
        Address: "123 Test St",
    }
    
    // Mock expectations
    mockRepo.On("Exists", ctx, course.Name, course.Address).Return(false, nil)
    mockRepo.On("Create", ctx, course, mock.AnythingOfType("*uint")).Return(nil)
    
    // Execute
    err := service.CreateCourse(ctx, course, &userID)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Repository Tests

```go
func TestCourseRepository_Create(t *testing.T) {
    // Setup test database
    testDB := testingPkg.NewTestDB(t)
    defer testDB.Close()
    
    repo := &courseRepository{db: testDB.DB}
    ctx := context.Background()
    
    // Test data
    course := Course{
        Name:    "Test Course",
        Address: "123 Test St",
    }
    
    // Execute
    err := repo.Create(ctx, course, &userID)
    
    // Assert
    assert.NoError(t, err)
    
    // Verify persistence
    courses, err := repo.GetAll(ctx)
    assert.NoError(t, err)
    assert.Len(t, courses, 1)
}
```

### Handler Tests

```go
func TestHandlers_CreateCourse(t *testing.T) {
    // Setup Echo
    e := echo.New()
    handler := NewHandlers()
    
    // Test data
    formData := "name=Test Course&address=123 Test St"
    req := httptest.NewRequest(http.MethodPost, "/courses", strings.NewReader(formData))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
    rec := httptest.NewRecorder()
    
    // Execute
    e.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
    assert.Contains(t, rec.Body.String(), "course created")
}
```

### Test Helpers

```go
// Use test helpers for common operations
func TestWithHelper(t *testing.T) {
    testingPkg.LogTestStart(t, "TestWithHelper")
    defer testingPkg.LogTestEnd(t, "TestWithHelper")
    
    ctx := testingPkg.TestContext(t)
    testDB := testingPkg.NewTestDB(t)
    defer testDB.Close()
    
    // Test logic here
    
    testingPkg.AssertNoError(t, err, "Operation should succeed")
}
```

## CI/CD Pipeline

### GitHub Actions Workflow

The automated testing pipeline runs on:
- **Push to main/develop branches**
- **Pull requests to main/develop**

### Pipeline Steps

1. **Test Job**
   - Sets up Go 1.24
   - Installs dependencies
   - Runs full test suite with race detection
   - Generates coverage report
   - Uploads coverage to Codecov

2. **Lint Job**
   - Runs golangci-lint
   - Checks code quality and style

3. **Security Job**
   - Runs Gosec security scanner
   - Uploads results to GitHub Security tab

4. **Build Job**
   - Builds application binary
   - Uploads build artifacts

### Local CI Simulation

```bash
# Run the same checks as CI
go test -v -race -cover ./...
golangci-lint run
gosec ./...
go build -v .
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Errors
```bash
# Error: no such table
# Solution: Ensure test database is properly migrated
go test -v ./services -run TestRepositorySuite
```

#### 2. Test Timeouts
```bash
# Error: test timed out
# Solution: Increase timeout or use -short flag
go test -timeout 60s ./...
go test -short ./...
```

#### 3. Race Conditions
```bash
# Error: race condition detected
# Solution: Fix concurrent access to shared resources
go test -race ./...
```

#### 4. Mock Expectations
```bash
# Error: mock expectations not met
# Solution: Verify all mock calls are properly set up
// In test code:
mockRepo.AssertExpectations(t)
```

### Debug Tests

```bash
# Run single test with verbose output
go test -v -run TestSpecificFunction ./package

# Run test with debugging
go test -v -run TestSpecificFunction ./package -args -test.v

# Run test with race detection
go test -race -run TestSpecificFunction ./package
```

### Test Environment Variables

```bash
# Set test environment
export GO_ENV=test

# Increase test timeout
export GO_TEST_TIMEOUT=60s

# Enable verbose logging
export LOG_LEVEL=debug
```

## Best Practices

### 1. Test Organization
- Group related tests in the same file
- Use descriptive test names
- Follow the pattern: `TestComponent_Method_Scenario`

### 2. Test Data
- Use test fixtures for consistent data
- Clean up test data after each test
- Use factories for creating test objects

### 3. Mocking
- Mock external dependencies
- Use interfaces for testability
- Verify mock expectations

### 4. Assertions
- Use descriptive assertion messages
- Test both success and failure cases
- Check error messages, not just error existence

### 5. Performance
- Use `-short` flag for fast feedback
- Run full test suite in CI
- Profile slow tests with benchmarks

## Quick Reference

### Essential Commands
```bash
# Development workflow
go test -short ./...              # Fast feedback
go test -cover ./...              # Coverage check
go test -race ./...               # Race detection
go test -v ./services             # Service tests
go test -v -run TestHandlers      # Handler tests

# CI/CD simulation
go test -v -race -cover ./...     # Full CI test
golangci-lint run                 # Linting
gosec ./...                       # Security scan
go build .                        # Build check
```

### Test Flags
- `-v`: Verbose output
- `-short`: Skip slow tests
- `-cover`: Coverage report
- `-race`: Race detection
- `-timeout`: Test timeout
- `-run`: Run specific tests
- `-bench`: Run benchmarks

### Coverage Commands
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

This testing infrastructure provides comprehensive coverage and ensures code quality throughout the development process. Use this guide to maintain and extend the test suite as the application grows.