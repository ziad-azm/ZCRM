# Story 03 — Frontend Project Structure (Angular) (Story: ZCRM-4)

## Prerequisites

- Story 02 completed: [02-story-backend-project-structure-ZCRM-3.md](02-story-backend-project-structure-ZCRM-3.md). `GET /health` on `:8080` is what this story calls end-to-end.
- **The ZCRM-3 pull request is still open as of this plan.** `feature/ZCRM-3-backend-structure` is pushed but not merged into `develop`, so `develop` has no `backend/` source. Either merge that PR first and branch from `develop`, or stack this story's branch on `feature/ZCRM-3-backend-structure` (task 1 covers both). **Do not** branch this story from `develop` while the backend is missing — the end-to-end check in task 11 has nothing to call.
- Toolchain verified present: `node v24.15.0`, `npm 11.12.1`, `@angular/cli 22.0.6`, `go1.26.7`.
- **`frontend/README.md` must be deleted before `ng new` runs.** Verified: `ng new frontend` into a folder already holding `README.md` aborts with `A merge conflicted on path "/frontend/README.md"`, and **`--force` does not override it** — both were confirmed by dry run. Task 2 handles this.

---

## Story Goal

Scaffold `frontend/` as an Angular 22 workspace with a feature-based structure, a themed component library, and a service layer that reaches the Go backend — ending with a browser page that displays live data from `GET /health`.

When this story is done:

1. `frontend/` is an Angular 22 workspace: standalone, **zoneless**, SCSS, `vitest` as the test runner.
2. `src/app/` is split into `core/`, `shared/`, and `features/`, and the router has a default route that lands on a real feature component.
3. Angular Material **22.1.3** is installed and themed; the shell uses its components rather than hand-rolled CSS.
4. `ApiService` reads its base URL from `src/environments/`, and a typed `Health` interface mirrors the Go `HealthResponse` struct.
5. `npm start` serves the app on `:4200`, and the dashboard shows the backend's real status, version, and uptime.

**Not in scope:** the backend CORS middleware and the Angular HTTP interceptor — both belong to **ZCRM-7**. This story reaches the backend through the **dev-server proxy** instead (task 8), which is why no backend change is needed here. No auth, no route guards, no PostgreSQL. No production deployment config; `environment.ts` gets the one field this story needs and **ZCRM-6** owns the rest. **No backend file is modified by this story.**

---

## Context — Read These Files First

1. `frontend/README.md` — all 9 lines, the placeholder written by Story 01. It **blocks `ng new`** and is deleted in task 2, then rewritten in task 12.
2. `README.md` (root) — read **lines 20–27** (prerequisites table; Node/npm/Angular CLI rows are 23–25) and **line 37** (`backend/` is a runnable service, `frontend/` still a placeholder "until **ZCRM-4**"), plus the **`### Backend`** block at **lines 39–45**. Task 13 edits line 37 and adds a matching `### Frontend` block.
3. `.gitignore` — read **lines 35–46** (the Node/Angular block). `node_modules/`, `.angular/`, `frontend/dist/`, `frontend/coverage/` are already ignored. **No `.gitignore` change is needed** — `ng new` also writes its own `frontend/.gitignore`, and both are kept.
4. `backend/internal/models/health.go` — **lines 5–11**, the `HealthResponse` struct. The JSON keys are `status`, `version`, **`uptime_seconds`** (snake_case), and `timestamp`. The TypeScript interface in task 7 must mirror these exactly.
5. `backend/internal/server/server.go` — **lines 18–26**. Only `/health` is registered, and **no CORS middleware is mounted**. This is the reason task 8 adds a dev-server proxy instead of calling `http://localhost:8080` directly from the browser.
6. [02-story-backend-project-structure-ZCRM-3.md](02-story-backend-project-structure-ZCRM-3.md) — match its shape: one file per task, a mandatory edge-case section, literal verification commands.
7. `.squad/stories/project-setup-foundation/ZCRM-4/intake.md` — the source work item. Its four tracker tasks map onto tasks 2+6 (`ng new` + default route), 5 (core/shared/features), 9 (UI library + theme), and 7 (`ApiService` + environment).

**Angular 22 facts verified against the installed CLI** — the generated code differs from older tutorials, so do not pattern-match from memory:

- **File naming is the 2025 style guide.** The root component is `src/app/app.ts` exporting class `App`, with `app.html`, `app.scss`, `app.config.ts`, `app.routes.ts`. There is **no** `.component.ts` suffix and **no** `AppModule`.
- **The workspace is zoneless.** The generated `package.json` has **no `zone.js`** dependency and `app.config.ts` has no `provideZoneChangeDetection`. **Do not add zone.js.** Async assertions use `await fixture.whenStable()`.
- **The test runner is `vitest` 4** (with `jsdom`), not Karma/Jasmine. `describe`/`it`/`expect` are globals; `npm test` runs `ng test`.
- **`provideHttpClient` is not in the generated `app.config.ts`** — task 7 adds it.
- **There is no `src/environments/` folder** until `ng generate environments` is run.

---

## Implementation tasks

**No backend changes required.** `backend/` is read-only for this story.

Run every command from `e:\Work\AZM\ZCRM\frontend` unless the block says otherwise.

### 1 — Branch

The ZCRM-3 PR is unmerged, so pick one:

**Preferred — merge ZCRM-3 first, then branch from `develop`:**

```bash
cd e:/Work/AZM/ZCRM
# merge the ZCRM-3 pull request into develop on GitHub first, then:
git checkout develop
git pull
git checkout -b feature/ZCRM-4-frontend-structure
```

**Fallback — stack on the open branch** (use when the PR cannot be merged yet):

```bash
cd e:/Work/AZM/ZCRM
git checkout feature/ZCRM-3-backend-structure
git checkout -b feature/ZCRM-4-frontend-structure
```

If the fallback is used, the ZCRM-4 PR **must** be merged after the ZCRM-3 PR, or its diff will show the backend files as new. Record which path was taken in the PR description.

### 2 — Remove the placeholder and run `ng new`

The placeholder README blocks the schematic, so delete it through git first:

```bash
cd e:/Work/AZM/ZCRM
git rm -q frontend/README.md
```

Then generate the workspace **from the repository root** (the `frontend` argument creates the folder):

```bash
ng new frontend --defaults --style=scss --routing --skip-git --package-manager=npm --ai-config=none
```

Every flag is load-bearing:

- `--defaults` — no interactive prompts.
- `--style=scss` — Angular Material's theming API is Sass; `--style=css` would make task 9 impossible without a follow-up conversion.
- `--routing` — writes `app.routes.ts`, which task 6 fills.
- `--skip-git` — **required.** The repo already has `.git` at the root; letting the schematic run `git init` would nest a second repository inside `frontend/`.
- `--package-manager=npm` — matches the `packageManager` field the schematic writes (`npm@11.12.1`).
- `--ai-config=none` — this option has no default and would otherwise prompt. `none` keeps AI tool files out of `frontend/`; the repo's Claude config already lives at the root in `.claude/`.

Expected output, verified by dry run: `frontend/{angular.json,package.json,tsconfig*.json,.editorconfig,.gitignore,.prettierrc}`, `frontend/.vscode/{extensions,launch,tasks}.json`, `frontend/src/{main.ts,index.html,styles.scss}`, `frontend/src/app/{app.ts,app.html,app.scss,app.spec.ts,app.config.ts,app.routes.ts}`, and `frontend/public/favicon.ico`.

### 3 — Baseline check before writing any code

```bash
cd e:/Work/AZM/ZCRM/frontend
npm install
npm run build
npm test -- --run
```

A green baseline here separates "the scaffold is broken" from "my code is broken" later. `npm test -- --run` forces vitest to exit instead of entering watch mode.

### 4 — Generate the environment files

```bash
ng generate environments
```

This creates `src/environments/environment.ts` and `src/environments/environment.development.ts` and adds a `fileReplacements` entry to `angular.json`.

**Note the direction, it is the opposite of the old convention:** `environment.ts` is the **production/default** file, and `environment.development.ts` **replaces** it in the `development` configuration. There is no `environment.prod.ts` in modern Angular — the ZCRM-6 intake still uses that older wording, and this is the file layout it should be planned against.

**File: `src/environments/environment.ts`**

```ts
export const environment = {
  production: true,
  apiBaseUrl: '/api',
};
```

**File: `src/environments/environment.development.ts`**

```ts
export const environment = {
  production: false,
  apiBaseUrl: '/api',
};
```

Both use `/api` — a **same-origin relative prefix**, never an absolute `http://localhost:8080`. In development the dev-server proxy (task 8) forwards it; in production a reverse proxy does. This is what keeps the browser from making a cross-origin request that the backend would reject, since it has no CORS middleware until ZCRM-7.

### 5 — Create the feature-based folders

```text
frontend/src/app/
├── core/                     # app-wide singletons: services, models, later guards
│   ├── api.service.ts
│   ├── api.service.spec.ts
│   └── models/
│       └── health.ts
├── shared/                   # reusable presentational pieces
│   └── README.md
├── features/                 # one folder per screen
│   └── dashboard/
│       ├── dashboard.ts
│       ├── dashboard.html
│       ├── dashboard.scss
│       └── dashboard.spec.ts
├── app.ts
├── app.html
├── app.scss
├── app.config.ts
└── app.routes.ts
```

`shared/` has nothing to hold yet; give it a one-line `README.md` stating that reusable components land there, so the folder survives commit (git does not track empty directories — the same constraint Story 01 hit with `backend/`).

Generate the two artefacts with the CLI rather than by hand, so naming matches the workspace style guide:

```bash
ng generate service core/api --type=service
ng generate component features/dashboard --type=""
```

`--type=service` is **required** to get `api.service.ts` with class `ApiService`; without it the 2025 style guide produces `api.ts` with class `Api`. Verify the class name after generating.

### 6 — Default route

**File: `src/app/app.routes.ts`** — replace the generated empty array:

```ts
import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
  {
    path: 'dashboard',
    loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
  },
  { path: '**', redirectTo: 'dashboard' },
];
```

The dashboard is **lazy-loaded** via `loadComponent`; that is the standalone-API equivalent of a lazy module and keeps the initial bundle small once more features arrive. The `**` catch-all must stay **last** — routes are matched in order, and a catch-all placed earlier swallows every route below it.

### 7 — `ApiService` and the typed health model

**File: `src/app/core/models/health.ts`**

Mirror `backend/internal/models/health.go` exactly, snake_case key included:

```ts
export interface Health {
  status: string;
  version: string;
  uptime_seconds: number;
  timestamp: string;
}
```

`timestamp` is `string`, not `Date` — `JSON.parse` produces the RFC 3339 string the Go `time.Time` marshals to; converting it is the caller's job.

**File: `src/app/core/api.service.ts`**

```ts
import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';

/** Base HTTP client for the ZCRM backend. All feature services build on this. */
@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = environment.apiBaseUrl;

  get<T>(path: string): Observable<T> {
    return this.http.get<T>(this.url(path));
  }

  post<T>(path: string, body: unknown): Observable<T> {
    return this.http.post<T>(this.url(path), body);
  }

  put<T>(path: string, body: unknown): Observable<T> {
    return this.http.put<T>(this.url(path), body);
  }

  delete<T>(path: string): Observable<T> {
    return this.http.delete<T>(this.url(path));
  }

  /** Joins the configured base URL with a leading-slash path. */
  private url(path: string): string {
    return `${this.baseUrl}${path.startsWith('/') ? path : `/${path}`}`;
  }
}
```

Use `inject()`, not a constructor parameter — that is the idiom in a standalone, zoneless workspace.

**File: `src/app/app.config.ts`** — add `provideHttpClient`. The generated file has only `provideBrowserGlobalErrorListeners()` and `provideRouter(routes)`:

```ts
import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideHttpClient(withFetch()),
    provideRouter(routes),
  ],
};
```

**Without `provideHttpClient`, every injection of `HttpClient` throws `NullInjectorError` at runtime** — the app compiles and the failure only appears when the dashboard loads.

### 8 — Dev-server proxy to the backend

The backend has no CORS middleware (ZCRM-7 adds it), so a browser at `:4200` calling `:8080` directly is blocked. Proxy instead: the browser sees one origin.

**Create file: `frontend/proxy.conf.json`**

```json
{
  "/api": {
    "target": "http://localhost:8080",
    "secure": false,
    "changeOrigin": true,
    "pathRewrite": { "^/api": "" }
  }
}
```

`pathRewrite` strips the `/api` prefix, so `/api/health` reaches the backend as `/health` — the only path `server.go` registers. Angular 22's Vite-based dev server translates `pathRewrite` into its own `rewrite` function (verified in `@angular/build`'s `load-proxy-config.js`), so the webpack-era JSON format is correct here.

**File: `frontend/angular.json`** — add `proxyConfig` to the `serve` target's `options` (sibling of `buildTarget`):

```json
"serve": {
  "builder": "@angular/build:dev-server",
  "options": {
    "proxyConfig": "proxy.conf.json"
  },
  "configurations": { }
}
```

Put it in `options`, **not** inside `configurations.development`, or `ng serve --configuration production` silently loses the proxy.

### 9 — Angular Material, installed and themed

```bash
ng add @angular/material@22 --theme=azure-blue --skip-confirmation
```

**UI library decision: Angular Material 22.1.3** over PrimeNG 22.1.0. It is first-party, versioned in lockstep with Angular majors (its peer range is `@angular/core: ^22.0.0 || ^23.0.0`), and its `ng add` wires the theme, the Roboto and Material Symbols fonts, and `@angular/cdk` in one step. The CDK also gives later stories overlay, a11y, and table primitives without another dependency.

`--theme=azure-blue` is the schematic's default palette pair; the four valid choices are `azure-blue`, `rose-red`, `magenta-violet`, `cyan-orange`. `--skip-confirmation` suppresses the install prompt. The schematic writes this into `src/styles.scss`:

```scss
@use '@angular/material' as mat;

html {
  height: 100%;
  @include mat.theme((
    color: (
      primary: mat.$azure-palette,
      tertiary: mat.$blue-palette,
    ),
    typography: Roboto,
    density: 0,
  ));
}
```

Leave that block as generated. Material 22 is theme-driven through `mat.theme()` and CSS system variables (`--mat-sys-*`); **do not** hand-write component colour overrides.

### 10 — The dashboard feature and the app shell

**File: `src/app/features/dashboard/dashboard.ts`**

```ts
import { Component, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { ApiService } from '../../core/api.service';
import { Health } from '../../core/models/health';

@Component({
  selector: 'app-dashboard',
  imports: [MatCardModule, MatButtonModule, MatProgressSpinnerModule],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class Dashboard {
  private readonly api = inject(ApiService);

  protected readonly health = signal<Health | null>(null);
  protected readonly error = signal<string | null>(null);
  protected readonly loading = signal(false);

  constructor() {
    this.load();
  }

  protected load(): void {
    this.loading.set(true);
    this.error.set(null);

    this.api.get<Health>('/health').subscribe({
      next: (health) => {
        this.health.set(health);
        this.loading.set(false);
      },
      error: (err: unknown) => {
        this.error.set(err instanceof Error ? err.message : 'Backend unreachable');
        this.loading.set(false);
      },
    });
  }
}
```

Signals, not `BehaviorSubject` — this is a zoneless workspace, and signals are what drive change detection reliably without `zone.js`. The `error` branch is **required**: the backend will be down at some point and a silent blank card is a bug report waiting to happen.

**File: `src/app/features/dashboard/dashboard.html`** — a `mat-card` with three states: the spinner while `loading()`, the error text plus a retry button when `error()`, and `status` / `version` / `uptime_seconds` / `timestamp` from `health()` otherwise. Use `@if` / `@else` control flow, not `*ngIf` — the new block syntax is the default in Angular 22 and needs no imports.

**File: `src/app/app.html`** — **replace the generated file entirely.** It is a ~20 KB Angular splash page. The replacement is a `mat-toolbar` with the title "ZCRM" and a `<router-outlet />` beneath it.

**File: `src/app/app.ts`** — add `MatToolbarModule` to `imports` alongside the existing `RouterOutlet`.

### 11 — End-to-end check

Two shells. Backend first:

```bash
cd e:/Work/AZM/ZCRM/backend
go run ./cmd/api
```

Then the frontend:

```bash
cd e:/Work/AZM/ZCRM/frontend
npm start
```

Open `http://localhost:4200`. The dashboard must show `status: ok`, `version: 0.1.0`, and a non-zero uptime. The backend shell must log one `http request` line with `"path":"/health"` and `"status":200` — that line is the proof the call reached Go rather than being served from a mock.

### 12 — Write `frontend/README.md`

The file was deleted in task 2 and `ng new` wrote its own generic Angular README in its place. **Replace that content** with:

1. `# Frontend (Angular)` and a line naming Angular 22, Angular Material 22.1.3, SCSS, and vitest.
2. **`## Layout`** — the tree from task 5, plus one line on the role of `core/` (singletons), `shared/` (reusable presentational pieces), and `features/` (one folder per screen).
3. **`## Run`** — `npm start` (`:4200`), and the explicit note that **the backend must be running on `:8080`** for the dashboard to populate, with the `go run ./cmd/api` command.
4. **`## Test`** — `npm test -- --run` (vitest, single pass) and `npm run build`.
5. **`## Backend calls`** — that `ApiService` reads `environment.apiBaseUrl` (`/api`), that `proxy.conf.json` maps `/api` → `http://localhost:8080` with the prefix stripped, and that this is why no CORS config exists yet.
6. **`## Conventions`** — 2025 file naming (no `.component` suffix), zoneless (no `zone.js`, use signals and `await fixture.whenStable()`), and `@if`/`@for` block control flow.
7. **`## Not here yet`** — the HTTP interceptor and CORS (**ZCRM-7**), full environment config (**ZCRM-6**), auth and guards (later).

### 13 — Update the root `README.md`

- **Line 37** — replace the "`frontend/` is still a placeholder … until **ZCRM-4**" clause: both apps are now runnable, each with its own README.
- Add a **`### Frontend`** block after the existing `### Backend` block (lines 39–45):

  ```bash
  cd frontend
  npm install
  npm start        # http://localhost:4200 — needs the backend running on :8080
  ```

- **Line 23–25** — no change needed; Node, npm, and Angular CLI rows are already accurate.

Leave `## Repository layout`, `## Branching`, `## Commit messages`, and `## Planning workflow` unchanged.

### 14 — Commit

```bash
cd e:/Work/AZM/ZCRM
git add frontend README.md
git status --short
```

`node_modules/`, `frontend/.angular/`, and `frontend/dist/` **must not** appear — they are ignored by root `.gitignore` lines 36–38. `frontend/package-lock.json` **must** appear; it is what makes `npm ci` reproducible.

```bash
git commit -m "feat(frontend): scaffold Angular workspace with Material and health dashboard (ZCRM-4)"
git push -u origin feature/ZCRM-4-frontend-structure
```

Open the pull request into **`develop`**.

---

## Edge Cases & Failure Modes

- **`ng new` aborts on the existing README.** Verified: `A merge conflicted on path "/frontend/README.md"`, and `--force` does **not** clear it. Enforced by the `git rm` in task 2. Recovery: delete the file and re-run.
- **A nested git repository inside `frontend/`.** Without `--skip-git` the schematic runs `git init`, and every frontend file then sits in a second repo invisible to the root one — `git status` at the root shows nothing to commit. Enforced by `--skip-git` in task 2. Recovery: `rm -rf frontend/.git`.
- **`NullInjectorError: No provider for HttpClient`.** The generated `app.config.ts` has no `provideHttpClient`; the app builds fine and only fails when the dashboard's `ApiService` is constructed. Enforced by task 7.
- **CORS failure from calling `:8080` directly.** Setting `apiBaseUrl` to `http://localhost:8080` makes the browser send a cross-origin request that the backend — which mounts no CORS middleware (`server.go` lines 20–23) — never permits. The browser console shows a CORS error while `curl` to the same URL succeeds, which reliably wastes an hour. Enforced by the relative `/api` in task 4 plus the proxy in task 8.
- **Proxy configured but the prefix not stripped.** Omitting `pathRewrite` forwards `/api/health` to the backend, which serves only `/health`, so the dashboard shows a `404` while the proxy itself works. Enforced by task 8.
- **`proxyConfig` placed under `configurations.development`.** It then applies only to that configuration and vanishes for any other serve invocation. Enforced by the explicit "in `options`" instruction in task 8.
- **Adding `zone.js` back.** Following an older tutorial into `provideZoneChangeDetection()` or importing `zone.js` in a workspace generated without it breaks change detection in confusing ways. The generated `package.json` has no `zone.js` — **leave it that way**.
- **Async test assertions without `whenStable`.** In a zoneless workspace, `fixture.detectChanges()` alone does not flush a pending HTTP response. The generated `app.spec.ts` already shows the correct idiom: `await fixture.whenStable()`.
- **Class name is `Api`, not `ApiService`.** Forgetting `--type=service` in task 5 produces `core/api.ts` exporting `Api`, and every import in task 10 fails. Verified against the service schematic's `--add-type-to-class-name` default of `true`.
- **Catch-all route ordering.** `{ path: '**' }` placed above `dashboard` in task 6 makes every URL redirect and the app appears to have one page. Routes match top-down.
- **Production bundle budget.** `angular.json` sets `maximumWarning: 500kB` / `maximumError: 1MB` on the initial bundle. Material plus the CDK can push a production build past the warning. A warning is acceptable; if `npm run build` **errors** on the budget, raise `maximumError` to `1.5MB` in `angular.json` and note it in the PR — do not delete the budget.
- **`npm test` hangs in CI.** `ng test` with vitest watches by default. Always `npm test -- --run` in scripted contexts.
- **`.vscode/` double-handling.** Root `.gitignore` deliberately does not ignore `.vscode/`, while the generated `frontend/.gitignore` ignores `frontend/.vscode/*` with negations for the four files `ng new` writes. Both are correct; leave them.

---

## Test Plan

Tests run with vitest via `npm test -- --run` from `frontend/`.

1. **`src/app/core/api.service.spec.ts`** (unit, replace the generated stub). Configure `TestBed` with `provideHttpClient()` and `provideHttpClientTesting()`. Assert that `api.get<Health>('/health')` issues exactly one `GET` to **`/api/health`** — proving the base URL is prepended — then flush a fixture `Health` object and assert the emitted value matches. Use `HttpTestingController.verify()` in `afterEach`.

2. **Path-joining assertion** (same file): `api.get('health')` — no leading slash — must also hit `/api/health`, covering the `url()` normalisation branch.

3. **`src/app/features/dashboard/dashboard.spec.ts`** (component). With `provideHttpClientTesting()`, create the `Dashboard` fixture, flush a `Health` response, `await fixture.whenStable()`, and assert the rendered text contains `ok` and `0.1.0`.

4. **Error-state assertion** (same file): flush an error (`{ status: 500, statusText: 'Server Error' }`) instead, then assert the error element renders and the spinner is gone. This is the regression test for the silent-blank-card failure mode.

5. **`src/app/app.spec.ts`** (generated). The generated second assertion expects `Hello, frontend` in an `<h1>` — task 10 replaces `app.html`, so **that assertion must be updated**, not deleted: assert the toolbar renders `ZCRM`. Leave the "should create the app" case as is.

6. **Manual smoke** — task 11, both servers up, checked in a browser. Not automated: there is no e2e runner in this workspace and adding one is out of scope.

---

## Verification Steps

Run from `e:\Work\AZM\ZCRM\frontend`.

1. **Install is clean:** `npm install` — exits 0; `frontend/package-lock.json` exists.
2. **Frontend builds:** `npm run build` — exits 0. Budget **warnings** are acceptable; a budget **error** is not.
3. **Unit tests pass:** `npm test -- --run` — all specs green, including the four assertions above.
4. **Dev server serves:** `npm start`, then `curl -s -o /dev/null -w "%{http_code}" http://localhost:4200` prints `200`.
5. **Proxy reaches the backend:** with both servers running, `curl -s http://localhost:4200/api/health` returns the backend's JSON with `"status":"ok"`, and the backend log shows a matching `"path":"/health"` line.
6. **Browser end-to-end:** task 11 — the dashboard renders live status, version, and uptime.
7. **Regression — backend untouched:** `git diff --stat develop -- backend` (or against `feature/ZCRM-3-backend-structure` when stacked) shows **no** backend changes.
8. **Regression — repo hygiene:** `git status --short` at the root lists no `node_modules`, `.angular`, or `dist` entries.

---

## Done Criteria

- [ ] `frontend/` is an Angular 22 workspace: `standalone`, zoneless (no `zone.js` in `package.json`), SCSS, vitest.
- [ ] `src/app/` contains `core/`, `shared/`, and `features/dashboard/`, and `shared/` holds a tracked `README.md` so the folder survives a clone.
- [ ] `app.routes.ts` redirects `''` → `dashboard`, lazy-loads `Dashboard` via `loadComponent`, and has `**` **last**.
- [ ] Angular Material 22.1.3 and `@angular/cdk` are in `package.json`, and `src/styles.scss` contains the generated `mat.theme(...)` block.
- [ ] `core/api.service.ts` exports class **`ApiService`** with `get`/`post`/`put`/`delete`, reading `environment.apiBaseUrl`.
- [ ] `core/models/health.ts` mirrors the Go struct: `status`, `version`, `uptime_seconds`, `timestamp`.
- [ ] `app.config.ts` provides `provideHttpClient(withFetch())`.
- [ ] `proxy.conf.json` maps `/api` → `http://localhost:8080` with `^/api` stripped, and `angular.json` references it from the `serve` target's `options`.
- [ ] `npm run build` and `npm test -- --run` both pass; no budget error.
- [ ] `curl http://localhost:4200/api/health` returns the backend's JSON, and the backend logs the matching request.
- [ ] The browser dashboard shows `ok`, `0.1.0`, and a live uptime; stopping the backend switches it to the error state with a retry button.
- [ ] `frontend/README.md` documents Layout, Run, Test, Backend calls, Conventions, and Not-here-yet.
- [ ] Root `README.md` line 37 no longer calls `frontend/` a placeholder, and a `### Frontend` quickstart follows `### Backend`.
- [ ] No file under `backend/` is modified.
- [ ] Committed on `feature/ZCRM-4-frontend-structure`, pushed, PR targets `develop`.
- [ ] `.squad/plans/project-setup-foundation/00-overview.md` contains the row for this story.

**STOP HERE. Report to the user and wait for confirmation before proceeding to Story 04 (ZCRM-5 — Database & Migrations).**
