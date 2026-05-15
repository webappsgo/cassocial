# Cassocial Implementation TODO

**Current Status: Project does NOT compile. Substantial scaffolding exists but has structural defects, 261 TODO/FIXME/HACK markers in src/, no tests, no templates.**

Last audit: 2026-05-15

## P0 — Compile blockers (must fix before anything else)

- [ ] `src/server/routes.go` was rewritten against `*mux.Router` but `Server.router` is `*http.ServeMux` (server.go:24) and `Server` has no `authHandler`/`publicHandler`/`profileHandler`/`linkHandler`/`serviceHandler`/`analyticsHandler`/`shortlinkHandler`/`qrHandler`/`adminHandler`/`middleware`/`importExportHandler`/`dashboardHandler`/`userHandler`/`graphqlHandler`/`swaggerHandler` fields. Either:
  - Switch `Server.router` to `*mux.Router` (gorilla/mux), add all handler fields, wire them in `NewServer()`, OR
  - Rewrite `routes.go` to use `http.ServeMux` patterns and the methods that actually exist on `Server`.
- [ ] `src/server/service/export_service.go:344,346` — `[]AnalyticsRecord` passed where `[]interface{}` expected; convert via explicit slice.
- [ ] `src/server/service/export_service.go:380` — `record` declared and unused.
- [ ] `go.mod` was missing `github.com/gorilla/mux` and `github.com/robfig/cron/v3`; `go mod tidy` adds them — verify and commit `go.mod`/`go.sum`.

## P1 — Server struct wiring

`src/server/server.go` `Server` struct only has `config`, `db`, `httpServer`, `router`, plus a few maps. Routes file references ~15 handler dependencies that do not exist on the struct.

- [ ] Add handler fields to `Server` struct.
- [ ] Construct each handler inside `NewServer()` with its dependencies (db/store, config, mailer, etc.).
- [ ] Replace `// TODO: Add remaining routes` (server.go:107) by actually delegating to `SetupRoutes()`.

## P2 — Stub code that lies about being implemented

261 `TODO`/`FIXME`/`HACK` markers in `src/`. The most load-bearing ones:

- [ ] `src/server/handler/profile.go` — `// TODO: Get user ID from session` (×2), `// TODO: Save to database`, `// TODO: Get profile from database`, `// TODO: Verify user owns this profile`, `// TODO: Update profile in database`. Handler currently returns canned data.
- [ ] `src/server/geoip.go:42,90,91` — entire GeoIP lookup is a TODO; PART-29 GeoIP analytics not functional.
- [ ] `src/server/tor.go:154` — vanity address generation TODO.
- [ ] `src/server/maintenance.go:70,180,184,188` — bypass list, reconnection, disk-space check, SQLite integrity check all TODO.
- [ ] `src/server/password_hash.go:48` — `// TODO: Rehash bcrypt passwords to Argon2id on successful login` (acceptable placeholder, but project rule is no bcrypt anywhere — confirm bcrypt path is dead and remove the rehash branch, OR implement it).
- [ ] `src/server/service/profile_service.go:341` — DNS verification TODO; PART-35 custom-domain feature not functional.
- [ ] `src/server/service/backup.go:189` — backup type parsing TODO.
- [ ] `src/main.go:45,48,51,52,98,187` — `--daemon`, `--service`, `--maintenance`, `--update`, `--status` flags accepted but not implemented; `_ = flag.Bool(...)` discards the values.
- [ ] `src/server/server.go:263,268,303` — node ID generation, cluster detection, version import all TODO.
- [ ] Audit the remaining ~240 TODO markers; either implement or open tracked issues. Per project rules, no `TODO/FIXME/HACK` may be committed in production code.

## P3 — Missing files required by spec

- [ ] `IDEA.md` — does not exist. AI.md / project rules expect it as the source of truth for project variables and business logic.
- [ ] `src/server/template/layout/` — directory empty. No `base.html`, `admin.html`, etc.
- [ ] `src/server/template/page/` — directory empty. No `login.html`, `register.html`, `profile.html`, `dashboard.html`, `admin.html`, `setup.html`, `home.html`.
- [ ] `src/server/template/partial/` — directory empty. No `header.html`, `footer.html`, `link_card.html`.
- [ ] `src/server/static/images/` — directory empty.
- [ ] `src/server/static/css/admin.css` — missing.
- [ ] `src/server/static/js/admin.js` — missing.
- [ ] `tests/` — directory exists but is empty. There are zero `*_test.go` files anywhere in the repo. PART 13 (testing) is wholly unsatisfied.

## P4 — CI/CD compliance

Third-party Actions are pinned to tags, not full commit SHAs. Per project rules this is a security finding (supply-chain).

- [ ] `.github/workflows/beta.yml`, `daily.yml`, `release.yml`, `docker.yml` — repin to full SHAs:
  - `actions/checkout@v4` → SHA
  - `actions/setup-go@v5` → SHA
  - `actions/upload-artifact@v4` → SHA
  - `actions/download-artifact@v4` → SHA
  - `softprops/action-gh-release@v1` → SHA
  - `docker/setup-qemu-action@v3` → SHA
  - `docker/setup-buildx-action@v3` → SHA
  - `docker/login-action@v3` → SHA
  - `docker/build-push-action@v5` → SHA
- [ ] Apply identical pinning to `.gitea/workflows/*`.
- [ ] Add explicit least-privilege `permissions:` block to each workflow.

## P5 — Tests

- [ ] `src/config/` — table tests for `ParseBool` covering the documented variations, plus config-load roundtrip.
- [ ] `src/server/store/` — CRUD tests for User/Profile/Link/Theme/Service/Settings against the modernc/sqlite driver in-memory.
- [ ] `src/server/` — Argon2id `HashPassword`/`VerifyPassword` tests, including invalid PHC string rejection.
- [ ] `src/server/handler/` — at minimum smoke tests for each route once P0/P1 land (httptest).
- [ ] `Makefile test` runs `go test -v -cover ./...` in Docker — verify it actually exercises tests once they exist.

## P6 — Documentation completeness

- [ ] Verify each file in `docs/` reflects current CLI flags and config keys, not the original spec aspirations.
- [ ] Add doc comments to every exported type/function in `src/server/handler/`, `src/server/service/`, `src/server/store/` (most lack them).
- [ ] `README.md` — re-validate install steps and feature list against what actually compiles after P0/P1.

## P7 — Spec compliance / project hygiene

- [ ] `binaries/` is committed with 9 binaries (~145 MB). Per project conventions build artifacts must not live in the repo. Add `binaries/` to `.gitignore` and remove from tracking.
- [ ] `src/service/example_usage.go` — entire body is a block comment. Either move to `docs/examples/smtp.md` or keep but rename/document why it lives in `package service`.
- [ ] `src/server/store/sqlite.go` previously contained TWO `package store` declarations (lines 1 and 609) and TWO `import` blocks — the file had been concatenated by mistake. Split into `sqlite.go` and `sqlite_profile.go` already done in this audit; verify no other files in the repo have the same defect (`grep -n "^package " src/**/*.go` should show one declaration per file).

## Done in this audit pass

- Aliased `model` as `models` in `src/service/{smtp,mailer,notifications}.go` so they match how the code references the package.
- Split `src/server/store/sqlite.go` (lines 609–1670 contained a second concatenated file) into `src/server/store/sqlite_profile.go`; removed unused `fmt` import there.
- Removed unused `strings` import from `src/swagger/swagger.go`.
- Identified that `gorilla/mux` and `robfig/cron/v3` were missing from `go.mod`; `go mod tidy` resolves them.
