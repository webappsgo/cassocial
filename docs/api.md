# API Reference

Cassocial serves its JSON API at `/api/{api_version}/server/`; the default version is `v1`. Public API paths in this guide are relative because each operator runs their own server. Use your instance URL when making a request.

```bash
curl -q -LSsf https://YOUR_SERVER/api/v1/server/healthz
```

## API documentation

The interactive OpenAPI explorer is available at `/server/docs/swagger`. The machine-readable OpenAPI document is served at `/api/{api_version}/server/swagger`; `/api/swagger` resolves directly to the configured version. The GraphQL interface is available at `/server/docs/graphql`, with its API at `/api/{api_version}/server/graphql`.

## Health

| Method | Path | Authentication | Description |
|--------|------|----------------|-------------|
| `GET` | `/server/healthz` | None | Server-rendered health status |
| `GET` | `/api/v1/server/healthz` | None | JSON health status |
| `GET` | `/api/healthz` | None | Direct alias for the configured API version |

## Authentication

Authentication endpoints live under `/server/auth/` and `/api/{api_version}/server/auth/`. Login is shared by account types; the server redirects a successful browser login to the appropriate account area. The API login response identifies the account type and may require a second-factor step.

| Method | API path | Purpose |
|--------|----------|---------|
| `POST` | `/api/v1/server/auth/register` | Register a Regular User when open registration is enabled |
| `POST` | `/api/v1/server/auth/login` | Sign in a Regular User or Server Admin |
| `POST` | `/api/v1/server/auth/logout` | End the current session |
| `POST` | `/api/v1/server/auth/2fa` | Complete a second-factor challenge |
| `POST` | `/api/v1/server/auth/password/forgot` | Request password recovery |
| `POST` | `/api/v1/server/auth/password/reset` | Set a new password from a recovery token |
| `POST` | `/api/v1/server/auth/verify` | Verify an email address |
| `POST` | `/api/v1/server/auth/refresh` | Refresh an authenticated session |

Additional passkey, invite, external identity, and recovery routes are documented by the OpenAPI document for the running API version.

## Authorization

Send an API token in the `Authorization` header:

```bash
curl -q -LSsf \
  -H 'Authorization: Bearer YOUR_API_TOKEN' \
  https://YOUR_SERVER/api/v1/server/auth/check
```

Server Admin tokens begin with `adm_`; Regular User tokens begin with `usr_`. Tokens are scoped to their account type. A Server Admin bearer token is required for `/api/{api_version}/server/{admin_path}/...`; browser session cookies do not authorize API calls.

## Server Admin API

The API mirrors the administration pages under `/api/{api_version}/server/{admin_path}/...`. This family covers server settings, Regular User management, security, backups, logs, email, scheduler, updates, network, and other administration sections. Requests require a Server Admin token.

## Public profile data

Public endpoints return only public profile data. Private Regular User profiles are excluded from lists and searches; requests that identify a private profile return 404. The profile owner can access their own profile with authentication. Server Admin access is through the authenticated administration panel only.

Profile link and shortlink destinations are validated by the server. Unsafe schemes and private or local destinations are rejected.

## Request and error behavior

The server validates request bodies and authorization before changing data. Responses use JSON for API routes and the negotiated HTML or JSON format for paired browser/API routes. A rate-limited request returns HTTP `429`; callers should respect any `Retry-After` header.

Use the OpenAPI document for exact request fields, response schemas, and status codes for the configured version. Do not put credentials in query parameters or report bodies.
