package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	origEnv := map[string]string{}
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			origEnv[parts[0]] = parts[1]
		}
	}
	
	// Cleanup function
	defer func() {
		// Clear environment
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Unsetenv(parts[0])
			}
		}
		// Restore original environment
		for key, value := range origEnv {
			os.Setenv(key, value)
		}
	}()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		// Clear all environment variables
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Unsetenv(parts[0])
			}
		}

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig should not fail with defaults: %v", err)
		}

		// Test default values
		if config.Environment != "development" {
			t.Errorf("Expected environment 'development', got '%s'", config.Environment)
		}
		if config.Server.Port != "8080" {
			t.Errorf("Expected port '8080', got '%s'", config.Server.Port)
		}
		if config.Database.Host != "localhost" {
			t.Errorf("Expected DB host 'localhost', got '%s'", config.Database.Host)
		}
	})

	t.Run("EnvironmentVariableOverrides", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("ENV", "testing")
		os.Setenv("PORT", "9000")
		os.Setenv("DB_HOST", "testhost")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("SESSION_SECRET", "test-session-secret-32-characters-minimum")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if config.Environment != "testing" {
			t.Errorf("Expected environment 'testing', got '%s'", config.Environment)
		}
		if config.Server.Port != "9000" {
			t.Errorf("Expected port '9000', got '%s'", config.Server.Port)
		}
		if config.Database.Host != "testhost" {
			t.Errorf("Expected DB host 'testhost', got '%s'", config.Database.Host)
		}
		if config.Database.Name != "testdb" {
			t.Errorf("Expected DB name 'testdb', got '%s'", config.Database.Name)
		}
	})

	t.Run("DurationParsing", func(t *testing.T) {
		os.Setenv("SERVER_READ_TIMEOUT", "60s")
		os.Setenv("SESSION_TIMEOUT", "2h")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if config.Server.ReadTimeout != 60*time.Second {
			t.Errorf("Expected read timeout 60s, got %v", config.Server.ReadTimeout)
		}
		if config.Security.SessionTimeout != 2*time.Hour {
			t.Errorf("Expected session timeout 2h, got %v", config.Security.SessionTimeout)
		}
	})

	t.Run("BooleanParsing", func(t *testing.T) {
		os.Setenv("SECURE_COOKIES", "true")
		os.Setenv("LOG_COMPRESS", "false")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if !config.Security.SecureCookies {
			t.Errorf("Expected secure cookies true, got %v", config.Security.SecureCookies)
		}
		if config.Logging.Compress {
			t.Errorf("Expected log compress false, got %v", config.Logging.Compress)
		}
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("ValidDevelopmentConfig", func(t *testing.T) {
		config := &Config{
			Environment: "development",
			Server: ServerConfig{
				Port: "8080",
			},
			Database: DatabaseConfig{
				Host: "localhost",
				Name: "testdb",
			},
			Security: SecurityConfig{
				SessionSecret: "development-secret-32-characters-minimum",
			},
		}

		if err := config.Validate(); err != nil {
			t.Errorf("Valid development config should not fail validation: %v", err)
		}
	})

	t.Run("InvalidEnvironment", func(t *testing.T) {
		config := &Config{
			Environment: "invalid",
			Server: ServerConfig{
				Port: "8080",
			},
			Database: DatabaseConfig{
				Host: "localhost",
				Name: "testdb",
			},
		}

		if err := config.Validate(); err == nil {
			t.Error("Invalid environment should fail validation")
		}
	})

	t.Run("ProductionValidation", func(t *testing.T) {
		config := &Config{
			Environment: "production",
			Server: ServerConfig{
				Port: "8080",
			},
			Database: DatabaseConfig{
				Host: "localhost",
				Name: "testdb",
				Password: "", // Missing password
			},
			Security: SecurityConfig{
				SessionSecret: "short", // Too short
			},
		}

		if err := config.Validate(); err == nil {
			t.Error("Production config with missing password and short secret should fail validation")
		}
	})

	t.Run("EmptyRequiredFields", func(t *testing.T) {
		config := &Config{
			Environment: "development",
			Server: ServerConfig{
				Port: "", // Empty port
			},
			Database: DatabaseConfig{
				Host: "", // Empty host
				Name: "testdb",
			},
		}

		if err := config.Validate(); err == nil {
			t.Error("Config with empty required fields should fail validation")
		}
	})
}

func TestConfigHelperMethods(t *testing.T) {
	t.Run("EnvironmentChecks", func(t *testing.T) {
		devConfig := &Config{Environment: "development"}
		prodConfig := &Config{Environment: "production"}
		testConfig := &Config{Environment: "testing"}

		if !devConfig.IsDevelopment() {
			t.Error("Development config should return true for IsDevelopment()")
		}
		if devConfig.IsProduction() {
			t.Error("Development config should return false for IsProduction()")
		}

		if !prodConfig.IsProduction() {
			t.Error("Production config should return true for IsProduction()")
		}
		if prodConfig.IsDevelopment() {
			t.Error("Production config should return false for IsDevelopment()")
		}

		if !testConfig.IsTesting() {
			t.Error("Testing config should return true for IsTesting()")
		}
	})

	t.Run("GetDatabaseDSN", func(t *testing.T) {
		config := &Config{
			Database: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
		}

		expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
		if dsn := config.GetDatabaseDSN(); dsn != expected {
			t.Errorf("Expected DSN '%s', got '%s'", expected, dsn)
		}
	})

	t.Run("GetServerAddress", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: "8080",
			},
		}

		expected := "localhost:8080"
		if addr := config.GetServerAddress(); addr != expected {
			t.Errorf("Expected address '%s', got '%s'", expected, addr)
		}
	})
}

func TestGetEnvHelpers(t *testing.T) {
	t.Run("getIntOrDefault", func(t *testing.T) {
		// Test with valid integer
		os.Setenv("TEST_INT", "42")
		if result := getIntOrDefault("TEST_INT", 10); result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}

		// Test with invalid integer
		os.Setenv("TEST_INT_INVALID", "not_a_number")
		if result := getIntOrDefault("TEST_INT_INVALID", 10); result != 10 {
			t.Errorf("Expected default 10, got %d", result)
		}

		// Test with missing environment variable
		os.Unsetenv("TEST_INT_MISSING")
		if result := getIntOrDefault("TEST_INT_MISSING", 10); result != 10 {
			t.Errorf("Expected default 10, got %d", result)
		}

		// Cleanup
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_INT_INVALID")
	})

	t.Run("getBoolOrDefault", func(t *testing.T) {
		// Test with valid boolean
		os.Setenv("TEST_BOOL", "true")
		if result := getBoolOrDefault("TEST_BOOL", false); !result {
			t.Errorf("Expected true, got %v", result)
		}

		// Test with invalid boolean
		os.Setenv("TEST_BOOL_INVALID", "not_a_bool")
		if result := getBoolOrDefault("TEST_BOOL_INVALID", false); result {
			t.Errorf("Expected default false, got %v", result)
		}

		// Cleanup
		os.Unsetenv("TEST_BOOL")
		os.Unsetenv("TEST_BOOL_INVALID")
	})

	t.Run("getDurationOrDefault", func(t *testing.T) {
		// Test with valid duration
		os.Setenv("TEST_DURATION", "30s")
		expected := 30 * time.Second
		if result := getDurationOrDefault("TEST_DURATION", 10*time.Second); result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}

		// Test with invalid duration
		os.Setenv("TEST_DURATION_INVALID", "not_a_duration")
		expected = 10 * time.Second
		if result := getDurationOrDefault("TEST_DURATION_INVALID", expected); result != expected {
			t.Errorf("Expected default %v, got %v", expected, result)
		}

		// Cleanup
		os.Unsetenv("TEST_DURATION")
		os.Unsetenv("TEST_DURATION_INVALID")
	})
}

func TestContainsHelper(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}
	
	if !contains(slice, "banana") {
		t.Error("Should find 'banana' in slice")
	}
	
	if contains(slice, "grape") {
		t.Error("Should not find 'grape' in slice")
	}
	
	if contains([]string{}, "anything") {
		t.Error("Should not find anything in empty slice")
	}
}