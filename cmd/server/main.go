package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"gin-mania-backend/internal/search"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/gins", func(c *gin.Context) {
		query := c.Query("q")
		results := search.Search(query)
		c.JSON(http.StatusOK, gin.H{
			"query":   query,
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
