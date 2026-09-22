package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/olegtemek/friend-seek-backend/internal/config"
	"github.com/olegtemek/friend-seek-backend/internal/logger"
	"github.com/olegtemek/friend-seek-backend/internal/repository/memory"
	"github.com/olegtemek/friend-seek-backend/internal/repository/sqlite"
	"github.com/olegtemek/friend-seek-backend/internal/transport/rest"
	"github.com/olegtemek/friend-seek-backend/internal/usecase"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	slog.SetDefault(logger.New(cfg))

	slog.Info("starting friend-seek")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	locations := memory.New()

	repo, err := sqlite.New(ctx, cfg.Database.GetSQLiteDSN(), cfg.Database.MaxConns)
	if err != nil {
		slog.Error("failed to open the database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer repo.Close()

	uc := usecase.New(repo, locations, cfg.Game.LocationTTL)

	srv := rest.New(cfg.HTTPServer, uc)

	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.Run() }()

	select {
	case err = <-serverErr:
		if err != nil {
			slog.Error("http server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("stopped")
}
