// Command server starts the kubevirt-management HTTP API and embedded SPA.
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

	"github.com/Exonical/kubevirt-management/internal/config"
	"github.com/Exonical/kubevirt-management/internal/server"
	"github.com/Exonical/kubevirt-management/internal/version"
)

func main() {
	logger := newLogger()
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "err", err)
		os.Exit(1)
	}

	logger.Info("starting kubevirt-management",
		"version", version.Version,
		"commit", version.Commit,
		"go_version", version.GoVersion(),
		"addr", cfg.HTTPAddr,
		"auth_mode", cfg.AuthMode,
		"impersonation_enabled", cfg.ImpersonationEnabled,
	)

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("build server", "err", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		logger.Error("http server error", "err", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown", "err", err)
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug", "DEBUG":
		level = slog.LevelDebug
	case "warn", "WARN":
		level = slog.LevelWarn
	case "error", "ERROR":
		level = slog.LevelError
	}
	format := os.Getenv("LOG_FORMAT")
	opts := &slog.HandlerOptions{Level: level}
	if format == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
