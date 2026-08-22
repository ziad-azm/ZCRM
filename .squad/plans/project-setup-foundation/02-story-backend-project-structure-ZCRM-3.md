# Story 02 — Backend Project Structure (Golang) (Story: ZCRM-3)

## Prerequisites

- Story 01 completed: [01-story-initialize-repository-ZCRM-2.md](01-story-initialize-repository-ZCRM-2.md). The monorepo exists at `e:\Work\AZM\ZCRM`, `origin` is `https://github.com/ziad-azm/ZCRM.git`, and `backend/` holds only a placeholder `README.md`.
- **Open item carried over from Story 01 — fix before opening this story's pull request.** `git ls-remote origin` currently returns only `refs/heads/develop`; **`main` was never pushed** and local `main` has no upstream (`git branch -vv` shows `main 6adffff` with no `[origin/...]`). A pull request from `feature/ZCRM-3-backend-structure` into `develop` works today, but the `develop` → `main` release path does not exist yet. Run `git push -u origin main` from a clean tree. If the remote rejects it because the `protect-main` ruleset is already active, temporarily set that ruleset's enforcement to **Disabled**, push, then set it back to **Active**.
- **Go must be installed — it is not on this machine.** `where.exe go` and `go version` both fail as of this plan. Install it before starting:

  ```powershell
  winget install --id GoLang.Go --accept-source-agreements --accept-package-agreements
  ```

  Then **open a new shell** (the installer edits `PATH`; existing shells keep the old one) and confirm `go version` prints `go1.23` or newer. Every command in this plan assumes `go` resolves.
- Network access to `proxy.golang.org` for `go get`. Behind a proxy that blocks it, set `GOPROXY=direct` before the `go get` in task 3.

---

## Story Goal

Scaffold the Go backend in `backend/` as a layered application — **handlers → services → repositories** — and end with an HTTP server that starts, serves `GET /health` with a `200`, logs every request as structured JSON, and shuts down cleanly on Ctrl+C without dropping in-flight requests.

Concretely, when this story is done:

1. `backend/` is a Go module named `github.com/ziad-azm/ZCRM/backend` with `chi v5` as the router.
2. The layered folders exist under `backend/internal/` and the health check flows through a real vertical slice: handler → service → response model. No layer reaches past its neighbour.
3. `GET /health` returns `200` and a JSON body; an unknown route returns `404`.
4. Logging goes through `log/slog` as JSON, with one line per request carrying method, path, status, duration, and request id.
5. `SIGINT`/`SIGTERM` triggers `http.Server.Shutdown` with a bounded grace period.

**Not in scope:** any database. `repositories/` ships as an empty package with a doc comment only — PostgreSQL, `pgxpool`, and the first real repository arrive in **ZCRM-5**. No full config loader (**ZCRM-6** owns that; this story reads only `PORT` via `os.Getenv` with a default). No CORS middleware and no auth (**ZCRM-7** and later). No Dockerfile, no CI workflow, no `Makefile` — `make` is not installed on this Windows machine, so every command in this plan is a plain `go` invocation.

---

## Context — Read These Files First

1. `backend/README.md` — all 9 lines. It is a placeholder that says *"Empty until ZCRM-3"* and lists what this story adds. Task 8 **replaces this file entirely**; do not leave the "Empty until" wording behind.
2. `README.md` (root) — read **lines 20–27** (the prerequisites table; the `Go | 1.23+` row is line 26 and carries the now-stale note *"not yet installed on the reference machine"*) and **line 37** (the paragraph declaring `backend/` and `frontend/` placeholders). Both are edited in task 9.
3. `README.md` (root) — read **lines 39–43** (`## Branching`). The convention is `feature/<TRACKER-ID>-<short-slug>` branched from `develop`; this story uses `feature/ZCRM-3-backend-structure`, the exact example given on line 43.
4. `.gitignore` — read **lines 17–33** (the Go block added by Story 01). `backend/bin/`, `*.exe`, `*.test`, `*.out`, `coverage.out`, `go.work` are already ignored, so a `go build -o backend/bin/api` produces nothing committable. **No `.gitignore` change is needed in this story** — confirm that rather than adding rules.
5. `.gitattributes` — read the whole file (18 lines). `* text=auto eol=lf` means every `.go` file is stored LF. `gofmt` output is therefore stable across machines; **do not** add a Go-specific attributes entry.
6. [01-story-initialize-repository-ZCRM-2.md](01-story-initialize-repository-ZCRM-2.md) — read `## Edge Cases & Failure Modes`. The Windows-path-casing rule applies here too: all Go package directories are **lowercase**.
7. `.squad/stories/project-setup-foundation/ZCRM-3/intake.md` — the source work item. Its four tracker tasks map onto tasks 2+3 (`go mod init` + router), 4 (layered folders + `cmd/`), 5 (`GET /health`), and 6+7 (structured logging + graceful shutdown).
8. Run `git ls-remote origin` and confirm the state described in Prerequisites before branching.

There is exactly one prior plan in this folder — Story 01. Match its shape: numbered tasks that each name a file, a mandatory edge-case section, and verification steps that are literal commands.

---

## Implementation tasks

No frontend changes required — `frontend/` is untouched by this story and stays a placeholder until ZCRM-4.

Run every command from `e:\Work\AZM\ZCRM\backend` unless the command block says otherwise.

### 1 — Branch from `develop`

From the repository root:

```bash
cd e:/Work/AZM/ZCRM
git checkout develop
git pull
git checkout -b feature/ZCRM-3-backend-structure
```

Confirm the tree is clean first. Two untracked directories may be present from a separate feature's scaffolding — `.squad/plans/authentication-user-management/` and `.squad/stories/authentication-user-management/`. **Leave them untracked.** They belong to ZCRM-9 and must not enter this story's commits.

### 2 — Initialize the Go module

```bash
cd e:/Work/AZM/ZCRM/backend
go mod init github.com/ziad-azm/ZCRM/backend
```

The module path **must** be `github.com/ziad-azm/ZCRM/backend` — it matches the `origin` remote (`https://github.com/ziad-azm/ZCRM.git`) plus the monorepo subfolder, so `go get` resolves for anyone outside the repo and every internal import below is correct as written. **Do not** use a bare `zcrm` or `backend` module path; that breaks importability and forces a rewrite later.

This creates `backend/go.mod` with a `go` directive matching the installed toolchain. If that line reads `go 1.24` or higher, **edit it down to `go 1.23`** so the module matches the minimum in the root README prerequisites table (line 26). Leave any `toolchain` line the command adds.

### 3 — Add the router: chi v5

```bash
go get github.com/go-chi/chi/v5@latest
```

**Router decision: `chi` v5** — the intake offers chi/gin/echo and this is the choice. Reasons that matter for this codebase:

- chi handlers **are** `http.HandlerFunc`. Handlers, middleware, and tests use only `net/http` and `httptest`, so the layered packages carry no framework types and swapping routers later touches only `internal/server`.
- Its middleware set (`RequestID`, `Recoverer`, `WrapResponseWriter`) is exactly what task 6's structured logging needs, without pulling a logging framework.
- gin and echo both wrap request/response in their own context types, which would leak a framework dependency into handler signatures.

`go get @latest` resolves the current v5 patch release and records it in `go.mod`/`go.sum`. **Do not** hand-write a version number into `go.mod`. Record the resolved version in `backend/README.md` in task 8 — read it from `go.mod` after this command.

### 4 — Create the layered package tree

Create exactly these directories and files. All names lowercase.

```text
backend/
├── cmd/
│   └── api/
│       └── main.go             # entrypoint: config, logger, server, shutdown
├── internal/
│   ├── handlers/
│   │   └── health.go           # HTTP layer: decode/encode only
│   ├── services/
│   │   └── health.go           # business layer: uptime, status
│   ├── repositories/
│   │   └── doc.go              # package doc only — no code until ZCRM-5
│   ├── models/
│   │   └── health.go           # response DTO
│   ├── middleware/
│   │   └── logging.go          # slog request logger
│   └── server/
│       └── server.go           # router construction + middleware chain
├── go.mod
├── go.sum
└── README.md
```

`internal/` is deliberate: Go's `internal` rule makes these packages unimportable from outside `github.com/ziad-azm/ZCRM/backend`, so the layering cannot be bypassed by a future package elsewhere in the monorepo. The intake's four folder names (`handlers`, `services`, `repositories`, `models`) all appear under it, plus `middleware` and `server` for wiring.

**The dependency direction is one-way and enforced by review:** `cmd` → `server` → `handlers` → `services` → `repositories`, with `models` importable by any layer. **`services` must not import `handlers`, and no layer under `internal/` may import `net/http` except `handlers`, `middleware`, and `server`.**

**Create file: `backend/internal/repositories/doc.go`**

```go
// Package repositories holds data-access implementations.
//
// It is intentionally empty until ZCRM-5 (Database & Migrations), which adds
// the pgxpool connection and the first concrete repository. Services depend on
// interfaces declared here; nothing in this package imports net/http.
package repositories
```

### 5 — The health vertical slice

**Create file: `backend/internal/models/health.go`**

```go
package models

import "time"

// HealthResponse is the JSON body returned by GET /health.
type HealthResponse struct {
	Status        string    `json:"status"`
	Version       string    `json:"version"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
}
```

**Create file: `backend/internal/services/health.go`**

The service owns the answer to "am I healthy"; it holds no HTTP types.

```go
package services

import (
	"time"

	"github.com/ziad-azm/ZCRM/backend/internal/models"
)

// HealthService reports process liveness.
type HealthService struct {
	startedAt time.Time
	version   string
	now       func() time.Time // injected for tests
}

// NewHealthService returns a HealthService that treats now as the process start.
func NewHealthService(version string) *HealthService {
	return &HealthService{startedAt: time.Now(), version: version, now: time.Now}
}

// Check returns the current health snapshot. It never fails: this story has no
// dependencies to probe. ZCRM-5 extends it to ping the database.
func (s *HealthService) Check() models.HealthResponse {
	now := s.now()
	return models.HealthResponse{
		Status:        "ok",
		Version:       s.version,
		UptimeSeconds: now.Sub(s.startedAt).Seconds(),
		Timestamp:     now.UTC(),
	}
}
```

The `now func() time.Time` field is **required**, not decoration — the service test in Test Plan step 2 substitutes it to assert uptime deterministically.

**Create file: `backend/internal/handlers/health.go`**

The handler does HTTP and nothing else: no time arithmetic, no status logic.

```go
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

// HealthHandler serves GET /health.
type HealthHandler struct {
	svc *services.HealthService
	log *slog.Logger
}

// NewHealthHandler wires the handler to its service.
func NewHealthHandler(svc *services.HealthService, log *slog.Logger) *HealthHandler {
	return &HealthHandler{svc: svc, log: log}
}

// Get responds 200 with the health snapshot as JSON.
func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := h.svc.Check()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Status is already written; log and return rather than double-writing.
		h.log.ErrorContext(r.Context(), "encode health response", slog.Any("error", err))
	}
}
```

**Set `Content-Type` and call `WriteHeader` before `Encode`** — writing the body first makes Go infer `text/plain` and locks the status to `200` implicitly.

### 6 — Structured logging with `log/slog`

Use the standard library's `log/slog` (Go 1.21+). **Do not** add zap, logrus, or zerolog — a JSON handler and a middleware are all this story needs, and a dependency here would spread through every later package.

**Create file: `backend/internal/middleware/logging.go`**

```go
// Package middleware holds HTTP middleware shared by all routes.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// RequestLogger logs one structured line per request after it completes.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				log.InfoContext(r.Context(), "http request",
					slog.String("request_id", chimw.GetReqID(r.Context())),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Duration("duration", time.Since(start)),
					slog.String("remote_addr", r.RemoteAddr),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
```

The log call sits in a **`defer`** on purpose: it still runs when a downstream handler panics, so `chi`'s `Recoverer` produces both a `500` and a log line for it.

**Create file: `backend/internal/server/server.go`**

```go
// Package server builds the HTTP router and its middleware chain.
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ziad-azm/ZCRM/backend/internal/handlers"
	"github.com/ziad-azm/ZCRM/backend/internal/middleware"
	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

// New returns the application router with all routes and middleware mounted.
func New(log *slog.Logger, version string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimw.Recoverer)

	health := handlers.NewHealthHandler(services.NewHealthService(version), log)
	r.Get("/health", health.Get)

	return r
}
```

**Middleware order is load-bearing.** `RequestID` must precede `RequestLogger` or `GetReqID` returns `""`. `RequestLogger` must precede `Recoverer` so a panic still produces a log line with a `500` status. **Do not reorder these four.**

CORS is **not** mounted here — ZCRM-7 adds it, and it belongs above `RequestID` in the chain when it arrives.

### 7 — Entrypoint with graceful shutdown

**Create file: `backend/cmd/api/main.go`**

```go
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

	"github.com/ziad-azm/ZCRM/backend/internal/server"
)

// version is the reported build version. ZCRM-6 replaces this with config, and
// a later story injects it at build time via -ldflags.
const version = "0.1.0"

const (
	defaultPort     = "8080"
	shutdownTimeout = 10 * time.Second
	readTimeout     = 10 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)

	// Minimal env read only. The full config loader is ZCRM-6.
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      server.New(log, version),
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
```

Three details that must survive review:

- **`errors.Is(err, http.ErrServerClosed)`** — `Shutdown` always makes `ListenAndServe` return that error. Treating it as a failure turns every clean stop into `exit 1`.
- **`serverErr` is buffered (capacity 1)** — an unbuffered channel leaks the goroutine when `main` returns via the `ctx.Done()` branch.
- **The `stop()` call inside `ctx.Done()`** — without it, a developer holding Ctrl+C during a slow drain cannot interrupt, because the signal stays captured.

### 8 — Rewrite `backend/README.md`

**File: `backend/README.md`** — replace all 9 lines. The current text says *"Empty until ZCRM-3"*, which this story makes false.

Required content:

1. `# Backend (Go)` and one line naming the module path `github.com/ziad-azm/ZCRM/backend` and the router (chi v5, with the resolved version read from `go.mod` after task 3).
2. **`## Layout`** — a fenced tree matching task 4, with the one-way dependency rule `cmd → server → handlers → services → repositories` stated as a line of prose beneath it.
3. **`## Run`**:

   ```bash
   cd backend
   go run ./cmd/api          # listens on :8080
   PORT=9000 go run ./cmd/api
   ```

   On PowerShell the env-var form is `$env:PORT="9000"; go run ./cmd/api` — include both, since the reference machine is Windows.
4. **`## Test`** — `go test ./...`, plus `go vet ./...` and `gofmt -l .` (expected output: nothing).
5. **`## Endpoints`** — a table with one row: `GET /health` → `200` and the exact JSON shape from `models.HealthResponse`.
6. **`## Not here yet`** — one line each for the database (ZCRM-5), config loader (ZCRM-6), and CORS (ZCRM-7), so the next reader does not add them here by accident.

### 9 — Update the root `README.md`

**File: `README.md`**

- **Line 26** — the `Go | 1.23+` row still reads *"not yet installed on the reference machine; required from ZCRM-3 onward"*. Replace that note with `backend build and tests`.
- **Line 37** — the paragraph declaring both folders placeholders is now half wrong. Replace it with a sentence that `backend/` is a runnable Go service (pointing at `backend/README.md` for its commands) and that `frontend/` remains a placeholder until **ZCRM-4**.
- Add a **`### Backend`** subsection under `## Getting started` with the three-line quickstart:

  ```bash
  cd backend
  go run ./cmd/api
  curl http://localhost:8080/health
  ```

Leave `## Repository layout` (lines 9–17), `## Branching`, `## Commit messages`, and `## Planning workflow` unchanged — the tree already lists `backend/`, and its `(scaffolded in ZCRM-3)` comment is what this story satisfies.

### 10 — Commit

```bash
cd e:/Work/AZM/ZCRM
git add backend README.md
git status --short
```

Expected additions: `backend/go.mod`, `backend/go.sum`, `backend/README.md` (modified), the eight `.go` files from tasks 4–7, the three test files from the Test Plan, and a modified root `README.md`. **`backend/bin/`, `*.exe`, and `coverage.out` must not appear** — they are ignored by `.gitignore` lines 19–33. The two `authentication-user-management` directories must not appear either.

```bash
git commit -m "feat(backend): scaffold layered Go service with health endpoint (ZCRM-3)"
git push -u origin feature/ZCRM-3-backend-structure
```

Open the pull request into **`develop`**, not `main`.

---

## Edge Cases & Failure Modes

- **Wrong module path.** `go mod init zcrm` (or `backend`) compiles fine locally, then every import line in tasks 5–7 fails to resolve and the fix is a rewrite of all eight files. Enforced by the explicit path in task 2. Recovery: `go mod edit -module github.com/ziad-azm/ZCRM/backend` followed by a find-and-replace across `backend/**/*.go`.
- **`go` not on PATH after install.** `winget` edits the machine `PATH`, but shells opened before the install keep the stale copy, so `go version` still fails and every task appears blocked. Enforced by the "open a new shell" instruction in Prerequisites.
- **`go.mod` `go` directive newer than the README minimum.** A Go 1.26 toolchain writes `go 1.26`, and a developer on 1.23 then gets `go.mod requires go >= 1.26` on `go build`. Enforced by the edit-down step in task 2. Recovery: `go mod edit -go=1.23`.
- **Panic in a handler produces no log line.** If `RequestLogger` is mounted *after* `Recoverer`, the recover happens below the logger and the request vanishes from the logs — the worst possible failure mode for a health endpoint. Enforced by the fixed middleware order in task 6 and asserted by Test Plan step 5.
- **Empty `request_id` in every log line.** Mounting `RequestLogger` before `chimw.RequestID` makes `GetReqID` return `""` silently — no error, just useless logs. Enforced by the fixed order in task 6.
- **Clean shutdown reported as a crash.** `srv.Shutdown` causes `ListenAndServe` to return `http.ErrServerClosed`; without the `errors.Is` guard in task 7 the process logs `server failed` and exits `1`, which will later make a container orchestrator restart-loop the pod. Asserted by Verification step 5.
- **Port already in use.** `ListenAndServe` returns `listen tcp :8080: bind: Only one usage of each socket address...` on Windows. The `serverErr` branch logs `server failed` and exits `1` — correct and intended. Recovery: `PORT=9000 go run ./cmd/api`. Note that Angular's dev server (ZCRM-4) defaults to `:4200`, so it will not collide.
- **`GET /health` with a trailing slash.** `chi` treats `/health/` as a distinct pattern; the route registered in task 6 is `/health` only, so `/health/` returns `404`. This is the accepted behaviour for this story — **do not** add `middleware.StripSlashes` to paper over it. Asserted by Test Plan step 4.
- **Wrong method on `/health`.** `r.Get` registers `GET` (and `HEAD`) only. `POST /health` returns `405 Method Not Allowed` from chi's own handler, not `404`. Asserted by Test Plan step 4.
- **Unicode or huge `PORT` value.** `os.Getenv("PORT")` is copied into `Addr` unvalidated, so `PORT=abc` fails at `ListenAndServe` with `unknown port` and exits `1` — a loud, immediate failure, which is acceptable here. Port **validation** belongs to the ZCRM-6 config loader; note it in a comment rather than adding validation now.
- **CRLF sneaking into `.go` files.** `.gitattributes` normalizes to LF on staging, but an editor writing CRLF plus `gofmt` can produce a confusing diff. Asserted by Verification step 3 (`gofmt -l .` prints nothing).

---

## Test Plan

All test files live beside the code they test, per Go convention. Run everything with `go test ./...` from `backend/`.

1. **Create file: `backend/internal/handlers/health_test.go`** (unit). Table-free, three assertions using `net/http/httptest`: build a `HealthHandler` with `services.NewHealthService("test")` and `slog.New(slog.NewJSONHandler(io.Discard, nil))`, call `Get` with `httptest.NewRequest(http.MethodGet, "/health", nil)` and a `httptest.NewRecorder()`, then assert (a) status is `200`, (b) `Content-Type` is `application/json`, (c) the body decodes into `models.HealthResponse` with `Status == "ok"` and `Version == "test"`.

2. **Create file: `backend/internal/services/health_test.go`** (unit). Construct `HealthService` directly (same package, so unexported fields are reachable) with `startedAt` fixed and `now` returning `startedAt.Add(90 * time.Second)`. Assert `Check().UptimeSeconds == 90` exactly and `Status == "ok"`. This is the test the injected `now` field exists for — **no `time.Sleep`**.

3. **Create file: `backend/internal/server/server_test.go`** (integration). Build the full router via `server.New(logger, "test")` and drive it with `httptest.NewServer`. Assert `GET /health` → `200`.

4. **Routing negatives** (same file as step 3): `GET /nope` → `404`; `GET /health/` → `404`; `POST /health` → `405`. These three lock in the trailing-slash and method behaviour called out in Edge Cases.

5. **Panic-logging test** (same file as step 3): mount a temporary route that panics onto a router built the same way as `server.New`, point the logger at a `bytes.Buffer` via `slog.NewJSONHandler`, issue a request, then assert the response is `500` **and** the buffer contains one `"http request"` line with `"status":500`. This is the regression test for the middleware-order failure mode.

6. **Manual smoke** — not automated: `go run ./cmd/api`, then `curl -i http://localhost:8080/health` in a second shell. Confirm `200`, a JSON body with a non-zero `uptime_seconds`, and exactly one JSON log line on the server's stdout containing `"method":"GET"` and `"status":200`.

No test touches `internal/repositories` — it has no code until ZCRM-5.

---

## Verification Steps

Run from `e:\Work\AZM\ZCRM\backend`.

1. **Dependencies resolve:** `go mod tidy` — exits 0, and `git diff --stat go.mod go.sum` afterwards shows no unexpected additions beyond chi.
2. **Backend builds:** `go build ./...` — no output, exit 0.
3. **Static checks clean:** `go vet ./...` prints nothing, and `gofmt -l .` prints nothing (any listed file is unformatted — run `gofmt -w .`).
4. **Unit tests pass:** `go test ./... -count=1` — all packages `ok` or `no test files`; Test Plan steps 1–5 green.
5. **Server runs and stops cleanly:** `go run ./cmd/api`, confirm the `server starting` JSON line on stdout, then press **Ctrl+C**. Expect `shutdown signal received` followed by `server stopped cleanly`, and an exit code of `0` (`echo $LASTEXITCODE` in PowerShell). A non-zero exit or a `server failed` line means the `http.ErrServerClosed` guard is wrong.
6. **Endpoint answers:** with the server running, `curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health` prints `200`, and `curl -s http://localhost:8080/health` returns JSON containing `"status":"ok"`.
7. **Regression — repo hygiene from Story 01 still holds:** from the repo root, `git status --short` lists no build artifacts, and `git check-ignore backend/bin/api` exits 0.

---

## Done Criteria

- [ ] `backend/go.mod` declares module `github.com/ziad-azm/ZCRM/backend` with a `go` directive of `1.23`, and `backend/go.sum` records `github.com/go-chi/chi/v5`.
- [ ] The package tree from task 4 exists exactly, all-lowercase, with `internal/repositories/doc.go` containing only a package comment.
- [ ] `GET /health` returns `200` with `Content-Type: application/json` and a body carrying `status`, `version`, `uptime_seconds`, and `timestamp`.
- [ ] `GET /nope` → `404`, `GET /health/` → `404`, `POST /health` → `405`.
- [ ] Every request produces exactly one `log/slog` JSON line with `request_id`, `method`, `path`, `status`, `bytes`, `duration`, and `remote_addr`; a panicking route still logs its line with `"status":500`.
- [ ] Ctrl+C logs `shutdown signal received` then `server stopped cleanly` and exits `0`.
- [ ] `PORT=9000 go run ./cmd/api` listens on `:9000`; unset `PORT` listens on `:8080`.
- [ ] `go build ./...`, `go vet ./...`, `go test ./... -count=1` all pass, and `gofmt -l .` prints nothing.
- [ ] No layer violation: `internal/services` and `internal/models` import neither `net/http` nor `internal/handlers` (check with `go list -deps ./internal/services`).
- [ ] `backend/README.md` no longer contains the words "Empty until", and documents Layout, Run, Test, Endpoints, and Not-here-yet.
- [ ] Root `README.md` line 26 no longer claims Go is uninstalled, and line 37 no longer calls `backend/` a placeholder.
- [ ] Work is committed on `feature/ZCRM-3-backend-structure`, pushed, and a pull request targets `develop`.
- [ ] `.squad/plans/project-setup-foundation/00-overview.md` contains the row for this story.

**STOP HERE. Report to the user and wait for confirmation before proceeding to Story 03 (ZCRM-4 — Frontend Project Structure).**
