package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gin-mania-backend/internal/search"
	"gin-mania-backend/pkg/database"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	db, err := database.OpenPostgres(context.Background(), database.Config{
		DSN:             resolveDatabaseURL(),
		MaxIdleConns:    10,
		MaxOpenConns:    50,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		log.Fatalf("failed to initialize database connection: %v", err)
	}

	searchService := search.NewService(search.NewRepository(db))

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/gins", func(c *gin.Context) {
		filter, err := parseSearchFilter(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results, err := searchService.Search(c.Request.Context(), filter)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, search.ErrInvalidPagination) {
				status = http.StatusBadRequest
			}

			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   filter.Query,
			"limit":   filter.Limit,
			"offset":  filter.Offset,
			"results": results,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Gin Mania server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
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

func resolveDatabaseURL() string {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}

	return "postgresql://gin_admin:gin_admin_password@localhost:5432/gin_mania?sslmode=disable"
}
