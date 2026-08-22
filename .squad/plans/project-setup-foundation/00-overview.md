# project-setup-foundation — plan overview

Entry point for the **project-setup-foundation** feature. Stories execute in order by their `NN` prefix.

## Stories

| NN | File | Title | Tracker id | Depends on |
|----|------|-------|------------|------------|
| 01 | [01-story-initialize-repository-ZCRM-2.md](01-story-initialize-repository-ZCRM-2.md) | Initialize the Repository | ZCRM-2 | — |
| 02 | [02-story-backend-project-structure-ZCRM-3.md](02-story-backend-project-structure-ZCRM-3.md) | Backend Project Structure (Golang) | ZCRM-3 | 01 |
| 03 | [03-story-frontend-project-structure-ZCRM-4.md](03-story-frontend-project-structure-ZCRM-4.md) | Frontend Project Structure (Angular) | ZCRM-4 | 02 |

## Dependency notes

- Stories run strictly in tracker order: **ZCRM-2** (repo) → **ZCRM-3** (Go backend) → **ZCRM-4** (Angular frontend) → **ZCRM-5** (PostgreSQL + migrations) → **ZCRM-6** (environment config) → **ZCRM-7** (CORS + frontend API client).
- Story 01 establishes the shared contracts every later story depends on: the `backend/` and `frontend/` folder split, the `.gitignore` env rules (`.env` ignored, `.env.example` committed — required by ZCRM-6), the `.gitattributes` LF normalization, and the `main` (protected) / `develop` branch model.
- Story 02 fixes the module path as `github.com/ziad-azm/ZCRM/backend` and the router as **chi v5**. Every later backend story imports through that path, and ZCRM-7 mounts CORS into the middleware chain built in `backend/internal/server/server.go`.
- Story 02 leaves `internal/repositories/` deliberately empty (package doc only) and reads `PORT` with `os.Getenv` rather than a config loader. **ZCRM-5** fills the repository layer; **ZCRM-6** replaces the env read. Neither should be pulled forward.
- Story 03 scaffolds `frontend/` as an Angular 22 workspace (standalone, **zoneless**, SCSS, vitest) with Angular Material 22.1.3. It reaches the backend through the Angular dev-server proxy (`frontend/proxy.conf.json`: `/api` -> `http://localhost:8080`, prefix stripped) rather than CORS, so **ZCRM-7** still owns the real CORS middleware and the HTTP interceptor.
- Story 03 fixes two frontend contracts later stories build on: `ApiService` in `src/app/core/api.service.ts` reading `environment.apiBaseUrl`, and the `core/` / `shared/` / `features/` split.
- **Environment file naming for ZCRM-6:** Angular 22 generates `src/environments/environment.ts` (production default) plus `environment.development.ts` (dev override via `fileReplacements`). There is no `environment.prod.ts`; the ZCRM-6 intake's wording predates this layout.
- Story 01's open item is closed: `main` was pushed to `origin` during the Story 02 implementation session, so the `develop` → `main` release path exists.
- Go 1.26.7 was installed (via `winget install GoLang.Go`) during the Story 02 implementation session. `go.mod` targets `go 1.23`.
- No cross-feature dependencies. A second feature folder (`authentication-user-management`, ZCRM-9) has been scaffolded under `.squad/stories/` but is not yet planned and is not indexed.
