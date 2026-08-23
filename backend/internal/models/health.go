package models

import "time"

// Database reachability states reported by GET /health.
const (
	DatabaseUp   = "up"
	DatabaseDown = "down"
)

// HealthResponse is the JSON body returned by GET /health.
type HealthResponse struct {
	Status        string    `json:"status"`
	Version       string    `json:"version"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
	Database      string    `json:"database"`
}
