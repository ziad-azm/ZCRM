package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// getString returns the trimmed value of key, or def when unset or blank.
func getString(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// getStringSlice splits a comma-separated value, trimming each element and
// dropping empties. Returns def when the variable is unset or has no usable
// elements.
func getStringSlice(key string, def []string) []string {
	raw := getString(key, "")
	if raw == "" {
		return def
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}

	if len(out) == 0 {
		return def
	}

	return out
}

func getDuration(key string, def time.Duration) (time.Duration, error) {
	raw := getString(key, "")
	if raw == "" {
		return def, nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not a duration (want e.g. 30s, 5m, 1h): %w", key, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s: must be positive, got %s", key, d)
	}

	return d, nil
}

func getInt32(key string, def int32) (int32, error) {
	raw := getString(key, "")
	if raw == "" {
		return def, nil
	}

	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not an integer: %w", key, raw, err)
	}

	return int32(n), nil
}

func getLogLevel(key string, def slog.Level) (slog.Level, error) {
	raw := getString(key, "")
	if raw == "" {
		return def, nil
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return 0, fmt.Errorf("%s: %q is not a log level (want debug, info, warn or error): %w", key, raw, err)
	}

	return level, nil
}

// databaseURL returns DATABASE_URL when set, otherwise composes a DSN from the
// same POSTGRES_* variables docker-compose.yml reads, so one .env drives both.
func databaseURL() string {
	if dsn := getString("DATABASE_URL", ""); dsn != "" {
		return dsn
	}

	user := getString("POSTGRES_USER", "zcrm")
	password := getString("POSTGRES_PASSWORD", "zcrm")
	host := getString("POSTGRES_HOST", "localhost")
	port := getString("POSTGRES_PORT", "5432")
	name := getString("POSTGRES_DB", "zcrm")
	sslMode := getString("POSTGRES_SSLMODE", "disable")

	// url.UserPassword escapes credentials: a password containing @, / or :
	// would otherwise produce an unparseable DSN whose failure looks like bad
	// credentials rather than a quoting bug.
	return fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s",
		url.UserPassword(user, password).String(),
		net.JoinHostPort(host, port),
		name,
		sslMode,
	)
}
