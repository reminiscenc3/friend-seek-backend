package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/olegtemek/friend-seek-backend/internal/config"
	"github.com/olegtemek/friend-seek-backend/internal/logger"
	"github.com/olegtemek/friend-seek-backend/internal/repository/sqlite"
)

//go:embed schema.sql
var schema string

func main() {

	cfg, err := config.New()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	slog.SetDefault(logger.New(cfg))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := sqlite.New(ctx, cfg.Database.GetSQLiteDSN(), cfg.Database.MaxConns)
	if err != nil {
		slog.Error("failed to open the database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Migrate(ctx, schema); err != nil {
		slog.Error("failed to apply schema", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("schema applied")
}
