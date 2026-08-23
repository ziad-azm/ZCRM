package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/models"
)

// fakePinger stands in for the database at the repositories.Pinger boundary, so
// every test in this file runs with PostgreSQL stopped.
type fakePinger struct {
	err     error
	block   bool // block until the context is cancelled, then return ctx.Err()
	callCnt int
}

func (f *fakePinger) Ping(ctx context.Context) error {
	f.callCnt++

	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}

	return f.err
}

func TestHealthServiceCheck(t *testing.T) {
	startedAt := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)

	svc := &HealthService{
		startedAt:   startedAt,
		version:     "test",
		now:         func() time.Time { return startedAt.Add(90 * time.Second) },
		db:          &fakePinger{},
		pingTimeout: 2 * time.Second,
	}

	got := svc.Check(context.Background())

	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q", got.Status, "ok")
	}
	if got.Version != "test" {
		t.Errorf("Version = %q, want %q", got.Version, "test")
	}
	if got.UptimeSeconds != 90 {
		t.Errorf("UptimeSeconds = %v, want 90", got.UptimeSeconds)
	}
	if !got.Timestamp.Equal(startedAt.Add(90 * time.Second)) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, startedAt.Add(90*time.Second))
	}
}

func TestHealthServiceDatabaseUp(t *testing.T) {
	svc := NewHealthService("test", 2*time.Second, &fakePinger{})

	got := svc.Check(context.Background())

	if got.Database != models.DatabaseUp {
		t.Errorf("Database = %q, want %q", got.Database, models.DatabaseUp)
	}
	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q", got.Status, "ok")
	}
}

// TestHealthServiceDatabaseDown pins the liveness-not-readiness decision: a
// failed ping must NOT degrade Status.
func TestHealthServiceDatabaseDown(t *testing.T) {
	svc := NewHealthService("test", 2*time.Second, &fakePinger{err: errors.New("boom")})

	got := svc.Check(context.Background())

	if got.Database != models.DatabaseDown {
		t.Errorf("Database = %q, want %q", got.Database, models.DatabaseDown)
	}
	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q — /health is a liveness probe", got.Status, "ok")
	}
}

func TestHealthServiceNilPinger(t *testing.T) {
	svc := NewHealthService("test", 2*time.Second, nil)

	got := svc.Check(context.Background())

	if got.Database != models.DatabaseDown {
		t.Errorf("Database = %q, want %q", got.Database, models.DatabaseDown)
	}
}

// TestHealthServicePingTimeout is the regression test for a hung database: the
// probe must be bounded by pingTimeout rather than blocking the request.
func TestHealthServicePingTimeout(t *testing.T) {
	svc := NewHealthService("test", 50*time.Millisecond, &fakePinger{block: true})

	start := time.Now()
	got := svc.Check(context.Background())
	elapsed := time.Since(start)

	if got.Database != models.DatabaseDown {
		t.Errorf("Database = %q, want %q", got.Database, models.DatabaseDown)
	}
	if elapsed > time.Second {
		t.Errorf("Check took %v, want it bounded by the 50ms pingTimeout", elapsed)
	}
}
