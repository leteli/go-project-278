package main

import (
	"code/internal/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default() // NB: already uses Logger and Recover middlewares
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
