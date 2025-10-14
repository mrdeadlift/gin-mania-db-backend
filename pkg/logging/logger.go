package logging

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config describes the configurable settings for the structured logger.
type Config struct {
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
	Development      bool
	DisableCaller    bool
}

// Validate ensures the logging configuration is internally consistent before initialization.
func (c Config) Validate() error {
	if _, err := zapcore.ParseLevel(strings.ToLower(c.Level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", c.Level, err)
	}
	switch strings.ToLower(c.Encoding) {
	case "json", "console":
	default:
		return fmt.Errorf("invalid log encoding %q: expected json or console", c.Encoding)
	}
	if len(c.OutputPaths) == 0 {
		return fmt.Errorf("at least one log output path is required")
	}
	if len(c.ErrorOutputPaths) == 0 {
		return fmt.Errorf("at least one log error output path is required")
	}
	return nil
}

// NewLogger constructs a zap.Logger using the provided configuration.
func NewLogger(cfg Config) (*zap.Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var zapCfg zap.Config
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	if err := zapCfg.Level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	zapCfg.Encoding = strings.ToLower(cfg.Encoding)
	zapCfg.OutputPaths = cfg.OutputPaths
	zapCfg.ErrorOutputPaths = cfg.ErrorOutputPaths

	// Ensure consistent ISO8601 timestamps and short caller paths.
	zapCfg.EncoderConfig.TimeKey = "time"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if cfg.DisableCaller {
		zapCfg.EncoderConfig.CallerKey = ""
	}

	return zapCfg.Build()
}
