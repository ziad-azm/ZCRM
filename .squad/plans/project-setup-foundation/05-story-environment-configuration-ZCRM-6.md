# Story 05 — Environment Configuration (Story: ZCRM-6)

## Prerequisites

- Story 04 completed: [04-story-database-migrations-ZCRM-5.md](04-story-database-migrations-ZCRM-5.md). This story replaces the `os.Getenv` reads and the hardcoded pool tunables that Story 04 left behind, and unifies the dev DSN that Story 04 deliberately duplicated in three places.
- **Branch state at planning time.** `feature/ZCRM-5-database-migrations` (commit `7737f21`) is pushed but **unmerged**, and it is stacked on `feature/ZCRM-3-backend-structure`, which is also unmerged. `develop` still has no Go source. Branch this story from `feature/ZCRM-5-database-migrations`, or from `develop` once ZCRM-3 and ZCRM-5 have merged in that order.
- **ZCRM-4 is unfinished and this story touches the frontend.** `feature/ZCRM-4-frontend-structure` has only `ng new` (commit `87dbae5`); `frontend/src/environments/` does not exist on any branch. Task 8 is written to work either way — it creates the environment files if they are absent and extends them if ZCRM-4 landed first. Read task 8 before touching `frontend/`.
- Tooling verified present: `go1.26.7`, `@angular/cli 22.0.6`, `docker 20.10.22` with Compose `v2.15.1`.
- On this machine port `5432` is held by a native `postgresql-x64-17` service, so the container runs on `55432`. That is exactly the kind of per-machine difference this story exists to capture in a `.env` file.

---

## Story Goal

Move every setting out of Go constants and into environment variables with one typed, validated loader, and give both apps a documented configuration surface.

When this story is done:

1. `backend/internal/config` exposes `Load()` returning a validated `*Config`; **no `os.Getenv` call survives outside that package**.
2. A committed `.env.example` at the repository root documents every variable, and the real `.env` stays untracked. Docker Compose and the Go service read **the same file**.
3. The dev DSN exists in exactly **one** place — `.env.example` — rather than the three copies Story 04 left.
4. Invalid configuration fails at startup with a precise message, not at first use.
5. The Angular app has a typed environment contract for `apiBaseUrl` and its production counterpart.

**Not in scope:** CORS origins and the HTTP interceptor (**ZCRM-7** owns both, and adds `CORS_ALLOWED_ORIGINS` to the loader then). No secret manager, no Vault, no per-environment `.env.production` committed anywhere. No auth implementation — `JWT_SECRET` is loaded and validated because the intake names it, but **nothing reads it until ZCRM-9**. No CI, no deployment manifests. No change to the migration workflow beyond having the scripts read the same DSN variable.

---

## Context — Read These Files First

1. `backend/cmd/api/main.go` — **lines 17–47**. `const version = "0.1.0"` on line 19 with the comment *"ZCRM-6 replaces this with config"* (line 17); the `const` block on lines 21–28 (`defaultPort`, `defaultDatabaseURL`, and the four timeouts); the `PORT` read on lines 38–41 and the `DATABASE_URL` read on lines 43–47, both flagged *"ZCRM-6 replaces both"*. **Every one of these moves into the loader.**
2. `backend/internal/repositories/postgres.go` — **lines 20–26**, the `const` block (`maxConns = 10`, `minConns = 2`, `maxConnLifetime`, `maxConnIdleTime`, `healthCheckPeriod`, `connectTimeout`) carrying the comment *"ZCRM-6 makes them configurable"*. `NewPool`'s signature at line 34 changes in task 4.
3. `backend/internal/services/health.go` — **line 12**, `const pingTimeout = 2 * time.Second`. Becomes a constructor parameter in task 5.
4. `.gitignore` — **lines 10–15**. `.env` (line 11) and `.env.*` (line 12) are ignored, and `!.env.example` (line 13) un-ignores the template. **This is already correct — no `.gitignore` change is needed.** Verify with `git check-ignore` rather than editing (see Verification step 1).
5. `docker-compose.yml` — the `environment:` block reads `${POSTGRES_DB:-zcrm}`, `${POSTGRES_USER:-zcrm}`, `${POSTGRES_PASSWORD:-zcrm}`, and `ports:` reads `${POSTGRES_PORT:-5432}`. Compose auto-loads `.env` from its own directory, which is why the canonical `.env` location is the **repository root** and why the Go loader must compose its DSN from these same variable names.
6. `backend/README.md` — the **`## Not here yet`** section's first bullet names this story and the three-way DSN duplication. Task 9 deletes that bullet.
7. [04-story-database-migrations-ZCRM-5.md](04-story-database-migrations-ZCRM-5.md) — match its shape. Note its layering rule: `cmd → server → handlers → services → repositories`, with `models` importable anywhere. Task 2 extends that rule to `config`.
8. `frontend/src/app/app.config.ts` and `frontend/src/` — confirm whether `src/environments/` exists before starting task 8. On `feature/ZCRM-4-frontend-structure` it does **not**.

**Verified dependency version:** `github.com/joho/godotenv` **v1.5.1** (resolved against the module proxy during planning).

**Naming correction, carried from the overview:** the intake asks for *"environment.ts and environment.prod.ts"*. Angular 22 does not use `environment.prod.ts`. `ng generate environments` produces **`environment.ts`** (the production default) and **`environment.development.ts`** (substituted into the `development` configuration via `fileReplacements` in `angular.json`). **Do not create an `environment.prod.ts`** — it would be dead code that no build configuration references.

---

## Implementation tasks

Run Go commands from `e:\Work\AZM\ZCRM\backend`, `ng` commands from `e:\Work\AZM\ZCRM\frontend`, and `docker compose` from `e:\Work\AZM\ZCRM`.

### 1 — Branch

```bash
cd e:/Work/AZM/ZCRM
git checkout feature/ZCRM-5-database-migrations
git checkout -b feature/ZCRM-6-environment-config
```

Use `develop` instead once ZCRM-3 and ZCRM-5 have merged. Record which base was used in the PR description.

### 2 — The config package

```bash
cd e:/Work/AZM/ZCRM/backend
go get github.com/joho/godotenv@v1.5.1
```

`godotenv` is the one new dependency and it earns its place: Docker Compose loads `.env` automatically, but `go run ./cmd/api` does not, so without it every developer must export a dozen variables by hand in every shell. **Do not** add viper — it pulls a large dependency tree for a feature set this project does not use.

**`config` is a leaf package: it imports nothing from `internal/`.** That makes it importable by any layer, exactly like `models`. Update the layering note in `backend/README.md` (task 9) to say so.

**Create file: `backend/internal/config/config.go`**

```go
// Package config loads and validates all application configuration from the
// environment. It is the only package that reads environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
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

	port := getString("PORT", "8080")
	readTimeout, err := getDuration("READ_TIMEOUT", 10*time.Second)
	collect(err)
	writeTimeout, err := getDuration("WRITE_TIMEOUT", 15*time.Second)
	collect(err)
	idleTimeout, err := getDuration("IDLE_TIMEOUT", 60*time.Second)
	collect(err)
	shutdownTimeout, err := getDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	collect(err)

	cfg.Server = Server{
		Port:            port,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		ShutdownTimeout: shutdownTimeout,
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
```

**Create file: `backend/internal/config/env.go`** — the typed readers and the DSN composition:

```go
package config

// getString returns the trimmed value of key, or def when unset or blank.
func getString(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
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

	return fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s",
		url.UserPassword(user, password).String(),
		net.JoinHostPort(host, port),
		name,
		sslMode,
	)
}
```

Add `"net"` to the import block. **`url.UserPassword(...).String()` is required, not decoration** — a password containing `@`, `/`, or `:` produces an unparseable DSN if interpolated raw, and that failure looks like a wrong password rather than a quoting bug.

**Create file: `backend/internal/config/dotenv.go`**

```go
package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// dotEnvCandidates are searched in order; the first existing file wins. The
// canonical location is the repository root, which is also where Docker
// Compose looks, so `docker compose` and `go run ./cmd/api` read one file.
// Running from backend/ finds it via the parent path.
var dotEnvCandidates = []string{".env", filepath.Join("..", ".env")}

// loadDotEnv loads a .env file when one exists. Real environment variables
// always win: godotenv.Load never overwrites what is already set.
//
// A missing file is not an error — production sets real environment variables
// and ships no .env at all.
func loadDotEnv() {
	if os.Getenv("APP_ENV") == EnvProduction {
		return
	}

	for _, path := range dotEnvCandidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		// Errors are deliberately ignored: a malformed .env must not stop the
		// service when real environment variables may already be sufficient.
		_ = godotenv.Load(path)
		return
	}
}
```

**`godotenv.Load` (not `Overload`) is required.** `Load` leaves existing variables untouched, so a real environment variable always beats the file — the behaviour every deployment expects.

**Create file: `backend/internal/config/validate.go`**

```go
package config

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

	// Secrets may be absent in development so a fresh clone runs with no setup.
	// In production a missing secret is a deployment error, not a default.
	if c.IsProduction() && c.Auth.JWTSecret == "" {
		errs = append(errs, errors.New("JWT_SECRET: required when APP_ENV=production"))
	}

	return errors.Join(errs...)
}
```

`errors.Join` means a developer with three mistakes sees three messages on the first run instead of three consecutive restarts.

### 3 — `.env.example`

**Create file: `.env.example`** (repository **root** — Compose reads `.env` from its own directory, and the Go loader looks there too).

```dotenv
# Copy to .env and adjust. .env is git-ignored; this file is committed.
# Real environment variables always win over anything set here.

# ---------- Application ----------
APP_ENV=development          # development | production
APP_VERSION=0.1.0
LOG_LEVEL=info               # debug | info | warn | error

# ---------- HTTP server ----------
PORT=8080
READ_TIMEOUT=10s
WRITE_TIMEOUT=15s
IDLE_TIMEOUT=60s
SHUTDOWN_TIMEOUT=10s

# ---------- PostgreSQL ----------
# docker-compose.yml reads POSTGRES_DB/USER/PASSWORD/PORT from this same file.
# The Go service composes DATABASE_URL from these unless DATABASE_URL is set.
POSTGRES_DB=zcrm
POSTGRES_USER=zcrm
POSTGRES_PASSWORD=zcrm
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable
# Set this to override the composed DSN entirely (managed database, pgbouncer):
# DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require

# ---------- Connection pool ----------
DB_MAX_CONNS=10
DB_MIN_CONNS=2
DB_MAX_CONN_LIFETIME=1h
DB_MAX_CONN_IDLE_TIME=30m
DB_HEALTHCHECK_PERIOD=1m
DB_CONNECT_TIMEOUT=5s
DB_PING_TIMEOUT=2s

# ---------- Secrets ----------
# Required when APP_ENV=production. Nothing reads it until ZCRM-9 (auth).
# Generate one with: openssl rand -base64 48
JWT_SECRET=
```

**Do not create `.env`** as part of this story — it is per-developer and untracked. Task 10 documents `cp .env.example .env`.

`POSTGRES_PORT` deserves a note in the README: on a machine where 5432 is already taken, setting it in `.env` now fixes Compose **and** the composed DSN in one place. That is the concrete payoff of this story.

### 4 — `NewPool` takes configuration

**File: `backend/internal/repositories/postgres.go`** — delete the `const` block on lines 20–26 entirely and change the signature:

```go
package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ziad-azm/ZCRM/backend/internal/config"
)

// Pinger reports whether a data store is reachable. Services depend on this
// interface rather than on *pgxpool.Pool so they stay testable without a
// database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// NewPool returns a connection pool configured from cfg.
//
// It does NOT establish a connection: pgxpool connects lazily, so a database
// that is down produces no error here and surfaces on the first query or Ping.
// That is intentional — the API must still start and report an unhealthy
// database rather than refusing to boot.
func NewPool(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	return pool, nil
}
```

The `time` import goes away with the constants. **Keep the error string `"parse database dsn"`** — `postgres_test.go` asserts on it.

### 5 — `pingTimeout` becomes a parameter

**File: `backend/internal/services/health.go`** — delete the `const pingTimeout` on line 12 and store it on the struct:

```go
// HealthService reports process liveness and dependency reachability.
type HealthService struct {
	startedAt   time.Time
	version     string
	now         func() time.Time // injected for tests
	db          repositories.Pinger
	pingTimeout time.Duration
}

// NewHealthService returns a HealthService that treats now as the process start.
// db may be nil, in which case the database is reported as down.
func NewHealthService(version string, pingTimeout time.Duration, db repositories.Pinger) *HealthService {
	return &HealthService{
		startedAt:   time.Now(),
		version:     version,
		now:         time.Now,
		db:          db,
		pingTimeout: pingTimeout,
	}
}
```

`databaseStatus` uses `s.pingTimeout` instead of the constant. **`services` must not import `config`** — passing a `time.Duration` keeps the service testable with a literal and keeps the layer free of configuration concerns.

### 6 — `server.New` takes the config

**File: `backend/internal/server/server.go`**:

```go
// New returns the application router with all routes and middleware mounted.
func New(log *slog.Logger, cfg *config.Config, db repositories.Pinger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimw.Recoverer)

	health := handlers.NewHealthHandler(
		services.NewHealthService(cfg.Version, cfg.Database.PingTimeout, db),
		log,
	)
	r.Get("/health", health.Get)

	return r
}
```

Add the `config` import. **The middleware order is unchanged and still load-bearing** — do not touch it.

### 7 — `main.go` reads only the config

**File: `backend/cmd/api/main.go`** — this is the task that deletes the most code.

- **Delete** `const version = "0.1.0"` (line 19) and the whole `const` block (lines 21–28). Every value now comes from `cfg`.
- **Delete** the `PORT` block (lines 38–41) and the `DATABASE_URL` block (lines 43–47).
- Load config **first**, before the logger, because the logger's level comes from it:

```go
func main() {
	cfg, err := config.Load()
	if err != nil {
		// The logger is not configured yet: write plainly to stderr and stop.
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Log.Level,
	}))
	slog.SetDefault(log)

	log.Info("configuration loaded",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
		slog.String("log_level", cfg.Log.Level.String()),
	)

	pool, err := repositories.NewPool(context.Background(), cfg.Database)
	if err != nil {
		log.Error("database pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()
	// ... eager probe unchanged ...

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      server.New(log, cfg, pool),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
```

Replace the two `shutdownTimeout` uses in the shutdown branch with `cfg.Server.ShutdownTimeout`, and `slog.String("version", version)` with `slog.String("version", cfg.Version)`.

**Never log the DSN or `JWT_SECRET`.** The `configuration loaded` line above logs env, version, and level only. A DSN carries the password, and logs get shipped to places secrets should not reach. Log `cfg.Server.Port` if a port is useful, never `cfg.Database.URL`.

Add `"fmt"` and the `config` import. `os` is already imported.

### 8 — Angular environments

**First check what exists:** `ls frontend/src/environments`. Then take exactly one path.

**Path A — the folder does not exist** (the state on every branch at planning time). Generate it, which also writes the `fileReplacements` entry into `angular.json`:

```bash
cd e:/Work/AZM/ZCRM/frontend
ng generate environments
```

**Path B — ZCRM-4 already created it.** Skip the generate step and edit the two files in place.

Either way, end with these three files.

**Create file: `frontend/src/environments/environment.model.ts`**

```ts
/** Shape shared by every environment file, so the two cannot drift apart. */
export interface Environment {
  production: boolean;
  /** Base path for API calls. Relative on purpose — see frontend/README.md. */
  apiBaseUrl: string;
}
```

**File: `frontend/src/environments/environment.ts`** — the **production** default:

```ts
import { Environment } from './environment.model';

export const environment: Environment = {
  production: true,
  apiBaseUrl: '/api',
};
```

**File: `frontend/src/environments/environment.development.ts`** — substituted by the `development` build configuration:

```ts
import { Environment } from './environment.model';

export const environment: Environment = {
  production: false,
  apiBaseUrl: '/api',
};
```

Three points that must not be "improved" away:

- **`environment.ts` is the production file.** Angular's `fileReplacements` swaps in `environment.development.ts` for the development configuration, which is the inverse of the old `environment.prod.ts` convention. There is **no** `environment.prod.ts`.
- **`apiBaseUrl` is relative (`/api`) in both.** Same-origin requests need no CORS; the dev-server proxy handles development and a reverse proxy handles production. An absolute `http://localhost:8080` would break the moment the app is served from anywhere else, and would need CORS that does not exist until ZCRM-7.
- **The typed `Environment` interface is the point of the model file.** Adding a field to one environment and forgetting the other becomes a compile error rather than a runtime `undefined`.

Angular has no `.env` mechanism: these files are compiled into the bundle, so **anything in them is public**. Never put a secret in `src/environments/`. That is stated in the README in task 10.

### 9 — Update `backend/README.md`

- Add a **`## Configuration`** section: `cp .env.example .env`, the precedence rule (real environment variables → `.env` → built-in defaults), the fact that `internal/config` is the only package reading the environment, and that invalid values fail at startup listing **all** problems at once.
- Note that `docker compose` and the Go service read the **same** root `.env`, and that setting `POSTGRES_PORT` there fixes both the container's published port and the composed DSN.
- Update the **`## Layout`** tree with `internal/config/`.
- Extend the layering paragraph: `config`, like `models`, imports nothing internal and is importable by any layer. `services` deliberately does **not** import it.
- In **`## Database`**, replace the hardcoded default-DSN paragraph with a pointer to `.env.example` and the `POSTGRES_*` composition rule.
- In **`## Not here yet`**, delete the config-loader bullet (the whole first bullet) and keep the domain-tables and CORS bullets.

### 10 — Update the root `README.md` and `frontend/README.md`

**File: `README.md`** — add a **`## Configuration`** section after `## Getting started`: `cp .env.example .env` as the first step after cloning, that `.env` is git-ignored while `.env.example` is committed, and that one root file feeds Compose and the Go service. Add the `cp .env.example .env` line to the `### Backend` quickstart before `go run ./cmd/api`.

**File: `frontend/README.md`** — if ZCRM-4 has already rewritten this file, add an `## Environments` section; if it is still the CLI-generated placeholder, leave it alone and let ZCRM-4 write the section (record that in the PR description). The section states: `environment.ts` is production, `environment.development.ts` is development, both typed by `Environment`, `apiBaseUrl` is relative, and **no secret may ever go in these files** because they are compiled into a public bundle.

### 11 — Commit

```bash
cd e:/Work/AZM/ZCRM
git add backend .env.example README.md frontend
git status --short
```

`.env.example` **must** appear as a new file. **`.env` must not appear at all** — if it does, `.gitignore` was damaged; restore it before committing. Verify with `git check-ignore -v .env`.

```bash
git commit -m "feat: load all configuration from the environment (ZCRM-6)"
git push -u origin feature/ZCRM-6-environment-config
```

Open the pull request into **`develop`**.

---

## Edge Cases & Failure Modes

- **`.env` committed by accident.** The single worst outcome of this story: real credentials in git history. `.gitignore` lines 11–13 already prevent it, and the staging check in task 11 plus Verification step 1 confirm it. Recovery if it ever happens: `git rm --cached .env`, rotate every secret it held, and treat the history as compromised.
- **`.env.example` accidentally ignored.** `.env.*` (line 12) matches `.env.example`, so the `!.env.example` negation on line 13 is what makes the template committable, and **order matters** — a negation before the wildcard has no effect. Verified by Verification step 1.
- **`godotenv.Overload` instead of `Load`.** `Overload` lets the file beat real environment variables, so a stale developer `.env` would silently override production settings. Enforced by the explicit `Load` in task 2.
- **A `.env` password containing `@`, `:`, or `/`.** Interpolating it raw into the DSN yields an unparseable URL, and the error reads like bad credentials. Enforced by `url.UserPassword(...).String()` in `databaseURL()`, and asserted by Test Plan step 7.
- **`DATABASE_URL` and `POSTGRES_*` both set and disagreeing.** `DATABASE_URL` wins, unconditionally and by design; the `POSTGRES_*` values are then dead. Documented in `.env.example` and asserted by Test Plan step 6. Diagnose with the `configuration loaded` log line plus the startup probe result — never by logging the DSN.
- **Compose and the Go service reading different files.** Compose loads `.env` from its own directory (the repository root); the Go loader tries `.env` then `../.env`. Running `go run ./cmd/api` from `backend/` therefore finds the root file, and running it from the root also works. A `.env` placed inside `backend/` would shadow the root one and silently diverge from Compose — **do not create one there**.
- **`APP_ENV=production` in a developer shell.** `loadDotEnv` returns immediately, so the `.env` file is ignored and every value falls back to a default or a real environment variable — and a missing `JWT_SECRET` then fails validation. This is intended, and the error message says exactly which variable is missing.
- **Typo'd variable name.** `POSTGRES_PORTT=55432` is silently ignored and the default `5432` is used, which surfaces as a connection failure rather than a config error. This is an accepted limitation: the loader validates values, not the set of names. `.env.example` is the defence — keep it exhaustive.
- **Zero or negative durations.** `READ_TIMEOUT=0` would disable the read deadline entirely and `-5s` is meaningless; `getDuration` rejects both with a message naming the variable. Asserted by Test Plan step 4.
- **`DB_MAX_CONNS` below `DB_MIN_CONNS`.** `pgxpool` fails at pool creation with a message that does not name the environment variables. `validate()` catches it first and names both. Asserted by Test Plan step 5.
- **Secret in an Angular environment file.** `src/environments/*.ts` is compiled into the JavaScript bundle and served to every visitor. Any secret placed there is public. Stated in task 8 and in the frontend README.
- **`environment.prod.ts` created out of habit.** No build configuration references that filename in Angular 22, so it would be dead code that appears to work while changing nothing. Enforced by the naming correction in Context.
- **Tests polluting each other's environment.** Config tests must use `t.Setenv`, which restores the previous value automatically and forbids `t.Parallel()`. A test that calls `os.Setenv` directly leaks into every later test in the package.
- **A stray `.env` breaking the test suite.** `Load()` calls `loadDotEnv()`, so a developer's real `.env` two directories up could leak into a test's expectations. Enforced by Test Plan step 1, which sets `APP_ENV=production` (skipping `.env` loading) or asserts only on explicitly-set variables.

---

## Test Plan

All tests run with `go test ./... -count=1` from `backend/` and must pass with PostgreSQL stopped. Use `t.Setenv` throughout — never `os.Setenv` — and do not call `t.Parallel()` in this package.

1. **`backend/internal/config/config_test.go`** (unit, new) — **defaults.** With `APP_ENV=production` and `JWT_SECRET=x` set (so `.env` loading is skipped and validation passes), assert `Server.Port == "8080"`, `Server.ReadTimeout == 10*time.Second`, `Database.MaxConns == 10`, `Database.MinConns == 2`, `Database.PingTimeout == 2*time.Second`, and `Log.Level == slog.LevelInfo`.

2. **Overrides** (same file): set `PORT=9999`, `READ_TIMEOUT=45s`, `DB_MAX_CONNS=25`, `LOG_LEVEL=debug` and assert each lands, including `Log.Level == slog.LevelDebug`.

3. **Invalid port** (same file): `PORT=abc` and `PORT=70000` each make `Load()` return an error mentioning `PORT`. Table-driven.

4. **Invalid durations** (same file): `READ_TIMEOUT=nonsense`, `READ_TIMEOUT=0`, and `READ_TIMEOUT=-5s` each error and name `READ_TIMEOUT`.

5. **Pool bounds** (same file): `DB_MAX_CONNS=1` with `DB_MIN_CONNS=5` errors and names both variables; `DB_MAX_CONNS=0` errors.

6. **DSN precedence** (same file): with `DATABASE_URL=postgres://a:b@example.com:6000/x?sslmode=require` **and** conflicting `POSTGRES_*` values set, assert `Database.URL` is exactly the `DATABASE_URL` value. Then unset it and assert the composed DSN equals `postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable`.

7. **DSN escaping** (same file): `POSTGRES_PASSWORD=p@ss:w/rd` with no `DATABASE_URL`. Assert the result parses with `net/url.Parse` **and** that `pgxpool.ParseConfig` accepts it — the second assertion is what proves the escaping is correct rather than merely URL-shaped. Regression test for the password-with-special-characters failure mode.

8. **Production requires a secret** (same file): `APP_ENV=production` with `JWT_SECRET` empty errors and names `JWT_SECRET`; the same with `JWT_SECRET=abc` succeeds. With `APP_ENV=development` and no secret, `Load()` succeeds — a fresh clone must run with no setup.

9. **Unknown `APP_ENV`** (same file): `APP_ENV=staging` errors and names `APP_ENV`.

10. **Multiple errors at once** (same file): `PORT=abc` **and** `APP_ENV=staging` together — assert the single returned error message contains **both** `PORT` and `APP_ENV`. This is the assertion that pins the `errors.Join` behaviour.

11. **`backend/internal/repositories/postgres_test.go`** (modify existing). `NewPool` now takes `config.Database`; update both cases to build one literally (`config.Database{URL: ..., MaxConns: 10, MinConns: 2, ConnectTimeout: 5 * time.Second}`). Keep the malformed-DSN assertion on `"parse database dsn"` and the lazy-connection assertion, and keep checking that `MaxConns`/`MinConns`/`ConnectTimeout` reach `pool.Config()` — now proving config propagates rather than that constants exist.

12. **`backend/internal/services/health_test.go`** (modify existing). `NewHealthService` gained a `pingTimeout` parameter: pass `2*time.Second` in the existing cases and set `pingTimeout` on the struct literal in `TestHealthServiceCheck`. The timeout test passes a short value (`50*time.Millisecond`) and asserts `Check` returns within a second — faster and stricter than before.

13. **`backend/internal/handlers/health_test.go`** and **`backend/internal/server/server_test.go`** (modify existing). Update the constructor calls: handlers pass a `pingTimeout`; server builds a `&config.Config{Version: "test", Database: config.Database{PingTimeout: 2 * time.Second}}`. The four routing verdicts and the panic-logging assertion are unchanged and must still pass.

14. **Frontend build** (not a unit test): `npm run build` from `frontend/` must succeed, proving the `Environment` interface and both environment files compile. Only applicable once ZCRM-4's workspace is in place — which it is, since `ng new` has landed.

---

## Verification Steps

1. **Ignore rules still correct:** from the repo root, `git check-ignore -v .env` matches `.gitignore:11`, and `git check-ignore .env.example` exits **1** with no output (not ignored).
2. **Backend builds and is clean:** from `backend/`, `go build ./...`, `go vet ./...`, `gofmt -l .` (silent).
3. **No stray environment reads:** `grep -rn "os.Getenv" --include=*.go .` from `backend/` returns hits **only** under `internal/config/`. This is the structural check that the story actually landed.
4. **Tests pass with the database stopped:** `docker compose stop postgres`, then `go test ./... -count=1` — all green.
5. **Defaults work with no `.env`:** with no `.env` present, `go run ./cmd/api` logs `configuration loaded` with `env=development`, then serves `200` on `/health`.
6. **`.env` is honoured:** `cp .env.example .env`, set `PORT=8081` and `POSTGRES_PORT=55432` in it, then `docker compose up -d` and `go run ./cmd/api` — the container publishes `55432`, the API listens on `8081` (`curl http://localhost:8081/health` returns `200` with `"database":"up"`), and the DSN was composed from the file. **Delete the `.env` afterwards or leave it — it is untracked either way.**
7. **Real environment variables beat the file:** with `PORT=8081` in `.env`, run `PORT=8082 go run ./cmd/api` and confirm it listens on `8082`.
8. **Invalid config fails loudly:** `PORT=abc APP_ENV=staging go run ./cmd/api` exits non-zero and prints one `configuration error:` line naming **both** `PORT` and `APP_ENV`.
9. **No secret leaks into logs:** `go run ./cmd/api 2>&1 | grep -i -E "password|secret|postgres://"` produces **no** match.
10. **Frontend compiles:** from `frontend/`, `npm run build` exits 0.

---

## Done Criteria

- [ ] `backend/internal/config/` exists with `config.go`, `env.go`, `dotenv.go`, `validate.go`, and `Load() (*Config, error)`.
- [ ] `grep -rn "os.Getenv" --include=*.go` inside `backend/` matches only files under `internal/config/`.
- [ ] `cmd/api/main.go` has no `const version` and no `const` block of defaults; every value comes from `cfg`.
- [ ] `repositories.NewPool` takes `config.Database`; the `const` block in `postgres.go` is gone.
- [ ] `services.NewHealthService` takes a `pingTimeout time.Duration`; the `const pingTimeout` is gone; `services` does **not** import `config`.
- [ ] `server.New(log, cfg, db)` builds the health service from `cfg`.
- [ ] `.env.example` is committed at the repository root and documents every variable the loader reads, including `JWT_SECRET`.
- [ ] `git check-ignore -v .env` matches `.gitignore:11`; `git check-ignore .env.example` exits 1; **no `.env` is committed**.
- [ ] `DATABASE_URL` overrides the composed DSN; with it unset, the DSN is composed from `POSTGRES_*` and a password containing `@:/` round-trips through `pgxpool.ParseConfig`.
- [ ] Invalid configuration exits non-zero before the logger is built, and reports **all** problems in one message.
- [ ] `APP_ENV=production` with no `JWT_SECRET` fails; `APP_ENV=development` with no `JWT_SECRET` succeeds.
- [ ] No log line contains the DSN, the password, or `JWT_SECRET`.
- [ ] `frontend/src/environments/environment.ts` (production) and `environment.development.ts` both satisfy the `Environment` interface in `environment.model.ts`; **no `environment.prod.ts` exists**; `npm run build` passes.
- [ ] `go build ./...`, `go vet ./...`, `go test ./... -count=1` pass with PostgreSQL stopped; `gofmt -l .` silent.
- [ ] `backend/README.md` has a `## Configuration` section and no longer lists the config loader under "Not here yet"; root `README.md` documents `cp .env.example .env`.
- [ ] Committed on `feature/ZCRM-6-environment-config`, pushed, PR targets `develop`.
- [ ] `.squad/plans/project-setup-foundation/00-overview.md` contains the row for this story.

**STOP HERE. Report to the user and wait for confirmation before proceeding to Story 06 (ZCRM-7 — CORS & Frontend API Client).**
