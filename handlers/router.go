package handlers

import (
	"code/config"
	db "code/db/generated"
	"database/sql"
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func SetupRouter(conn *sql.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default() // NB: already uses Logger and Recover middlewares
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	RegisterCommonRoutes(router)

	queries := db.New(conn)

	linkHandler := NewLinkHandler(queries, cfg)
	linksRouter := router.Group("api/links")
	linkHandler.Register(linksRouter)

	return router
}

func RegisterCommonRoutes(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}
