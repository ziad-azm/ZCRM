package services

import (
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/models"
)

// HealthService reports process liveness.
type HealthService struct {
	startedAt time.Time
	version   string
	now       func() time.Time // injected for tests
}

// NewHealthService returns a HealthService that treats now as the process start.
func NewHealthService(version string) *HealthService {
	return &HealthService{startedAt: time.Now(), version: version, now: time.Now}
}

// Check returns the current health snapshot. It never fails: this story has no
// dependencies to probe. ZCRM-5 extends it to ping the database.
func (s *HealthService) Check() models.HealthResponse {
	now := s.now()
	return models.HealthResponse{
		Status:        "ok",
		Version:       s.version,
		UptimeSeconds: now.Sub(s.startedAt).Seconds(),
		Timestamp:     now.UTC(),
	}
}
