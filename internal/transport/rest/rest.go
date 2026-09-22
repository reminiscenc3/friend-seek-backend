package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/olegtemek/friend-seek-backend/internal/config"
)

type Rest struct {
	uc      UseCase
	timeout time.Duration
	srv     *http.Server
}

func New(cfg config.HTTPServer, uc UseCase) *Rest {
	s := &Rest{uc: uc, timeout: cfg.DefaultTimeout}

	s.srv = &http.Server{
		Addr:         cfg.Address,
		Handler:      s.router(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return s
}

func (s *Rest) Run() error {
	slog.Info("http server started", slog.String("address", s.srv.Addr))

	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("rest.Run: %w", err)
	}

	return nil
}

func (s *Rest) Shutdown(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("rest.Shutdown: %w", err)
	}

	return nil
}
