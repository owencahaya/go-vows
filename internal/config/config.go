package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	AppPort    string
	AppBaseURL string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Meta WhatsApp
	MetaAccessToken   string
	MetaPhoneNumberID string
	MetaVerifyToken   string
	MetaAPIVersion    string
}

// Load reads configuration from environment variables, applying sensible
// defaults for local development.
func Load() *Config {
	return &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "vows"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "password"),

		MetaAccessToken:   getEnv("META_ACCESS_TOKEN", ""),
		MetaPhoneNumberID: getEnv("META_PHONE_NUMBER_ID", ""),
		MetaVerifyToken:   getEnv("META_VERIFY_TOKEN", ""),
		MetaAPIVersion:    getEnv("META_API_VERSION", "v20.0"),
	}
}

// DSN builds the MySQL data source name for GORM.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
