package handlers

import (
	"errors"
	"net/http"

	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrorInvalidID         = errors.New("invalid id")
	ErrorOriginalURLExists = errors.New("original url already exists")
	ErrorShortNameExists   = errors.New("short name already in use")
)

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Bad Request",
		"message": err.Error(),
	})
}

func notFound(c *gin.Context, err error) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "Not Found",
		"message": err.Error(),
	})
}

func conflict(c *gin.Context, err error) {
	c.JSON(http.StatusConflict, gin.H{
		"error":   "Conflict",
		"message": err.Error(),
	})
}

func internalServerError(c *gin.Context, err error) {
	log.Printf("internal server error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "Internal Server Error",
		"message": "Something went wrong",
	})
}

func handleDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			switch pgErr.ConstraintName {
			case "links_original_url_unique":
				return ErrorOriginalURLExists
			case "links_short_name_unique":
				return ErrorShortNameExists
			default:
				return fmt.Errorf("link conflict: %w", err)
			}
		default:
			return fmt.Errorf("db error: %w", err)
		}
	}
	return err
}

func isLinkConflict(err error) bool {
	return errors.Is(err, ErrorOriginalURLExists) || errors.Is(err, ErrorShortNameExists)
}
