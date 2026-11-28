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
alias `/api/swagger` serves the configured version. Both OpenAPI routes mount
the same handler; the served document uses the request's public origin and
the configured API-version path. Swagger UI and GraphiQL assets are pinned,
vendored, and served from this application.

The GraphiQL page is `/server/docs/graphql`. The public, read-only GraphQL
endpoint is `/api/{api_version}/server/graphql` and accepts POST JSON or GET
queries; `/api/graphql` is its direct alias. GraphQL exposes anonymous public
reads only and does not forward cookies or bearer credentials. A GET without a
`query` parameter returns a short usage message. Binary reads remain available
through their documented REST URL and are excluded from GraphQL fields.

Public profile and public profile-link responses have selectable GraphQL
fields. For example:

```graphql
query PublicProfile($slug: String!) {
  getPublicProfile(slug: $slug) {
    slug
    displayName
    bio
  }
}
```

```graphql
query PublicLinks($slug: String!) {
  listPublicProfileLinks(slug: $slug) {
    title
    url
    username
  }
}
```

The GraphQL profile types expose only public display data; owner identifiers,
profile ownership, and analytics counters stay private. Other public read
operations still return a JSON scalar and do not yet support nested field
selection.

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
| `POST` | `/api/{api_version}/server/auth/register` | Register a Regular User when multi-user mode is enabled and registration mode is `open`; unavailable in `private` mode |
| `POST` | `/api/{api_version}/server/auth/forgot-password` | Request password recovery |
| `POST` | `/api/{api_version}/server/auth/reset-password` | Set a password using a recovery token |
| `GET` | `/api/{api_version}/server/auth/verify-email/{token}` | Verify an email address |
| `POST` | `/api/{api_version}/server/auth/logout` | Revoke the authenticated API session |
| `POST` | `/api/{api_version}/server/auth/refresh` | Refresh the authenticated API session |
| `POST` | `/api/{api_version}/server/auth/2fa/enable` | Enable TOTP for the authenticated account |
| `POST` | `/api/{api_version}/server/auth/2fa/verify` | Verify a TOTP code |
| `POST` | `/api/{api_version}/server/auth/2fa/disable` | Disable TOTP for the authenticated account |

Registration depends on `server.users.enabled` and
`server.users.registration.mode`. Email verification and password recovery
depend on their settings and working SMTP delivery. Browser forms use the
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
| `GET`, `PATCH`, `DELETE` | `/api/{api_version}/users/profiles/{profile_id}` | Regular User | Read or update an owned profile; DELETE returns `204` |
| `POST` | `/api/{api_version}/users/profiles/{id}/duplicate` | Regular User | Duplicate an owned profile |
| `GET` | `/api/{api_version}/users/profiles/{id}/qr` | Regular User | Generate a QR image for an owned profile; accepts the documented `format` and `size` options |
| `POST` | `/api/{api_version}/users/profiles/{id}/verify-domain` | Regular User | Verify ownership of a custom domain for an owned profile |
| `GET`, `POST` | `/api/{api_version}/users/profiles/{profile_id}/links` | Regular User | List or add links to an owned profile |
| `POST` | `/api/{api_version}/users/profiles/{profile_id}/links/order` | Regular User | Reorder links for an owned profile |
| `POST` | `/api/{api_version}/users/profiles/import` | Regular User | Start an import for the caller |
| `GET` | `/api/{api_version}/users/profiles/import/{job_id}` | Regular User | Read import job status |
| `GET` | `/api/{api_version}/users/profiles/export/{profile_id}` | Regular User | Export an owned profile |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics` | Regular User | Read view and click totals, optional country breakdown, device type, and referrer aggregates |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics/links` | Regular User | Read click counts by link |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics/export` | Regular User | Export analytics |
| `PUT`, `DELETE` | `/api/{api_version}/users/links/{link_id}` | Regular User | Update an owned link; DELETE returns `204` |
| `POST` | `/api/{api_version}/users/links/{id}/toggle` | Regular User | Toggle an owned link's enabled state |
| `GET`, `PATCH` | `/api/{api_version}/users/settings` | Regular User | Read or update account settings |
| `GET` | `/api/{api_version}/users/notifications` | Regular User | List stored notifications; add `?unread=true` to filter unread entries |
| `POST` | `/api/{api_version}/users/notifications/{id}/read` | Regular User | Mark an owned notification as read |
| `POST` | `/api/{api_version}/users/notifications/read-all` | Regular User | Mark all of the caller's notifications as read |
| `GET`, `POST` | `/api/{api_version}/users/tokens` | Regular User | List or create API tokens |
| `GET`, `DELETE` | `/api/{api_version}/users/tokens/{token_id}` | Regular User | Read or revoke an owned token |
| `POST` | `/api/{api_version}/users/tokens/{token_id}/rotate` | Regular User | Rotate an owned token |
| `GET` | `/api/{api_version}/users/{username}` | Optional | Read a public account profile; private accounts return 404 to other callers |
| `GET` | `/api/{api_version}/users/{username}.txt` | Optional | Read the same public account profile as plain text |
| `GET` | `/api/{api_version}/profiles/{slug}` | Optional | Read public display fields (`slug`, `display_name`, `bio`, `avatar_url`, `header_image_url`, and `show_usernames`); private profiles return 404 to other users |
| `GET` | `/api/{api_version}/profiles/{slug}/links` | Optional | Read active public card fields (`title`, `username`, `url`, icon/colors, and display order); owner IDs, service IDs, and analytics counters are omitted |
| `GET` | `/api/{api_version}/profiles/{slug}/qr` | Optional | Generate a profile QR image; accepts `format=png` (default) or `format=svg`, and `size=128`, `256` (default), `512`, or `1024` |
| `GET` | `/api/{api_version}/services` | None | List the service catalog |
| `GET` | `/api/{api_version}/services/search` | None | Search the service catalog |
| `GET` | `/api/{api_version}/services/categories` | None | List service categories |
| `GET` | `/api/{api_version}/services/popular` | None | List popular services |
| `GET` | `/api/{api_version}/services/{service_id}` | None | Read a service entry |
| `GET`, `POST` | `/api/{api_version}/users/shortlinks` | Regular User | List or create the caller's short links |
| `DELETE` | `/api/{api_version}/users/shortlinks/{code}` | Regular User | Delete an owned short link; returns `204` |
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

## MCP server

Cassocial exposes an always-on remote Streamable HTTP endpoint over JSON-RPC 2.0.
Every request requires the same API Bearer token used by REST; there is no
anonymous MCP access, separate MCP credential, or session. The token scope
filters the tools available to the caller. The server builds tools from its
OpenAPI operations and dispatches tool calls through the existing API router.

| Method | Path | Authentication | Purpose |
|--------|------|----------------|---------|
| `POST` | `/api/{api_version}/server/mcp` | Bearer token | Canonical MCP endpoint |
| `POST` | `/api/mcp` | Bearer token | Direct alias for the configured API version |
| `POST` | `/mcp` | Bearer token | Root alias for the configured API version |
| `GET` | `/server/help/mcp` | None | Server-rendered connection guide and live tool list |
| `GET` | `/api/{api_version}/server/help/mcp` | None | JSON connection guide and tool list |

The MCP endpoint accepts `POST` only. It has no initialize handshake or
session state. Requests carry `MCP-Protocol-Version: 2026-07-28` and a
`Mcp-Method` matching the JSON-RPC method; tool calls also carry `Mcp-Name`.
Each alias returns the same JSON responses and never redirects. Use the
[MCP connection guide](mcp.md) for a copyable client configuration and request.

## Guest preferences

Guest preference endpoints configure cookies for the requesting browser; they
do not require an account or bearer token. API responses use the standard
`ok`/`data` envelope. Preference writes are gated by the browser's accepted
Preferences cookie category. The API mirror does not use CSRF tokens because
it accepts only UI preference cookies and has no ambient authentication.

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/{api_version}/server/preferences` | Read effective theme, language, and documented guest preferences |
| `POST` | `/api/{api_version}/server/preferences` | Apply a partial preference update from JSON or form data |
| `GET` | `/api/{api_version}/server/preferences/sync` | Read effective preferences and their shareable URL and short code |
| `GET` | `/api/{api_version}/server/preferences/sync/import` | Import valid preferences from query values or a shared code |

Import accepts a shared URL, query string, or padded or unpadded base64url
code. It is additive: omitted settings keep their existing cookies. Invalid
values are ignored, and preference-cookie writes still require local consent.
The import endpoint is intentionally a `GET` so shared links can be opened
directly; it can change only these guest-scoped UI cookies.

## Server Admin API

Server Admin endpoints use
`/api/{api_version}/server/{admin_path}/...` and require an `adm_` bearer
token. The currently mounted API includes:

- `GET /users`, `GET /users/{id}`, `PUT /users/{id}`, and `DELETE /users/{id}`
- `GET /stats`
- `GET /settings` and `PUT /settings`
- `GET /backup`, `POST /backup`, `GET /backup/{id}`, `DELETE /backup/{id}`, and `GET /backup/{id}/download`
- `POST /backup/restore` accepts `backup_file`, `confirm` (`RESTORE`), and an optional `password`; it verifies the archive before queueing a live restore and returns `202 Accepted`. Encrypted archives require `password`. The server enters maintenance mode, drains requests, restores the backup, and exits with the restart-required status so its supervisor can start it again.
- `POST /services/import` and `POST /cache/clear`
- `GET /email/config`, `PUT /email/config`, and `POST /email/test`
- `GET /{admin_username}/notifications` to list history (`?unread=true` filters unread items)
- `POST /{admin_username}/notifications/{id}/read` and `POST /{admin_username}/notifications/read-all` to mark entries read
- `GET` and `PATCH /{admin_username}/preferences` for the authenticated admin's appearance settings
- `GET` and `PATCH /{admin_username}/notifications/preferences` for the authenticated admin's notification settings
- `GET /security/allowlist`, `POST /security/allowlist`, `DELETE /security/allowlist/{id}`, and `GET /security/allowlist/check/{ip}`
- `GET /security/ratelimit` and `PATCH /security/ratelimit`
- `GET /security/tokens`, `DELETE /security/tokens/{id}`, and `POST /security/tokens/{id}/rotate`
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
