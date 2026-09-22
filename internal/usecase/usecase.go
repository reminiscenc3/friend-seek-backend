package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

type DatabaseRepository interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, login string) (models.User, error)
	UserByLogin(ctx context.Context, login string) (models.User, error)
	Users(ctx context.Context, exceptLogin string) ([]models.User, error)
	UpdateNote(ctx context.Context, login, note string) error

	SaveLocation(ctx context.Context, login string, loc models.Location) error
	LastLocation(ctx context.Context, login string) (models.Location, error)
}

type LocationCache interface {
	Published(ctx context.Context, login string) (loc models.Location, publishedAt time.Time, err error)
	Publish(ctx context.Context, login string, loc models.Location, at time.Time) error
	Drop(ctx context.Context, login string) error
}

type UseCase struct {
	repo        DatabaseRepository
	locations   LocationCache
	locationTTL time.Duration
}

func New(repo DatabaseRepository, locations LocationCache, locationTTL time.Duration) *UseCase {
	return &UseCase{
		repo:        repo,
		locations:   locations,
		locationTTL: locationTTL,
	}
}

func (u *UseCase) Health(ctx context.Context) error {
	const op = "usecase.Health"

	if err := u.repo.Ping(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *UseCase) Login(ctx context.Context, req models.LoginRequest) (models.User, error) {
	const op = "usecase.Login"

	if err := req.NormalizeLogin(); err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := u.repo.UserByLogin(ctx, req.Login)
	if err == nil {
		return user, err
	}

	if !errors.Is(err, models.ErrUserNotFound) {
		err = fmt.Errorf("%s: %w", op, err)
		return user, err
	}

	user, err = u.repo.CreateUser(ctx, req.Login)
	switch {
	case err == nil:
		slog.Info("user registered", slog.String("login", req.Login))
		return user, err

	case errors.Is(err, models.ErrLoginTaken):
		if user, err = u.repo.UserByLogin(ctx, req.Login); err != nil {
			return user, err
		}

		return user, nil
	default:
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
}

func (u *UseCase) Profile(ctx context.Context, login string) (models.User, error) {
	const op = "usecase.Profile"

	user, err := u.repo.UserByLogin(ctx, login)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (u *UseCase) Users(ctx context.Context, exceptLogin string) ([]models.User, error) {
	const op = "usecase.Users"

	users, err := u.repo.Users(ctx, exceptLogin)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (u *UseCase) UpdateNote(ctx context.Context, login, note string) (models.User, error) {
	const op = "usecase.UpdateNote"

	if utf8.RuneCountInString(note) > models.MaxNoteLength {
		return models.User{}, fmt.Errorf("%s: %w", op, models.ErrNoteTooLong)
	}

	if err := u.repo.UpdateNote(ctx, login, note); err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := u.repo.UserByLogin(ctx, login)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

const clockSkew = time.Minute

func (u *UseCase) ReportLocation(ctx context.Context, login string, req models.ReportLocationRequest) error {
	const op = "usecase.ReportLocation"

	loc, err := req.Location()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	loc.UpdatedAt = stamp(loc.UpdatedAt)

	if err = u.repo.SaveLocation(ctx, login, loc); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.Debug("location reported",
		slog.String("login", login),
		slog.Time("ts", loc.UpdatedAt),
	)

	return nil
}

func stamp(ts time.Time) time.Time {
	now := time.Now().UTC()

	if ts.IsZero() || !ts.Before(now.Add(clockSkew)) {
		return now
	}

	return ts
}

func (u *UseCase) position(ctx context.Context, login string) (models.Location, time.Time, error) {
	now := time.Now().UTC()

	published, publishedAt, err := u.locations.Published(ctx, login)
	cached := err == nil

	switch {
	case cached && now.Sub(publishedAt) < u.locationTTL:
		return published, publishedAt, nil
	case !cached && !errors.Is(err, models.ErrLocationNotFound):
		return models.Location{}, time.Time{}, err
	}

	stored, err := u.repo.LastLocation(ctx, login)
	switch {
	case err == nil:
		if !cached || stored.UpdatedAt.After(published.UpdatedAt) {
			published = stored
		}
	case errors.Is(err, models.ErrLocationNotFound) && cached:
	default:
		return models.Location{}, time.Time{}, err
	}

	if err = u.locations.Publish(ctx, login, published, now); err != nil {
		return models.Location{}, time.Time{}, err
	}

	return published, now, nil
}

func (u *UseCase) Track(ctx context.Context, target string) (models.TargetLocation, error) {
	const op = "usecase.Track"

	target, err := models.NormalizeLogin(target)
	if err != nil {
		return models.TargetLocation{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := u.repo.UserByLogin(ctx, target)
	if err != nil {
		return models.TargetLocation{}, fmt.Errorf("%s: %w", op, err)
	}

	position, publishedAt, err := u.position(ctx, target)
	if err != nil {
		return models.TargetLocation{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.TargetLocation{
		Login:    user.Login,
		Note:     user.Note,
		NextAt:   publishedAt.Add(u.locationTTL),
		Location: position,
	}, nil
}
