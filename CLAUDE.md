# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

This is a Go-based golf course management system that allows users to track course ratings, reviews, and scores.

### Key Architecture Components

- **Backend**: Go with Echo framework for REST API and server-side rendering
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript for map functionality
- **Data Storage**: JSON files for course data (with planned PostgreSQL migration)
- **Authentication**: Google OAuth2 integration
- **Templates**: HTML templates with Go templating engine

### Core Models

- **Course**: Contains course information, rankings, holes, and scores
- **User**: Handles authentication and user sessions
- **Ranking**: Structured ratings system for various course aspects
- **Score**: Individual user scores with handicap tracking

## Development Commands

- **Run the application**: `go run .` (Air is used for hot reloading)
- **Install dependencies**: `go mod download`
- **Database migration**: Run scripts in `scripts/` directory
- **Test environment**: Set up with `.env` file (copy from `.env.example`)

## Important Development Rules

### Technology Stack Constraints
- **Frontend**: Use HTMX and vanilla JavaScript only
- **Backend**: Go with Echo framework
- **Database**: Currently JSON files, migrating to PostgreSQL
- **NO React, Vue, Svelte, TypeScript, or Angular**

### Development Practices
- Air is used for hot reloading - do not manually restart the application
- Build the application with `go build .` to ensure no errors
- Write code but avoid running/starting the app unless specifically requested
- Place all documentation in the `docs/` folder
- Use existing patterns found in handlers.go and models.go. If they are insuffiecient, create new patterns that align with the existing codebase and deprecate or update the old patters.

### Key Files and Directories

- `main.go`: Application entry point and server setup
- `models.go`: Data structures for courses, users, rankings
- `handlers.go`: HTTP request handlers
- `config.go`: Configuration management
- `auth_service.go`: Google OAuth authentication
- `course_service.go`: Course data management
- `views/`: HTML templates
- `static/`: Static assets (CSS, JS, images)
- `courses/`: JSON course data files
- `scripts/`: Database migration and utility scripts
- `docs/`: Documentation files

### Database Architecture

Currently transitioning from JSON files to PostgreSQL with the following planned schema:
- `users`: User profiles with handicap tracking
- `courses`: Course data with JSONB for complex course information
- `course_reviews`: Multi-user course reviews and ratings
- `user_scores`: Individual user scores per course
- `activities`: Activity feed for social features

## Common Patterns

- Use Echo's context for request handling
- Implement proper error handling with HTTP status codes
- Follow existing JSON structure for course data
- Use HTMX attributes for dynamic frontend interactions
- Maintain session state through Echo's session middleware