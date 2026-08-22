# project-setup-foundation — plan overview

Entry point for the **project-setup-foundation** feature. Stories execute in order by their `NN` prefix.

## Stories

| NN | File | Title | Tracker id | Depends on |
|----|------|-------|------------|------------|
| 01 | [01-story-initialize-repository-ZCRM-2.md](01-story-initialize-repository-ZCRM-2.md) | Initialize the Repository | ZCRM-2 | — |

## Dependency notes

- Stories run strictly in tracker order: **ZCRM-2** (repo) → **ZCRM-3** (Go backend) → **ZCRM-4** (Angular frontend) → **ZCRM-5** (PostgreSQL + migrations) → **ZCRM-6** (environment config) → **ZCRM-7** (CORS + frontend API client).
- Story 01 establishes the shared contracts every later story depends on: the `backend/` and `frontend/` folder split, the `.gitignore` env rules (`.env` ignored, `.env.example` committed — required by ZCRM-6), the `.gitattributes` LF normalization, and the `main` (protected) / `develop` branch model.
- Story 01 does **not** install Go. `go` is absent from PATH on the reference machine and must be installed before ZCRM-3 starts.
- No cross-feature dependencies — this is currently the only feature folder.
