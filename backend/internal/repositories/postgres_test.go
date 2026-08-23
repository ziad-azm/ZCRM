package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/config"
)

// testDatabaseConfig mirrors the production defaults so the assertions below
// prove configuration propagates into the pool.
func testDatabaseConfig(dsn string) config.Database {
	return config.Database{
		URL:            dsn,
		MaxConns:       10,
		MinConns:       2,
		ConnectTimeout: 5 * time.Second,
	}
}

func TestNewPoolRejectsMalformedDSN(t *testing.T) {
	pool, err := NewPool(context.Background(), testDatabaseConfig("not-a-dsn"))

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
	pool, err := NewPool(context.Background(), testDatabaseConfig("postgres://u:p@localhost:1/db?sslmode=disable"))
	if err != nil {
		t.Fatalf("NewPool with no listener returned error %v, want nil (pgxpool connects lazily)", err)
	}
	if pool == nil {
		t.Fatal("NewPool returned a nil pool with no error")
	}
	defer pool.Close()

	if got := pool.Config().MaxConns; got != 10 {
		t.Errorf("MaxConns = %d, want %d", got, 10)
	}
	if got := pool.Config().MinConns; got != 2 {
		t.Errorf("MinConns = %d, want %d", got, 2)
	}
	if got := pool.Config().ConnConfig.ConnectTimeout; got != 5*time.Second {
		t.Errorf("ConnectTimeout = %v, want %v", got, 5*time.Second)
	}
}
