# ZCRM

ZCRM is a monorepo: the Angular frontend and the Golang backend live side by side in one repository, sharing tooling, versioning, and a single commit history, while staying cleanly separated in their own top-level folders. One clone gives you the whole system — there is no second repository to find, and a change that spans both layers is one pull request.

## Repository layout

```text
ZCRM/
├── backend/        # Golang API service          (scaffolded in ZCRM-3)
├── frontend/       # Angular SPA                 (scaffolded in ZCRM-4)
├── .claude/        # Claude Code slash commands
├── .squad/         # squad-kit stories & plans
├── .gitattributes  # line-ending normalization
├── .gitignore
└── README.md
```

## Prerequisites

| Tool | Minimum version | Used for |
| ---- | --------------- | -------- |
| Git | 2.54+ | version control (2.54.0 in use) |
| Node.js | 24.x | frontend toolchain (v24.15.0 in use) |
| npm | 11.x | frontend package management (11.12.1 in use) |
| Angular CLI | 22.x | `ng` commands for `frontend/` (22.0.6 in use) |
| Go | 1.25+ | backend build and tests (1.26.7 in use; pgx v5.10.0 sets the 1.25 floor) |
| Docker Desktop | latest | local PostgreSQL via `docker compose up -d` (20.10.22 / Compose v2.15.1 in use) |

## Getting started

```bash
git clone <repo-url> ZCRM
cd ZCRM
git checkout develop
```

`backend/` is a runnable Go service — see [backend/README.md](backend/README.md) for its layout, commands, and endpoints. `frontend/` is still a placeholder holding only a `README.md` until **ZCRM-4** (Angular frontend) scaffolds it.

### Backend

```bash
cd backend
go run ./cmd/api
curl http://localhost:8080/health
```

## Branching

- **`main`** — protected and release-ready. **Direct pushes to `main` are rejected by a branch ruleset — `main` only advances through a pull request from `develop`.**
- **`develop`** — the integration branch. Day-to-day work starts here and merges back here.
- **`feature/<TRACKER-ID>-<short-slug>`** — one branch per story, branched from `develop` and merged into `develop` via pull request. Example: `feature/ZCRM-3-backend-structure`.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/) with the tracker id in the subject:

```text
feat(backend): add health endpoint (ZCRM-3)
fix(frontend): correct API base URL in production build (ZCRM-7)
chore: initialize ZCRM monorepo (ZCRM-2)
```

## Planning workflow

- `.squad/README.md` — the intake → plan → implement loop this project uses (squad-kit).
- `.squad/plans/00-index.md` — index of every planned feature and its story sequence.
- Jira board: <https://ziadhosny007.atlassian.net>
