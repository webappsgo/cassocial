# TODO.AI.md

Findings from a compliance audit (2026-08-02) not yet fixed. Each item needs
a design decision before implementation, so it is logged here rather than
guessed at.

## 1. NotificationManager never constructed or wired (AI.md PART 18)

`Mailer` is now constructed at startup (`src/server/mailer_setup.go`) and
wired into `AuthHandlers` (welcome, password reset, email verification,
2FA-state-change emails) — done. `src/service/notifications.go`
(`NotificationManager`) is still fully implemented but `NewNotificationManager`
is never called outside its own tests, so admin-facing alerts (backup
failure, domain-verification result, cluster/security notices) are never
sent.

Needs deciding: where the `NotificationManager` instance lives (likely
alongside the `Mailer` in `mailer_setup.go`, threaded into whichever
services/handlers trigger the backup, domain-verification, and
cluster/security-alert flows), the admin recipient address source
(`server.admin_email` vs primary admin's account email), and which existing
call sites (backup service, domain verification scheduler task) must call
`NotificationManager.Notify*` methods.

## 2. `login_alert` email template never wired (AI.md PART 18)

PART 18's account-email table includes `login_alert` (security notification
for a new/suspicious login), and PART 18 states security notifications
(login alerts, password/2FA changes) are always ON — but no handler in
`src/server/handler/auth_handlers.go` (`Login`, `LoginWith2FA`) calls
`Mailer.SendNotification`/an equivalent login-alert method, and no such
method currently exists on `Mailer`.

Needs deciding: the trigger condition (every login vs. new-device/new-IP
only — PART 18 doesn't mandate detection heuristics), and whether a new
`Mailer.SendLoginAlert` method is added or `SendNotification` is reused with
a `login_alert` template key.

## 3. Frontend `/reset-password` page missing (AI.md PART 16 CRUD parity)

`ForgotPassword` (in `auth_handlers.go`) emails a link to
`{siteURL}/reset-password?token=...`, and `POST {prefix}/reset-password` is
a registered API route (`router.go:141`), but there is no frontend
`GET /reset-password` route/template rendering an actual form — the emailed
link has nowhere to land in a browser, violating PART 16's user-facing
API-route/frontend-route parity rule.

Needs deciding: template location (alongside other `public.tmpl`-layout auth
pages), whether the token is validated server-side before rendering the form
(vs. only on submit), and the no-JS-required form submission target (should
POST directly to the API route or to a frontend-owned handler that proxies
to it).

## 4. ImportService (Linktree/Linkstack/Carrd/CSV/JSON import) unreachable

Backend wiring done: `ImportExportHandler.HandleImport`/`HandleImportStatus`
now delegate to `service.ImportService` (`ImportData`, `GetImportJob`) and
are registered in `router.go` under `/api/profiles/import` (POST to submit,
GET with `?job_id=` to poll status — a path-segment job ID conflicts with
the existing `GET /api/profiles/{id}/qr` pattern under Go 1.22 ServeMux
route-conflict detection, so status polling uses a query param instead,
matching `HandleExport`'s `?profile_id=` convention). Covered by
`import_export_test.go` (`TestHandleImport` per-source subtests,
`TestHandleImportStatus`).

Still open: no frontend page for import exists yet. `DashboardHandler`
(`dashboard.go`) is itself still all hardcoded/stub JSON responses with no
routes registered in `router.go` for any of its methods (`HandleDashboard`,
`HandleProfileList`, `HandleProfileCreate`, `HandleProfileEdit`,
`HandleAnalyticsOverview`, `HandleAccountSettings`, `HandleNotifications`,
`HandleRecentActivity`) — a real import UI belongs in that same dashboard
surface once it's backed by real data, not bolted on ahead of it. Needs
deciding: whether to wire the dashboard's real routes/data first (separate,
larger task) before adding an import form, or add a minimal standalone
import form now; and the no-JS-required submission target given the API is
JSON-body-only (a plain HTML form posts urlencoded/multipart, so either the
API needs to accept form-encoded bodies too, or a frontend-owned handler
translates form fields to the `ImportRequest` JSON shape server-side).

## 5. `docker/Dockerfile` builder stage violates PART 27 (found by go-lint)

FIXED: builder stage now uses `FROM casjaysdev/go:latest`; the `go build`
invocation now sets `GOFLAGS=-buildvcs=false` and passes `-trimpath`.
Verified via a full `docker build` of `docker/Dockerfile` — builds cleanly.

## 6. `docker/Dockerfile` bakes `LABEL` blocks in at build time (PART 27)

FIXED: both `LABEL` blocks removed from `docker/Dockerfile`, along with
the now-unused `VERSION`/`BUILD_DATE`/`COMMIT_ID`/`LICENSE` runtime-stage
`ARG`s that only existed to feed them. The same key/value pairs are now
supplied as `labels:`/`annotations:` on the new `docker.yml` workflow's
`docker/build-push-action` step (see item 7). Verified with a full
`docker build` — still builds and runs cleanly with no labels baked in.

## 7. `docker.yml` (and `beta.yml`, `daily.yml`) workflows missing (AI.md PART 28)

`docker.yml` FIXED: created `.github/workflows/docker.yml`, copied
verbatim from AI.md PART 28's reference workflow (`build-standard` job
only — cassocial has no `docker/Dockerfile.aio`, so the reference's
`build-aio` job doesn't apply here). Registry is `ghcr.io` per AI.md
PART 28 ("Registry: ghcr.io") and PART 2 line 5307, authenticated via
`secrets.GITHUB_TOKEN` — no extra registry secret needed, matching the
reference spec exactly. Validated with `act --list -W
.github/workflows/docker.yml`; triggered CI build confirmed green.

Still open: `beta.yml` (push to `beta` branch) and `daily.yml` (3am UTC
cron + push to main) are still missing per `cicd-rules.md`'s required
workflow list. Smaller, separate follow-up — same reference-spec pattern
as `docker.yml`/`ci.yml`/`release.yml`, no new decisions needed, just not
done in this pass since it wasn't part of what item 6 required.

