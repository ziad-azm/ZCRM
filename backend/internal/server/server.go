// Package server builds the HTTP router and its middleware chain.
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ziad-azm/ZCRM/backend/internal/config"
	"github.com/ziad-azm/ZCRM/backend/internal/handlers"
	"github.com/ziad-azm/ZCRM/backend/internal/middleware"
	"github.com/ziad-azm/ZCRM/backend/internal/repositories"
	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

// New returns the application router with all routes and middleware mounted.
func New(log *slog.Logger, cfg *config.Config, db repositories.Pinger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	// CORS sits BELOW RequestLogger on purpose: cors.Handler answers preflight
	// OPTIONS without calling the next handler, and RequestLogger logs from a
	// defer, so mounting it here is what makes preflight requests visible in
	// the logs. Above RequestID they would be invisible.
	//
	// AllowCredentials stays false: auth will use a bearer token in the
	// Authorization header, not a cookie. It flips only if ZCRM-9 chooses
	// cookie sessions, and the wildcard origin must be removed in that change.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: false,
		MaxAge:           cfg.CORS.MaxAge,
	}))
	r.Use(chimw.Recoverer)

	health := handlers.NewHealthHandler(
		services.NewHealthService(cfg.Version, cfg.Database.PingTimeout, db),
		log,
	)
	r.Get("/health", health.Get)

	// No OPTIONS catch-all: cors.Handler answers every genuine browser preflight
	// (OPTIONS carrying Origin + Access-Control-Request-Method) and stops the
	// chain, so one is unnecessary. Registering r.Options("/*") would also make
	// chi answer 405 instead of 404 for unknown paths, breaking that contract.

	return r
}
