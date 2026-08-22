# Backend (Go)

Golang API service. Empty until **ZCRM-3 — Backend Project Structure (Golang)**, which adds:

- `go.mod` (via `go mod init`) and the HTTP router
- the layered folders `handlers/`, `services/`, `repositories/`, `models/`
- the `cmd/` entrypoint and a `GET /health` endpoint

Do not add Go source here before that story — the module path and router choice are decided there.
