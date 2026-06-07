package handlers

import (
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default() // NB: already uses Logger and Recover middlewares
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	h := &Handlers{}
	h.Register(router)
	return router
}

type Handlers struct{}

func (h *Handlers) Register(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}
