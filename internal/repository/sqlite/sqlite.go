package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string, maxConns int) (*Repository, error) {
	const op = "sqlite.New"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: open: %w", op, err)
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)

	if err = db.PingContext(ctx); err != nil {
		db.Close()

		return nil, fmt.Errorf("%s: ping: %w", op, err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Migrate(ctx context.Context, schema string) error {
	const op = "sqlite.Migrate"

	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) Close() {
	if err := r.db.Close(); err != nil {
		slog.Error("failed to close the database", slog.String("error", err.Error()))
	}
}
