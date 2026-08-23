package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/repositories"
	"github.com/ziad-azm/ZCRM/backend/internal/server"
)

// version is the reported build version. ZCRM-6 replaces this with config, and
// a later story injects it at build time via -ldflags.
const version = "0.1.0"

const (
	defaultPort        = "8080"
	defaultDatabaseURL = "postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable"
	shutdownTimeout    = 10 * time.Second
	readTimeout        = 10 * time.Second
	writeTimeout       = 15 * time.Second
	idleTimeout        = 60 * time.Second
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)

	// Minimal env read only. The full config loader is ZCRM-6, which also owns
	// validating this value instead of letting ListenAndServe reject it.
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Minimal env read only, matching the PORT pattern above. ZCRM-6 replaces both.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}

	pool, err := repositories.NewPool(context.Background(), dsn)
	if err != nil {
		// A malformed DSN is a configuration error, not a transient outage.
		log.Error("database pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	// One eager probe so startup logs say plainly whether the database answered.
	// A failure is not fatal: /health reports it and the API keeps serving.
	probeCtx, cancelProbe := context.WithTimeout(context.Background(), 5*time.Second)
	if err := pool.Ping(probeCtx); err != nil {
		log.Warn("database unreachable at startup", slog.Any("error", err))
	} else {
		log.Info("database connected")
	}
	cancelProbe()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      server.New(log, version, pool),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Cancelled on Ctrl+C (SIGINT) or SIGTERM from a container runtime.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server starting", slog.String("addr", srv.Addr), slog.String("version", version))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			log.Error("server failed", slog.Any("error", err))
			os.Exit(1)
		}
	case <-ctx.Done():
		log.Info("shutdown signal received", slog.Duration("grace_period", shutdownTimeout))
		stop() // restore default signal handling: a second Ctrl+C kills immediately

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", slog.Any("error", err))
			os.Exit(1)
		}
		log.Info("server stopped cleanly")
	}
}
