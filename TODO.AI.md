# TODO.AI.md

Findings from a compliance audit (2026-08-02) not yet fixed. Each item needs
a design decision before implementation, so it is logged here rather than
guessed at.

## 1. Mailer/NotificationManager never constructed or called (AI.md PART 18)

`src/service/mailer.go` (`Mailer`) and `src/service/notifications.go`
(`NotificationManager`) are fully implemented (welcome, password reset,
email verification, 2FA code, team invite, backup/domain-verification
notices) but `NewMailer()` / `NewNotificationManager()` are never called
outside their own tests, and none of `SendWelcome`/`SendPasswordReset`/
`SendEmailVerification`/etc. are called from any registration or
password-reset handler. Net effect: account emails required by PART 18 are
never actually sent.

Needs deciding: where SMTP config is loaded from (server.yml vs DB-backed
settings), where the `Mailer`/`NotificationManager` instances live
(constructed once at startup and threaded through to auth handlers via the
handler/router struct), and which handlers in `src/server/handler/` must
call which `Mailer` methods (registration, password-reset request, email
verification, 2FA setup, admin notifications). Also revisit
`scheduler/tasks.go` `ProcessEmailQueue()`, which currently always returns
nil — no queue is actually processed.

## 2. ImportService (Linktree/Linkstack/Carrd/CSV/JSON import) unreachable

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

## 3. `docker/Dockerfile` builder stage violates PART 27 (found by go-lint)

Builder stage uses `FROM golang:alpine` instead of the required
`casjaysdev/go:latest` (PART 27, Dockerfile Requirements). The `go build`
invocation is also missing the `-buildvcs=false -trimpath` flags required
for Docker builds against a mounted `.git` directory.

Needs deciding: nothing — this is a mechanical fix (swap the builder base
image, add the two missing build flags), no design decision required. Not
fixed inline when found because it was outside the scope of the task in
progress at the time (RestoreBackup verification pipeline).

