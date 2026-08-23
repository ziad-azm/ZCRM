package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// validate reports every configuration problem at once.
func (c *Config) validate() error {
	var errs []error

	switch c.Env {
	case EnvDevelopment, EnvProduction:
	default:
		errs = append(errs, fmt.Errorf("APP_ENV: %q is not a known environment (want %s or %s)",
			c.Env, EnvDevelopment, EnvProduction))
	}

	if port, err := strconv.Atoi(c.Server.Port); err != nil || port < 1 || port > 65535 {
		errs = append(errs, fmt.Errorf("PORT: %q is not a valid TCP port (want 1-65535)", c.Server.Port))
	}

	if c.Database.MinConns < 0 {
		errs = append(errs, fmt.Errorf("DB_MIN_CONNS: must not be negative, got %d", c.Database.MinConns))
	}
	if c.Database.MaxConns < 1 {
		errs = append(errs, fmt.Errorf("DB_MAX_CONNS: must be at least 1, got %d", c.Database.MaxConns))
	}
	if c.Database.MaxConns < c.Database.MinConns {
		errs = append(errs, fmt.Errorf("DB_MAX_CONNS (%d) must be >= DB_MIN_CONNS (%d)",
			c.Database.MaxConns, c.Database.MinConns))
	}

	if len(c.CORS.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS: must list at least one origin"))
	}

	// A wildcard in production would let any site call the API with a user's
	// browser. Development keeps the escape hatch.
	if c.IsProduction() && c.CORS.AllowsAnyOrigin() {
		errs = append(errs, errors.New(`CORS_ALLOWED_ORIGINS: "*" is not allowed when APP_ENV=production`))
	}

	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			continue
		}
		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not an absolute origin (want e.g. https://app.example.com)", origin))
			continue
		}
		// Browsers send Origin with no path, so a configured path never matches.
		if u.Path != "" && u.Path != "/" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q must not include a path", origin))
		}
	}

	// Secrets may be absent in development so a fresh clone runs with no setup.
	// In production a missing secret is a deployment error, not a default.
	if c.IsProduction() && c.Auth.JWTSecret == "" {
		errs = append(errs, errors.New("JWT_SECRET: required when APP_ENV=production"))
	}

	return errors.Join(errs...)
}
