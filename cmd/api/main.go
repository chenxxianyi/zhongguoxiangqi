package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xiangqi-lab/internal/analysis"
	"xiangqi-lab/internal/config"
	"xiangqi-lab/internal/database"
	"xiangqi-lab/internal/engine/builtin"
	"xiangqi-lab/internal/game"
	"xiangqi-lab/internal/learning"
	"xiangqi-lab/internal/observability"
	"xiangqi-lab/internal/records"
	"xiangqi-lab/internal/transport/httpapi"
)

func main() {
	logger := observability.NewLogger()
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	// ── 数据库连接 ──
	var db *sql.DB
	if cfg.DataMode == "mysql" {
		logger.Info("connecting to MySQL",
			"host", cfg.DBHost,
			"port", cfg.DBPort,
			"database", cfg.DBName,
		)
		db, err = database.AutoCreateDatabase(cfg)
		if err != nil {
			logger.Warn("MySQL connection failed, falling back to in-memory", "error", err)
			cfg.DataMode = "memory"
		} else {
			if err := database.Migrate(db, logger); err != nil {
				logger.Error("database migration failed", "error", err)
				os.Exit(1)
			}
			defer db.Close()
			logger.Info("MySQL connected and migrated successfully",
				"host", cfg.DBHost,
				"port", cfg.DBPort,
				"database", cfg.DBName,
			)
		}
	}

	searchEngine := builtin.New()
	defer searchEngine.Close()

	recordService := records.NewService()
	learningService := learning.NewService(recordService)
	events := game.NewEventBus()
	matchService := game.NewService(
		game.NewMemoryRepository(), searchEngine, events, cfg.EngineMoveTime,
	)
	matchService.SetBookAdvisor(learningService)
	analysisService := analysis.NewService(matchService, searchEngine)
	api := httpapi.NewServer(
		cfg, logger, matchService, events, recordService,
		learningService, analysisService, searchEngine,
	)

	server := &http.Server{
		Addr: cfg.HTTPAddr, Handler: api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
		// WriteTimeout remains zero because match streams are long-lived.
	}
	if err := run(server, cfg.ShutdownTimeout, logger); err != nil {
		logger.Error("api terminated", "error", err)
		os.Exit(1)
	}
}

func run(server *http.Server, shutdownTimeout time.Duration, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errs := make(chan error, 1)
	go func() {
		logger.Info("api listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("api stopped")
	return nil
}
