package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/olegtemek/friend-seek-backend/internal/models"
	"github.com/olegtemek/friend-seek-backend/internal/repository/memory"
)

const ttl = 3 * time.Minute

type stubRepo struct {
	last   models.Location
	lastOK bool
	saved  []models.Location
}

func (s *stubRepo) Ping(context.Context) error { return nil }

func (s *stubRepo) CreateUser(_ context.Context, login string) (models.User, error) {
	return models.User{Login: login}, nil
}

func (s *stubRepo) UserByLogin(_ context.Context, login string) (models.User, error) {
	return models.User{Login: login}, nil
}

func (s *stubRepo) Users(context.Context, string) ([]models.User, error) { return nil, nil }

func (s *stubRepo) UpdateNote(context.Context, string, string) error { return nil }

func (s *stubRepo) SaveLocation(_ context.Context, _ string, loc models.Location) error {
	s.saved = append(s.saved, loc)
	s.last, s.lastOK = loc, true

	return nil
}

func (s *stubRepo) LastLocation(context.Context, string) (models.Location, error) {
	if !s.lastOK {
		return models.Location{}, models.ErrLocationNotFound
	}

	return s.last, nil
}

func ptr[T any](v T) *T { return &v }

func at(lat float64, age time.Duration) models.Location {
	return models.Location{Latitude: lat, Longitude: 37.61, UpdatedAt: time.Now().UTC().Add(-age)}
}

func track(t *testing.T, u *UseCase, target string) models.TargetLocation {
	t.Helper()

	location, err := u.Track(context.Background(), target)
	if err != nil {
		t.Fatalf("Track: %v", err)
	}

	return location
}

func report(lat, lon float64) models.ReportLocationRequest {
	return models.ReportLocationRequest{Latitude: &lat, Longitude: &lon}
}

func TestPositionStaysFrozenInsideTTL(t *testing.T) {
	ctx := context.Background()
	repo := &stubRepo{last: at(55.80, 0), lastOK: true}
	cache := memory.New()
	u := New(repo, cache, ttl)

	frozen := at(55.75, 2*time.Minute)
	publishedAt := time.Now().UTC().Add(-2 * time.Minute)

	if err := cache.Publish(ctx, "target", frozen, publishedAt); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	tracked := track(t, u, "target")

	if !tracked.UpdatedAt.Equal(frozen.UpdatedAt) {
		t.Errorf("position moved inside the ttl: %v, want %v", tracked.UpdatedAt, frozen.UpdatedAt)
	}

	if !tracked.NextAt.Equal(publishedAt.Add(ttl)) {
		t.Errorf("next_at = %v, want %v", tracked.NextAt, publishedAt.Add(ttl))
	}

	got, gotAt, err := cache.Published(ctx, "target")
	if err != nil {
		t.Fatalf("Published: %v", err)
	}

	if got.Latitude != frozen.Latitude || !gotAt.Equal(publishedAt) {
		t.Errorf("cache changed inside the ttl: %v at %v", got.Latitude, gotAt)
	}
}

func TestPositionRefreshesFromTheTableAfterTTL(t *testing.T) {
	ctx := context.Background()
	fresh := at(55.80, 0)
	repo := &stubRepo{last: fresh, lastOK: true}
	cache := memory.New()
	u := New(repo, cache, ttl)

	stale := at(55.75, 4*time.Minute)
	if err := cache.Publish(ctx, "target", stale, time.Now().UTC().Add(-4*time.Minute)); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	before := time.Now().UTC()
	tracked := track(t, u, "target")

	if !tracked.UpdatedAt.Equal(fresh.UpdatedAt) {
		t.Errorf("position = %v, want the row from the table %v", tracked.UpdatedAt, fresh.UpdatedAt)
	}

	got, gotAt, err := cache.Published(ctx, "target")
	if err != nil {
		t.Fatalf("Published: %v", err)
	}

	if got.Latitude != fresh.Latitude {
		t.Errorf("cache = %v, want the refreshed %v", got.Latitude, fresh.Latitude)
	}

	if gotAt.Before(before) {
		t.Errorf("publishedAt = %v, want it moved to the moment of the lookup", gotAt)
	}
}

func TestPositionFallsBackToTheTableOnColdCache(t *testing.T) {
	stored := at(55.80, time.Minute)
	repo := &stubRepo{last: stored, lastOK: true}
	u := New(repo, memory.New(), ttl)

	tracked := track(t, u, "target")

	if !tracked.UpdatedAt.Equal(stored.UpdatedAt) {
		t.Errorf("position = %v, want %v", tracked.UpdatedAt, stored.UpdatedAt)
	}
}

func TestPositionKeepsTheStaleSnapshotWhenTheTableIsEmpty(t *testing.T) {
	ctx := context.Background()
	repo := &stubRepo{}
	cache := memory.New()
	u := New(repo, cache, ttl)

	stale := at(55.75, 4*time.Minute)
	if err := cache.Publish(ctx, "target", stale, time.Now().UTC().Add(-4*time.Minute)); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	tracked := track(t, u, "target")

	if !tracked.UpdatedAt.Equal(stale.UpdatedAt) {
		t.Errorf("position = %v, want the last known %v", tracked.UpdatedAt, stale.UpdatedAt)
	}
}

func TestTrackWithoutAnyLocation(t *testing.T) {
	u := New(&stubRepo{}, memory.New(), ttl)

	_, err := u.Track(context.Background(), "target")
	if !errors.Is(err, models.ErrLocationNotFound) {
		t.Errorf("Track = %v, want ErrLocationNotFound", err)
	}
}

func TestReportLocationWritesToTheTable(t *testing.T) {
	repo := &stubRepo{}
	u := New(repo, memory.New(), ttl)

	req := report(55.75, 37.61)
	req.Accuracy = 12

	if err := u.ReportLocation(context.Background(), "oleg", req); err != nil {
		t.Fatalf("ReportLocation: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("saved %d rows, want 1", len(repo.saved))
	}

	if repo.saved[0].Accuracy != 12 || repo.saved[0].UpdatedAt.IsZero() {
		t.Errorf("saved row = %+v, want the accuracy and a timestamp", repo.saved[0])
	}
}

func TestReportDoesNotDisturbTheFrozenPosition(t *testing.T) {
	ctx := context.Background()
	repo := &stubRepo{}
	cache := memory.New()
	u := New(repo, cache, ttl)

	frozen := at(55.75, time.Minute)
	if err := cache.Publish(ctx, "target", frozen, time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	if err := u.ReportLocation(ctx, "target", report(55.99, 37.61)); err != nil {
		t.Fatalf("ReportLocation: %v", err)
	}

	got, _, err := cache.Published(ctx, "target")
	if err != nil {
		t.Fatalf("Published: %v", err)
	}

	if got.Latitude != frozen.Latitude {
		t.Errorf("published position = %v, want the frozen %v", got.Latitude, frozen.Latitude)
	}
}

func TestReportKeepsTheReportedTimestamp(t *testing.T) {
	repo := &stubRepo{}
	u := New(repo, memory.New(), ttl)

	ts := time.Now().UTC().Add(-30 * time.Second).Truncate(time.Millisecond)

	req := report(55.751244, 37.618423)
	req.Azimuth, req.Timestamp, req.Accuracy = ptr(137.4), &ts, 8

	if err := u.ReportLocation(context.Background(), "oleg", req); err != nil {
		t.Fatalf("ReportLocation: %v", err)
	}

	saved := repo.saved[0]

	if !saved.UpdatedAt.Equal(ts) {
		t.Errorf("ts = %v, want the reported %v", saved.UpdatedAt, ts)
	}

	if saved.Azimuth != 137.4 || saved.Accuracy != 8 {
		t.Errorf("saved row = %+v, want azimuth 137.4 and accuracy 8", saved)
	}
}

func TestReportDropsAFutureTimestamp(t *testing.T) {
	repo := &stubRepo{}
	u := New(repo, memory.New(), ttl)

	req := report(55.75, 37.61)
	req.Timestamp = ptr(time.Now().UTC().Add(time.Hour))

	before := time.Now().UTC()

	if err := u.ReportLocation(context.Background(), "oleg", req); err != nil {
		t.Fatalf("ReportLocation: %v", err)
	}

	if saved := repo.saved[0]; saved.UpdatedAt.Before(before) || saved.UpdatedAt.After(time.Now().UTC()) {
		t.Errorf("ts = %v, want the server clock", saved.UpdatedAt)
	}
}

func TestReportNormalizesTheAzimuth(t *testing.T) {
	repo := &stubRepo{}
	u := New(repo, memory.New(), ttl)

	req := report(55.75, 37.61)
	req.Azimuth = ptr(-10.0)

	if err := u.ReportLocation(context.Background(), "oleg", req); err != nil {
		t.Fatalf("ReportLocation: %v", err)
	}

	if got := repo.saved[0].Azimuth; got != 350 {
		t.Errorf("azimuth = %v, want 350", got)
	}
}

func TestReportRejectsCoordinatesOffTheGlobe(t *testing.T) {
	u := New(&stubRepo{}, memory.New(), ttl)

	err := u.ReportLocation(context.Background(), "oleg", report(91, 37.61))
	if !errors.Is(err, models.ErrInvalidLocation) {
		t.Errorf("ReportLocation = %v, want ErrInvalidLocation", err)
	}
}

func TestTrackReturnsTheStoredFix(t *testing.T) {
	stored := models.Location{
		Latitude:  55.751244,
		Longitude: 37.618423,
		Azimuth:   137.4,
		Accuracy:  8,
		UpdatedAt: time.Now().UTC().Add(-time.Minute),
	}

	u := New(&stubRepo{last: stored, lastOK: true}, memory.New(), ttl)

	got := track(t, u, "target")

	if got.Location != stored {
		t.Errorf("location = %+v, want %+v", got.Location, stored)
	}

	if got.Login != "target" {
		t.Errorf("login = %q, want %q", got.Login, "target")
	}
}
