package memory

import (
	"context"
	"sync"
	"time"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

type snapshot struct {
	location    models.Location
	publishedAt time.Time
}

type Repository struct {
	mu      sync.RWMutex
	entries map[string]snapshot
}

func New() *Repository {
	return &Repository{entries: make(map[string]snapshot)}
}

func (r *Repository) Published(_ context.Context, login string) (models.Location, time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.entries[login]
	if !ok {
		return models.Location{}, time.Time{}, models.ErrLocationNotFound
	}

	return s.location, s.publishedAt, nil
}

func (r *Repository) Publish(_ context.Context, login string, loc models.Location, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[login] = snapshot{location: loc, publishedAt: at.UTC()}

	return nil
}

func (r *Repository) Drop(_ context.Context, login string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.entries, login)

	return nil
}
