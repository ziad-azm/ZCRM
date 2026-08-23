package config

import (
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Every test uses t.Setenv, which restores the previous value and forbids
// t.Parallel(). Tests that assert defaults set APP_ENV=production so
// loadDotEnv() is skipped and a developer's .env cannot leak in.

// productionBaseline sets the minimum needed for a valid production config, so
// a test can assert on defaults without tripping validation.
func productionBaseline(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("JWT_SECRET", "test-secret")
}

func TestLoadDefaults(t *testing.T) {
	productionBaseline(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", cfg.Version, "0.1.0")
	}
	if cfg.Log.Level != slog.LevelInfo {
		t.Errorf("Log.Level = %v, want %v", cfg.Log.Level, slog.LevelInfo)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "8080")
	}
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.Server.ReadTimeout, 10*time.Second)
	}
	if cfg.Server.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", cfg.Server.WriteTimeout, 15*time.Second)
	}
	if cfg.Server.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.Server.ShutdownTimeout, 10*time.Second)
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("MaxConns = %d, want 10", cfg.Database.MaxConns)
	}
	if cfg.Database.MinConns != 2 {
		t.Errorf("MinConns = %d, want 2", cfg.Database.MinConns)
	}
	if cfg.Database.PingTimeout != 2*time.Second {
		t.Errorf("PingTimeout = %v, want %v", cfg.Database.PingTimeout, 2*time.Second)
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
}

func TestLoadOverrides(t *testing.T) {
	productionBaseline(t)
	t.Setenv("PORT", "9999")
	t.Setenv("READ_TIMEOUT", "45s")
	t.Setenv("DB_MAX_CONNS", "25")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("APP_VERSION", "9.9.9")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Server.Port != "9999" {
		t.Errorf("Port = %q, want %q", cfg.Server.Port, "9999")
	}
	if cfg.Server.ReadTimeout != 45*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.Server.ReadTimeout, 45*time.Second)
	}
	if cfg.Database.MaxConns != 25 {
		t.Errorf("MaxConns = %d, want 25", cfg.Database.MaxConns)
	}
	if cfg.Log.Level != slog.LevelDebug {
		t.Errorf("Log.Level = %v, want %v", cfg.Log.Level, slog.LevelDebug)
	}
	if cfg.Version != "9.9.9" {
		t.Errorf("Version = %q, want %q", cfg.Version, "9.9.9")
	}
}

func TestLoadInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "70000", "0", "-1"} {
		t.Run(port, func(t *testing.T) {
			productionBaseline(t)
			t.Setenv("PORT", port)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() with PORT=%q = nil error, want an error", port)
			}
			if !strings.Contains(err.Error(), "PORT") {
				t.Errorf("error = %q, want it to name PORT", err.Error())
			}
		})
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	for _, value := range []string{"nonsense", "0", "-5s"} {
		t.Run(value, func(t *testing.T) {
			productionBaseline(t)
			t.Setenv("READ_TIMEOUT", value)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() with READ_TIMEOUT=%q = nil error, want an error", value)
			}
			if !strings.Contains(err.Error(), "READ_TIMEOUT") {
				t.Errorf("error = %q, want it to name READ_TIMEOUT", err.Error())
			}
		})
	}
}

func TestLoadPoolBounds(t *testing.T) {
	t.Run("max below min", func(t *testing.T) {
		productionBaseline(t)
		t.Setenv("DB_MAX_CONNS", "1")
		t.Setenv("DB_MIN_CONNS", "5")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() = nil error, want an error")
		}
		for _, want := range []string{"DB_MAX_CONNS", "DB_MIN_CONNS"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error = %q, want it to name %s", err.Error(), want)
			}
		}
	})

	t.Run("max zero", func(t *testing.T) {
		productionBaseline(t)
		t.Setenv("DB_MAX_CONNS", "0")

		if _, err := Load(); err == nil {
			t.Fatal("Load() with DB_MAX_CONNS=0 = nil error, want an error")
		}
	})
}

func TestDatabaseURLPrecedence(t *testing.T) {
	t.Run("DATABASE_URL wins", func(t *testing.T) {
		productionBaseline(t)
		explicit := "postgres://a:b@example.com:6000/x?sslmode=require"
		t.Setenv("DATABASE_URL", explicit)
		t.Setenv("POSTGRES_USER", "ignored")
		t.Setenv("POSTGRES_PORT", "9999")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Database.URL != explicit {
			t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, explicit)
		}
	})

	t.Run("composed from POSTGRES_*", func(t *testing.T) {
		productionBaseline(t)
		t.Setenv("DATABASE_URL", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		want := "postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable"
		if cfg.Database.URL != want {
			t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, want)
		}
	})
}

// TestDatabaseURLEscapesCredentials is the regression test for a password
// containing URL-significant characters: the DSN must survive both net/url and
// pgx's own parser, which is what actually opens the connection.
func TestDatabaseURLEscapesCredentials(t *testing.T) {
	productionBaseline(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_PASSWORD", "p@ss:w/rd")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	parsed, err := url.Parse(cfg.Database.URL)
	if err != nil {
		t.Fatalf("net/url could not parse %q: %v", cfg.Database.URL, err)
	}
	if got, _ := parsed.User.Password(); got != "p@ss:w/rd" {
		t.Errorf("round-tripped password = %q, want %q", got, "p@ss:w/rd")
	}
	if parsed.Host != "localhost:5432" {
		t.Errorf("host = %q, want %q", parsed.Host, "localhost:5432")
	}

	if _, err := pgxpool.ParseConfig(cfg.Database.URL); err != nil {
		t.Errorf("pgxpool.ParseConfig(%q) = %v, want nil", cfg.Database.URL, err)
	}
}

func TestJWTSecretRequiredInProduction(t *testing.T) {
	t.Run("production without secret fails", func(t *testing.T) {
		t.Setenv("APP_ENV", EnvProduction)
		t.Setenv("JWT_SECRET", "")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() = nil error, want an error")
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Errorf("error = %q, want it to name JWT_SECRET", err.Error())
		}
	})

	t.Run("production with secret succeeds", func(t *testing.T) {
		t.Setenv("APP_ENV", EnvProduction)
		t.Setenv("JWT_SECRET", "abc")

		if _, err := Load(); err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
	})

	// A fresh clone must run with no setup at all.
	t.Run("development without secret succeeds", func(t *testing.T) {
		t.Setenv("APP_ENV", EnvDevelopment)
		t.Setenv("JWT_SECRET", "")

		if _, err := Load(); err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
	})
}

func TestUnknownAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "staging")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() with APP_ENV=staging = nil error, want an error")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Errorf("error = %q, want it to name APP_ENV", err.Error())
	}
}

// TestLoadReportsEveryProblem pins the errors.Join behaviour: three mistakes
// must surface in one run, not across three restarts.
func TestLoadReportsEveryProblem(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() = nil error, want an error")
	}
	for _, want := range []string{"APP_ENV", "PORT"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to name %s", err.Error(), want)
		}
	}
}

func TestCORSDefaults(t *testing.T) {
	productionBaseline(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if want := []string{"http://localhost:4200"}; !slices.Equal(cfg.CORS.AllowedOrigins, want) {
		t.Errorf("AllowedOrigins = %v, want %v", cfg.CORS.AllowedOrigins, want)
	}
	if !slices.Contains(cfg.CORS.AllowedMethods, "OPTIONS") {
		t.Errorf("AllowedMethods = %v, want it to contain OPTIONS", cfg.CORS.AllowedMethods)
	}
	// ZCRM-9 sends a bearer token; a missing entry fails preflight later.
	if !slices.Contains(cfg.CORS.AllowedHeaders, "Authorization") {
		t.Errorf("AllowedHeaders = %v, want it to contain Authorization", cfg.CORS.AllowedHeaders)
	}
	if cfg.CORS.MaxAge != 300 {
		t.Errorf("MaxAge = %d, want 300", cfg.CORS.MaxAge)
	}
	if cfg.CORS.AllowsAnyOrigin() {
		t.Error("AllowsAnyOrigin() = true for the default policy, want false")
	}
}

func TestCORSOriginsParsing(t *testing.T) {
	t.Run("trims and drops empties", func(t *testing.T) {
		productionBaseline(t)
		t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.test, http://b.test ,")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		want := []string{"http://a.test", "http://b.test"}
		if !slices.Equal(cfg.CORS.AllowedOrigins, want) {
			t.Errorf("AllowedOrigins = %v, want %v", cfg.CORS.AllowedOrigins, want)
		}
	})

	// An all-empty value must fall back to the default rather than produce an
	// empty policy that silently rejects every origin.
	for _, raw := range []string{",,", "", "  "} {
		t.Run("falls back on "+raw, func(t *testing.T) {
			productionBaseline(t)
			t.Setenv("CORS_ALLOWED_ORIGINS", raw)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if want := []string{"http://localhost:4200"}; !slices.Equal(cfg.CORS.AllowedOrigins, want) {
				t.Errorf("AllowedOrigins = %v, want %v", cfg.CORS.AllowedOrigins, want)
			}
		})
	}
}

func TestCORSWildcardRejectedInProduction(t *testing.T) {
	t.Run("production rejects", func(t *testing.T) {
		productionBaseline(t)
		t.Setenv("CORS_ALLOWED_ORIGINS", "*")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() with a wildcard origin in production = nil error, want an error")
		}
		if !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS") {
			t.Errorf("error = %q, want it to name CORS_ALLOWED_ORIGINS", err.Error())
		}
	})

	t.Run("development allows", func(t *testing.T) {
		t.Setenv("APP_ENV", EnvDevelopment)
		t.Setenv("CORS_ALLOWED_ORIGINS", "*")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v, want nil in development", err)
		}
		if !cfg.CORS.AllowsAnyOrigin() {
			t.Error("AllowsAnyOrigin() = false, want true")
		}
	})
}

func TestCORSOriginValidation(t *testing.T) {
	cases := []struct {
		origin  string
		wantErr bool
	}{
		{"http://app.test", false},
		{"https://app.test:8443", false},
		{"https://app.test/", false},
		{"not-a-url", true},
		{"localhost:4200", true},
		{"http://app.test/path", true},
	}

	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			productionBaseline(t)
			t.Setenv("CORS_ALLOWED_ORIGINS", tc.origin)

			_, err := Load()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Load() with origin %q = nil error, want an error", tc.origin)
				}
				if !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS") {
					t.Errorf("error = %q, want it to name CORS_ALLOWED_ORIGINS", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() with origin %q error = %v, want nil", tc.origin, err)
			}
		})
	}
}
