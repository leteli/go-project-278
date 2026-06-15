//go:build integration

package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"code/config"
	migrate "code/db"
	db "code/db/generated"

	"os"

	"encoding/json"

	_ "github.com/jackc/pgx/v5/stdlib"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()
	pg, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("secret"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Printf("start pg: %v", err)
		return 1
	}
	defer func() { _ = pg.Terminate(ctx) }()

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("postgres dsn: %v", err)
		return 1
	}

	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("open db: %v", err)
		return 1
	}
	defer testDB.Close()

	ctxPing, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := testDB.PingContext(ctxPing); err != nil {
		cancel()
		log.Printf("ping db: %v", err)
		return 1
	}
	cancel()
	if err := migrate.MigrateUp(testDB); err != nil {
		log.Printf("migrate up: %v", err)
		return 1
	}
	return m.Run()
}

func TestPingRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterCommonRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestCreateLink(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodPost, "/api/links", `{
			"original_url": "https://example.com",
			"short_name": "example"
		}`)

		require.Equal(t, http.StatusCreated, w.Code)
		res := decodeLinkResponse(t, w)
		assert.Positive(t, res.ID)
		assert.Equal(t, "https://example.com", res.OriginalUrl)
		assert.Equal(t, "example", res.ShortName)
		assert.Equal(t, "http://localhost:8080/example", res.ShortUrl)
	})
	t.Run("success - no short name", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodPost, "/api/links", `{
			"original_url": "https://example.com"
		}`)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		res := decodeLinkResponse(t, w)
		assert.Positive(t, res.ID)
		assert.Equal(t, "https://example.com", res.OriginalUrl)
		assert.NotEmpty(t, res.ShortName)
		assert.Equal(t, "http://localhost:8080/"+res.ShortName, res.ShortUrl)
	})

	t.Run("invalid body", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodPost, "/api/links", `{
			"short_name": "example"
		}`)

		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})

	t.Run("duplicate short name", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		createLink(t, router, "https://example.com/one", "example")

		second := performRequest(router, http.MethodPost, "/api/links", `{
			"original_url": "https://example.com/two",
			"short_name": "example"
		}`)

		assert.Equal(t, http.StatusConflict, second.Code, second.Body.String())
	})

	t.Run("duplicate original url", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		createLink(t, router, "https://example.com/one", "example")

		second := performRequest(router, http.MethodPost, "/api/links", `{
			"original_url": "https://example.com/one",
			"short_name": "new"
		}`)

		assert.Equal(t, http.StatusConflict, second.Code, second.Body.String())
	})
}

func TestGetLinks(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		first := createLink(t, router, "https://example.com/one", "one")
		second := createLink(t, router, "https://example.com/two", "two")

		w := performRequest(router, http.MethodGet, "/api/links", "")

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		res := decodeLinksResponse(t, w)
		require.Len(t, res, 2)
		assert.Equal(t, first, res[0])
		assert.Equal(t, second, res[1])
	})

	t.Run("with limit", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		first := createLink(t, router, "https://example.com/one", "one")
		createLink(t, router, "https://example.com/two", "two")

		w := performRequest(router, http.MethodGet, "/api/links?limit=1&offset=0", "")

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		res := decodeLinksResponse(t, w)
		require.Len(t, res, 1)
		assert.Equal(t, first, res[0])
	})
}

func TestGetLinkByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		created := createLink(t, router, "https://example.com", "example")

		w := performRequest(router, http.MethodGet, fmt.Sprintf("/api/links/%d", created.ID), "")

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		res := decodeLinkResponse(t, w)
		assert.Equal(t, created, res)
	})

	t.Run("not found", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodGet, "/api/links/999999", "")

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("invalid id", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodGet, "/api/links/abc", "")

		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})
}

func TestUpdateLink(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		created := createLink(t, router, "https://example.com/old", "old")

		w := performRequest(router, http.MethodPut, fmt.Sprintf("/api/links/%d", created.ID), `{
			"original_url": "https://example.com/new",
			"short_name": "new"
		}`)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		res := decodeLinkResponse(t, w)
		assert.Equal(t, created.ID, res.ID)
		assert.Equal(t, "https://example.com/new", res.OriginalUrl)
		assert.Equal(t, "new", res.ShortName)
		assert.Equal(t, "http://localhost:8080/new", res.ShortUrl)
	})

	t.Run("empty body", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		created := createLink(t, router, "https://example.com", "example")

		w := performRequest(router, http.MethodPut, fmt.Sprintf("/api/links/%d", created.ID), `{}`)

		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})

	t.Run("duplicate short name", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		createLink(t, router, "https://example.com/one", "one")
		second := createLink(t, router, "https://example.com/two", "two")

		w := performRequest(router, http.MethodPut, fmt.Sprintf("/api/links/%d", second.ID), `{
			"short_name": "one"
		}`)

		assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	})

	t.Run("duplicate original url", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		createLink(t, router, "https://example.com/one", "one")
		second := createLink(t, router, "https://example.com/two", "two")

		w := performRequest(router, http.MethodPut, fmt.Sprintf("/api/links/%d", second.ID), `{
			"original_url": "https://example.com/one"
		}`)

		assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	})

	t.Run("not found", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodPut, "/api/links/999999", `{
			"short_name": "missing"
		}`)

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("invalid id", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodPut, "/api/links/abc", `{
			"original_url": "https://example.com/one"
		}`)

		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})
}

func TestDeleteLink(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		created := createLink(t, router, "https://example.com", "example")

		w := performRequest(router, http.MethodDelete, fmt.Sprintf("/api/links/%d", created.ID), "")

		assert.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

		get := performRequest(router, http.MethodGet, fmt.Sprintf("/api/links/%d", created.ID), "")
		assert.Equal(t, http.StatusNotFound, get.Code, get.Body.String())
	})

	t.Run("not found", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodDelete, "/api/links/999999", "")

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("invalid id", func(t *testing.T) {
		router := setupTestRouterWithTx(t)

		w := performRequest(router, http.MethodDelete, "/api/links/abc", "")

		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})
}

func setupTestRouter(tx db.DBTX, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterCommonRoutes(router)

	q := db.New(tx)

	h := NewLinkHandler(q, cfg)
	linksRouter := router.Group("api/links")
	h.Register(linksRouter)
	return router
}

func setupTestRouterWithTx(t *testing.T) *gin.Engine {
	t.Helper()
	tx, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
	})
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	return setupTestRouter(tx, cfg)
}

func performRequest(router http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()

	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	router.ServeHTTP(w, req)
	return w
}

func createLink(t *testing.T, router http.Handler, originalURL, shortName string) LinkResponse {
	t.Helper()

	body := fmt.Sprintf(`{
		"original_url": %q,
		"short_name": %q
	}`, originalURL, shortName)

	w := performRequest(router, http.MethodPost, "/api/links", body)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	return decodeLinkResponse(t, w)
}

func decodeLinksResponse(t *testing.T, w *httptest.ResponseRecorder) []LinkResponse {
	t.Helper()

	var res []LinkResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res), w.Body.String())
	return res
}

func decodeLinkResponse(t *testing.T, w *httptest.ResponseRecorder) LinkResponse {
	t.Helper()

	var res LinkResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res), w.Body.String())
	return res
}
