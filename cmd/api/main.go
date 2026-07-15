package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xiangqi-lab/internal/analysis"
	"xiangqi-lab/internal/config"
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
	searchEngine := builtin.New()
	defer searchEngine.Close()

	events := game.NewEventBus()
	matchService := game.NewService(
		game.NewMemoryRepository(), searchEngine, events, cfg.EngineMoveTime,
	)
	recordService := records.NewService()
	learningService := learning.NewService(recordService)
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
		logger.Info("api listening", "address", server.Addr, "dataMode", "memory")
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
