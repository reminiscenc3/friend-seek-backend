package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

var statuses = []struct {
	err    error
	status int
}{
	{errUnauthorized, http.StatusUnauthorized},

	{models.ErrBadJSON, http.StatusBadRequest},
	{models.ErrNoLogin, http.StatusBadRequest},
	{models.ErrNoDatabase, http.StatusServiceUnavailable},
	{models.ErrInvalidLogin, http.StatusBadRequest},
	{models.ErrNoteTooLong, http.StatusBadRequest},
	{models.ErrNoCoordinates, http.StatusBadRequest},
	{models.ErrInvalidLocation, http.StatusBadRequest},
	{models.ErrInvalidAzimuth, http.StatusBadRequest},
	{models.ErrUserNotFound, http.StatusNotFound},
	{models.ErrLocationNotFound, http.StatusNotFound},
	{models.ErrLoginTaken, http.StatusConflict},
}

func respond(w http.ResponseWriter, status int, msg string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(models.Response{Msg: msg, Data: data}); err != nil {
		slog.Error("failed to write response", slog.String("error", err.Error()))
	}
}

func ok(w http.ResponseWriter, status int, data any) {
	respond(w, status, "", data)
}

func fail(w http.ResponseWriter, err error) {
	for _, known := range statuses {
		if errors.Is(err, known.err) {
			respond(w, known.status, known.err.Error(), nil)

			return
		}
	}

	slog.Error("internal error", slog.String("error", err.Error()))
	respond(w, http.StatusInternalServerError, "internal error", nil)
}
