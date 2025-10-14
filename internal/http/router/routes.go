package router

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gin-mania-backend/internal/search"
)

func registerRoutes(engine *gin.Engine, deps Dependencies) {
	engine.GET("/healthz", healthHandler)
	engine.GET("/gins", ginsHandler(deps.SearchService))
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func ginsHandler(service *search.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter, err := parseSearchFilter(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results, err := service.Search(c.Request.Context(), filter)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, search.ErrInvalidPagination) {
				status = http.StatusBadRequest
			}

			c.Error(err)
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   filter.Query,
			"limit":   filter.Limit,
			"offset":  filter.Offset,
			"results": results,
		})
	}
}

func parseSearchFilter(c *gin.Context) (search.SearchFilter, error) {
	filter := search.SearchFilter{
		Query: c.Query("q"),
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return search.SearchFilter{}, errors.New("limit must be a non-negative integer")
		}
		filter.Limit = limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return search.SearchFilter{}, errors.New("offset must be a non-negative integer")
		}
		filter.Offset = offset
	}

	return filter, nil
}
