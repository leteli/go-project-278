package handlers

import (
	db "code/db/generated"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"errors"

	"code/config"
	"net/url"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	Querier db.Querier
	Config  *config.Config
	Service *LinkService
}

func NewLinkHandler(q db.Querier, cfg *config.Config) *LinkHandler {
	return &LinkHandler{Querier: q, Config: cfg, Service: NewLinkService(q)}
}

func (h *LinkHandler) Register(router *gin.RouterGroup) {
	router.GET("", h.List)
	router.POST("", h.Create)
	router.GET("/:id", h.Get)
	router.PUT("/:id", h.Update)
	router.DELETE("/:id", h.Delete)
}

func (h *LinkHandler) List(c *gin.Context) {
	var params GetLinksDTO
	if err := c.ShouldBindQuery(&params); err != nil {
		badRequest(c, err)
		return
	}
	links, err := h.Querier.GetLinks(c.Request.Context(), db.GetLinksParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		internalServerError(c, err)
		return
	}
	linksResponse := make([]LinkResponse, len(links))
	for i, l := range links {
		linkRes, err := h.toLinkResponse(l)
		if err != nil {
			internalServerError(c, err)
			return
		}
		linksResponse[i] = linkRes
	}
	c.JSON(http.StatusOK, linksResponse)
}

func (h *LinkHandler) Create(c *gin.Context) {
	var body CreateLinkDTO

	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err)
		return
	}
	link, err := h.Service.Create(c.Request.Context(), body)
	if err != nil {
		fmt.Errorf("create err: %w", err)
		if isLinkConflict(err) {
			conflict(c, err)
			return
		}
		internalServerError(c, err)
		return
	}
	linksRes, err := h.toLinkResponse(link)
	if err != nil {
		fmt.Errorf("create toLinkResponse err: %w", err)
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusCreated, linksRes)
}

func (h *LinkHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		badRequest(c, fmt.Errorf("get link: %w", err))
		return
	}
	link, err := h.Querier.GetLink(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(c, errors.New("get link: not found"))
			return
		}
		internalServerError(c, err)
		return
	}
	linkResponse, err := h.toLinkResponse(link)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, linkResponse)
}

func (h *LinkHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		badRequest(c, fmt.Errorf("update link: %w", err))
		return
	}
	var body UpdateLinkDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err)
		return
	}
	if body.OriginalUrl == nil && body.ShortName == nil {
		badRequest(c, errors.New("at least one field must be updated"))
		return
	}
	link, err := h.Service.Update(c.Request.Context(), id, body)
	if err != nil {
		if isLinkConflict(err) {
			conflict(c, err)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			notFound(c, errors.New("update link: not found"))
			return
		}
		internalServerError(c, err)
		return
	}
	linkResponse, err := h.toLinkResponse(link)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, linkResponse)
}

func (h *LinkHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		badRequest(c, fmt.Errorf("delete link: %w", err))
		return
	}
	rowsCount, err := h.Querier.DeleteLink(c.Request.Context(), id)
	if err != nil {
		internalServerError(c, err)
		return
	}
	if rowsCount == 0 {
		notFound(c, errors.New("delete link: not found"))
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrorInvalidID
	}
	return id, nil
}

func (h *LinkHandler) toLinkResponse(link db.Link) (LinkResponse, error) {
	if h.Config.BaseURL == "" {
		return LinkResponse{}, errors.New("base url not set")
	}
	shortUrl, err := url.JoinPath(h.Config.BaseURL, link.ShortName)
	if err != nil {
		return LinkResponse{}, errors.New("failed to build short url")
	}
	return LinkResponse{
		ID:          link.ID,
		OriginalUrl: link.OriginalUrl,
		ShortName:   link.ShortName,
		ShortUrl:    shortUrl,
	}, nil
}
