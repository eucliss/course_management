# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

This is a Go-based golf course management system that allows users to track course ratings, reviews, and scores.

### Key Architecture Components

- **Backend**: Go with Echo framework for REST API and server-side rendering
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript for map functionality
- **Data Storage**: PostgreSQL with GORM ORM (migrated from JSON files)
- **Authentication**: Google OAuth2 integration
- **Templates**: HTML templates with Go templating engine
- **Cache**: Redis/in-memory caching for performance optimization
- **Hot Reload**: Air configuration for development

### Core Models

- **Course**: Contains course information, rankings, holes, and scores
- **User**: Handles authentication and user sessions
- **Ranking**: Structured ratings system for various course aspects
- **Score**: Individual user scores with handicap tracking
- **Reviews**: Multi-user course reviews and ratings

## Development Commands

- **Hot reload development**: Air is running automatically (see `.air.toml`)
- **Install dependencies**: `go mod download`
- **Run tests**: `go test ./...` (see detailed options below)
- **Database migration**: Run scripts in `scripts/` directory
- **Build application**: `go build .`

### Test Commands

Use these commands from the comprehensive test suite:

```bash
# Basic test commands
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -cover ./...             # Coverage report
go test -race ./...              # Race detection
go test -short ./...             # Skip slow integration tests

# Specific test categories
go test -v ./services            # Service layer tests
go test -v -run TestHandlers     # Handler tests
go test -v ./services -run TestRepository  # Repository tests

# Test with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Performance and debugging
go test -bench=. ./...           # Run benchmarks
go test -timeout 30s ./...       # Custom timeout
go test -count=1 ./...           # Disable cache
```

## Important Development Rules

### Technology Stack Constraints
- **Frontend**: Use HTMX and vanilla JavaScript only
- **Backend**: Go with Echo framework
- **Database**: PostgreSQL with GORM ORM
- **NO React, Vue, Svelte, TypeScript, or Angular**

### Development Practices
- **Air is used for hot reloading** - do not manually restart the application
- **CRITICAL**: After making any code changes, ALWAYS run `go build .` followed by `go test ./...` to ensure the application builds and all tests pass
- **Test Flexibility**: Use test flags as needed (see test commands above)
- If tests fail, fix the code to make them pass OR update the tests to reflect the updated logic in the codebase
- Write code but avoid running/starting the app unless specifically requested
- Place all documentation in the `docs/` folder
- Use existing patterns found in handlers.go and models.go. If they are insufficient, create new patterns that align with the existing codebase and deprecate or update the old patterns.

### Key Files and Directories

- `main.go`: Application entry point and server setup
- `models.go`: Data structures for courses, users, rankings
- `handlers.go`: HTTP request handlers
- `config/`: Configuration management with environment-specific settings
- `services/`: Service layer with business logic and repositories
- `auth_service.go`: Google OAuth authentication
- `course_service.go`: Course data management
- `database.go`: Database initialization and connection management
- `cache_service.go`: Redis/in-memory caching implementation
- `views/`: HTML templates
- `static/`: Static assets (CSS, JS, images)
- `scripts/`: Database migration and utility scripts
- `docs/`: Documentation files
- `testing/`: Test utilities and helpers

### Database Architecture

Uses PostgreSQL with GORM ORM:
- `users`: User profiles with handicap tracking
- `courses`: Course data with JSONB for complex course information
- `course_reviews`: Multi-user course reviews and ratings
- `user_scores`: Individual user scores per course
- `activities`: Activity feed for social features

### Service Layer Architecture

The application uses a service-oriented architecture:
- **Services**: Business logic layer (`services/` directory)
- **Repositories**: Data access layer with interfaces
- **Handlers**: HTTP layer that delegates to services
- **Models**: Shared data structures

## Common Patterns

- Use Echo's context for request handling
- Implement proper error handling with HTTP status codes
- Follow existing service layer patterns for business logic
- Use HTMX attributes for dynamic frontend interactions
- Maintain session state through Echo's session middleware
- Cache frequently accessed data using the cache service
- Use middleware for authentication and authorization checks