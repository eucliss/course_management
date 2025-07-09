package main

import (
	"testing"
)

func TestValidator_ValidateRequired(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"Valid value", "test", false},
		{"Empty string", "", true},
		{"Whitespace only", "   ", true},
		{"Valid with spaces", "  test  ", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRequired("test_field", tt.value, "Test Field")
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
		})
	}
}

func TestValidator_ValidateLength(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name        string
		value       string
		min         int
		max         int
		shouldError bool
	}{
		{"Valid length", "test", 2, 10, false},
		{"Too short", "a", 2, 10, true},
		{"Too long", "this is too long", 2, 10, true},
		{"Exact min", "ab", 2, 10, false},
		{"Exact max", "1234567890", 2, 10, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLength("test_field", tt.value, "Test Field", tt.min, tt.max)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
		})
	}
}

func TestValidator_ValidateInt(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name        string
		value       string
		min         int
		max         int
		expected    int
		shouldError bool
	}{
		{"Valid integer", "5", 1, 10, 5, false},
		{"Empty string", "", 1, 10, 0, true},
		{"Invalid integer", "abc", 1, 10, 0, true},
		{"Below minimum", "0", 1, 10, 0, true},
		{"Above maximum", "15", 1, 10, 0, true},
		{"Exact minimum", "1", 1, 10, 1, false},
		{"Exact maximum", "10", 1, 10, 10, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateInt("test_field", tt.value, "Test Field", tt.min, tt.max)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
			if !tt.shouldError && result != tt.expected {
				t.Errorf("Expected result %d for %s, but got %d", tt.expected, tt.name, result)
			}
		})
	}
}

func TestValidator_ValidateFloat(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name        string
		value       string
		min         float64
		max         float64
		expected    float64
		shouldError bool
	}{
		{"Valid float", "5.5", 1.0, 10.0, 5.5, false},
		{"Empty string", "", 1.0, 10.0, 0, true},
		{"Invalid float", "abc", 1.0, 10.0, 0, true},
		{"Below minimum", "0.5", 1.0, 10.0, 0, true},
		{"Above maximum", "15.0", 1.0, 10.0, 0, true},
		{"Exact minimum", "1.0", 1.0, 10.0, 1.0, false},
		{"Exact maximum", "10.0", 1.0, 10.0, 10.0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateFloat("test_field", tt.value, "Test Field", tt.min, tt.max)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
			if !tt.shouldError && result != tt.expected {
				t.Errorf("Expected result %f for %s, but got %f", tt.expected, tt.name, result)
			}
		})
	}
}

func TestValidator_ValidateInList(t *testing.T) {
	validator := NewValidator()
	allowedValues := []string{"A+", "A", "A-", "B+", "B", "B-"}
	
	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"Valid value", "A+", false},
		{"Another valid value", "B-", false},
		{"Invalid value", "C+", true},
		{"Empty string", "", false}, // Empty is allowed when not using ValidateRequired
		{"Case sensitive", "a+", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateInList("test_field", tt.value, "Test Field", allowedValues)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
		})
	}
}

func TestValidator_ValidateHandicap(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name        string
		value       string
		expected    float64
		shouldError bool
	}{
		{"Valid handicap", "15.2", 15.2, false},
		{"Zero handicap", "0", 0, false},
		{"Maximum handicap", "54", 54, false},
		{"Above maximum", "55", 0, true},
		{"Negative handicap", "-1", 0, true},
		{"Invalid format", "abc", 0, true},
		{"Empty string", "", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateHandicap(tt.value)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
			if !tt.shouldError && result != tt.expected {
				t.Errorf("Expected result %f for %s, but got %f", tt.expected, tt.name, result)
			}
		})
	}
}

func TestValidator_ValidateDisplayName(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"Valid name", "John Doe", false},
		{"Empty name", "", false}, // Empty is allowed to clear display name
		{"Too short", "J", true},
		{"Too long", "This name is way too long for a display name and exceeds the maximum allowed length", true},
		{"Contains admin", "AdminUser", true},
		{"Contains system", "SystemUser", true},
		{"Valid with numbers", "User123", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDisplayName(tt.value)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
		})
	}
}