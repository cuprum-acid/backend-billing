// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Tracing  TracingConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	DSN      string // Full connection string (computed)
}

// TracingConfig holds OpenTelemetry tracing configuration
type TracingConfig struct {
	Endpoint string
	Service  string
}

// Load reads configuration from environment and config files
func Load() (*Config, error) {
	// Set defaults
	setDefaults()

	// Read from environment variables
	viper.AutomaticEnv()

	// Try to read from config file (optional)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		// Ignore error if file doesn't exist - config will be loaded from env vars
		_ = err
	}

	config := &Config{
		Server: ServerConfig{
			Port:         viper.GetString("server.port"),
			ReadTimeout:  viper.GetDuration("server.readTimeout"),
			WriteTimeout: viper.GetDuration("server.writeTimeout"),
			IdleTimeout:  viper.GetDuration("server.idleTimeout"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		Tracing: TracingConfig{
			Endpoint: viper.GetString("tracing.endpoint"),
			Service:  viper.GetString("tracing.service"),
		},
	}

	// Build DSN from components
	config.Database.DSN = buildDSN(config.Database)

	return config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.readTimeout", 15*time.Second)
	viper.SetDefault("server.writeTimeout", 15*time.Second)
	viper.SetDefault("server.idleTimeout", 60*time.Second)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "billing")
	viper.SetDefault("database.sslmode", "disable")

	// Tracing defaults
	viper.SetDefault("tracing.endpoint", "jaeger:4318")
	viper.SetDefault("tracing.service", "billing-api")
}

// buildDSN constructs PostgreSQL connection string
func buildDSN(dbConfig DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.Port,
		dbConfig.SSLMode,
	)
}

// GetDSN returns the database connection string
// Supports both DATABASE_URL env var and component-based config
func (c *Config) GetDSN() string {
	// Check if DATABASE_URL is set directly (for docker-compose compatibility)
	if dbURL := viper.GetString("database_url"); dbURL != "" {
		return dbURL
	}
	return c.Database.DSN
}
