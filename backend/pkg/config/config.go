package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port    string
	Env     string
	DBURL   string
	DBMURL  string // migration URL (may differ from pool DSN)
	RedisURL string

	JWTSecret     string
	JWTExpiry     time.Duration
	RefreshExpiry time.Duration

	StripeKey       string
	StripeWebhookSecret string

	StorageType string // local, s3
	AWSBucket   string
	AWSRegion   string
	UploadDir   string

	OpenAIKey string

	TelegramToken  string
	TelegramChatID string

	FrontendURL string
	SMTPHost    string
	SMTPPort    string

	PgwebURL          string
	RedisCommanderURL string
	MinioURL          string
	GrafanaURL        string
	PrometheusURL     string

	// Rate limiting
	RateLimitAPI      int // requests per minute for general API (default 60)
	RateLimitAuth     int // requests per minute for auth endpoints (default 10)
	RateLimitRegister int // requests per hour for register (default 5)
	RateLimitForgot   int // requests per hour for forgot password (default 3)

	// Email verification
	EmailVerificationRequired bool // require email verification on register (default false)

	// Security
	MaxRequestBodyMB int // max request body in MB (default 10)
	BcryptCost       int // bcrypt cost for password hashing (default 12)
}

func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "8080"),
		Env:     getEnv("ENV", "development"),
		DBURL:   getEnv("DATABASE_URL", "postgres://blueprint:blueprint@localhost:5432/blueprint?sslmode=disable"),
		DBMURL:  getEnv("DATABASE_MIGRATION_URL", getEnv("DATABASE_URL", "postgres://blueprint:blueprint@localhost:5432/blueprint?sslmode=disable")),
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:     getDuration("JWT_EXPIRY", 15*time.Minute),
		RefreshExpiry: getDuration("REFRESH_EXPIRY", 7*24*time.Hour),

		StripeKey:           getEnv("STRIPE_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),

		StorageType: getEnv("STORAGE_TYPE", "local"),
		AWSBucket:   getEnv("AWS_BUCKET", ""),
		AWSRegion:   getEnv("AWS_REGION", "us-east-1"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),

		OpenAIKey: getEnv("OPENAI_KEY", ""),

		TelegramToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID: getEnv("TELEGRAM_CHAT_ID", ""),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		SMTPHost:    getEnv("SMTP_HOST", "localhost"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),

		PgwebURL:          getEnv("PGWEB_URL", ""),
		RedisCommanderURL: getEnv("REDIS_COMMANDER_URL", ""),
		MinioURL:          getEnv("MINIO_URL", ""),
		GrafanaURL:        getEnv("GRAFANA_URL", ""),
		PrometheusURL:     getEnv("PROMETHEUS_URL", ""),

		RateLimitAPI:              getEnvInt("RATE_LIMIT_API", 60),
		RateLimitAuth:             getEnvInt("RATE_LIMIT_AUTH", 10),
		RateLimitRegister:         getEnvInt("RATE_LIMIT_REGISTER", 5),
		RateLimitForgot:           getEnvInt("RATE_LIMIT_FORGOT", 3),
		EmailVerificationRequired: getEnv("EMAIL_VERIFICATION_REQUIRED", "false") == "true",
		MaxRequestBodyMB:          getEnvInt("MAX_REQUEST_BODY_MB", 10),
		BcryptCost:                getEnvInt("BCRYPT_COST", 12),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if mins, err := strconv.Atoi(value); err == nil {
			return time.Duration(mins) * time.Minute
		}
	}
	return defaultValue
}
