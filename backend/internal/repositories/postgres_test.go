package repositories

import (
	"context"
	"strings"
	"testing"
)

func TestNewPoolRejectsMalformedDSN(t *testing.T) {
	pool, err := NewPool(context.Background(), "not-a-dsn")

	if err == nil {
		pool.Close()
		t.Fatal("NewPool(\"not-a-dsn\") = nil error, want a parse error")
	}
	if !strings.Contains(err.Error(), "parse database dsn") {
		t.Errorf("error = %q, want it to mention %q", err.Error(), "parse database dsn")
	}
}

// TestNewPoolIsLazy pins the deliberate behaviour that the pool is created even
// when nothing is listening: the API must start and report an unhealthy
// database rather than refusing to boot. Port 1 has no listener.
func TestNewPoolIsLazy(t *testing.T) {
	pool, err := NewPool(context.Background(), "postgres://u:p@localhost:1/db?sslmode=disable")
	if err != nil {
		t.Fatalf("NewPool with no listener returned error %v, want nil (pgxpool connects lazily)", err)
	}
	if pool == nil {
		t.Fatal("NewPool returned a nil pool with no error")
	}
	defer pool.Close()

	if got := pool.Config().MaxConns; got != maxConns {
		t.Errorf("MaxConns = %d, want %d", got, maxConns)
	}
	if got := pool.Config().MinConns; got != minConns {
		t.Errorf("MinConns = %d, want %d", got, minConns)
	}
	if got := pool.Config().ConnConfig.ConnectTimeout; got != connectTimeout {
		t.Errorf("ConnectTimeout = %v, want %v", got, connectTimeout)
	}
}
