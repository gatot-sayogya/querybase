package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	Dialect  string `mapstructure:"dialect"` // postgresql or mysql
}

// RedisConfig represents the Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig represents the JWT configuration
type JWTConfig struct {
	Secret      string        `mapstructure:"secret"`
	ExpireHours time.Duration `mapstructure:"expire_hours"`
	Issuer      string        `mapstructure:"issuer"`
}

// CORSConfig represents the CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a comma-separated list of origins allowed to make requests
	// Example: "http://localhost:3000,https://app.example.com"
	AllowedOrigins string `mapstructure:"allowed_origins"`

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool `mapstructure:"allow_credentials"`

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int `mapstructure:"max_age"`
}

// Load loads the configuration from file and environment variables
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30*time.Second)
	viper.SetDefault("server.write_timeout", 30*time.Second)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.expire_hours", 24*time.Hour)
	viper.SetDefault("jwt.issuer", "querybase")
	viper.SetDefault("cors.allowed_origins", "http://localhost:3000,http://localhost:3001")
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400) // 24 hours

	// Allow environment variables to override config
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetDatabaseDSN returns the database connection string
func (c *DatabaseConfig) GetDatabaseDSN() string {
	// Default to PostgreSQL if dialect not specified
	dialect := c.Dialect
	if dialect == "" {
		dialect = "postgresql"
	}

	switch dialect {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.Name)
	case "postgresql", "":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
	default:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
	}
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetAllowedOrigins parses the comma-separated origins into a slice
func (c *CORSConfig) GetAllowedOrigins() []string {
	if c.AllowedOrigins == "" {
		return []string{"http://localhost:3000"}
	}

	// Parse comma-separated values
	origins := []string{}
	for _, origin := range splitAndTrim(c.AllowedOrigins, ",") {
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	// Fallback to defaults if no valid origins
	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}

	return origins
}

// splitAndTrim splits a string by separator and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString splits a string by separator (simple implementation)
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	if sep == "" {
		return []string{s}
	}

	result := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
