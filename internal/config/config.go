package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gin-mania-backend/pkg/logging"
)

const (
	defaultDatabaseURL = "postgresql://gin_admin:gin_admin_password@localhost:5432/gin_mania?sslmode=disable"
	defaultRedisURL    = "redis://localhost:6379/0"
)

// Config aggregates application settings sourced from environment variables.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Logging  logging.Config
}

// AppConfig captures process-wide flags that influence behavior.
type AppConfig struct {
	Environment string
}

// ServerConfig contains HTTP server parameters and middleware options.
type ServerConfig struct {
	Address         string
	GinMode         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

// DatabaseConfig defines PostgreSQL connection configuration.
type DatabaseConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig describes the Redis endpoint used for cache and rate limiting.
type RedisConfig struct {
	URL string
}

// AuthConfig captures Auth0 integration toggles and metadata.
type AuthConfig struct {
	Enabled  bool
	Domain   string
	Audience string
}

// Load constructs a Config instance by reading environment variables and applying defaults.
func Load() (*Config, error) {
	appEnv := valueOrDefault("APP_ENV", "development")

	server, err := loadServerConfig(appEnv)
	if err != nil {
		return nil, err
	}

	database, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	redisCfg, err := loadRedisConfig()
	if err != nil {
		return nil, err
	}

	authCfg, err := loadAuthConfig()
	if err != nil {
		return nil, err
	}

	loggingCfg, err := loadLoggingConfig(appEnv)
	if err != nil {
		return nil, err
	}

	return &Config{
		App: AppConfig{
			Environment: appEnv,
		},
		Server:   server,
		Database: database,
		Redis:    redisCfg,
		Auth:     authCfg,
		Logging:  loggingCfg,
	}, nil
}

func loadServerConfig(appEnv string) (ServerConfig, error) {
	var cfg ServerConfig

	address := strings.TrimSpace(os.Getenv("SERVER_ADDRESS"))
	if address == "" {
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			port = "8080"
		}
		if strings.Contains(port, ":") {
			address = port
		} else {
			address = ":" + port
		}
	}

	ginMode := strings.TrimSpace(os.Getenv("GIN_MODE"))
	if ginMode == "" {
		if strings.EqualFold(appEnv, "production") {
			ginMode = gin.ReleaseMode
		} else {
			ginMode = gin.DebugMode
		}
	}
	if !isValidGinMode(ginMode) {
		return ServerConfig{}, fmt.Errorf("invalid GIN_MODE: %s", ginMode)
	}

	readTimeout, err := parseDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	writeTimeout, err := parseDuration("SERVER_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	shutdownTimeout, err := parseDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}

	allowedOrigins := parseCSVEnv("CORS_ALLOWED_ORIGINS", []string{"*"})

	cfg.Address = address
	cfg.GinMode = ginMode
	cfg.ReadTimeout = readTimeout
	cfg.WriteTimeout = writeTimeout
	cfg.ShutdownTimeout = shutdownTimeout
	cfg.AllowedOrigins = allowedOrigins

	return cfg, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	dsn := strings.TrimSpace(valueOrDefault("DATABASE_URL", defaultDatabaseURL))
	if dsn == "" {
		return DatabaseConfig{}, errors.New("DATABASE_URL must not be empty")
	}
	if _, err := url.Parse(dsn); err != nil {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL parse error: %w", err)
	}

	maxIdle, err := parseInt("DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return DatabaseConfig{}, err
	}
	if maxIdle < 0 {
		return DatabaseConfig{}, errors.New("DB_MAX_IDLE_CONNS must be non-negative")
	}

	maxOpen, err := parseInt("DB_MAX_OPEN_CONNS", 50)
	if err != nil {
		return DatabaseConfig{}, err
	}
	if maxOpen < 0 {
		return DatabaseConfig{}, errors.New("DB_MAX_OPEN_CONNS must be non-negative")
	}

	connMaxLifetime, err := parseDuration("DB_CONN_MAX_LIFETIME", time.Hour)
	if err != nil {
		return DatabaseConfig{}, err
	}

	connMaxIdleTime, err := parseDuration("DB_CONN_MAX_IDLE_TIME", 30*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		DSN:             dsn,
		MaxIdleConns:    maxIdle,
		MaxOpenConns:    maxOpen,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}, nil
}

func loadRedisConfig() (RedisConfig, error) {
	redisURL := strings.TrimSpace(valueOrDefault("REDIS_URL", defaultRedisURL))
	if redisURL == "" {
		return RedisConfig{}, errors.New("REDIS_URL must not be empty")
	}
	parsed, err := url.Parse(redisURL)
	if err != nil {
		return RedisConfig{}, fmt.Errorf("REDIS_URL parse error: %w", err)
	}
	if parsed.Scheme != "redis" && parsed.Scheme != "rediss" {
		return RedisConfig{}, fmt.Errorf("REDIS_URL must use redis or rediss scheme, got %q", parsed.Scheme)
	}

	return RedisConfig{URL: redisURL}, nil
}

func loadAuthConfig() (AuthConfig, error) {
	enabled := parseBool("AUTH0_ENABLED", false)
	domain := strings.TrimSpace(os.Getenv("AUTH0_DOMAIN"))
	audience := strings.TrimSpace(os.Getenv("AUTH0_AUDIENCE"))

	if !enabled {
		enabled = domain != "" && audience != ""
	}

	if enabled {
		if domain == "" {
			return AuthConfig{}, errors.New("AUTH0_DOMAIN is required when Auth0 is enabled")
		}
		if audience == "" {
			return AuthConfig{}, errors.New("AUTH0_AUDIENCE is required when Auth0 is enabled")
		}
		if strings.Contains(domain, "://") {
			return AuthConfig{}, errors.New("AUTH0_DOMAIN should not include a scheme (https://)")
		}
	}

	return AuthConfig{
		Enabled:  enabled,
		Domain:   domain,
		Audience: audience,
	}, nil
}

func loadLoggingConfig(appEnv string) (logging.Config, error) {
	level := strings.TrimSpace(valueOrDefault("LOG_LEVEL", "info"))
	encoding := strings.TrimSpace(valueOrDefault("LOG_ENCODING", "json"))

	outputPaths := parseCSVEnv("LOG_OUTPUT_PATHS", []string{"stdout"})
	errorPaths := parseCSVEnv("LOG_ERROR_OUTPUT_PATHS", []string{"stderr"})

	development := parseBool("LOG_DEVELOPMENT", !strings.EqualFold(appEnv, "production"))

	loggerCfg := logging.Config{
		Level:            level,
		Encoding:         encoding,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorPaths,
		Development:      development,
	}

	if err := loggerCfg.Validate(); err != nil {
		return logging.Config{}, err
	}

	return loggerCfg, nil
}

func valueOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func parseInt(key string, fallback int) (int, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as integer: %w", key, err)
	}
	return parsed, nil
}

func parseDuration(key string, fallback time.Duration) (time.Duration, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as duration: %w", key, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("%s must not be negative", key)
	}
	return d, nil
}

func parseBool(key string, fallback bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseCSVEnv(key string, fallback []string) []string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}

	parts := strings.Split(val, ",")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			results = append(results, trimmed)
		}
	}
	if len(results) == 0 {
		return fallback
	}
	return results
}

func isValidGinMode(mode string) bool {
	switch mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		return true
	default:
		return false
	}
}
