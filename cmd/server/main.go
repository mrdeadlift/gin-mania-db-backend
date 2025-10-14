package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"gin-mania-backend/internal/config"
	httpRouter "gin-mania-backend/internal/http/router"
	"gin-mania-backend/internal/search"
	"gin-mania-backend/pkg/database"
	"gin-mania-backend/pkg/logging"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	db, err := database.OpenPostgres(ctx, database.Config{
		DSN:             cfg.Database.DSN,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("retrieve sql DB: %w", err)
	}
	defer sqlDB.Close()

	searchService := search.NewService(search.NewRepository(db))

	engine, err := httpRouter.New(cfg, logger, httpRouter.Dependencies{
		SearchService: searchService,
	})
	if err != nil {
		return fmt.Errorf("initialize router: %w", err)
	}

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	logger.Info("starting Gin Mania server",
		zap.String("address", server.Addr),
		zap.String("environment", cfg.App.Environment),
	)

	if err := startServer(ctx, server, logger, cfg.Server.ShutdownTimeout); err != nil {
		return err
	}

	return nil
}

func startServer(ctx context.Context, server *http.Server, logger *zap.Logger, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	case sig := <-sigCh:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("server shutdown timed out", zap.Duration("timeout", shutdownTimeout))
		}
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}
