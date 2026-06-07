package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string
	HTTPPort  string
	SentryDSN string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	config := &Config{
		AppEnv:    getEnv("APP_ENV", "local"),
		HTTPPort:  getEnv("PORT", "8080"),
		SentryDSN: getEnv("SENTRY_DSN", ""),
	}
	return config, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
