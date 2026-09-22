package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

func (r *Repository) CreateUser(ctx context.Context, login string) (models.User, error) {
	const op = "sqlite.CreateUser"

	const query = `
		INSERT INTO users (login)
		VALUES (?)
		ON CONFLICT (login) DO NOTHING
		RETURNING login, note, created_at`

	var user models.User

	err := r.db.QueryRowContext(ctx, query, login).Scan(&user.Login, &user.Note, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, models.ErrLoginTaken)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *Repository) UserByLogin(ctx context.Context, login string) (models.User, error) {
	const op = "sqlite.UserByLogin"

	const query = `SELECT login, note, created_at FROM users WHERE login = ?`

	var user models.User

	err := r.db.QueryRowContext(ctx, query, login).Scan(&user.Login, &user.Note, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *Repository) Users(ctx context.Context, exceptLogin string) ([]models.User, error) {
	const op = "sqlite.Users"

	const query = `SELECT login, note, created_at FROM users WHERE login != ? ORDER BY login`

	rows, err := r.db.QueryContext(ctx, query, exceptLogin)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var user models.User

		if err = rows.Scan(&user.Login, &user.Note, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (r *Repository) UpdateNote(ctx context.Context, login, note string) error {
	const op = "sqlite.UpdateNote"

	const query = `UPDATE users SET note = ? WHERE login = ?`

	result, err := r.db.ExecContext(ctx, query, note, login)
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
