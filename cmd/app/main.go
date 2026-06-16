package main

import (
	"code/config"
	"code/db"
	"code/handlers"
	"log"
	"time"

	"context"
	"fmt"

	"database/sql"

	"github.com/getsentry/sentry-go"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	if cfg.DatabaseURL == "" {
		log.Fatal("db url not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := db.MigrateUp(conn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	router := handlers.SetupRouter(conn, cfg)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
