# Story 04 — Database & Migrations (PostgreSQL) (Story: ZCRM-5)

## Prerequisites

- Story 02 completed: [02-story-backend-project-structure-ZCRM-3.md](02-story-backend-project-structure-ZCRM-3.md). This story extends `backend/internal/repositories`, `internal/services/health.go`, `internal/server/server.go`, and `cmd/api/main.go`.
- **Branch state at planning time — read before branching.** `feature/ZCRM-3-backend-structure` (ZCRM-3) is pushed but **unmerged**, and `feature/ZCRM-4-frontend-structure` (ZCRM-4) is **mid-flight**: only `ng new` has landed there (commit `87dbae5`); its tasks 4–13 are unfinished. This story touches **no frontend file**, so branch it from `develop` once the ZCRM-3 PR merges, or stack it on `feature/ZCRM-3-backend-structure`. **Do not branch it from `feature/ZCRM-4-frontend-structure`** — that would tangle the unfinished Angular work into a backend PR.
- Tooling verified present: `go1.26.7`, `docker 20.10.22`, `docker compose v2.15.1`.
- **The `migrate` CLI is not installed.** Task 5 installs it with `go install`; it lands in `C:\Users\user\go\bin` (verified `GOPATH`), which must be on `PATH`.
- Port `5432` must be free. A local PostgreSQL service already listening there will make `docker compose up` fail with a port-binding error — see Edge Cases for the override.

---

## Story Goal

Give the backend a real database and a versioned way to change its schema.

When this story is done:

1. `docker compose up -d` starts PostgreSQL 18 with a named volume, and reports healthy via its own healthcheck.
2. The Go service holds a **pooled** `pgxpool.Pool` built from `DATABASE_URL`, with explicit connection limits and lifetimes.
3. `migrate` applies and rolls back schema changes from `backend/migrations/`, and the first migration exists and round-trips cleanly.
4. `GET /health` reports database reachability as a new `database` field, without changing its status code.
5. Scripts wrap the migrate invocations so nobody has to remember a DSN.

**Not in scope:** any domain table. The first migration installs a shared `set_updated_at()` trigger function and nothing else — `contacts`, `accounts`, `users`, and every other CRM table belongs to its own feature story, and `users` specifically belongs to the authentication feature (**ZCRM-9**). No repository for a domain entity, so `repositories/` gains the pool and a `Pinger` interface only. No config loader — `DATABASE_URL` is read with `os.Getenv` and a default, exactly as `PORT` is today; **ZCRM-6** owns replacing both. No `.env` / `.env.example` file (**ZCRM-6** owns those). No CI, no production deployment, no connection retry/backoff loop.

---

## Context — Read These Files First

1. `backend/internal/repositories/doc.go` — all 6 lines. The package comment promises exactly what this story delivers: *"the pgxpool connection and the first concrete repository. Services depend on interfaces declared here."* Task 3 rewrites this comment, since the package stops being empty.
2. `backend/internal/services/health.go` — **lines 9–31**. Note `HealthService` fields (`startedAt`, `version`, `now`) on lines 10–14, the `NewHealthService(version string)` signature on line 17, and the comment on **line 22**: *"ZCRM-5 extends it to ping the database."* That is this story's instruction. `Check()` on lines 23–31 becomes context-aware.
3. `backend/internal/models/health.go` — **lines 6–11**, the `HealthResponse` struct. Task 4 adds one field; the existing four keys (`status`, `version`, `uptime_seconds`, `timestamp`) **must not** change name or type.
4. `backend/internal/server/server.go` — **lines 17–29**. `New(log *slog.Logger, version string)` on line 17 constructs the health service inline on **line 25**. That signature changes in task 6, so `server_test.go` changes with it.
5. `backend/cmd/api/main.go` — **lines 18–22** (the `version` and `defaultPort` constants), **lines 36–38** (the `os.Getenv("PORT")` pattern this story mirrors for `DATABASE_URL`), **line 43** (`Handler: server.New(log, version)` — the call site that gains an argument), and **lines 70–78** (the shutdown branch, where the pool must be closed).
6. `.gitignore` — **lines 10–15** (env block: `.env` ignored, `.env.example` negated) and **lines 17–33** (Go block). PostgreSQL data lives in a **named Docker volume**, not a bind mount, so there is nothing new to ignore. **No `.gitignore` change is needed in this story.**
7. `.gitattributes` — `*.ps1` is `text eol=crlf` and everything else normalizes to LF. The PowerShell script in task 7 therefore lands CRLF and the `.sh` sibling LF, which is correct for both.
8. [02-story-backend-project-structure-ZCRM-3.md](02-story-backend-project-structure-ZCRM-3.md) — match its shape and its layering rule: `cmd → server → handlers → services → repositories`, `models` importable anywhere.

**Verified dependency versions** (resolved against the module proxy during planning — do not hand-write different ones):

- `github.com/jackc/pgx/v5` **v5.10.0**
- `github.com/golang-migrate/migrate/v4` **v4.19.1**
- Docker image `postgres:18-alpine` (tag confirmed present on Docker Hub)

---

## Implementation tasks

**No frontend changes in this story.** See task 10 for the one-line follow-up ZCRM-4 must absorb when it resumes.

Run Go commands from `e:\Work\AZM\ZCRM\backend` and `docker compose` commands from `e:\Work\AZM\ZCRM`.

### 1 — Branch

```bash
cd e:/Work/AZM/ZCRM
git checkout develop && git pull          # after the ZCRM-3 PR merges
git checkout -b feature/ZCRM-5-database-migrations
```

If the ZCRM-3 PR is still open, branch from `feature/ZCRM-3-backend-structure` instead and merge the PRs in tracker order. **Never** branch from `feature/ZCRM-4-frontend-structure`.

### 2 — PostgreSQL in Docker Compose

**Create file: `docker-compose.yml`** (repository **root**, not `backend/` — it is monorepo infrastructure, and ZCRM-6 will add more services beside it).

```yaml
services:
  postgres:
    image: postgres:18-alpine
    container_name: zcrm-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-zcrm}
      POSTGRES_USER: ${POSTGRES_USER:-zcrm}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-zcrm}
      # Keeps initdb deterministic across developer machines.
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=C"
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
    volumes:
      - zcrm-postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-zcrm} -d ${POSTGRES_DB:-zcrm}"]
      interval: 5s
      timeout: 5s
      retries: 10
      start_period: 10s

volumes:
  zcrm-postgres-data:
```

Three deliberate choices:

- **`${VAR:-default}` everywhere.** The file works with **no `.env` present at all**, which is the state of the repo today. When ZCRM-6 adds `.env`, Compose picks it up automatically from the same directory and the defaults become overrides — no edit to this file.
- **Named volume, not a bind mount.** A bind mount under the repo would put database files inside the working tree, where `.gitignore` has no rule for them and Windows file permissions break `initdb`.
- **A real healthcheck.** `depends_on: service_healthy` in later stories needs it, and `docker compose ps` becomes a truthful readiness signal instead of "container started".

Start it and confirm:

```bash
cd e:/Work/AZM/ZCRM
docker compose up -d
docker compose ps
```

`docker compose ps` must show `zcrm-postgres` as `healthy` (not merely `running`) before continuing.

### 3 — The pooled connection

```bash
cd e:/Work/AZM/ZCRM/backend
go get github.com/jackc/pgx/v5@v5.10.0
```

**File: `backend/internal/repositories/doc.go`** — replace the package comment; the package is no longer empty:

```go
// Package repositories holds data-access implementations.
//
// It owns the PostgreSQL connection pool (postgres.go) and the interfaces that
// services depend on. Concrete per-entity repositories are added by the feature
// stories that need them. Nothing in this package imports net/http.
package repositories
```

**Create file: `backend/internal/repositories/postgres.go`**

```go
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pinger reports whether a data store is reachable. Services depend on this
// interface rather than on *pgxpool.Pool so they stay testable without a
// database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Pool configuration. These are deliberate defaults for a single API instance;
// ZCRM-6 makes them configurable.
const (
	maxConns          = 10
	minConns          = 2
	maxConnLifetime   = time.Hour
	maxConnIdleTime   = 30 * time.Minute
	healthCheckPeriod = time.Minute
	connectTimeout    = 5 * time.Second
)

// NewPool parses dsn and returns a configured connection pool.
//
// It does NOT establish a connection: pgxpool connects lazily, so a database
// that is down produces no error here and surfaces on the first query or Ping.
// That is intentional — the API must still start and report an unhealthy
// database rather than refusing to boot.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.HealthCheckPeriod = healthCheckPeriod
	cfg.ConnConfig.ConnectTimeout = connectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	return pool, nil
}
```

`*pgxpool.Pool` already has a `Ping(context.Context) error` method, so it satisfies `Pinger` with no adapter.

### 4 — Health reports the database

**File: `backend/internal/models/health.go`** — add one field. **Do not** rename or retype the existing four:

```go
package models

import "time"

// Database reachability states reported by GET /health.
const (
	DatabaseUp   = "up"
	DatabaseDown = "down"
)

// HealthResponse is the JSON body returned by GET /health.
type HealthResponse struct {
	Status        string    `json:"status"`
	Version       string    `json:"version"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
	Database      string    `json:"database"`
}
```

**File: `backend/internal/services/health.go`** — the service takes a `Pinger` and `Check` becomes context-aware:

```go
package services

import (
	"context"
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/models"
	"github.com/ziad-azm/ZCRM/backend/internal/repositories"
)

// pingTimeout bounds the database probe so a hung database cannot hang /health.
const pingTimeout = 2 * time.Second

// HealthService reports process liveness and dependency reachability.
type HealthService struct {
	startedAt time.Time
	version   string
	now       func() time.Time // injected for tests
	db        repositories.Pinger
}

// NewHealthService returns a HealthService that treats now as the process start.
// db may be nil, in which case the database is reported as down.
func NewHealthService(version string, db repositories.Pinger) *HealthService {
	return &HealthService{startedAt: time.Now(), version: version, now: time.Now, db: db}
}

// Check returns the current health snapshot. It always succeeds: an unreachable
// database is reported in the Database field, not as an error.
func (s *HealthService) Check(ctx context.Context) models.HealthResponse {
	now := s.now()

	return models.HealthResponse{
		Status:        "ok",
		Version:       s.version,
		UptimeSeconds: now.Sub(s.startedAt).Seconds(),
		Timestamp:     now.UTC(),
		Database:      s.databaseStatus(ctx),
	}
}

func (s *HealthService) databaseStatus(ctx context.Context) string {
	if s.db == nil {
		return models.DatabaseDown
	}

	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		return models.DatabaseDown
	}

	return models.DatabaseUp
}
```

**`Status` stays `"ok"` and the handler keeps returning `200` even when the database is down.** `/health` is a **liveness** probe: the process is answering. Readiness semantics — a `503` when a dependency is down — are deliberately **out of scope**; introducing them here would break the `200` assertions in `handlers/health_test.go` and `server_test.go` and the ZCRM-4 dashboard. Note it in a comment and leave it.

**File: `backend/internal/handlers/health.go`** — pass the request context into the now-context-aware call:

```go
resp := h.svc.Check(r.Context())
```

That is the only change in this file.

### 5 — golang-migrate and the first migration

Install the CLI with the PostgreSQL driver compiled in:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
migrate -version
```

`-tags 'postgres'` is **required** — without it the binary builds with no database drivers and every command fails with `unknown driver postgres`. The binary lands in `C:\Users\user\go\bin`; add that to `PATH` if `migrate -version` is not found.

Create `backend/migrations/` with the first pair. golang-migrate requires the exact `{version}_{title}.{up|down}.sql` naming.

**Create file: `backend/migrations/000001_init.up.sql`**

```sql
-- Shared trigger function: every table with an updated_at column attaches this
-- so the application never has to set the timestamp by hand.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

**Create file: `backend/migrations/000001_init.down.sql`**

```sql
DROP FUNCTION IF EXISTS set_updated_at();
```

This migration deliberately creates **no table**. It exists to prove the migration workflow end-to-end (apply, record version, roll back cleanly) and to install the one piece of schema every future table will reuse. Inventing a domain table here would pre-empt the feature stories that own them.

`gen_random_uuid()` is built into PostgreSQL 13+, so **no `pgcrypto` extension is needed** — do not add one.

### 6 — Wire the pool through the server

**File: `backend/internal/server/server.go`** — `New` accepts the pinger and passes it to the health service. Line 17's signature and line 25's construction both change:

```go
// New returns the application router with all routes and middleware mounted.
func New(log *slog.Logger, version string, db repositories.Pinger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimw.Recoverer)

	health := handlers.NewHealthHandler(services.NewHealthService(version, db), log)
	r.Get("/health", health.Get)

	return r
}
```

Add `"github.com/ziad-azm/ZCRM/backend/internal/repositories"` to the import block. **The middleware order is unchanged and still load-bearing** — do not touch lines 20–23.

**File: `backend/cmd/api/main.go`** — three edits.

Add the DSN constant beside `defaultPort` (line 21):

```go
const (
	defaultPort        = "8080"
	defaultDatabaseURL = "postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable"
	shutdownTimeout    = 10 * time.Second
	readTimeout        = 10 * time.Second
	writeTimeout       = 15 * time.Second
	idleTimeout        = 60 * time.Second
)
```

After the `PORT` read (lines 36–38), build the pool and probe it once:

```go
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
```

Then pass the pool at line 43: `Handler: server.New(log, version, pool),`.

`defer pool.Close()` in `main` runs after either `select` branch returns, so the pool closes on both clean shutdown and startup failure. **Do not** put `pool.Close()` inside the `ctx.Done()` branch only — the `serverErr` path would leak it. Note that `os.Exit(1)` skips deferred calls; that is acceptable in the DSN-parse branch because no pool exists yet.

Add `"github.com/ziad-azm/ZCRM/backend/internal/repositories"` to main's imports. `context` is already imported.

### 7 — Migration scripts

Wrap the DSN so nobody types it. Both scripts take the migrate subcommand through to the CLI.

**Create file: `backend/scripts/migrate.ps1`**

```powershell
#Requires -Version 5.1
<#
.SYNOPSIS
  Runs golang-migrate against the local ZCRM database.
.EXAMPLE
  ./scripts/migrate.ps1 up
  ./scripts/migrate.ps1 down 1
  ./scripts/migrate.ps1 version
#>
param(
    [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
    [string[]]$MigrateArgs
)

$ErrorActionPreference = 'Stop'

if ($env:DATABASE_URL) { $dsn = $env:DATABASE_URL }
else { $dsn = 'postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable' }

$migrationsDir = Join-Path $PSScriptRoot '..\migrations'

& migrate -path $migrationsDir -database $dsn @MigrateArgs
exit $LASTEXITCODE
```

**Create file: `backend/scripts/migrate.sh`**

```bash
#!/usr/bin/env bash
# Runs golang-migrate against the local ZCRM database.
#   ./scripts/migrate.sh up
#   ./scripts/migrate.sh down 1
set -euo pipefail

DSN="${DATABASE_URL:-postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable}"
MIGRATIONS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../migrations" && pwd)"

exec migrate -path "$MIGRATIONS_DIR" -database "$DSN" "$@"
```

Both honour an existing `DATABASE_URL` and fall back to the same default DSN as `main.go`. Keep the three DSN copies (compose defaults, `main.go`, scripts) in sync until ZCRM-6 centralises them — that duplication is a known, temporary cost, and ZCRM-6's job is to remove it.

### 8 — Apply and roll back

```bash
cd e:/Work/AZM/ZCRM/backend
./scripts/migrate.sh up
./scripts/migrate.sh version      # expect: 1
./scripts/migrate.sh down 1       # answer y at the confirmation prompt
./scripts/migrate.sh up
```

The round-trip is the point: a `down` that fails leaves the schema unrecoverable without manual SQL.

### 9 — Update `backend/README.md`

**File: `backend/README.md`** — the current file has a **`## Not here yet`** section whose first bullet says the database lands in ZCRM-5. That bullet is now false.

- Add a **`## Database`** section: `docker compose up -d` from the repo root, the `postgres:18-alpine` service, the default DSN, and the `DATABASE_URL` override.
- Add a **`## Migrations`** section: the `go install -tags 'postgres' …@v4.19.1` line, `./scripts/migrate.sh up`, `down 1`, `version`, the `{version}_{title}.{up|down}.sql` naming rule, and the requirement that every migration ships a working `down`.
- Update the `GET /health` row in **`## Endpoints`** to include the new `database` field, and state that the status code stays `200` when the database is down (liveness, not readiness).
- In **`## Not here yet`**, delete the database bullet and keep the config-loader (ZCRM-6) and CORS (ZCRM-7) bullets.
- Add `internal/repositories/postgres.go`, `migrations/`, and `scripts/` to the **`## Layout`** tree.

### 10 — Follow-up for ZCRM-4 (do not implement here)

`GET /health` now returns a fifth field. The frontend `Health` interface that ZCRM-4 defines at `frontend/src/app/core/models/health.ts` must gain `database: string`. ZCRM-4 is unfinished, so **record this in the ZCRM-5 PR description** and leave the frontend untouched. Adding a JSON field is backward compatible — an older TypeScript interface simply ignores it — so nothing breaks in the meantime.

### 11 — Commit

```bash
cd e:/Work/AZM/ZCRM
git add backend docker-compose.yml
git status --short
```

Expected: new `docker-compose.yml`, `backend/internal/repositories/postgres.go`, `backend/migrations/000001_init.{up,down}.sql`, `backend/scripts/migrate.{ps1,sh}`, and modified `backend/go.mod`, `go.sum`, `internal/repositories/doc.go`, `internal/models/health.go`, `internal/services/health.go`, `internal/handlers/health.go`, `internal/server/server.go`, `cmd/api/main.go`, `backend/README.md`, plus the test files from the Test Plan. **No `frontend/` path may appear.**

```bash
git commit -m "feat(backend): add PostgreSQL pool and golang-migrate workflow (ZCRM-5)"
git push -u origin feature/ZCRM-5-database-migrations
```

Open the pull request into **`develop`**.

---

## Edge Cases & Failure Modes

- **Port 5432 already bound.** A locally installed PostgreSQL makes `docker compose up -d` fail with `Bind for 0.0.0.0:5432 failed: port is already allocated`. The compose file reads `${POSTGRES_PORT:-5432}`, so recovery is `POSTGRES_PORT=55432 docker compose up -d` plus a matching `DATABASE_URL`. Enforced by the parameterised port in task 2.
- **`migrate` built without the postgres driver.** Omitting `-tags 'postgres'` in task 5 produces a binary that fails every command with `unknown driver postgres` — a confusing error, since the CLI itself runs fine. Enforced by the explicit tag.
- **`migrate` not on `PATH`.** `go install` writes to `C:\Users\user\go\bin`; if that is not on `PATH`, both scripts fail with "command not found" while `go install` reported success. Enforced by the `migrate -version` check in task 5.
- **Migrating before Postgres is healthy.** `docker compose up -d` returns as soon as the container starts, but `initdb` runs for a few seconds on a fresh volume, so an immediate `migrate up` fails with `connection refused`. Enforced by the "must show `healthy`" gate in task 2.
- **Dirty migration state.** A migration that fails halfway leaves `schema_migrations.dirty = true`, and every later `migrate` call refuses with `Dirty database version 1. Fix and force version.` Recovery: fix the SQL, then `migrate -path ... -database ... force 0` followed by `up`. This is the single most common golang-migrate trap — the `down` round-trip in task 8 is what catches a broken migration before it reaches anyone else.
- **Database down at startup.** `pgxpool.New` connects lazily, so a stopped container produces **no** startup error. The eager probe in task 6 logs `database unreachable at startup` as a warning and the API keeps serving with `/health` reporting `"database":"down"`. This is deliberate: failing fast would make the ZCRM-4 dashboard undemonstrable whenever Docker is not running.
- **Malformed `DATABASE_URL`.** `pgxpool.ParseConfig` fails immediately and task 6 exits `1` with a logged error — a configuration mistake, not a transient outage, so a loud crash is correct. Note that `os.Exit(1)` skips `defer`; no pool exists at that point, so nothing leaks.
- **Hung database hanging `/health`.** Without `pingTimeout`, a database accepting TCP but never answering would block the health request until the client gave up, and a load balancer would read that as a dead process. Enforced by the 2-second `context.WithTimeout` in `databaseStatus`, and by `ConnectTimeout` on the pool.
- **`pool.Close()` in the wrong place.** Putting it in the `ctx.Done()` branch only leaks connections on the `serverErr` path. Enforced by `defer pool.Close()` in `main` immediately after the pool is built.
- **Named volume survives `docker compose down`.** `down` removes containers but **keeps** `zcrm-postgres-data`, so a "clean" restart still has the old schema and a stale `schema_migrations` row. To truly reset: `docker compose down -v`. Do not add `-v` to any documented default command — it destroys data.
- **`POSTGRES_*` env vars are read only by `initdb`.** Changing `POSTGRES_PASSWORD` after the volume exists has **no effect**; the old password persists and connections fail with `password authentication failed`. Recovery is `docker compose down -v` and recreate.
- **Windows line endings in `migrate.sh`.** `.gitattributes` normalizes `.sh` to LF, so the script runs inside Git Bash and containers. `.ps1` is pinned to CRLF by the same file. No action needed — but do not add a `.sh` exception.
- **Test suite requires no database.** Every test in the Test Plan uses a fake `Pinger`. If a test ever needs real PostgreSQL it must be build-tagged and skipped by default, or `go test ./...` stops working for anyone without Docker running.

---

## Test Plan

All tests run with `go test ./... -count=1` from `backend/` and **must pass with PostgreSQL stopped** — the database is faked at the `Pinger` boundary.

1. **`backend/internal/services/health_test.go`** (unit, modify existing). The current test constructs `HealthService` with `startedAt`/`version`/`now` and calls `Check()` with no argument; it will not compile after task 4. Update it to pass a fake pinger and `context.Background()`, and keep the existing exact `UptimeSeconds == 90` assertion.

2. **Database-up case** (same file): a fake whose `Ping` returns `nil` → `Check(ctx).Database == models.DatabaseUp` and `Status == "ok"`.

3. **Database-down case** (same file): a fake whose `Ping` returns `errors.New("boom")` → `Database == models.DatabaseDown`, and `Status` is still `"ok"`. This is the assertion that pins the liveness-not-readiness decision.

4. **Nil-pinger case** (same file): `NewHealthService("test", nil)` → `Database == models.DatabaseDown` with no panic. Guards the `s.db == nil` branch.

5. **Ping timeout case** (same file): a fake whose `Ping` blocks until its context is cancelled, returning `ctx.Err()` → `Database == models.DatabaseDown` and `Check` returns in well under `pingTimeout + 1s`. Assert with a `time.Now()` delta around the call. Regression test for the hung-database failure mode.

6. **`backend/internal/handlers/health_test.go`** (unit, modify existing). `NewHealthService` gained a parameter, so the constructor call needs the fake. Add an assertion that the decoded body's `Database` field equals `"up"`, and keep the existing `200` and `application/json` assertions.

7. **`backend/internal/server/server_test.go`** (integration, modify existing). Every `server.New(logger, "test")` call becomes `server.New(logger, "test", fake)`. The four routing verdicts (`/health` → `200`, `/nope` → `404`, `/health/` → `404`, `POST /health` → `405`) and the panic-logging test are unchanged and must still pass.

8. **`backend/internal/repositories/postgres_test.go`** (unit, new). No database: assert `NewPool(ctx, "not-a-dsn")` returns an error mentioning `parse database dsn`, and that `NewPool(ctx, "postgres://u:p@localhost:5432/db?sslmode=disable")` returns a non-nil pool and **no** error even with nothing listening — the explicit regression test for lazy connection behaviour. Close the pool in the success case.

9. **Migration round-trip** (manual, not automated): task 8. With the container healthy, `up` → `version` reports `1` → `down 1` → `up`. Then `docker compose exec postgres psql -U zcrm -d zcrm -c "\df set_updated_at"` lists the function, confirming the migration applied to the real database rather than only recording a version row.

---

## Verification Steps

1. **Database is healthy:** from `e:\Work\AZM\ZCRM`, `docker compose up -d` then `docker compose ps` — `zcrm-postgres` shows `healthy`.
2. **Backend builds and is clean:** from `backend/`, `go build ./...`, `go vet ./...`, and `gofmt -l .` (no output).
3. **Tests pass without a database:** `docker compose stop postgres`, then `go test ./... -count=1` — all green. Restart with `docker compose start postgres`.
4. **Migrations apply and roll back:** Test Plan step 9 passes, including the `psql \df set_updated_at` check.
5. **Health reports the database up:** `go run ./cmd/api` logs `database connected`; `curl -s http://localhost:8080/health` returns `"database":"up"` and the status code is `200`.
6. **Health reports the database down:** with the API still running, `docker compose stop postgres`, then `curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health` still prints **`200`** while the body shows `"database":"down"`. Restart Postgres and confirm the body flips back to `"up"` without restarting the API — that is the proof the pool reconnects on its own.
7. **Shutdown still clean:** Ctrl+C the API — `shutdown signal received` then `server stopped cleanly`, exit `0`, no pool errors logged.
8. **Regression — frontend untouched:** `git diff --stat -- frontend` is empty.

---

## Done Criteria

- [ ] `docker-compose.yml` at the repo root runs `postgres:18-alpine` with a named volume `zcrm-postgres-data`, a `pg_isready` healthcheck, and `${VAR:-default}` values that work with no `.env` file present.
- [ ] `docker compose ps` reports `zcrm-postgres` as `healthy`.
- [ ] `backend/internal/repositories/postgres.go` exports `NewPool(ctx, dsn) (*pgxpool.Pool, error)` and the `Pinger` interface, with `MaxConns`, `MinConns`, lifetimes, and `ConnectTimeout` all set explicitly.
- [ ] `doc.go`'s package comment no longer says the package is empty.
- [ ] `go.mod` requires `github.com/jackc/pgx/v5 v5.10.0`.
- [ ] `models.HealthResponse` has a fifth field `database`; the original four JSON keys are unchanged.
- [ ] `HealthService` takes a `repositories.Pinger`, `Check` takes a `context.Context`, and the probe is bounded by a 2-second timeout.
- [ ] `GET /health` returns **`200`** with `"database":"up"` when Postgres is up and **`200`** with `"database":"down"` when it is stopped.
- [ ] `backend/migrations/000001_init.{up,down}.sql` exist; `up` creates `set_updated_at()` and `down` drops it.
- [ ] `migrate` (installed with `-tags 'postgres'`) runs `up`, `version` → `1`, `down 1`, and `up` again without leaving a dirty state.
- [ ] `backend/scripts/migrate.ps1` and `migrate.sh` both work and both honour `DATABASE_URL`.
- [ ] `go build ./...`, `go vet ./...`, `go test ./... -count=1` all pass **with PostgreSQL stopped**, and `gofmt -l .` is silent.
- [ ] `backend/README.md` documents Database and Migrations, updates the `/health` row, and no longer lists the database under "Not here yet".
- [ ] No file under `frontend/` is modified; the `Health` TypeScript interface follow-up is recorded in the PR description.
- [ ] Committed on `feature/ZCRM-5-database-migrations`, pushed, PR targets `develop`.
- [ ] `.squad/plans/project-setup-foundation/00-overview.md` contains the row for this story.

**STOP HERE. Report to the user and wait for confirmation before proceeding to Story 05 (ZCRM-6 — Environment Configuration).**
