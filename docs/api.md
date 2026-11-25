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
| `GET` | `/server/healthz` | None | HTML health status |
| `GET` | `/api/{api_version}/server/healthz` | None | JSON health status |
| `GET` | `/api/healthz` | None | Alias for the configured JSON health endpoint |

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

## Profiles, links, and short links

Regular User profile and link operations require a `usr_` bearer token.
Server Admin API operations require an `adm_` bearer token. A browser session
cookie does not authorize API requests.

| Method | Path | Authentication | Purpose |
|--------|------|----------------|---------|
| `GET`, `POST` | `/api/{api_version}/users/profiles` | Regular User | List or create profiles owned by the caller |
| `GET`, `PATCH`, `DELETE` | `/api/{api_version}/users/profiles/{id}` | Regular User | Read, update, or delete an owned profile |
| `POST` | `/api/{api_version}/users/profiles/{id}/duplicate` | Regular User | Duplicate an owned profile |
| `GET`, `POST` | `/api/{api_version}/users/profiles/{id}/links` | Regular User | List or add links to an owned profile |
| `POST` | `/api/{api_version}/users/profiles/{id}/links/order` | Regular User | Reorder links for an owned profile |
| `PUT`, `DELETE` | `/api/{api_version}/users/links/{id}` | Regular User | Update or delete an owned link |
| `GET` | `/api/{api_version}/users/{username}` | Optional | Read a public account profile; private accounts return 404 to other callers |
| `GET` | `/api/{api_version}/users/{username}.txt` | Optional | Read the same public account profile as plain text |
| `GET` | `/api/{api_version}/profiles/{username}` | Optional | Read a public profile; private profiles return 404 to other users |
| `GET` | `/api/{api_version}/profiles/{username}/links` | Optional | Read public profile links |
| `GET` | `/api/{api_version}/profiles/{username}/qr` | Optional | Generate a profile QR code |
| `GET` | `/api/{api_version}/services` | None | List the service catalog |
| `GET` | `/api/{api_version}/services/search` | None | Search the service catalog |
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

## Server Admin API

Server Admin endpoints use
`/api/{api_version}/server/{admin_path}/...` and require an `adm_` bearer
token. The currently mounted API includes:

- `GET /users`, `GET /users/{id}`, `PUT /users/{id}`, and `DELETE /users/{id}`
- `GET /stats`
- `GET /settings` and `PUT /settings`
- `POST /backup`
- `POST /services/import` and `POST /cache/clear`
- `GET /smtp/config`, `PUT /smtp/config`, and `POST /smtp/test`
- `GET /notifications/preferences` and `PUT /notifications/preferences`

The setup API is under `/api/{api_version}/server/{admin_path}/setup/`. Setup
pages and the operations available after setup are described in
[Administration](admin.md).

## Privacy and validation

Public profile endpoints return public profiles only. Private profiles return
404 to visitors; the owner may access their own profile. Profiles do not use
password protection. The server validates input and authorization before
changing data. Do not put credentials in query parameters or request logs.

For exact request fields and response schemas, use the OpenAPI document served
by the running version.
