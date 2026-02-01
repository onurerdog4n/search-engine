package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig `validate:"required"`
	Redis    RedisConfig    `validate:"required"`
	Server   ServerConfig   `validate:"required"`
	Sync     SyncConfig     `validate:"required"`
	Cache    CacheConfig    `validate:"required"`
	Logger   LoggerConfig   `validate:"required"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string `validate:"required,url"`
	MaxOpenConns    int    `validate:"min=1,max=100"`
	MaxIdleConns    int    `validate:"min=1,max=50"`
	ConnMaxLifetime int    `validate:"min=60"` // seconds
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL string `validate:"required"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port               string `validate:"required"`
	RateLimitPerMinute int    `validate:"min=1,max=1000"`
	ReadTimeout        int    `validate:"min=1"` // seconds
	WriteTimeout       int    `validate:"min=1"` // seconds
}

// SyncConfig holds sync configuration
type SyncConfig struct {
	IntervalSeconds int `validate:"min=60"` // minimum 1 minute
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	TTLSeconds int `validate:"min=1,max=3600"` // 1 second to 1 hour
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `validate:"required,oneof=debug info warn error"`
	Encoding   string `validate:"required,oneof=json console"`
	OutputPath string `validate:"required"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/search_engine?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "localhost:6379"),
		},
		Server: ServerConfig{
			Port:               getEnv("PORT", "8080"),
			RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
			ReadTimeout:        getEnvAsInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout:       getEnvAsInt("SERVER_WRITE_TIMEOUT", 15),
		},
		Sync: SyncConfig{
			IntervalSeconds: getEnvAsInt("SYNC_INTERVAL", 3600),
		},
		Cache: CacheConfig{
			TTLSeconds: getEnvAsInt("CACHE_TTL_SECONDS", 60),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Encoding:   getEnv("LOG_ENCODING", "json"),
			OutputPath: getEnv("LOG_OUTPUT", "stdout"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// getEnv gets an environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
