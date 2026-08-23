package services

import (
	"context"
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/models"
	"github.com/ziad-azm/ZCRM/backend/internal/repositories"
)

// HealthService reports process liveness and dependency reachability.
type HealthService struct {
	startedAt   time.Time
	version     string
	now         func() time.Time // injected for tests
	db          repositories.Pinger
	pingTimeout time.Duration
}

// NewHealthService returns a HealthService that treats now as the process start.
// db may be nil, in which case the database is reported as down.
func NewHealthService(version string, pingTimeout time.Duration, db repositories.Pinger) *HealthService {
	return &HealthService{
		startedAt:   time.Now(),
		version:     version,
		now:         time.Now,
		db:          db,
		pingTimeout: pingTimeout,
	}
}

// Check returns the current health snapshot. It always succeeds: an unreachable
// database is reported in the Database field, not as an error.
//
// Status stays "ok" and the handler keeps returning 200 even when the database
// is down: /health is a liveness probe, not a readiness probe. Readiness
// semantics (503 on a failed dependency) are deliberately out of scope.
func (s *HealthService) Check(ctx context.Context) models.HealthResponse {
	now := s.now()

	return models.HealthResponse{
		Status:        "ok",
		Version:       s.version,
		UptimeSeconds: now.Sub(s.startedAt).Seconds(),
		Timestamp:     now.UTC(),
		Database:      s.databaseStatus(ctx),
	}
}

func (s *HealthService) databaseStatus(ctx context.Context) string {
	if s.db == nil {
		return models.DatabaseDown
	}

	ctx, cancel := context.WithTimeout(ctx, s.pingTimeout)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		return models.DatabaseDown
	}

	return models.DatabaseUp
}
