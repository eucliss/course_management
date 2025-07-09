package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Google      GoogleConfig   `mapstructure:"google"`
	Security    SecurityConfig `mapstructure:"security"`
	Mapbox      MapboxConfig   `mapstructure:"mapbox"`
	Logging     LoggingConfig  `mapstructure:"logging"`
	Paths       PathsConfig    `mapstructure:"paths"`
	Cache       CacheConfig    `mapstructure:"cache"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	MaxRequestSize  int64         `mapstructure:"max_request_size"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// GoogleConfig contains Google OAuth configuration
type GoogleConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	SessionSecret    string        `mapstructure:"session_secret"`
	CSRFSecret       string        `mapstructure:"csrf_secret"`
	JWTSecret        string        `mapstructure:"jwt_secret"`
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`
	RateLimitPerMin  int           `mapstructure:"rate_limit_per_min"`
	BcryptCost       int           `mapstructure:"bcrypt_cost"`
	SecureCookies    bool          `mapstructure:"secure_cookies"`
	TrustedProxies   []string      `mapstructure:"trusted_proxies"`
}

// MapboxConfig contains Mapbox configuration
type MapboxConfig struct {
	AccessToken string `mapstructure:"access_token"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// PathsConfig contains file system path configuration
type PathsConfig struct {
	ViewsDir    string `mapstructure:"views_dir"`
	StaticDir   string `mapstructure:"static_dir"`
	UploadsDir  string `mapstructure:"uploads_dir"`
	TemplatesDir string `mapstructure:"templates_dir"`
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	RedisURL     string        `mapstructure:"redis_url"`
	EnableRedis  bool          `mapstructure:"enable_redis"`
	EnableMemory bool          `mapstructure:"enable_memory"`
	DefaultTTL   time.Duration `mapstructure:"default_ttl"`
	MaxMemoryMB  int           `mapstructure:"max_memory_mb"`
}

// LoadConfig loads configuration from environment variables with validation
func LoadConfig() (*Config, error) {
	config := &Config{
		Environment: getEnvOrDefault("ENV", "development"),
		Server: ServerConfig{
			Port:            getEnvOrDefault("PORT", "8080"),
			Host:            getEnvOrDefault("HOST", "localhost"),
			ReadTimeout:     getDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationOrDefault("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			MaxRequestSize:  getInt64OrDefault("MAX_REQUEST_SIZE", 32<<20), // 32MB
		},
		Database: DatabaseConfig{
			Host:            getEnvOrDefault("DB_HOST", "localhost"),
			Port:            getEnvOrDefault("DB_PORT", "5432"),
			User:            getEnvOrDefault("DB_USER", "postgres"),
			Password:        getEnvOrDefault("DB_PASSWORD", ""),
			Name:            getEnvOrDefault("DB_NAME", "course_management"),
			SSLMode:         getEnvOrDefault("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntOrDefault("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntOrDefault("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Google: GoogleConfig{
			ClientID:     getEnvOrDefault("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnvOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		},
		Security: SecurityConfig{
			SessionSecret:    getEnvOrDefault("SESSION_SECRET", ""),
			CSRFSecret:       getEnvOrDefault("CSRF_SECRET", ""),
			JWTSecret:        getEnvOrDefault("JWT_SECRET", ""),
			SessionTimeout:   getDurationOrDefault("SESSION_TIMEOUT", 24*time.Hour),
			RateLimitPerMin:  getIntOrDefault("RATE_LIMIT_PER_MIN", 60),
			BcryptCost:       getIntOrDefault("BCRYPT_COST", 12),
			SecureCookies:    getBoolOrDefault("SECURE_COOKIES", false),
			TrustedProxies:   getStringSliceOrDefault("TRUSTED_PROXIES", []string{}),
		},
		Mapbox: MapboxConfig{
			AccessToken: getEnvOrDefault("MAPBOX_ACCESS_TOKEN", ""),
		},
		Logging: LoggingConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			Output:     getEnvOrDefault("LOG_OUTPUT", "stdout"),
			MaxSize:    getIntOrDefault("LOG_MAX_SIZE", 100),
			MaxBackups: getIntOrDefault("LOG_MAX_BACKUPS", 3),
			MaxAge:     getIntOrDefault("LOG_MAX_AGE", 28),
			Compress:   getBoolOrDefault("LOG_COMPRESS", true),
		},
		Paths: PathsConfig{
			ViewsDir:     getEnvOrDefault("VIEWS_DIR", "views"),
			StaticDir:    getEnvOrDefault("STATIC_DIR", "static"),
			UploadsDir:   getEnvOrDefault("UPLOADS_DIR", "uploads"),
			TemplatesDir: getEnvOrDefault("TEMPLATES_DIR", "templates"),
		},
		Cache: CacheConfig{
			RedisURL:     getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
			EnableRedis:  getBoolOrDefault("CACHE_ENABLE_REDIS", true),
			EnableMemory: getBoolOrDefault("CACHE_ENABLE_MEMORY", true),
			DefaultTTL:   getDurationOrDefault("CACHE_DEFAULT_TTL", 30*time.Minute),
			MaxMemoryMB:  getIntOrDefault("CACHE_MAX_MEMORY_MB", 100),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	// Validate environment
	validEnvs := []string{"development", "testing", "staging", "production"}
	if !contains(validEnvs, c.Environment) {
		errors = append(errors, fmt.Sprintf("invalid environment '%s', must be one of: %s", c.Environment, strings.Join(validEnvs, ", ")))
	}

	// Validate required fields for production
	if c.Environment == "production" {
		if c.Security.SessionSecret == "" {
			errors = append(errors, "SESSION_SECRET is required in production")
		}
		if len(c.Security.SessionSecret) < 32 {
			errors = append(errors, "SESSION_SECRET must be at least 32 characters in production")
		}
		if c.Database.Password == "" {
			errors = append(errors, "DB_PASSWORD is required in production")
		}
		if c.Google.ClientID == "" {
			errors = append(errors, "GOOGLE_CLIENT_ID is required in production")
		}
		if c.Google.ClientSecret == "" {
			errors = append(errors, "GOOGLE_CLIENT_SECRET is required in production")
		}
	}

	// Validate server configuration
	if c.Server.Port == "" {
		errors = append(errors, "server port cannot be empty")
	}

	// Validate database configuration
	if c.Database.Host == "" {
		errors = append(errors, "database host cannot be empty")
	}
	if c.Database.Name == "" {
		errors = append(errors, "database name cannot be empty")
	}

	// Validate paths exist (except in testing)
	if c.Environment != "testing" {
		pathChecks := map[string]string{
			"views directory":     c.Paths.ViewsDir,
			"static directory":    c.Paths.StaticDir,
			"templates directory": c.Paths.TemplatesDir,
		}

		for name, path := range pathChecks {
			if path != "" {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					log.Printf("Warning: %s '%s' does not exist", name, path)
				}
			}
		}

		// Create uploads directory if it doesn't exist
		if c.Paths.UploadsDir != "" {
			if err := os.MkdirAll(c.Paths.UploadsDir, 0755); err != nil {
				errors = append(errors, fmt.Sprintf("failed to create uploads directory '%s': %v", c.Paths.UploadsDir, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTesting returns true if the environment is testing
func (c *Config) IsTesting() bool {
	return c.Environment == "testing"
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// Helper functions for loading environment variables with defaults

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: invalid integer value for %s: %s, using default %d", key, value, defaultValue)
	}
	return defaultValue
}

func getInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: invalid int64 value for %s: %s, using default %d", key, value, defaultValue)
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Warning: invalid boolean value for %s: %s, using default %t", key, value, defaultValue)
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: invalid duration value for %s: %s, using default %v", key, value, defaultValue)
	}
	return defaultValue
}

func getStringSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}