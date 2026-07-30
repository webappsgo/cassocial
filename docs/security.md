# Security

This page documents the security model of the shipped product: how credentials
are stored, how sessions work, which endpoints are public, and how to report a
vulnerability.

## Password Storage

Passwords are hashed with **Argon2id** (`golang.org/x/crypto/argon2`), using the
OWASP 2023-recommended parameters:

| Parameter | Value |
|-----------|-------|
| Iterations (time) | 3 |
| Memory | 64 MB |
| Parallelism | 4 |
| Key length | 32 bytes |
| Salt length | 16 bytes |

Hashes are stored in PHC string format (`$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>`).
Verification uses a constant-time comparison (`crypto/subtle`) so timing does not
leak information about a partially-correct hash.

## Sessions & Tokens

- Authentication issues a signed **JWT** (`HS256`) via `github.com/golang-jwt/jwt/v5`.
- The signing secret comes from the `JWT_SECRET` environment variable. If it is
  not set, a random secret is generated at process startup — sessions do not
  survive a server restart in that case, so set `JWT_SECRET` explicitly in
  production.
- Session lifetime is controlled by the `session_timeout_minutes` setting
  (default: 1440 minutes / 24 hours), editable from the admin panel.
- Tokens are validated on every authenticated request; the `RequireAuth`,
  `RequireAdmin`, `RequireRole`, and `RequireActiveUser` middleware wrap the
  routes that need them.

## Two-Factor Authentication (2FA / TOTP)

- 2FA uses TOTP (RFC 6238), 30-second period, 6-digit codes.
- Enrollment (`POST /api/auth/2fa/enable` or `/api/v1/auth/2fa/enable`) generates
  a base32 secret and 10 single-use backup codes (backup codes are stored as
  hashes, shown to the user only once at enrollment time).
- `POST /api/auth/login/2fa` completes login for accounts with 2FA enabled.
- 2FA can be disabled via `POST /api/auth/2fa/disable` (authenticated).

## API Keys & Scopes

The data model (`src/server/model/api.go`) defines API keys with per-key
scopes (`profile:read`, `profile:write`, `link:read`, `link:write`,
`analytics:read`, `user:read`, `user:write`) and expiration. As of this
writing, API key issuance/management is **not yet wired to an HTTP endpoint** —
the admin UI documents an `/admin/api-keys` page (see
[Admin Panel](admin.md)), but authenticated requests currently use the JWT
session token described above, not standalone API keys.

## Security Headers

Every response (via the `SecurityHeaders` middleware) includes:

```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;
```

## Rate Limiting

Authenticated write endpoints can be wrapped with `RateLimitByUser`, a
per-user, per-window rate limiter. Requests over the configured limit receive
`429 Too Many Requests`.

## Public / Health Endpoints

| Path | Purpose |
|------|---------|
| `GET /server/healthz` | Human-readable HTML health page |
| `GET /api/v1/server/healthz` | JSON health check (`status`, `uptime`, `version`, `timestamp`) |
| `GET /health` | Legacy JSON health alias (kept for backward compatibility) |
| `GET /health/ready` | Readiness probe (JSON `{"status":"ready"}`) |
| `GET /health/live` | Liveness probe (JSON `{"status":"alive"}`) |
| `GET /metrics` | Prometheus-format metrics — **restricted to `127.0.0.1`/`::1`/`localhost`**; all other callers get `404` |
| `GET /robots.txt` | Crawler policy (disallows `/api/` and `/admin`) |
| `GET /sitemap.xml` | Sitemap containing the home page |
| `GET /.well-known/security.txt` | Security contact info (RFC 9116) |

None of the database connection details, credentials, or internal error
messages are exposed by these endpoints; the health checks only ping the
database and report `healthy`/`unhealthy`.

## Security Reporting

Cassocial publishes an RFC 9116 `security.txt` at `/.well-known/security.txt`,
which points researchers to the project's GitHub security page
(`https://github.com/casapps/cassocial/security`) for coordinated disclosure.
There is currently no separate in-app `/server/security` reporting page or
`security_id` contact form — use the GitHub security advisories flow linked
from `security.txt` to report a vulnerability.

## Well-Known Namespace

Only one `/.well-known/*` entry is currently enabled:

| Path | Status |
|------|--------|
| `/.well-known/security.txt` | Enabled — see above |

Any other `/.well-known/*` path (e.g. WebFinger, OpenID Provider Metadata,
`assetlinks.json`, `apple-app-site-association`, MTA-STS) returns `404` — none
of those are implemented in the current codebase. See
[Integrations](integrations.md) for what would need to change for any of them
to be enabled.

## Reserved Slugs

To prevent profile slugs from shadowing infrastructure routes, the following
top-level paths are reserved and cannot be claimed as a profile slug: `api`,
`admin`, `auth`, `setup`, `dashboard`, `static`, `health`, `metrics`,
`manifest.json`, `service-worker.js`, `robots.txt`, `sitemap.xml`, `s`, `l`,
`.well-known`, `server`.
