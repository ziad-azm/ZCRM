package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ziad-azm/ZCRM/backend/internal/models"
	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

func TestHealthHandlerGet(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	h := NewHealthHandler(services.NewHealthService("test"), log)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var got models.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q", got.Status, "ok")
	}
	if got.Version != "test" {
		t.Errorf("Version = %q, want %q", got.Version, "test")
	}
	if got.Timestamp.IsZero() {
		t.Error("Timestamp is zero, want a populated time")
	}
}
