package services

import (
	"testing"
	"time"
)

// TestHealthServiceCheck asserts uptime is derived from the injected clock, so
// the assertion is exact and needs no sleep.
func TestHealthServiceCheck(t *testing.T) {
	startedAt := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)

	svc := &HealthService{
		startedAt: startedAt,
		version:   "test",
		now:       func() time.Time { return startedAt.Add(90 * time.Second) },
	}

	got := svc.Check()

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
