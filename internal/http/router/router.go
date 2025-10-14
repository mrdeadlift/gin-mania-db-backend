package router

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"gin-mania-backend/internal/config"
	"gin-mania-backend/internal/search"
)

const (
	ContextKeyRequestID = "request_id"
	requestIDHeader     = "X-Request-ID"
)

// Dependencies aggregates external services required by the router.
type Dependencies struct {
	SearchService *search.Service
}

var (
	// ErrMissingConfig indicates configuration was not provided when constructing the router.
	ErrMissingConfig = errors.New("router config is required")
	// ErrMissingLogger indicates a logger dependency was not supplied.
	ErrMissingLogger = errors.New("router logger is required")
	// ErrMissingSearchService indicates the search service dependency was missing.
	ErrMissingSearchService = errors.New("search service is required")
)

// New constructs a gin.Engine with shared middleware and registered routes.
func New(cfg *config.Config, logger *zap.Logger, deps Dependencies) (*gin.Engine, error) {
	if cfg == nil {
		return nil, ErrMissingConfig
	}
	if logger == nil {
		return nil, ErrMissingLogger
	}
	if deps.SearchService == nil {
		return nil, ErrMissingSearchService
	}

	gin.SetMode(cfg.Server.GinMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestIDMiddleware())
	engine.Use(loggingMiddleware(logger))
	engine.Use(corsMiddleware(cfg.Server.AllowedOrigins))

	registerRoutes(engine, deps)

	return engine, nil
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(ContextKeyRequestID, requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)

		c.Next()
	}
}

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(ContextKeyRequestID)
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("raw_path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
		}
		if requestIDStr, ok := requestID.(string); ok && requestIDStr != "" {
			fields = append(fields, zap.String("request_id", requestIDStr))
		}
		if query := c.Request.URL.RawQuery; query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if len(c.Errors) > 0 {
			logger.Error("request failed", append(fields, zap.String("error", c.Errors.String()))...)
			return
		}

		logger.Info("request completed", fields...)
	}
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := resolveAllowedOrigin(origin, allowedOrigins)

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			if allowOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")

		requestHeaders := c.GetHeader("Access-Control-Request-Headers")
		if requestHeaders == "" {
			requestHeaders = "Authorization,Content-Type"
		}
		c.Header("Access-Control-Allow-Headers", requestHeaders)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func resolveAllowedOrigin(origin string, allowed []string) string {
	if len(allowed) == 0 {
		return ""
	}

	if origin == "" {
		if containsWildcard(allowed) {
			return "*"
		}
		return ""
	}

	if containsWildcard(allowed) {
		return origin
	}

	for _, candidate := range allowed {
		if strings.EqualFold(candidate, origin) {
			return origin
		}
	}

	return ""
}

func containsWildcard(list []string) bool {
	for _, item := range list {
		if item == "*" {
			return true
		}
	}
	return false
}
