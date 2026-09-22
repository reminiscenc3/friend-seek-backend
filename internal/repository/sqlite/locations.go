package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

const timestampLayout = "2006-01-02T15:04:05.000Z"

func (r *Repository) SaveLocation(ctx context.Context, login string, loc models.Location) error {
	const op = "sqlite.SaveLocation"

	const query = `
		INSERT INTO locations (login, latitude, longitude, azimuth, accuracy, updated_at)
		SELECT ?, ?, ?, ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM users WHERE login = ?)`

	result, err := r.db.ExecContext(ctx, query,
		login, loc.Latitude, loc.Longitude, loc.Azimuth, loc.Accuracy,
		loc.UpdatedAt.UTC().Format(timestampLayout), login,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if affected == 0 {
		return fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
	}

	return nil
}

func (r *Repository) LastLocation(ctx context.Context, login string) (models.Location, error) {
	const op = "sqlite.LastLocation"

	const query = `
		SELECT latitude, longitude, azimuth, accuracy, updated_at
		FROM locations
		WHERE login = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`

	var loc models.Location

	err := r.db.QueryRowContext(ctx, query, login).
		Scan(&loc.Latitude, &loc.Longitude, &loc.Azimuth, &loc.Accuracy, &loc.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Location{}, fmt.Errorf("%s: %w", op, models.ErrLocationNotFound)
		}

		return models.Location{}, fmt.Errorf("%s: %w", op, err)
	}

	loc.UpdatedAt = loc.UpdatedAt.UTC()

	return loc, nil
}
