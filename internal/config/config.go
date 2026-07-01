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

	// Twilio WhatsApp
	TwilioAccountSID     string
	TwilioAuthToken      string
	TwilioWhatsAppNumber string // sender, e.g. whatsapp:+14155238886

	// Optional approved Content Template SIDs (business-initiated messages).
	// When empty, the corresponding message is sent as freeform Body text.
	TwilioContentSidInvitation string
	TwilioContentSidReminder   string
	TwilioContentSidQR         string

	// Optional public URL Twilio calls back with delivery/status updates.
	TwilioStatusCallbackURL string
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

		TwilioAccountSID:     getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:      getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioWhatsAppNumber: getEnv("TWILIO_WHATSAPP_NUMBER", ""),

		TwilioContentSidInvitation: getEnv("TWILIO_CONTENT_SID_INVITATION", ""),
		TwilioContentSidReminder:   getEnv("TWILIO_CONTENT_SID_REMINDER", ""),
		TwilioContentSidQR:         getEnv("TWILIO_CONTENT_SID_QR", ""),

		TwilioStatusCallbackURL: getEnv("TWILIO_STATUS_CALLBACK_URL", ""),
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
