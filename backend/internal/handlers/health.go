package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

// HealthHandler serves GET /health.
type HealthHandler struct {
	svc *services.HealthService
	log *slog.Logger
}

// NewHealthHandler wires the handler to its service.
func NewHealthHandler(svc *services.HealthService, log *slog.Logger) *HealthHandler {
	return &HealthHandler{svc: svc, log: log}
}

// Get responds 200 with the health snapshot as JSON.
func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := h.svc.Check()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Status is already written; log and return rather than double-writing.
		h.log.ErrorContext(r.Context(), "encode health response", slog.Any("error", err))
	}
}
