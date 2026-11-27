# API Reference

Cassocial's default API version is `v1`; the server can change it with
`server.api_version`. The default Server Admin path is `administration`, set by
`server.admin_path`.

Use your server address in requests:

```sh
curl -q -LSsf https://YOUR_SERVER/api/v1/server/healthz
```

## API documentation

The OpenAPI UI is available at `/server/docs/swagger`. The machine-readable
OpenAPI document is served at `/api/{api_version}/server/swagger`; the direct
alias `/api/swagger` serves the configured version. Cassocial does not provide
a GraphQL endpoint or GraphiQL page.

## Health

| Method | Path | Authentication | Description |
|--------|------|----------------|-------------|
| `GET` | `/server/healthz` | None | Frontend health page; negotiates HTML, JSON, or human-readable text |
| `GET` | `/healthz` | None | Optional direct alias for `/server/healthz`; disabled by default |
| `GET` | `/api/{api_version}/server/healthz` | None | API health JSON; accepts text with `Accept: text/plain` |
| `GET` | `/api/{api_version}/server/healthz.txt` | None | Dotted key/value API health text |
| `GET` | `/api/healthz` | None | Unversioned JSON health alias |
| `GET` | `/api/healthz.txt` | None | Unversioned dotted key/value API health text |

Health responses include public project and build information, runtime mode,
overlay feature state, component checks, and aggregate request statistics. A
`healthy`, `degraded`, or `restart_required` response uses HTTP 200;
`unhealthy`, `maintenance`, and `shutting_down` use HTTP 503. Overlay feature
statuses are limited to `disabled`, `starting`, `healthy`, or `unavailable`.
Health output never includes credentials, private addresses, filesystem paths,
or provider error details. The optional root alias is enabled with
`server.healthz.root.enabled: true`.

## Autodiscovery and CLI downloads

| Method | Path | Authentication | Description |
|--------|------|----------------|-------------|
| `GET` | `/api/autodiscover` | None | Returns the current API version, public origins, exposed overlay addresses, CLI release versions, and configured minimum client versions in the standard `ok`/`data` envelope. |
| `GET` | `/cli/binaries/{filename}` | None by default; optional bearer token | Streams a platform CLI binary only when its SHA-256 matches the operator-provisioned release manifest. |

CLI files are read from `{data_dir}/cli` by default. See
[Configuration](configuration.md) for the `yaml.d/maintenance.yml` settings.

## Authentication

Browser sign-in and account forms are under `/server/auth/`. API sign-in is
under `/api/{api_version}/server/auth/`.

| Method | API path | Purpose |
|--------|----------|---------|
| `POST` | `/api/{api_version}/server/auth/login` | Sign in a Server Admin or Regular User |
| `POST` | `/api/{api_version}/server/auth/login/2fa` | Complete a second-factor challenge |
| `POST` | `/api/{api_version}/server/auth/register` | Register a Regular User when registration is open |
| `POST` | `/api/{api_version}/server/auth/forgot-password` | Request password recovery |
| `POST` | `/api/{api_version}/server/auth/reset-password` | Set a password using a recovery token |
| `GET` | `/api/{api_version}/server/auth/verify-email/{token}` | Verify an email address |
| `POST` | `/api/{api_version}/server/auth/logout` | Revoke the authenticated API session |
| `POST` | `/api/{api_version}/server/auth/refresh` | Refresh the authenticated API session |
| `POST` | `/api/{api_version}/server/auth/2fa/enable` | Enable TOTP for the authenticated account |
| `POST` | `/api/{api_version}/server/auth/2fa/verify` | Verify a TOTP code |
| `POST` | `/api/{api_version}/server/auth/2fa/disable` | Disable TOTP for the authenticated account |

Registration, email verification, and password recovery depend on the
deployment's settings and working SMTP delivery. Browser forms use the
server-rendered pages and session cookies; API routes use bearer tokens.
Requests using a revoked opaque API token return HTTP `401` with
`{"error":"TOKEN_REVOKED", ...}`; expired tokens return `TOKEN_EXPIRED`.
Clients should discard the cached credential and require an explicit sign-in.

## Profiles, links, and short links

Regular User profile and link operations require a `usr_` bearer token.
Server Admin API operations require an `adm_` bearer token. A browser session
cookie does not authorize API requests.

| Method | Path | Authentication | Purpose |
|--------|------|----------------|---------|
| `GET` | `/api/{api_version}/users` | Regular User | Read the caller's account summary |
| `GET`, `POST` | `/api/{api_version}/users/profiles` | Regular User | List or create profiles owned by the caller |
| `GET`, `PATCH`, `DELETE` | `/api/{api_version}/users/profiles/{id}` | Regular User | Read, update, or delete an owned profile |
| `POST` | `/api/{api_version}/users/profiles/{id}/duplicate` | Regular User | Duplicate an owned profile |
| `GET`, `POST` | `/api/{api_version}/users/profiles/{id}/links` | Regular User | List or add links to an owned profile |
| `POST` | `/api/{api_version}/users/profiles/{id}/links/order` | Regular User | Reorder links for an owned profile |
| `POST` | `/api/{api_version}/users/profiles/import` | Regular User | Start an import for the caller |
| `GET` | `/api/{api_version}/users/profiles/import/{job_id}` | Regular User | Read import job status |
| `GET` | `/api/{api_version}/users/profiles/export/{profile_id}` | Regular User | Export an owned profile |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics` | Regular User | Read profile analytics |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics/links` | Regular User | Read link analytics |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics/export` | Regular User | Export analytics |
| `PUT`, `DELETE` | `/api/{api_version}/users/links/{id}` | Regular User | Update or delete an owned link |
| `POST` | `/api/{api_version}/users/links/{id}/toggle` | Regular User | Toggle an owned link's enabled state |
| `GET`, `PATCH` | `/api/{api_version}/users/settings` | Regular User | Read or update account settings |
| `GET`, `POST` | `/api/{api_version}/users/tokens` | Regular User | List or create API tokens |
| `GET`, `DELETE` | `/api/{api_version}/users/tokens/{token_id}` | Regular User | Read or revoke an owned token |
| `POST` | `/api/{api_version}/users/tokens/{token_id}/rotate` | Regular User | Rotate an owned token |
| `GET` | `/api/{api_version}/users/{username}` | Optional | Read a public account profile; private accounts return 404 to other callers |
| `GET` | `/api/{api_version}/users/{username}.txt` | Optional | Read the same public account profile as plain text |
| `GET` | `/api/{api_version}/profiles/{username}` | Optional | Read a public profile; private profiles return 404 to other users |
| `GET` | `/api/{api_version}/profiles/{username}/links` | Optional | Read public profile links |
| `GET` | `/api/{api_version}/profiles/{username}/qr` | Optional | Generate a profile QR image; accepts `format=png` (default) or `format=svg`, and `size=128`, `256` (default), `512`, or `1024` |
| `GET` | `/api/{api_version}/services` | None | List the service catalog |
| `GET` | `/api/{api_version}/services/search` | None | Search the service catalog |
| `GET` | `/api/{api_version}/services/categories` | None | List service categories |
| `GET` | `/api/{api_version}/services/popular` | None | List popular services |
| `GET` | `/api/{api_version}/services/{id}` | None | Read a service entry |
| `GET`, `POST` | `/api/{api_version}/users/shortlinks` | Regular User | List or create the caller's short links |
| `DELETE` | `/api/{api_version}/users/shortlinks/{code}` | Regular User | Delete an owned short link |
| `GET` | `/api/{api_version}/shortlinks/{code}` | None | Read public short-link information |

Public account profiles omit email unless the user explicitly enables
`show_email` in privacy settings. Private account profiles return `404` to
other visitors. The browser renders `/users/{username}` as HTML and responds
to `/users/{username}.txt` or `Accept: text/plain` with the same profile in
plain text.

The public short-link redirect is `GET /s/{code}`. The click redirect is
`GET /links/{id}/click`.

## Public server content API

These endpoints expose the operator-configured public content also shown on
server-rendered information pages:

| Method | Path | Purpose |
|--------|------|---------|
| `GET`, `POST` | `/api/{api_version}/server/contact` | Read contact details or submit a contact report |
| `GET` | `/api/{api_version}/server/about` | Read About content |
| `GET` | `/api/{api_version}/server/help` | Read Help content |
| `GET` | `/api/{api_version}/server/privacy` | Read privacy information |
| `GET` | `/api/{api_version}/server/terms` | Read terms |
| `GET` | `/api/{api_version}/server/security` | Read security information |
| `GET` | `/api/{api_version}/server/security/policy` | Read the security reporting policy |
| `GET` | `/api/{api_version}/server/security/thanks` | Read the security report receipt page |

The account and profile endpoints are described in the [Regular User guide](users.md).

## Server Admin API

Server Admin endpoints use
`/api/{api_version}/server/{admin_path}/...` and require an `adm_` bearer
token. The currently mounted API includes:

- `GET /users`, `GET /users/{id}`, `PUT /users/{id}`, and `DELETE /users/{id}`
- `GET /stats`
- `GET /settings` and `PUT /settings`
- `POST /backup`
- `POST /services/import` and `POST /cache/clear`
- `GET /email/config`, `PUT /email/config`, and `POST /email/test`
- `GET /notifications/preferences` and `PUT /notifications/preferences`
- `GET /security/allowlist`, `POST /security/allowlist`, `DELETE /security/allowlist/{id}`, and `GET /security/allowlist/check/{ip}`
- `GET /network/blocklists`, `PATCH /network/blocklists`, and `GET /network/blocklists/stats`
- `POST /network/blocklists/sources`, `PATCH /network/blocklists/sources/{name}`, and `DELETE /network/blocklists/sources/{name}`
- `POST /network/blocklists/update`, `POST /network/blocklists/update/{name}`, and `GET /network/blocklists/check/{ip}`
- `GET` and `PATCH /network/i2p`, plus `POST /network/i2p/validate`, `/network/i2p/restart`, and `/network/i2p/regenerate`

The setup API is under `/api/{api_version}/server/{admin_path}/setup/`. The
server prints the one-time setup token on first run; `POST /setup/verify`
validates it and issues a short-lived setup session cookie. That cookie
authorizes the remaining setup steps until completion. The server-rendered
forms post to same-origin CSRF-protected routes which call the same step
handlers.

| Method | Path suffix | Step |
|--------|-------------|------|
| `GET` | `/setup` | Read setup status |
| `POST` | `/setup/verify` | Verify the one-time setup token |
| `POST` | `/setup/account` | Create the primary Server Admin |
| `POST` | `/setup/token` | Create an admin-bound API token; raw value returned once |
| `POST` | `/setup/config` | Save app branding, domain, mode, and timezone |
| `POST` | `/setup/security` | Save optional backup encryption and admin TOTP settings |
| `POST` | `/setup/services` | Save HTTPS and Regular User registration settings |
| `POST` | `/setup/complete` | Invalidate setup token and create the admin browser session |

Setup pages and the operations available after setup are described in
[Administration](admin.md).

## Privacy and validation

Public profile endpoints return public profiles only. Private profiles return
404 to visitors; the owner may access their own profile. Profiles do not use
password protection. The server validates input and authorization before
changing data. Do not put credentials in query parameters or request logs.

For exact request fields and response schemas, use the OpenAPI document served
by the running version.
