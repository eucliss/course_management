package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LoadConfigFromFile loads configuration from environment-specific files
func LoadConfigFromFile(environment string) (*Config, error) {
	// Load environment file if it exists
	envFile := fmt.Sprintf("config/%s.env", environment)
	if err := loadEnvFile(envFile); err != nil {
		log.Printf("Warning: Could not load environment file %s: %v", envFile, err)
	}

	// Also try to load from .env file as fallback
	if err := loadEnvFile(".env"); err == nil {
		log.Printf("Loaded additional configuration from .env file")
	}

	// Load configuration from environment variables
	return LoadConfig()
}

// LoadConfigForTesting loads configuration specifically for testing
func LoadConfigForTesting() (*Config, error) {
	// Set testing environment
	os.Setenv("ENV", "testing")
	
	// Load testing environment file
	if err := loadEnvFile("config/testing.env"); err != nil {
		log.Printf("Warning: Could not load testing.env: %v", err)
	}
	
	return LoadConfig()
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(filename string) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("Warning: Invalid line %d in %s: %s", lineNumber, filename, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Expand environment variables in the value
		value = os.ExpandEnv(value)

		// Only set if not already set (environment variables take precedence)
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				log.Printf("Warning: Could not set environment variable %s: %v", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	log.Printf("Loaded environment configuration from %s", filename)
	return nil
}

// GetConfigPath returns the path to configuration files
func GetConfigPath() string {
	// Check if we're in a subdirectory and need to go up
	if _, err := os.Stat("config"); err == nil {
		return "config"
	}
	
	// Try relative paths
	paths := []string{
		"./config",
		"../config",
		"../../config",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}
	
	return "config"
}

// SaveConfigTemplate creates a template configuration file
func SaveConfigTemplate(filename string) error {
	template := `# Course Management System Configuration
# Copy this file and rename to match your environment (development.env, production.env, etc.)

# Environment (development, testing, staging, production)
ENV=development

# Server Configuration
PORT=8080
HOST=localhost
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_SHUTDOWN_TIMEOUT=10s
MAX_REQUEST_SIZE=33554432

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=course_management
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=5m

# Session Configuration
SESSION_SECRET=your-very-secure-session-secret-key-32-characters-minimum
CSRF_SECRET=your-csrf-secret-key
JWT_SECRET=your-jwt-secret-key
SESSION_TIMEOUT=24h
SECURE_COOKIES=false

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your_google_client_id_here
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback

# Mapbox Configuration
MAPBOX_ACCESS_TOKEN=your_mapbox_token_here

# Security Configuration
RATE_LIMIT_PER_MIN=60
BCRYPT_COST=12

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=3
LOG_MAX_AGE=28
LOG_COMPRESS=true

# Path Configuration
VIEWS_DIR=views
STATIC_DIR=static
UPLOADS_DIR=uploads
TEMPLATES_DIR=templates
`

	return os.WriteFile(filename, []byte(template), 0644)
}

// ValidateEnvironment validates that all required files exist for an environment
func ValidateEnvironment(environment string) error {
	var errors []string

	// Check if environment file exists
	envFile := fmt.Sprintf("config/%s.env", environment)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("environment file %s does not exist", envFile))
	}

	// Check for required directories in non-testing environments
	if environment != "testing" {
		requiredDirs := []string{"views", "static"}
		for _, dir := range requiredDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("required directory %s does not exist", dir))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("environment validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetAvailableEnvironments returns a list of available environment configurations
func GetAvailableEnvironments() []string {
	var environments []string
	
	configDir := GetConfigPath()
	if entries, err := os.ReadDir(configDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".env") {
				env := strings.TrimSuffix(entry.Name(), ".env")
				environments = append(environments, env)
			}
		}
	}
	
	return environments
}