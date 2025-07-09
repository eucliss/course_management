package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator provides input validation methods
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequired checks if a field is not empty
func (v *Validator) ValidateRequired(field, value, fieldName string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s is required", fieldName),
		}
	}
	return nil
}

// ValidateLength checks if a string is within the specified length range
func (v *Validator) ValidateLength(field, value, fieldName string, min, max int) *ValidationError {
	length := utf8.RuneCountInString(value)
	if length < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be at least %d characters long", fieldName, min),
		}
	}
	if length > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be no more than %d characters long", fieldName, max),
		}
	}
	return nil
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(field, value, fieldName string) *ValidationError {
	if value == "" {
		return nil // Use ValidateRequired separately if needed
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid email address", fieldName),
		}
	}
	return nil
}

// ValidateInt validates and parses an integer within a range
func (v *Validator) ValidateInt(field, value, fieldName string, min, max int) (int, *ValidationError) {
	if value == "" {
		return 0, &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s is required", fieldName),
		}
	}
	
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid number", fieldName),
		}
	}
	
	if intVal < min || intVal > max {
		return 0, &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be between %d and %d", fieldName, min, max),
		}
	}
	
	return intVal, nil
}

// ValidateFloat validates and parses a float within a range
func (v *Validator) ValidateFloat(field, value, fieldName string, min, max float64) (float64, *ValidationError) {
	if value == "" {
		return 0, &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s is required", fieldName),
		}
	}
	
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid number", fieldName),
		}
	}
	
	if floatVal < min || floatVal > max {
		return 0, &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be between %.1f and %.1f", fieldName, min, max),
		}
	}
	
	return floatVal, nil
}

// ValidateInList checks if value is in the allowed list
func (v *Validator) ValidateInList(field, value, fieldName string, allowedValues []string) *ValidationError {
	if value == "" {
		return nil // Use ValidateRequired separately if needed
	}
	
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	
	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(allowedValues, ", ")),
	}
}

// ValidatePattern validates against a regex pattern
func (v *Validator) ValidatePattern(field, value, fieldName, pattern string) *ValidationError {
	if value == "" {
		return nil // Use ValidateRequired separately if needed
	}
	
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s validation pattern error", fieldName),
		}
	}
	
	if !matched {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s format is invalid", fieldName),
		}
	}
	
	return nil
}

// Course validation functions

// ValidateCourseData validates course form data
func (v *Validator) ValidateCourseData(c echo.Context) (CourseFormData, ValidationErrors) {
	var errors ValidationErrors
	var data CourseFormData
	
	// Required fields
	name := c.FormValue("name")
	if err := v.ValidateRequired("name", name, "Course name"); err != nil {
		errors = append(errors, *err)
	} else if err := v.ValidateLength("name", name, "Course name", 3, 100); err != nil {
		errors = append(errors, *err)
	} else {
		data.Name = name
	}
	
	description := c.FormValue("description")
	if err := v.ValidateRequired("description", description, "Description"); err != nil {
		errors = append(errors, *err)
	} else if err := v.ValidateLength("description", description, "Description", 10, 500); err != nil {
		errors = append(errors, *err)
	} else {
		data.Description = description
	}
	
	address := c.FormValue("address")
	if err := v.ValidateRequired("address", address, "Address"); err != nil {
		errors = append(errors, *err)
	} else if err := v.ValidateLength("address", address, "Address", 10, 200); err != nil {
		errors = append(errors, *err)
	} else {
		data.Address = address
	}
	
	// Rating validation
	overallRating := c.FormValue("overallRating")
	allowedRatings := []string{"A+", "A", "A-", "B+", "B", "B-", "C+", "C", "C-", "D+", "D", "F"}
	if err := v.ValidateRequired("overallRating", overallRating, "Overall rating"); err != nil {
		errors = append(errors, *err)
	} else if err := v.ValidateInList("overallRating", overallRating, "Overall rating", allowedRatings); err != nil {
		errors = append(errors, *err)
	} else {
		data.OverallRating = overallRating
	}
	
	// Optional fields with validation
	price := c.FormValue("price")
	if price != "" {
		if err := v.ValidateInList("price", price, "Price", []string{"$", "$$", "$$$", "$$$$"}); err != nil {
			errors = append(errors, *err)
		} else {
			data.Price = price
		}
	}
	
	// Difficulty ratings (1-5)
	if handicapDiff, err := v.ValidateInt("handicapDifficulty", c.FormValue("handicapDifficulty"), "Handicap difficulty", 1, 5); err != nil {
		errors = append(errors, *err)
	} else {
		data.HandicapDifficulty = handicapDiff
	}
	
	if hazardDiff, err := v.ValidateInt("hazardDifficulty", c.FormValue("hazardDifficulty"), "Hazard difficulty", 1, 5); err != nil {
		errors = append(errors, *err)
	} else {
		data.HazardDifficulty = hazardDiff
	}
	
	// Optional text fields
	data.Condition = c.FormValue("condition")
	data.Merch = c.FormValue("merch")
	data.EnjoymentRating = c.FormValue("enjoymentRating")
	data.Vibe = c.FormValue("vibe")
	data.Range = c.FormValue("range")
	data.Amenities = c.FormValue("amenities")
	data.Glizzies = c.FormValue("glizzies")
	
	// Review text validation
	review := c.FormValue("review")
	if review != "" {
		if err := v.ValidateLength("review", review, "Review", 10, 2000); err != nil {
			errors = append(errors, *err)
		} else {
			data.Review = review
		}
	}
	
	return data, errors
}

// ValidateHandicap validates handicap input
func (v *Validator) ValidateHandicap(handicapStr string) (float64, *ValidationError) {
	return v.ValidateFloat("handicap", handicapStr, "Handicap", 0, 54)
}

// ValidateScore validates score input
func (v *Validator) ValidateScore(totalScoreStr, outScoreStr, inScoreStr, handicapStr string) (ScoreValidationResult, ValidationErrors) {
	var errors ValidationErrors
	var result ScoreValidationResult
	
	// Total score (required)
	if totalScore, err := v.ValidateInt("totalScore", totalScoreStr, "Total score", 18, 200); err != nil {
		errors = append(errors, *err)
	} else {
		result.TotalScore = totalScore
	}
	
	// Out and in scores (optional but must be valid if provided)
	if outScoreStr != "" {
		if outScore, err := v.ValidateInt("outScore", outScoreStr, "Out score", 9, 100); err != nil {
			errors = append(errors, *err)
		} else {
			result.OutScore = outScore
		}
	}
	
	if inScoreStr != "" {
		if inScore, err := v.ValidateInt("inScore", inScoreStr, "In score", 9, 100); err != nil {
			errors = append(errors, *err)
		} else {
			result.InScore = inScore
		}
	}
	
	// Handicap (optional)
	if handicapStr != "" {
		if handicap, err := v.ValidateFloat("handicap", handicapStr, "Handicap", -5, 40); err != nil {
			errors = append(errors, *err)
		} else {
			result.Handicap = handicap
		}
	}
	
	return result, errors
}

// CourseFormData represents validated course data
type CourseFormData struct {
	Name               string
	Description        string
	Address            string
	OverallRating      string
	Price              string
	HandicapDifficulty int
	HazardDifficulty   int
	Condition          string
	Merch              string
	EnjoymentRating    string
	Vibe               string
	Range              string
	Amenities          string
	Glizzies           string
	Review             string
}

// ScoreValidationResult represents validated score data
type ScoreValidationResult struct {
	TotalScore int
	OutScore   int
	InScore    int
	Handicap   float64
}

// ValidateDisplayName validates display name input
func (v *Validator) ValidateDisplayName(displayName string) *ValidationError {
	if displayName == "" {
		return nil // Empty is allowed to clear display name
	}
	
	// Check length
	if err := v.ValidateLength("display_name", displayName, "Display name", 2, 50); err != nil {
		return err
	}
	
	// Check for inappropriate content (basic check)
	inappropriate := []string{"admin", "system", "test", "null", "undefined"}
	lower := strings.ToLower(displayName)
	for _, word := range inappropriate {
		if strings.Contains(lower, word) {
			return &ValidationError{
				Field:   "display_name",
				Message: "Display name contains inappropriate content",
			}
		}
	}
	
	return nil
}

// Middleware functions

// ValidationMiddleware provides common validation middleware
func ValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add validator to context for use in handlers
			c.Set("validator", NewValidator())
			return next(c)
		}
	}
}

// RequestSizeMiddleware validates request size
func RequestSizeMiddleware(maxSize int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().ContentLength > maxSize {
				return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "Request too large")
			}
			return next(c)
		}
	}
}

// ContentTypeValidationMiddleware validates content type for specific routes
func ContentTypeValidationMiddleware(allowedTypes []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			contentType := c.Request().Header.Get("Content-Type")
			if contentType == "" {
				return next(c) // Allow empty content type for GET requests
			}
			
			// Extract main content type (ignore charset, etc.)
			mainType := strings.Split(contentType, ";")[0]
			mainType = strings.TrimSpace(mainType)
			
			for _, allowed := range allowedTypes {
				if mainType == allowed {
					return next(c)
				}
			}
			
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Unsupported content type")
		}
	}
}

// CSRFValidationMiddleware provides basic CSRF protection
func CSRFValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip CSRF for GET, HEAD, OPTIONS
			method := c.Request().Method
			if method == "GET" || method == "HEAD" || method == "OPTIONS" {
				return next(c)
			}
			
			// For POST, PUT, DELETE, check for CSRF token or valid referrer
			referer := c.Request().Header.Get("Referer")
			if referer == "" {
				return echo.NewHTTPError(http.StatusForbidden, "Missing referrer")
			}
			
			// Basic referrer validation - should be from same origin
			if !strings.Contains(referer, c.Request().Host) {
				return echo.NewHTTPError(http.StatusForbidden, "Invalid referrer")
			}
			
			return next(c)
		}
	}
}