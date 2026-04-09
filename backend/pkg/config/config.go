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
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
