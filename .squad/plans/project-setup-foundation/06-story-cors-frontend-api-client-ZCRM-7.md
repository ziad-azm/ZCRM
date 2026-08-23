# Story 06 — CORS & Frontend API Client (Story: ZCRM-7)

## Prerequisites

- Story 05 completed: [05-story-environment-configuration-ZCRM-6.md](05-story-environment-configuration-ZCRM-6.md). This story adds `CORS_ALLOWED_ORIGINS` to the loader that story built, so `backend/internal/config` must exist first.
- **Story 05 landed without its test suite.** Commit `40bc9e0` on `feature/ZCRM-6-environment-config` contains `internal/config/{config,env,dotenv,validate}.go` but **no `config_test.go`**, and its build/vet/test run was never completed. Before starting this story, run `go build ./... && go vet ./... && go test ./... -count=1` from `backend/` and fix whatever fails. Do **not** layer CORS onto an unverified config package.
- **Branch state.** `feature/ZCRM-6-environment-config` is stacked on `feature/ZCRM-5-database-migrations` → `feature/ZCRM-3-backend-structure`, and **none of the three has merged**. `develop` still has no Go source. Branch this story from `feature/ZCRM-6-environment-config`, or from `develop` once ZCRM-3 → ZCRM-5 → ZCRM-6 have merged in that order.
- **The Angular half is blocked on ZCRM-4, exactly as in Story 05.** The Angular workspace (`frontend/angular.json`, `frontend/src/`) exists **only** on `feature/ZCRM-4-frontend-structure`, which still has just `ng new`. On any branch stacked on ZCRM-3, `frontend/` holds only a `README.md`, so `ng generate interceptor` cannot run. Tasks 6–8 are gated on that; read task 6 before touching `frontend/`.
- Tooling verified present: `go1.26.7`, `@angular/cli 22.0.6`, `docker 20.10.22` with Compose `v2.15.1`. The Postgres container currently publishes **55432** because a native `postgresql-x64-17` service holds 5432.

---

## Story Goal

Let the Angular app and the Go API talk to each other over a real cross-origin request, and give the frontend one place where every outgoing request is shaped.

When this story is done:

1. The backend answers CORS preflight requests, with the allowed origins driven by `CORS_ALLOWED_ORIGINS` — no hardcoded origin anywhere.
2. A wildcard origin is **rejected at startup** when `APP_ENV=production`.
3. Preflight and actual cross-origin requests both appear in the structured logs, so a CORS failure is diagnosable.
4. The Angular app has a functional `HttpInterceptorFn` that sets shared headers on API requests and is the single place the auth token will later be attached.
5. `GET /health` is verified from the browser **two ways**: through the dev-server proxy (same-origin) and as a direct cross-origin call to `:8080`.

**Not in scope:** authentication. The interceptor sets base headers and carries a clearly marked insertion point for the bearer token, but **no token logic, no login, no refresh** — that is **ZCRM-9**. No cookie-based sessions, so `AllowCredentials` stays `false` (see task 2 for why that is deliberate). No rate limiting, no CSRF middleware, no error-mapping interceptor. No change to the database or migrations.

---

## Context — Read These Files First

1. `backend/internal/server/server.go` — all 34 lines. The chain is built on **lines 22–25**: `chimw.RequestID`, `chimw.RealIP`, `middleware.RequestLogger(log)`, `chimw.Recoverer`. `New` already takes `*config.Config` (line 19), so CORS needs no signature change. Task 3 inserts one `r.Use` into this block.
2. `backend/internal/config/config.go` — the `Server` struct at **lines 34–41** and the `cfg.Server = Server{...}` assignment inside `Load` (around **lines 93–99**). Task 2 adds a `CORS` struct beside `Server` and one more assignment.
3. `backend/internal/config/env.go` — the helpers `getString` (line 15), `getDuration` (line 22), `getInt32` (line 39), `getLogLevel` (line 53). **There is no slice helper yet**; task 2 adds `getStringSlice`.
4. `backend/internal/config/validate.go` — **lines 14–38**. Match the existing style exactly: append to `errs`, name the environment variable in every message, and return `errors.Join(errs...)` so all problems surface at once.
5. `backend/internal/middleware/logging.go` — `RequestLogger` logs in a **`defer`**, which is why it still records a request that a middleware below it short-circuits. That is what makes preflight logging work in task 3.
6. `backend/README.md` — the **`## Endpoints`** section's middleware paragraph states the chain is `RequestID → RealIP → RequestLogger → Recoverer`, and **`## Not here yet`** has a CORS bullet claiming CORS "mounts above `RequestID`". **Both statements change in this story** — see the correction below. Task 9 rewrites them.
7. `.env.example` — the `# ---------- HTTP server ----------` block. Task 2 adds the CORS variables there.
8. [05-story-environment-configuration-ZCRM-6.md](05-story-environment-configuration-ZCRM-6.md) — task 8 of that story defines the `Environment` interface and `apiBaseUrl`. This story's interceptor and its direct-origin verification build on it, and it is **also** still unimplemented on the frontend.

**Verified dependency version:** `github.com/go-chi/cors` **v1.2.2** (resolved against the module proxy during planning).

**Correction to earlier guidance — read this before task 3.** The ZCRM-4 plan and the current `backend/README.md` both say CORS "mounts above `RequestID`". **That is wrong and this story overrides it.** `cors.Handler` answers a preflight `OPTIONS` and returns without calling the next handler, so anything mounted *below* it never sees a preflight. Mounted above `RequestID`, every preflight would be invisible in the logs — and preflight failures are precisely what CORS debugging needs to see. The correct position is **after `RequestLogger` and before `Recoverer`**, because `RequestLogger` logs from a `defer` and therefore still records the short-circuited request with its real status.

---

## Implementation tasks

Run Go commands from `e:\Work\AZM\ZCRM\backend`, `ng` from `e:\Work\AZM\ZCRM\frontend`, `docker compose` from `e:\Work\AZM\ZCRM`.

### 1 — Branch, after verifying Story 05

```bash
cd e:/Work/AZM/ZCRM/backend
go build ./... && go vet ./... && go test ./... -count=1
```

**Fix any failure before continuing** — Story 05 shipped unverified. Then:

```bash
cd e:/Work/AZM/ZCRM
git checkout feature/ZCRM-6-environment-config
git checkout -b feature/ZCRM-7-cors-api-client
```

### 2 — CORS configuration

```bash
cd e:/Work/AZM/ZCRM/backend
go get github.com/go-chi/cors@v1.2.2
```

`go-chi/cors` is the router's companion package: a plain `func(http.Handler) http.Handler`, no framework types, and it handles preflight, `Vary`, and header negotiation correctly. Hand-rolling CORS is the classic source of "works in curl, fails in the browser" bugs.

**File: `backend/internal/config/env.go`** — add the slice helper beside the others:

```go
// getStringSlice splits a comma-separated value, trimming each element and
// dropping empties. Returns def when the variable is unset or has no usable
// elements.
func getStringSlice(key string, def []string) []string {
	raw := getString(key, "")
	if raw == "" {
		return def
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}

	if len(out) == 0 {
		return def
	}

	return out
}
```

`strings` is already imported.

**File: `backend/internal/config/config.go`** — add the struct after `Server` (line 41) and wire it in `Load`:

```go
// CORS holds cross-origin request policy.
type CORS struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// AllowsAnyOrigin reports whether the policy contains a wildcard.
func (c CORS) AllowsAnyOrigin() bool {
	for _, o := range c.AllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}
```

Add `CORS CORS` to the `Config` struct, and in `Load`, after the `cfg.Server = Server{...}` assignment:

```go
	corsMaxAge, err := getInt32("CORS_MAX_AGE", 300)
	collect(err)

	cfg.CORS = CORS{
		// The Angular dev server's default origin. Production sets its real one.
		AllowedOrigins: getStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:4200"}),
		AllowedMethods: getStringSlice("CORS_ALLOWED_METHODS",
			[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		AllowedHeaders: getStringSlice("CORS_ALLOWED_HEADERS",
			[]string{"Accept", "Authorization", "Content-Type", "X-Requested-With"}),
		MaxAge: int(corsMaxAge),
	}
```

`Authorization` is in the default allowed headers on purpose: ZCRM-9 sends a bearer token, and a missing entry there produces a preflight rejection that looks like a broken login.

**File: `backend/internal/config/validate.go`** — add two checks in the existing style:

```go
	if len(c.CORS.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS: must list at least one origin"))
	}

	// A wildcard in production would let any site call the API with a user's
	// browser. Development keeps the escape hatch.
	if c.IsProduction() && c.CORS.AllowsAnyOrigin() {
		errs = append(errs, errors.New(`CORS_ALLOWED_ORIGINS: "*" is not allowed when APP_ENV=production`))
	}

	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			continue
		}
		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not an absolute origin (want e.g. https://app.example.com)", origin))
			continue
		}
		if u.Path != "" && u.Path != "/" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q must not include a path", origin))
		}
	}
```

Add `"net/url"` to the imports. **The path check matters**: browsers send `Origin: https://app.example.com` with no trailing path, so a configured `https://app.example.com/` never matches and the failure is invisible from the server side.

**File: `.env.example`** — add after the `SHUTDOWN_TIMEOUT` line:

```dotenv
# ---------- CORS ----------
# Comma-separated. The Angular dev server runs on :4200.
# "*" is rejected when APP_ENV=production.
CORS_ALLOWED_ORIGINS=http://localhost:4200
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Accept,Authorization,Content-Type,X-Requested-With
CORS_MAX_AGE=300
```

### 3 — Mount the middleware

**File: `backend/internal/server/server.go`** — insert one `r.Use` between `RequestLogger` (line 24) and `Recoverer` (line 25):

```go
import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ziad-azm/ZCRM/backend/internal/config"
	"github.com/ziad-azm/ZCRM/backend/internal/handlers"
	"github.com/ziad-azm/ZCRM/backend/internal/middleware"
	"github.com/ziad-azm/ZCRM/backend/internal/repositories"
	"github.com/ziad-azm/ZCRM/backend/internal/services"
)

// New returns the application router with all routes and middleware mounted.
func New(log *slog.Logger, cfg *config.Config, db repositories.Pinger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	// CORS sits BELOW RequestLogger on purpose: cors.Handler answers preflight
	// OPTIONS without calling the next handler, and RequestLogger logs from a
	// defer, so mounting it here is what makes preflight requests visible in
	// the logs. Above RequestID they would be invisible.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: false,
		MaxAge:           cfg.CORS.MaxAge,
	}))
	r.Use(chimw.Recoverer)

	health := handlers.NewHealthHandler(
		services.NewHealthService(cfg.Version, cfg.Database.PingTimeout, db),
		log,
	)
	r.Get("/health", health.Get)

	return r
}
```

**`AllowCredentials: false` is deliberate.** The app will authenticate with a bearer token in the `Authorization` header, not a cookie, so credentials are unnecessary — and `AllowCredentials: true` combined with a wildcard origin is rejected by every browser, a trap worth not walking into. When ZCRM-9 chooses cookie sessions instead, that flag flips **and** the wildcard escape hatch must be removed at the same time.

**The four surrounding `r.Use` calls keep their order.** `RequestID` before `RequestLogger`, `RequestLogger` before `Recoverer` — both still load-bearing, both still covered by existing tests.

### 4 — `OPTIONS` route for preflight on unmatched paths

chi returns `405` for an `OPTIONS` request to a path that registers only `GET`, which stops the CORS middleware from answering some preflights. Register a catch-all so preflight always resolves:

```go
	r.Options("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
```

Place it immediately after `r.Get("/health", health.Get)`. The CORS middleware has already written the `Access-Control-*` headers by the time this handler runs; it only supplies the status. Without it, `OPTIONS /health` returns `405` and the browser reports a CORS failure that looks like a server misconfiguration.

### 5 — Backend end-to-end check

```bash
cd e:/Work/AZM/ZCRM
POSTGRES_PORT=55432 docker compose up -d
cd backend
DATABASE_URL="postgres://zcrm:zcrm@localhost:55432/zcrm?sslmode=disable" go run ./cmd/api
```

In a second shell, simulate what the browser sends:

```bash
# Preflight from the allowed origin
curl -i -X OPTIONS http://localhost:8080/health \
  -H "Origin: http://localhost:4200" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: content-type"

# Actual cross-origin GET
curl -i http://localhost:8080/health -H "Origin: http://localhost:4200"

# Disallowed origin — must NOT echo an allow header
curl -i http://localhost:8080/health -H "Origin: http://evil.example.com"
```

Expected: the first returns `204` with `Access-Control-Allow-Origin: http://localhost:4200` and an `Access-Control-Max-Age` of `300`; the second returns `200` with the allow header and the health JSON; the third returns `200` **with no `Access-Control-Allow-Origin` header at all** — the browser is what blocks it, so the body still arrives on the wire. That last point is the one that confuses people: absence of the header *is* the rejection.

### 6 — Angular interceptor (gated on the workspace)

**Check first:** `ls frontend/angular.json`. If it is missing, the Angular workspace is not on this branch — **stop here**, record tasks 6–8 as deferred in the PR description, and finish with tasks 9–10. Do **not** create a partial Angular tree by hand.

If it exists, also confirm `frontend/src/environments/environment.ts` is present (Story 05 task 8). If it is not, create it per that story's task 8 before continuing, since the interceptor's tests import `environment`.

```bash
cd e:/Work/AZM/ZCRM/frontend
ng generate interceptor core/interceptors/api
```

The schematic's `--functional` default is `true`, so this produces an `HttpInterceptorFn` — a function, not a class. Class-based interceptors are the pre-standalone pattern; do not convert it.

**File: `frontend/src/app/core/interceptors/api.interceptor.ts`**

```ts
import { HttpInterceptorFn } from '@angular/common/http';

import { environment } from '../../../environments/environment';

/**
 * Shapes every request to our own API: shared headers now, the auth token
 * later (ZCRM-9). Requests to third-party URLs pass through untouched.
 */
export const apiInterceptor: HttpInterceptorFn = (req, next) => {
  if (!req.url.startsWith(environment.apiBaseUrl)) {
    return next(req);
  }

  const headers = req.headers
    .set('Accept', 'application/json')
    .set('X-Requested-With', 'XMLHttpRequest');

  // ZCRM-9 attaches the bearer token here:
  //   const token = inject(AuthService).token();
  //   if (token) headers = headers.set('Authorization', `Bearer ${token}`);

  return next(req.clone({ headers }));
};
```

Two details that matter:

- **The `apiBaseUrl` guard is required.** An interceptor that stamps `X-Requested-With` onto every request also stamps it onto third-party calls, which adds a preflight to requests that did not need one and can leak an `Authorization` header to another host once ZCRM-9 lands. That last consequence is why the guard goes in now rather than later.
- **`req.clone({ headers })`, never mutation.** `HttpRequest` is immutable; assigning to `req.headers` silently does nothing.

**File: `frontend/src/app/app.config.ts`** — register it:

```ts
import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';

import { apiInterceptor } from './core/interceptors/api.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideHttpClient(withFetch(), withInterceptors([apiInterceptor])),
    provideRouter(routes),
  ],
};
```

**`withInterceptors` must be passed to `provideHttpClient`** — an interceptor that is merely exported and never registered runs never, and nothing warns about it.

### 7 — Direct cross-origin mode (gated on the workspace)

The proxy from ZCRM-4 makes browser calls same-origin, which means the proxy path never exercises CORS. Add a second, explicit configuration so the cross-origin path is genuinely testable.

**File: `frontend/src/environments/environment.development.ts`** — leave `apiBaseUrl: '/api'` as the default (proxied).

**Create file: `frontend/src/environments/environment.direct.ts`**

```ts
import { Environment } from './environment.model';

/**
 * Bypasses the dev-server proxy and calls the backend cross-origin, so the
 * backend's CORS configuration is exercised for real. Used by
 * `npm run start:direct`.
 */
export const environment: Environment = {
  production: false,
  apiBaseUrl: 'http://localhost:8080',
};
```

**File: `frontend/angular.json`** — add a `direct` build configuration that replaces the environment file, and a matching serve configuration:

- Under `projects.frontend.architect.build.configurations`, add `"direct"` with the same `fileReplacements` shape as `development` but pointing at `environment.direct.ts`, plus `"optimization": false` and `"sourceMap": true`.
- Under `projects.frontend.architect.serve.configurations`, add `"direct": { "buildTarget": "frontend:build:direct" }`.

**File: `frontend/package.json`** — add the script:

```json
"start:direct": "ng serve --configuration direct"
```

This is the configuration that proves task 2 works from a browser. `npm start` continues to use the proxy.

### 8 — Frontend end-to-end verification (gated on the workspace)

With Postgres and the API running (task 5), in two more shells:

```bash
cd frontend && npm start            # proxied, same-origin
cd frontend && npm run start:direct # cross-origin, exercises CORS
```

For each, open `http://localhost:4200`, confirm the dashboard shows `status: ok` and a live uptime, and check the browser devtools **Network** tab:

- **Proxied run:** the request goes to `http://localhost:4200/api/health`, carries `Accept: application/json` and `X-Requested-With` (proof the interceptor ran), and there is **no** preflight.
- **Direct run:** the request goes to `http://localhost:8080/health`, is **preceded by an `OPTIONS` preflight**, and the response carries `Access-Control-Allow-Origin: http://localhost:4200`.

The backend log must show the preflight and the `GET` as two separate `http request` lines in the direct run. That pair of log lines is the deliverable of task 3.

### 9 — Update `backend/README.md`

- In **`## Endpoints`**, update the middleware paragraph: the chain is now `RequestID → RealIP → RequestLogger → CORS → Recoverer`, and state why CORS sits below the logger (preflight visibility).
- Add a **`## CORS`** section: the four `CORS_*` variables, the default `http://localhost:4200`, that `"*"` is rejected when `APP_ENV=production`, that origins must be absolute and path-free, and that `AllowCredentials` is `false` because auth will use a bearer token.
- Include the three `curl` commands from task 5, and the note that a disallowed origin yields a `200` **without** the allow header rather than an error status.
- In **`## Not here yet`**, **delete the CORS bullet** (it is now false, and its "mounts above `RequestID`" claim is wrong). Keep the domain-tables bullet. Add a bullet for auth (**ZCRM-9**) noting the interceptor's token insertion point and that `AllowCredentials` flips only if cookie sessions are chosen.

### 10 — Update `frontend/README.md` and the root `README.md`

**`frontend/README.md`** — only if ZCRM-4 has rewritten it from the CLI placeholder. Add an **`## API calls`** section: `ApiService` reads `environment.apiBaseUrl`; `apiInterceptor` sets shared headers on requests to that base URL only; `npm start` is proxied and same-origin while `npm run start:direct` is cross-origin and exercises the backend's CORS; the auth token will be attached in the interceptor by ZCRM-9.

**`README.md`** (root) — one line in the `### Frontend` block noting `npm run start:direct` as the cross-origin variant. No other change.

### 11 — Commit

```bash
cd e:/Work/AZM/ZCRM
git add backend .env.example README.md frontend
git status --short
```

`.env` must not appear. If tasks 6–8 were skipped, no `frontend/` path should appear either — say so explicitly in the PR description.

```bash
git commit -m "feat: add configurable CORS and the Angular API interceptor (ZCRM-7)"
git push -u origin feature/ZCRM-7-cors-api-client
```

Open the pull request into **`develop`**.

---

## Edge Cases & Failure Modes

- **CORS mounted above `RequestID`.** Preflight requests are answered and returned before the logger is reached, so `OPTIONS` never appears in the logs and a CORS failure is undiagnosable from the server side. Enforced by the placement and comment in task 3, asserted by Test Plan step 6.
- **`OPTIONS` returning `405`.** chi answers `405` for a method a route did not register, which pre-empts the CORS middleware for some preflights, and the browser reports a generic CORS error. Enforced by the catch-all in task 4, asserted by Test Plan step 5.
- **Wildcard origin in production.** `CORS_ALLOWED_ORIGINS=*` would let any website call the API using a visitor's browser. Rejected at startup by `validate()` when `APP_ENV=production`, and asserted by Test Plan step 3. Development keeps the wildcard for convenience.
- **Origin configured with a trailing slash or path.** Browsers send `Origin: https://app.example.com` with no path, so `https://app.example.com/` never matches and every request is silently rejected with no server-side error. Enforced by the path check in task 2, asserted by Test Plan step 4.
- **`Authorization` missing from `AllowedHeaders`.** The preflight for an authenticated request fails, and the symptom looks like a broken login rather than a CORS misconfiguration. Enforced by including it in the defaults in task 2.
- **`AllowCredentials: true` with a wildcard origin.** Every browser rejects that combination outright. Avoided by keeping `AllowCredentials: false`; the flag flips only if ZCRM-9 chooses cookie sessions, and the wildcard must be removed in the same change.
- **A disallowed origin still returns `200`.** CORS is enforced by the *browser*, not the server: the response arrives with no `Access-Control-Allow-Origin` header and the browser discards it. `curl` therefore shows a normal `200` body. Anyone testing with `curl` alone will wrongly conclude CORS is broken — documented in task 5 and the README.
- **The proxy hides CORS entirely in development.** `npm start` routes through the dev-server proxy, so requests are same-origin and CORS is never exercised; a broken configuration would ship undetected. That is the whole reason task 7 adds `start:direct`.
- **Interceptor exported but not registered.** Without `withInterceptors([apiInterceptor])` in `provideHttpClient`, the interceptor never runs and no warning is emitted. Enforced by task 6, asserted by Test Plan step 8.
- **Interceptor applied to third-party URLs.** Stamping `X-Requested-With` on external requests forces needless preflights, and once ZCRM-9 adds the token it would leak an `Authorization` header to another host. Enforced by the `apiBaseUrl` guard, asserted by Test Plan step 9.
- **Mutating `req.headers`.** `HttpRequest` is immutable; assignment silently no-ops and the headers never reach the server. Enforced by `req.clone({ headers })`.
- **`environment.direct.ts` shipped to production.** It hardcodes `http://localhost:8080`. It is referenced **only** by the `direct` build configuration, never by `production`; do not add it to any other configuration, and never point `environment.ts` at an absolute localhost URL.
- **`getStringSlice` given an empty value.** `CORS_ALLOWED_ORIGINS=` or `,,` yields no usable elements and falls back to the default rather than producing an empty policy that rejects everything. Enforced by the `len(out) == 0` guard, asserted by Test Plan step 2.

---

## Test Plan

Go tests run with `go test ./... -count=1` from `backend/` and must pass with PostgreSQL stopped. Use `t.Setenv`, never `os.Setenv`, and no `t.Parallel()` in the config package.

1. **`backend/internal/config/config_test.go`** (extend — **note this file does not exist yet**, see Prerequisites): CORS defaults. Assert `CORS.AllowedOrigins == []string{"http://localhost:4200"}`, that `AllowedMethods` contains `OPTIONS`, that `AllowedHeaders` contains `Authorization`, and `MaxAge == 300`.

2. **Slice parsing** (same file): `CORS_ALLOWED_ORIGINS="http://a.test, http://b.test ,"` yields exactly `["http://a.test","http://b.test"]` — trimmed, empties dropped. Then `CORS_ALLOWED_ORIGINS=",,"` and `CORS_ALLOWED_ORIGINS=""` both fall back to the default.

3. **Wildcard rejected in production** (same file): `APP_ENV=production`, `JWT_SECRET=x`, `CORS_ALLOWED_ORIGINS=*` → error naming `CORS_ALLOWED_ORIGINS`. The same with `APP_ENV=development` → **no** error.

4. **Malformed origins** (same file), table-driven: `not-a-url`, `localhost:4200` (no scheme), and `http://app.test/path` each produce an error naming `CORS_ALLOWED_ORIGINS`; `http://app.test` and `https://app.test:8443` are accepted.

5. **`backend/internal/server/server_test.go`** (extend existing): preflight. `OPTIONS /health` with `Origin: http://localhost:4200` and `Access-Control-Request-Method: GET` returns `204` (not `405`) and `Access-Control-Allow-Origin: http://localhost:4200`. Add a second case for `OPTIONS /anything-else` also returning `204`, covering the task 4 catch-all.

6. **Preflight is logged** (same file): with the logger writing to the existing `lockedBuffer`, issue the preflight and assert exactly one `"msg":"http request"` line containing `"method":"OPTIONS"`. **This is the regression test for the middleware-ordering decision** — it fails if CORS is moved above `RequestLogger`.

7. **Origin enforcement** (same file): `GET /health` with `Origin: http://localhost:4200` returns `200` **with** the allow header; with `Origin: http://evil.test` it returns `200` **without** any `Access-Control-Allow-Origin` header. The existing four routing verdicts and the panic-logging test must still pass unchanged.

8. **`frontend/src/app/core/interceptors/api.interceptor.spec.ts`** (new, gated on the workspace): with `provideHttpClient(withInterceptors([apiInterceptor]))` and `provideHttpClientTesting()`, issue a `GET` to `` `${environment.apiBaseUrl}/health` `` and assert the intercepted request carries `Accept: application/json` and `X-Requested-With: XMLHttpRequest`.

9. **Third-party passthrough** (same file): a `GET` to `https://third-party.test/data` must arrive with **no** `X-Requested-With` header. Regression test for the guard.

10. **`frontend/src/app/core/api.service.spec.ts`** (existing, from ZCRM-4): unchanged and must still pass — the interceptor must not alter the URL, only headers.

11. **Manual browser verification**: task 8, both serve modes, checked in devtools. Not automated; there is no e2e runner in this workspace and adding one stays out of scope.

---

## Verification Steps

1. **Backend builds and is clean:** from `backend/`, `go build ./...`, `go vet ./...`, `gofmt -l .` (silent).
2. **Tests pass with the database stopped:** `docker compose stop postgres`, then `go test ./... -count=1` — all green, including Test Plan steps 1–7.
3. **Preflight succeeds:** the `OPTIONS` curl from task 5 returns `204` with `Access-Control-Allow-Origin: http://localhost:4200` and `Access-Control-Max-Age: 300`.
4. **Cross-origin GET succeeds:** the second curl returns `200`, the allow header, and `"status":"ok"`.
5. **Disallowed origin is not allowed:** the third curl returns `200` with **no** `Access-Control-Allow-Origin` header.
6. **Preflight is visible in the logs:** the API's stdout shows an `http request` line with `"method":"OPTIONS"` and `"path":"/health"`.
7. **Wildcard is refused in production:** `APP_ENV=production JWT_SECRET=x CORS_ALLOWED_ORIGINS=* go run ./cmd/api` exits non-zero with a `configuration error:` line naming `CORS_ALLOWED_ORIGINS`.
8. **No secret leaks:** `go run ./cmd/api 2>&1 | grep -i -E "password|secret|postgres://"` produces no match.
9. **Frontend (gated):** `npm run build` exits 0; `npm test -- --run` passes including Test Plan steps 8–10; both browser runs from task 8 behave as described.
10. **Regression:** `git diff --stat -- backend/migrations backend/scripts` is empty — this story touches neither.

---

## Done Criteria

- [ ] `go.mod` requires `github.com/go-chi/cors v1.2.2`.
- [ ] `config.CORS` exists with `AllowedOrigins`, `AllowedMethods`, `AllowedHeaders`, `MaxAge`, and `AllowsAnyOrigin()`; `getStringSlice` is in `env.go`.
- [ ] Defaults are `http://localhost:4200`, methods including `OPTIONS`, headers including `Authorization`, and `MaxAge` `300`.
- [ ] `validate()` rejects an empty origin list, a wildcard when `APP_ENV=production`, a non-absolute origin, and an origin carrying a path — each message naming `CORS_ALLOWED_ORIGINS`.
- [ ] `.env.example` documents all four `CORS_*` variables.
- [ ] The middleware chain is `RequestID → RealIP → RequestLogger → CORS → Recoverer`, with the ordering rationale in a code comment.
- [ ] `OPTIONS /health` and `OPTIONS` on an unregistered path both return `204`, never `405`.
- [ ] A preflight produces exactly one `http request` log line with `"method":"OPTIONS"`.
- [ ] `GET /health` with an allowed `Origin` carries `Access-Control-Allow-Origin`; with a disallowed one it returns `200` and no allow header.
- [ ] `AllowCredentials` is `false`, with the reason recorded in a comment.
- [ ] `go build ./...`, `go vet ./...`, `go test ./... -count=1` pass with PostgreSQL stopped; `gofmt -l .` silent.
- [ ] `backend/README.md` documents CORS, states the corrected middleware order, and no longer lists CORS under "Not here yet".
- [ ] **Gated on the Angular workspace:** `apiInterceptor` is an `HttpInterceptorFn` registered via `withInterceptors`, sets `Accept` and `X-Requested-With` on `apiBaseUrl` requests only, leaves third-party URLs untouched, and carries the marked token insertion point for ZCRM-9.
- [ ] **Gated on the Angular workspace:** `environment.direct.ts`, the `direct` build/serve configurations, and `npm run start:direct` exist; the browser shows a preflight plus the allow header in direct mode and no preflight in proxied mode.
- [ ] If tasks 6–8 were deferred for lack of the workspace, the PR description says so explicitly and no `frontend/` file is modified.
- [ ] Committed on `feature/ZCRM-7-cors-api-client`, pushed, PR targets `develop`.
- [ ] `.squad/plans/project-setup-foundation/00-overview.md` contains the row for this story.

**STOP HERE. This is the final story in the project-setup-foundation feature. Report to the user and wait for confirmation before starting the authentication-user-management feature (ZCRM-9).**
