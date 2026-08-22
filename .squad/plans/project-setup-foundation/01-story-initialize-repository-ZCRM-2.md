# Story 01 — Initialize the Repository (Story: ZCRM-2)

## Prerequisites

- None. This is the first story in the **project-setup-foundation** feature and the first story in the global sequence.
- Tooling verified present on this machine: `git 2.54.0.windows.1`, `node v24.15.0`, `npm 11.12.1`, `@angular/cli 22.0.6`.
- Tooling **not** present and **not** required by this story: `go` (needed from Story 02 / ZCRM-3) and the GitHub CLI `gh`. Because `gh` is missing, the remote repository is created through the GitHub web UI and wired up with `git remote add` — **do not** use `gh repo create`.
- A GitHub account with permission to create a repository under the target owner, and permission to edit repository rulesets (needed for **protect main**).

---

## Story Goal

Turn the existing local folder `e:\Work\AZM\ZCRM` — which today holds only `.squad/`, `.claude/`, and a squad-kit-managed `.gitignore` — into a real monorepo published on GitHub, with:

1. A tracked git history on branch **main**, plus a **develop** branch, both pushed to `origin`.
2. Top-level `backend/` and `frontend/` folders that exist in git (each with a placeholder `README.md`) so Stories 02 and 03 have a defined home and nothing has to be re-negotiated later.
3. A root `README.md` complete enough that a developer can clone the repo and understand the layout, prerequisites, and branch workflow **without asking anyone**.
4. A `.gitignore` covering Go, Node/Angular, and environment files, appended **below** the existing squad-kit block.
5. **main** protected on the remote so it cannot be pushed to directly or force-pushed.

**Not in scope:** running `go mod init`, `ng new`, or creating any application source. No Go module, no Angular workspace, no `docker-compose.yml`, no CI workflow, no `.env` loader. `backend/` and `frontend/` end this story containing only a placeholder `README.md` each. Those folders are filled by ZCRM-3 (backend), ZCRM-4 (frontend), ZCRM-5 (database), ZCRM-6 (env config), and ZCRM-7 (CORS / API client).

---

## Context — Read These Files First

1. `.gitignore` — the **entire** file is 8 lines / 201 bytes today. Lines 1–8 are a squad-kit managed block delimited by `# Managed by squad-kit — do not edit this block` (line 1) and `# End squad-kit block` (line 8). **Every new ignore rule goes on line 9 and below.** Do not reorder, reformat, or re-indent lines 1–8; squad-kit rewrites that block on upgrade and will clobber edits made inside it.
2. `.squad/config.yaml` — read `project.projectRoots` (`- .`, i.e. the monorepo root is the repo root, not a subfolder) and `tracker.workspace` (`ziadhosny007.atlassian.net`, used for the tracker link in the README).
3. `.squad/README.md` — the squad-kit workflow section (Intake → Plan → Implement). The root `README.md` written in task 5 links here instead of restating it.
4. `.squad/stories/project-setup-foundation/ZCRM-2/intake.md` — the source work item. Its four tracker tasks map onto tasks 1, 2+3, 4, and 6+7 below.
5. `.squad/plans/project-setup-foundation/00-overview.md` — the feature overview table referenced in task 8.
6. Run `git status` in `e:\Work\AZM\ZCRM` and confirm it prints `fatal: not a git repository`. If it prints anything else, a repo already exists — **stop and report** rather than re-initializing over it.
7. Run `git config --global init.defaultBranch` and confirm it prints nothing (unset). This is why task 1 uses `git init -b main` explicitly; without `-b main`, git 2.54 creates `master` and emits a hint.

There is no prior plan in `.squad/plans/` to follow for precedent — this file is the first, and it sets the pattern for Stories 02–06.

---

## Implementation tasks

No backend code changes required (there is no backend yet). No frontend code changes required (there is no Angular workspace yet). This story is repository plumbing only.

Run every command below from `e:\Work\AZM\ZCRM` unless stated otherwise.

### 1 — Initialize the local repository on `main`

```bash
cd e:/Work/AZM/ZCRM
git init -b main
```

`-b main` is **mandatory**: `init.defaultBranch` is unset globally, so plain `git init` produces `master`.

Set the commit identity **for this repository** so commits are not attributed to the wrong address. The global identity is `Ziad Hosny <ziad.hosny@azm.sa>`; the working address for this project is `ziadhosny@azmsquad.com`:

```bash
git config user.name "Ziad Hosny"
git config user.email "ziadhosny@azmsquad.com"
git config core.autocrlf false
```

`core.autocrlf false` leaves the working tree untouched and lets `.gitattributes` (task 4) decide what lands in the index.

**Do not commit yet.** Committing before task 3 lands the ignore rules would track junk that later has to be `git rm --cached`-ed out.

### 2 — Create the `backend/` and `frontend/` folders

Git cannot track an empty directory, so each folder gets a real placeholder file. Use a `README.md` rather than a `.gitkeep` — it does the same job and tells the next developer what the folder is for.

**Create file: `backend/README.md`**

```markdown
# Backend (Go)

Golang API service. Empty until **ZCRM-3 — Backend Project Structure (Golang)**, which adds:

- `go.mod` (via `go mod init`) and the HTTP router
- the layered folders `handlers/`, `services/`, `repositories/`, `models/`
- the `cmd/` entrypoint and a `GET /health` endpoint

Do not add Go source here before that story — the module path and router choice are decided there.
```

**Create file: `frontend/README.md`**

```markdown
# Frontend (Angular)

Angular single-page application. Empty until **ZCRM-4 — Frontend Project Structure (Angular)**, which adds:

- the Angular workspace (via `ng new`) and the default route
- the `core/`, `shared/`, and `features/` folders
- the UI component library and its theme
- the base `ApiService`

Do not add Angular source here before that story — the workspace name and UI library are decided there.
```

**Do not** create `.gitkeep` files in addition to these READMEs.

### 3 — Extend `.gitignore` for Go, Node/Angular, and env files

**File: `.gitignore`**

Append the block below starting at **line 9**, immediately after `# End squad-kit block`. Leave lines 1–8 byte-for-byte unchanged.

```gitignore

# ---------- Environment & secrets ----------
.env
.env.*
!.env.example
*.pem
*.key

# ---------- Go (backend/) ----------
# Compiled binaries and test artifacts
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
backend/bin/
backend/tmp/
# Go workspace files (per-developer, never shared)
go.work
go.work.sum
# Coverage
coverage.out
coverage.html

# ---------- Node / Angular (frontend/) ----------
node_modules/
frontend/dist/
frontend/.angular/
frontend/tmp/
frontend/out-tsc/
npm-debug.log*
yarn-error.log*
pnpm-debug.log*
.npm/
frontend/coverage/
testem.log

# ---------- Editors & OS ----------
.idea/
.vs/
*.swp
*.swo
.DS_Store
Thumbs.db
desktop.ini
```

Three rules that matter and must not be "simplified" away:

- **`.env.*` followed by `!.env.example`** — the negation must come **after** the wildcard, or `.env.example` stays ignored, and ZCRM-6 depends on that file being committed.
- **`.vscode/` is deliberately NOT ignored.** Shared editor settings for a Go + Angular monorepo are worth committing. `.idea/` and `.vs/` (per-developer IDE state) are ignored.
- **`node_modules/` is unanchored on purpose** so it matches at any depth (root, `frontend/`, and any future package).

### 4 — Add `.gitattributes` to pin line endings

**Create file: `.gitattributes`**

This file is **not** named in the tracker tasks. It is added deliberately: development happens on Windows 11 while Go tooling, shell scripts, and Docker builds assume LF. Without it, the first `go fmt` or `ng generate` in Stories 02–03 produces a diff that is pure CRLF noise, and any `backend/scripts/*.sh` file breaks inside a container.

```gitattributes
# Normalize all text files to LF in the repository
* text=auto eol=lf

# Windows-only files keep CRLF in the working tree
*.bat  text eol=crlf
*.cmd  text eol=crlf
*.ps1  text eol=crlf

# Binaries — never touch
*.png   binary
*.jpg   binary
*.jpeg  binary
*.gif   binary
*.ico   binary
*.pdf   binary
*.woff  binary
*.woff2 binary
```

### 5 — Write the root `README.md`

**Create file: `README.md`**

The bar is the story's own wording: *"any developer can clone and understand from the README alone."* Write real, runnable content for what exists **today**, and mark what does not exist yet with its tracker id — **do not** document commands that only work after a later story.

Required sections, in this order:

1. **`# ZCRM`** — one paragraph: a monorepo holding the Angular frontend and the Golang backend side by side, sharing tooling and versioning while staying cleanly separated.
2. **`## Repository layout`** — this fenced tree:

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

3. **`## Prerequisites`** — a table of tool, minimum version, and purpose. Pin the versions actually in use: **Git 2.54+**, **Node.js 24.x** (v24.15.0 in use), **npm 11.x** (11.12.1 in use), **Angular CLI 22.x** (22.0.6 in use), **Go 1.23+** (*not yet installed on the reference machine — required from ZCRM-3 onward*), **Docker Desktop** (*required from ZCRM-5 onward for PostgreSQL*).
4. **`## Getting started`** — the clone-to-ready path that works today:

```bash
git clone <repo-url> ZCRM
cd ZCRM
git checkout develop
```

   Follow it with one line stating plainly that `backend/` and `frontend/` are placeholders until ZCRM-3 and ZCRM-4, and that per-folder setup steps are added to this README by those stories.
5. **`## Branching`** — `main` is protected and release-ready; `develop` is the integration branch; feature branches use `feature/<TRACKER-ID>-<short-slug>` (e.g. `feature/ZCRM-3-backend-structure`) and merge into `develop` via pull request. State explicitly: **direct pushes to `main` are rejected by a branch ruleset — `main` only advances through a pull request from `develop`.**
6. **`## Commit messages`** — Conventional Commits with the tracker id in the subject: `feat(backend): add health endpoint (ZCRM-3)`.
7. **`## Planning workflow`** — three lines pointing at `.squad/README.md` for the intake → plan → implement loop, `.squad/plans/00-index.md` for the plan index, and `https://ziadhosny007.atlassian.net` for the Jira board.

### 6 — First commit, remote, and both branches

Stage, then verify the staged list **before** committing:

```bash
git add -A
git status --short
```

The staged set must be exactly: `.claude/commands/squad-new-story.md`, `.claude/commands/squad-plan.md`, `.gitattributes`, `.gitignore`, `README.md`, `backend/README.md`, `frontend/README.md`, `.squad/config.yaml`, `.squad/README.md`, `.squad/plans/**`, and `.squad/stories/**/intake.md`.

`.squad/secrets.yaml` **must not appear.** If it does, the squad-kit block in `.gitignore` was damaged in task 3 — fix `.gitignore`, run `git rm --cached .squad/secrets.yaml`, and re-stage.

```bash
git commit -m "chore: initialize ZCRM monorepo (ZCRM-2)"
```

Create the remote repository in the **GitHub web UI** (`gh` is not installed): new repository named `ZCRM`, visibility **Private**, and **no** README, `.gitignore`, or license — an initialized remote creates an unrelated root commit and the first push then fails with `! [rejected] ... fetch first`.

```bash
git remote add origin <repo-url>
git push -u origin main
git checkout -b develop
git push -u origin develop
```

Leave the local checkout on **develop** — `main` is protected and no work happens on it directly.

### 7 — Protect `main` on the remote

In the GitHub web UI: **Settings → Rules → Rulesets → New branch ruleset**.

- Ruleset name: `protect-main`
- Enforcement status: **Active**
- Target branches: **Include default branch** (`main`)
- Enable: **Restrict deletions**, **Block force pushes**, and **Require a pull request before merging** with **Required approvals: 1**
- Leave **Require status checks to pass** off — there is no CI workflow yet. ZCRM-3 and ZCRM-4 add build checks; enable the rule then.
- If the repository is Private on a Free plan and rulesets are unavailable, use **Settings → Branches → Add branch protection rule** for `main` with **Require a pull request before merging** and **Do not allow bypassing the above settings** instead. Record which of the two mechanisms was used in the PR description.

### 8 — Confirm the squad-kit plan tables

**File: `.squad/plans/project-setup-foundation/00-overview.md`** — must contain a row for this story: NN `01`, this file name, title, `ZCRM-2`, depends on `—`.

**File: `.squad/plans/00-index.md`** — must contain a `project-setup-foundation` row linking to that overview.

Both were written by the planning session that produced this plan. **Confirm they are present and correct; do not re-author them.**

---

## Edge Cases & Failure Modes

- **`git init` without `-b main`.** `init.defaultBranch` is unset globally (verified), so git creates `master`, and `git push -u origin main` then fails with `src refspec main does not match any`. Enforced by the explicit `-b main` in task 1. Recovery: `git branch -m master main`.
- **Squad-kit block in `.gitignore` damaged.** Lines 1–8 of `.gitignore` are machine-managed. Editing inside them de-ignores `.squad/secrets.yaml`, which is a real secrets file present in the working tree today — committing it leaks credentials. Enforced by the "line 9 and below" rule in task 3 and by the `git status --short` inspection in task 6. Recovery: restore lines 1–8 verbatim, then `git rm --cached .squad/secrets.yaml`.
- **`.env.example` swallowed by `.env.*`.** If the `!.env.example` negation is placed before `.env.*`, or omitted, ZCRM-6 cannot commit its example file and the failure only surfaces four stories later. Enforced by rule ordering in task 3, verified by Verification step 4.
- **Remote initialized with a README.** GitHub's "Add a README file" checkbox creates a root commit unrelated to the local history; `git push -u origin main` is then rejected with `fetch first`, and `git pull` refuses with `refusing to merge unrelated histories`. Enforced by the explicit "no README, no .gitignore, no license" instruction in task 6. Recovery: delete and recreate the empty remote.
- **Branch protection enabled before the first push.** A ruleset that blocks direct pushes to `main` also blocks the initial `git push -u origin main`. Task order is load-bearing: push in task 6, protect in task 7. **Do not reorder.**
- **`.gitattributes` added after files are committed.** `* text=auto eol=lf` applies only to files as they are next staged. Adding it in task 4 — before the first commit in task 6 — makes the whole history LF from the start. Added later, it requires `git add --renormalize .` and produces a whole-repo diff.
- **Empty `backend/` or `frontend/` in a fresh clone.** Git does not track empty directories. Without the placeholder READMEs from task 2, a `git clone` produces neither folder and Story 02 starts by guessing. Enforced by the fresh-clone check in Test Plan step 5.
- **Wrong commit author.** The global identity is `ziad.hosny@azm.sa`, which is not the working address for this project. The repo-local `git config user.email` in task 1 prevents commits landing under the wrong account; verified by Verification step 2.
- **Windows path casing.** The working tree lives on `e:\Work\AZM\ZCRM`, a case-insensitive filesystem. Creating `Backend/` and later referring to `backend/` works locally and breaks on the Linux runners used by later stories. All folder names in this story are **lowercase**: `backend/`, `frontend/`.

---

## Test Plan

**No automated unit or integration tests are added in this story** — there is no application code, no test runner, and no build to hook them into. The first suites arrive with ZCRM-3 (Go) and ZCRM-4 (Angular). What replaces them here is a set of repository-hygiene checks that must be run and shown to pass.

1. **Ignore-rule check (manual, deterministic).** Run in `e:\Work\AZM\ZCRM`:

```bash
git check-ignore -v .squad/secrets.yaml .env frontend/node_modules/x backend/bin/api .angular/cache
```

   Every path must print a matching rule and its source line. `.squad/secrets.yaml` must match `.gitignore:2`.

2. **Negation check (manual).** `git check-ignore -v .env.example` must exit non-zero and print **nothing** — the file is *not* ignored. This is the single most important ignore assertion in the story; ZCRM-6 depends on it.

3. **Tracked-file inventory (manual).** `git ls-files` must list the root-level files from task 6 and **must not** contain `secrets`, `node_modules`, or `.env`:

```bash
git ls-files | grep -Ei "secrets|node_modules|(^|/)\.env$" ; echo "exit=$?"
```

   Expected: no output, `exit=1`.

4. **Line-ending check (manual).** `git ls-files --eol README.md .gitignore` must report `i/lf` for the index on both files.

5. **Fresh-clone smoke test (smoke).** The real acceptance test for "any developer can clone and understand this":

```bash
git clone <repo-url> "$TEMP/zcrm-smoke"
cd "$TEMP/zcrm-smoke"
git branch -a
ls backend frontend README.md
```

   `git branch -a` must show `remotes/origin/main` and `remotes/origin/develop`; `backend/README.md` and `frontend/README.md` must both exist. Delete the clone afterwards.

6. **Protection check (smoke — this push must FAIL).** From the smoke clone:

```bash
git checkout main
git commit --allow-empty -m "test: protection check"
git push origin main
```

   The push **must be rejected** by the remote (`GH013` / `protected branch hook declined`). A successful push means task 7 did not take effect. Delete the smoke clone either way; never push this commit.

---

## Verification Steps

1. **Repo initialized:** `git -C e:/Work/AZM/ZCRM rev-parse --abbrev-ref HEAD` prints `develop`, and `git log --oneline` shows exactly one commit, `chore: initialize ZCRM monorepo (ZCRM-2)`.
2. **Identity correct:** `git -C e:/Work/AZM/ZCRM log -1 --format="%an <%ae>"` prints `Ziad Hosny <ziadhosny@azmsquad.com>`.
3. **Working tree clean:** `git -C e:/Work/AZM/ZCRM status --short` prints nothing.
4. **Ignore rules behave:** Test Plan steps 1–4 all pass, in particular `git check-ignore -v .env.example` printing nothing.
5. **Remote and branches:** `git ls-remote --heads origin` lists exactly `refs/heads/main` and `refs/heads/develop`.
6. **Regression / clone-ability:** Test Plan step 5 passes — a fresh clone has both branches, both placeholder folders, and a root `README.md`.
7. **main is protected:** Test Plan step 6 passes, i.e. the direct push to `main` is **rejected** by the remote.

---

## Done Criteria

- [ ] `e:\Work\AZM\ZCRM` is a git repository whose first commit is on branch `main`, authored by `ziadhosny@azmsquad.com`.
- [ ] A private GitHub repository exists, `origin` points at it, and both `main` and `develop` are pushed (`git ls-remote --heads origin` shows exactly those two).
- [ ] `backend/README.md` and `frontend/README.md` exist and are tracked, so both folders survive a fresh clone.
- [ ] Root `README.md` contains all seven sections from task 5, including the repository tree, pinned prerequisite versions, the `feature/<TRACKER-ID>-<slug>` branch convention, and the statement that `main` only advances via pull request.
- [ ] `.gitignore` lines 1–8 are byte-for-byte unchanged, and the Go / Node-Angular / env / editor blocks are appended from line 9 onward.
- [ ] `git check-ignore -v .squad/secrets.yaml` matches `.gitignore:2`; `git check-ignore -v .env.example` prints nothing.
- [ ] `.gitattributes` exists with `* text=auto eol=lf`, and `git ls-files --eol README.md` reports `i/lf`.
- [ ] A direct `git push origin main` from a fresh clone is rejected by the remote ruleset.
- [ ] `.squad/plans/project-setup-foundation/00-overview.md` and `.squad/plans/00-index.md` both contain a row for this story.

**STOP HERE. Report to the user and wait for confirmation before proceeding to Story 02 (ZCRM-3 — Backend Project Structure).**
