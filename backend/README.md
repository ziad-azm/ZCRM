# Backend (Go)

Golang API service. Module `github.com/ziad-azm/ZCRM/backend`, routed by [chi](https://github.com/go-chi/chi) **v5.3.2**, talking to PostgreSQL through [pgx](https://github.com/jackc/pgx) **v5.10.0**. Requires **Go 1.25+** (`go.mod` targets `go 1.25.0` — pgx v5.10.0 declares that floor; developed against toolchain 1.26.7).

## Layout

```text
backend/
├── cmd/
│   └── api/
│       └── main.go             # entrypoint: logger, server, graceful shutdown
├── internal/
│   ├── handlers/               # HTTP layer: decode/encode only
│   ├── services/               # business layer
│   ├── repositories/           # data access: postgres.go holds the pgxpool
│   ├── models/                 # response DTOs
│   ├── middleware/             # slog request logger
│   └── server/                 # router construction + middleware chain
├── migrations/                 # golang-migrate SQL, applied in version order
├── scripts/                    # migrate.ps1 / migrate.sh wrappers
├── go.mod
├── go.sum
└── README.md
```

Dependencies run one way only: `cmd → server → handlers → services → repositories`, with `models` importable by any layer. A service must never import a handler, and nothing under `internal/` imports `net/http` except `handlers`, `middleware`, and `server`. `internal/` is not importable from outside this module, so the layering cannot be bypassed.

The middleware chain in `internal/server/server.go` is order-sensitive: `RequestID → RealIP → RequestLogger → Recoverer`. `RequestID` must precede the logger or `request_id` logs empty, and the logger must precede `Recoverer` or a panicking route produces no log line.

## Run

```bash
cd backend
go run ./cmd/api          # listens on :8080
PORT=9000 go run ./cmd/api
```

On PowerShell, set the environment variable separately:

```powershell
cd backend
$env:PORT="9000"; go run ./cmd/api
```

Stop with **Ctrl+C** — the server drains in-flight requests for up to 10 seconds and logs `server stopped cleanly`.

## Test

```bash
go test ./... -count=1
go vet ./...
gofmt -l .                # expected output: nothing
```

`go test -race` requires cgo and a C toolchain, which is not installed on the reference Windows machine.

## Database

PostgreSQL 18 runs in Docker. From the **repository root**:

```bash
docker compose up -d          # starts postgres:18-alpine
docker compose ps             # wait for STATUS = healthy
docker compose down           # stop, keeping data
docker compose down -v        # stop and DESTROY the data volume
```

Default DSN, matching the compose defaults:

```text
postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable
```

Override it with `DATABASE_URL`; the API and both migrate scripts read that variable.

**If port 5432 is already taken** — a locally installed PostgreSQL service is the usual cause — pick another host port and point the DSN at it:

```bash
POSTGRES_PORT=55432 docker compose up -d
DATABASE_URL="postgres://zcrm:zcrm@localhost:55432/zcrm?sslmode=disable" go run ./cmd/api
```

The API starts even when the database is unreachable: it logs `database unreachable at startup` and `/health` reports `"database":"down"`. The pool reconnects on its own once the database returns — no API restart needed.

Note that the data lives in the named volume `zcrm-postgres-data`, mounted at `/var/lib/postgresql` (**not** `.../data` — PostgreSQL 18+ images store data in a major-version subdirectory and refuse to start if the older mount point is used). `docker compose down` keeps that volume; only `down -v` deletes it. Changing `POSTGRES_PASSWORD` after the volume exists has no effect, since those variables are read only by `initdb`.

## Migrations

Managed by [golang-migrate](https://github.com/golang-migrate/migrate) **v4.19.1**. Install the CLI once — the `postgres` build tag is **required**, or every command fails with `unknown driver postgres`:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
```

It installs into `$(go env GOPATH)/bin` — `C:\Users\<you>\go\bin` on Windows — which must be on `PATH`.

```bash
cd backend
./scripts/migrate.sh up        # apply everything pending
./scripts/migrate.sh version   # current version
./scripts/migrate.sh down 1    # roll back one migration
```

On PowerShell use `./scripts/migrate.ps1 up` — same arguments. Both honour `DATABASE_URL` and fall back to the default DSN above.

Files live in `migrations/` and **must** be named `{version}_{title}.up.sql` and `{version}_{title}.down.sql`. Every migration ships a `down` that actually reverses it: always test `up → down → up` before opening a pull request.

If a migration fails halfway, golang-migrate marks the schema **dirty** and refuses to continue (`Dirty database version N. Fix and force version.`). Fix the SQL, then `migrate -path migrations -database "$DATABASE_URL" force <N-1>` and run `up` again.

## Endpoints

| Method | Path | Status | Body |
| ------ | ---- | ------ | ---- |
| `GET` | `/health` | `200` | `{"status":"ok","version":"0.1.0","uptime_seconds":0.07,"timestamp":"2026-08-23T12:10:01Z","database":"up"}` |

`database` is `"up"` or `"down"`. **The status code stays `200` either way and `status` stays `"ok"`** — `/health` is a *liveness* probe reporting that the process answers, not a readiness probe. A `503` on a failed dependency is deliberately not implemented.

`GET /health/` (trailing slash) returns `404` and `POST /health` returns `405` — chi treats them as distinct from the registered route, and that is intentional.

Every request emits one JSON log line carrying `request_id`, `method`, `path`, `status`, `bytes`, `duration`, and `remote_addr`. On Windows, `duration` reads `0` for sub-millisecond requests: the platform's monotonic clock granularity is coarser than the request itself.

## Not here yet

- **Config loader** — `main.go` reads `PORT` and `DATABASE_URL` with `os.Getenv` and nothing else, and the dev DSN is duplicated in three places (`docker-compose.yml` defaults, `cmd/api/main.go`, `scripts/migrate.*`). **ZCRM-6** replaces that with a real loader and removes the duplication.
- **Domain tables** — the only migration installs the shared `set_updated_at()` trigger function. Every real table (including `users`, which belongs to **ZCRM-9**) arrives with its own feature story.
- **CORS** — **ZCRM-7** adds the middleware; it mounts above `RequestID` in the chain.
