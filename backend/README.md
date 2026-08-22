# Backend (Go)

Golang API service. Module `github.com/ziad-azm/ZCRM/backend`, routed by [chi](https://github.com/go-chi/chi) **v5.3.2**. Requires **Go 1.23+** (`go.mod` targets `go 1.23`; developed against toolchain 1.26.7).

## Layout

```text
backend/
├── cmd/
│   └── api/
│       └── main.go             # entrypoint: logger, server, graceful shutdown
├── internal/
│   ├── handlers/               # HTTP layer: decode/encode only
│   ├── services/               # business layer
│   ├── repositories/           # data access — empty until ZCRM-5
│   ├── models/                 # response DTOs
│   ├── middleware/             # slog request logger
│   └── server/                 # router construction + middleware chain
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

## Endpoints

| Method | Path | Status | Body |
| ------ | ---- | ------ | ---- |
| `GET` | `/health` | `200` | `{"status":"ok","version":"0.1.0","uptime_seconds":0.07,"timestamp":"2026-08-22T20:39:41Z"}` |

`GET /health/` (trailing slash) returns `404` and `POST /health` returns `405` — chi treats them as distinct from the registered route, and that is intentional.

Every request emits one JSON log line carrying `request_id`, `method`, `path`, `status`, `bytes`, `duration`, and `remote_addr`. On Windows, `duration` reads `0` for sub-millisecond requests: the platform's monotonic clock granularity is coarser than the request itself.

## Not here yet

- **Database** — PostgreSQL, `pgxpool`, and the first concrete repository land in **ZCRM-5**. `internal/repositories` holds only a package doc until then.
- **Config loader** — `main.go` reads `PORT` with `os.Getenv` and nothing else. **ZCRM-6** replaces that with a real loader, and owns validating the value.
- **CORS** — **ZCRM-7** adds the middleware; it mounts above `RequestID` in the chain.
