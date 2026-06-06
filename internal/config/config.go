package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	HTTPPort string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	config := &Config{
		AppEnv:   getEnv("APP_ENV", "local"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
	}
	return config, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
