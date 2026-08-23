package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/config"
	"github.com/ziad-azm/ZCRM/backend/internal/repositories"
	"github.com/ziad-azm/ZCRM/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// The logger is not configured yet: write plainly to stderr and stop.
		fmt.Fprintf(os.Stderr, "configuration error: %v
", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Log.Level,
	}))
	slog.SetDefault(log)

	// Never log the DSN or JWT_SECRET: logs travel further than secrets should.
	log.Info("configuration loaded",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
		slog.String("log_level", cfg.Log.Level.String()),
	)

	pool, err := repositories.NewPool(context.Background(), cfg.Database)
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
		Addr:         ":" + cfg.Server.Port,
		Handler:      server.New(log, cfg, pool),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Cancelled on Ctrl+C (SIGINT) or SIGTERM from a container runtime.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server starting", slog.String("addr", srv.Addr), slog.String("version", cfg.Version))
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
		log.Info("shutdown signal received", slog.Duration("grace_period", cfg.Server.ShutdownTimeout))
		stop() // restore default signal handling: a second Ctrl+C kills immediately

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", slog.Any("error", err))
			os.Exit(1)
		}
		log.Info("server stopped cleanly")
	}
}
