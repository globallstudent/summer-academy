package config

import "os"

// Config holds application-wide configuration
type Config struct {
	Environment string
	Database    DatabaseConfig
	Redis       RedisConfig
	Auth        AuthConfig
	WBFY        WBFYConfig
	Telegram    TelegramConfig
}

// DatabaseConfig holds database connection information
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis connection information
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret    string
	CookieName   string
	CookieMaxAge int
}

// WBFYConfig holds configuration for WBFY terminal integration
type WBFYConfig struct {
	BinaryPath string
	BaseURL    string
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken   string
	WebhookURL string
}

// New creates a new Config instance populated from environment variables
func New() *Config {
	return &Config{
		Environment: getEnv("ENV", "development"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "academy"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Auth: AuthConfig{
			JWTSecret:    getEnv("JWT_SECRET", "supersecret"),
			CookieName:   getEnv("COOKIE_NAME", "academy_session"),
			CookieMaxAge: 86400, // 24 hours
		},
		WBFY: WBFYConfig{
			BinaryPath: getEnv("WBFY_PATH", "../wbfy/wbfy"),
			BaseURL:    getEnv("WBFY_URL", "http://localhost:8081"),
		},
		Telegram: TelegramConfig{
			BotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
			WebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),
		},
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
