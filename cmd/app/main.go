package main

import (
	"code/internal/config"
	"code/internal/handlers"
	"log"
	"time"

	"fmt"

	"github.com/getsentry/sentry-go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.AppEnv,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	defer sentry.Flush(2 * time.Second)

	router := handlers.SetupRouter()
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
