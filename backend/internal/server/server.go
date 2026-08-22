// Package server builds the HTTP router and its middleware chain.
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ziad-azm/ZCRM/backend/internal/handlers"
	"github.com/ziad-azm/ZCRM/backend/internal/middleware"
	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

// New returns the application router with all routes and middleware mounted.
func New(log *slog.Logger, version string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimw.Recoverer)

	health := handlers.NewHealthHandler(services.NewHealthService(version), log)
	r.Get("/health", health.Get)

	return r
}
