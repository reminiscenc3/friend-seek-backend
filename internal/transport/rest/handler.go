package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

type UseCase interface {
	Health(ctx context.Context) error

	Login(ctx context.Context, req models.LoginRequest) (models.User, error)
	Profile(ctx context.Context, login string) (models.User, error)
	Users(ctx context.Context, exceptLogin string) ([]models.User, error)
	UpdateNote(ctx context.Context, login, note string) (models.User, error)

	ReportLocation(ctx context.Context, login string, req models.ReportLocationRequest) error
	Track(ctx context.Context, target string) (models.TargetLocation, error)
}

func (s *Rest) router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Get("/health", s.health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", s.login)

		r.Group(func(r chi.Router) {
			r.Use(auth)

			r.Get("/users", s.users)
			r.Get("/users/profile", s.profile)
			r.Put("/users/profile/note", s.updateProfileNote)

			r.Post("/location", s.reportLocation)
			r.Get("/track/{login}", s.track)
		})
	})

	return r
}

func (s *Rest) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	if err := s.uc.Health(ctx); err != nil {
		slog.Error("health check failed", slog.String("error", err.Error()))
		fail(w, models.ErrNoDatabase)

		return
	}

	ok(w, http.StatusOK, models.HealthResponse{Status: "ok", Database: "ok"})
}

func (s *Rest) login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, models.ErrBadJSON)

		return
	}

	user, err := s.uc.Login(ctx, req)
	if err != nil {
		fail(w, err)

		return
	}

	ok(w, http.StatusOK, user)
}

func (s *Rest) users(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	users, err := s.uc.Users(ctx, getUserLoginFromCtx(ctx))
	if err != nil {
		fail(w, err)

		return
	}

	ok(w, http.StatusOK, users)
}

func (s *Rest) profile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	user, err := s.uc.Profile(ctx, getUserLoginFromCtx(ctx))
	if err != nil {
		fail(w, err)

		return
	}

	ok(w, http.StatusOK, user)
}

func (s *Rest) updateProfileNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	var req models.UpdateNoteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, models.ErrBadJSON)

		return
	}

	user, err := s.uc.UpdateNote(ctx, getUserLoginFromCtx(ctx), req.Note)
	if err != nil {
		fail(w, err)

		return
	}

	ok(w, http.StatusOK, user)
}

func (s *Rest) reportLocation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	var req models.ReportLocationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, models.ErrBadJSON)

		return
	}

	if err := req.Validate(); err != nil {
		fail(w, err)

		return
	}

	if err := s.uc.ReportLocation(ctx, getUserLoginFromCtx(ctx), req); err != nil {
		fail(w, err)

		return
	}

	ok(w, http.StatusAccepted, nil)
}

func (s *Rest) track(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	login := chi.URLParam(r, "login")
	if login == "" {
		fail(w, models.ErrNoLogin)
		return
	}

	location, err := s.uc.Track(ctx, login)
	if err != nil {
		fail(w, err)

		return
	}

	ok(w, http.StatusOK, location)
}
