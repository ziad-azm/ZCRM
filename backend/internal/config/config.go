// Package config loads and validates all application configuration from the
// environment. It is the only package that reads environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Environment names recognised by APP_ENV.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Config is the fully resolved application configuration.
type Config struct {
	Env      string
	Version  string
	Log      Log
	Server   Server
	CORS     CORS
	Database Database
	Auth     Auth
}

// Log holds logging configuration.
type Log struct {
	Level slog.Level
}

// Server holds HTTP server configuration.
type Server struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// CORS holds cross-origin request policy.
type CORS struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// AllowsAnyOrigin reports whether the policy contains a wildcard.
func (c CORS) AllowsAnyOrigin() bool {
	for _, o := range c.AllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}

// Database holds PostgreSQL connection and pool configuration.
type Database struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
	PingTimeout       time.Duration
}

// Auth holds authentication secrets. Nothing reads JWTSecret until ZCRM-9; it
// is loaded and validated now so the deployment surface is complete.
type Auth struct {
	JWTSecret string
}

// IsProduction reports whether the service is running in production.
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

// Load reads configuration from the environment, applies defaults, and
// validates the result. It returns every validation problem at once rather
// than failing on the first.
func Load() (*Config, error) {
	loadDotEnv()

	var errs []error
	collect := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	cfg := &Config{
		Env:     getString("APP_ENV", EnvDevelopment),
		Version: getString("APP_VERSION", "0.1.0"),
	}

	level, err := getLogLevel("LOG_LEVEL", slog.LevelInfo)
	collect(err)
	cfg.Log = Log{Level: level}

	readTimeout, err := getDuration("READ_TIMEOUT", 10*time.Second)
	collect(err)
	writeTimeout, err := getDuration("WRITE_TIMEOUT", 15*time.Second)
	collect(err)
	idleTimeout, err := getDuration("IDLE_TIMEOUT", 60*time.Second)
	collect(err)
	shutdownTimeout, err := getDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	collect(err)

	cfg.Server = Server{
		Port:            getString("PORT", "8080"),
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		ShutdownTimeout: shutdownTimeout,
	}

	corsMaxAge, err := getInt32("CORS_MAX_AGE", 300)
	collect(err)

	cfg.CORS = CORS{
		// The Angular dev server's default origin. Production sets its real one.
		AllowedOrigins: getStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:4200"}),
		AllowedMethods: getStringSlice("CORS_ALLOWED_METHODS",
			[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		// Authorization is allowed up front: ZCRM-9 sends a bearer token, and a
		// missing entry here fails preflight in a way that looks like broken login.
		AllowedHeaders: getStringSlice("CORS_ALLOWED_HEADERS",
			[]string{"Accept", "Authorization", "Content-Type", "X-Requested-With"}),
		MaxAge: int(corsMaxAge),
	}

	maxConns, err := getInt32("DB_MAX_CONNS", 10)
	collect(err)
	minConns, err := getInt32("DB_MIN_CONNS", 2)
	collect(err)
	maxConnLifetime, err := getDuration("DB_MAX_CONN_LIFETIME", time.Hour)
	collect(err)
	maxConnIdleTime, err := getDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute)
	collect(err)
	healthCheckPeriod, err := getDuration("DB_HEALTHCHECK_PERIOD", time.Minute)
	collect(err)
	connectTimeout, err := getDuration("DB_CONNECT_TIMEOUT", 5*time.Second)
	collect(err)
	pingTimeout, err := getDuration("DB_PING_TIMEOUT", 2*time.Second)
	collect(err)

	cfg.Database = Database{
		URL:               databaseURL(),
		MaxConns:          maxConns,
		MinConns:          minConns,
		MaxConnLifetime:   maxConnLifetime,
		MaxConnIdleTime:   maxConnIdleTime,
		HealthCheckPeriod: healthCheckPeriod,
		ConnectTimeout:    connectTimeout,
		PingTimeout:       pingTimeout,
	}

	cfg.Auth = Auth{JWTSecret: getString("JWT_SECRET", "")}

	if err := cfg.validate(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid configuration: %w", errors.Join(errs...))
	}

	return cfg, nil
}
