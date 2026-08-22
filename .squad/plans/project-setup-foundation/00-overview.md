# project-setup-foundation — plan overview

Entry point for the **project-setup-foundation** feature. Stories execute in order by their `NN` prefix.

## Stories

| NN | File | Title | Tracker id | Depends on |
|----|------|-------|------------|------------|
| 01 | [01-story-initialize-repository-ZCRM-2.md](01-story-initialize-repository-ZCRM-2.md) | Initialize the Repository | ZCRM-2 | — |
| 02 | [02-story-backend-project-structure-ZCRM-3.md](02-story-backend-project-structure-ZCRM-3.md) | Backend Project Structure (Golang) | ZCRM-3 | 01 |

## Dependency notes

- Stories run strictly in tracker order: **ZCRM-2** (repo) → **ZCRM-3** (Go backend) → **ZCRM-4** (Angular frontend) → **ZCRM-5** (PostgreSQL + migrations) → **ZCRM-6** (environment config) → **ZCRM-7** (CORS + frontend API client).
- Story 01 establishes the shared contracts every later story depends on: the `backend/` and `frontend/` folder split, the `.gitignore` env rules (`.env` ignored, `.env.example` committed — required by ZCRM-6), the `.gitattributes` LF normalization, and the `main` (protected) / `develop` branch model.
- Story 02 fixes the module path as `github.com/ziad-azm/ZCRM/backend` and the router as **chi v5**. Every later backend story imports through that path, and ZCRM-7 mounts CORS into the middleware chain built in `backend/internal/server/server.go`.
- Story 02 leaves `internal/repositories/` deliberately empty (package doc only) and reads `PORT` with `os.Getenv` rather than a config loader. **ZCRM-5** fills the repository layer; **ZCRM-6** replaces the env read. Neither should be pulled forward.
- **Open item from Story 01, blocking the `develop` → `main` release path:** `git ls-remote origin` returns only `refs/heads/develop`. `main` was never pushed and has no upstream. Push it before the first release PR — see the Prerequisites section of Story 02.
- Go is still not installed on the reference machine. Story 02 carries the `winget install GoLang.Go` step in its Prerequisites; nothing in ZCRM-3 onward runs without it.
- No cross-feature dependencies. A second feature folder (`authentication-user-management`, ZCRM-9) has been scaffolded under `.squad/stories/` but is not yet planned and is not indexed.
