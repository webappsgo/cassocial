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

`src/server/service/import_service.go` implements `ImportService` with
`ImportData`, `GetImportJob`, and per-source importers (Linktree,
Linkstack, Carrd, about.me, CSV, JSON) but `NewImportService` is never
called and no route in `src/server/handler/router.go` reaches it. The
`AdminHandlers.ImportServices` handler is an unrelated admin-config-import
feature, not this one.

Needs deciding: the route path under PART 14's `/users/*` scope (e.g.
`/api/{api_version}/users/import`) plus a matching frontend route/page
(PART 16 CRUD-parity rule), and how job status polling
(`GetImportJob`) should be exposed.

## 5. `docker/Dockerfile` builder stage violates PART 27 (found by go-lint)

Builder stage uses `FROM golang:alpine` instead of the required
`casjaysdev/go:latest` (PART 27, Dockerfile Requirements). The `go build`
invocation is also missing the `-buildvcs=false -trimpath` flags required
for Docker builds against a mounted `.git` directory.

Needs deciding: nothing — this is a mechanical fix (swap the builder base
image, add the two missing build flags), no design decision required. Not
fixed inline when found because it was outside the scope of the task in
progress at the time (RestoreBackup verification pipeline).

