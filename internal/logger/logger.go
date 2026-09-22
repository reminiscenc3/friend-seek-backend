package logger

import (
	"log/slog"
	"os"

	"github.com/olegtemek/friend-seek-backend/internal/config"
)

func New(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level(cfg.Env)}

	var handler slog.Handler

	handler = slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler).With(slog.String("env", cfg.Env))
}

func level(env string) slog.Level {
	if env == config.EnvProduction {
		return slog.LevelError
	}

	return slog.LevelDebug
}
