package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config captures optional tuning knobs when establishing a PostgreSQL connection.
type Config struct {
	DSN             string
	Logger          logger.Interface
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// OpenPostgres initializes a GORM PostgreSQL connection using the supplied context and configuration.
func OpenPostgres(ctx context.Context, cfg Config) (*gorm.DB, error) {
	if strings.TrimSpace(cfg.DSN) == "" {
		return nil, errors.New("postgres DSN is required")
	}

	gormCfg := &gorm.Config{}
	if cfg.Logger != nil {
		gormCfg.Logger = cfg.Logger
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	pingCtx := ctx
	if pingCtx == nil {
		pingCtx = context.Background()
	}

	deadlineCtx, cancel := context.WithTimeout(pingCtx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(deadlineCtx); err != nil {
		return nil, err
	}

	return db, nil
}
