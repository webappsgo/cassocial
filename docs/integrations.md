# Integrations

This page documents external identity providers, discovery endpoints, and
platform-level integrations. Where a category is not implemented in the
shipped product, that is stated explicitly rather than omitted.

## External Identity

**None enabled.** Cassocial only supports local username/password
authentication (Argon2id-hashed, optionally combined with TOTP 2FA — see
[Security](security.md)). There is no OIDC, LDAP, or SAML provider
integration in the current codebase, so there is nothing to configure here.

## Discovery & Protocol Endpoints

| Path | Status |
|------|--------|
| `GET /robots.txt` | Enabled |
| `GET /sitemap.xml` | Enabled |
| `GET /manifest.json` | Enabled — PWA manifest |
| `GET /service-worker.js` | Enabled — PWA service worker |
| `/server/docs/swagger`, `/api/swagger` | Not implemented |
| `/graphql`, `/api/graphql` | Scaffolded but **not mounted** — a `src/graphql` package and a GraphiQL playground generator (`generateGraphiQLHTML` in `src/server/handler/import_export.go`) exist and are covered by unit tests, but no `/graphql` route is registered in `src/server/handler/router.go`. There is currently no live GraphQL endpoint. |
| `/api/v1/server/healthz`, `/health` | See [Security](security.md#public--health-endpoints) |

## Platform Integrations

- **Android App Links / `assetlinks.json`** — not implemented.
- **Apple Universal Links / `apple-app-site-association`** — not implemented.
- **Webhooks** — a data model (`APIWebhook` in `src/server/model/api.go`) exists
  with event constants (`profile.created`, `profile.updated`, `profile.deleted`,
  `link.created`, `link.updated`, `link.deleted`, `link.clicked`) and
  failure-count auto-disable logic, but there is **no HTTP endpoint** to
  create, list, or trigger webhooks yet — the model is not wired into the
  router.
- **Federation / ActivityPub / autodiscovery** — not implemented.

## Email (SMTP)

SMTP is the one integration that is fully wired end-to-end:

- Configured via the setup wizard (`/api/setup/email`) or the admin panel
  (`GET`/`PUT /api/admin/smtp/config`, `POST /api/admin/smtp/test`).
- `src/service/smtp.go` implements the client (TLS and plain connections,
  provider host lookup, retry-on-send).
- `src/service/mailer.go` and `src/service/templates.go` build and send
  transactional email (password reset, email verification, notifications).
- Used for password-reset (`POST /api/auth/forgot-password`,
  `/api/auth/reset-password`) and email verification
  (`GET /api/auth/verify-email/{token}`).

## Short Links

`/s/{code}` is a public redirect endpoint for short links created via
`POST /api/v1/shortlinks`. This is an internal Cassocial feature, not a
third-party integration, but is listed here since it is a public-facing
redirect surface.

## Operator Notes

- No integration in this document requires configuration before first run —
  Cassocial works standalone with local auth and (optionally) SMTP.
- SMTP is optional; without it, password reset and email verification links
  cannot be delivered, and the setup wizard will note this during the
  Email step.
- If external identity, webhooks, or GraphQL are added in a future release,
  update this page alongside the code change (see PART 30 documentation
  requirements).
