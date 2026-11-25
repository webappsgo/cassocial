# Integrations

Cassocial currently supports SMTP delivery and the standard browser discovery
endpoints. External identity providers, webhooks, GraphQL, and platform app-link
verification are not wired into the running server.

## SMTP email

SMTP supports email verification, password recovery, and notification mail.
Settings are stored under `server.notifications.email.smtp` in `server.yml`;
the password is stored separately in `{config_dir}/secrets.yml`.

Server Admins can review and update SMTP settings at
`/server/{admin_path}/smtp`. The JSON API is:

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/{api_version}/server/{admin_path}/smtp/config` | Read SMTP settings without returning the password |
| `PUT` | `/api/{api_version}/server/{admin_path}/smtp/config` | Update SMTP settings |
| `POST` | `/api/{api_version}/server/{admin_path}/smtp/test` | Test the configured relay |

The SMTP password is never echoed in the configuration response. For
autodetection, TLS modes, and environment variables, see the
[SMTP example](examples/smtp.md) and [Configuration](configuration.md).

## Discovery endpoints

| Path | Status |
|------|--------|
| `GET /robots.txt` | Available |
| `GET /sitemap.xml` | Available |
| `GET /manifest.json` | Available — PWA manifest |
| `GET /service-worker.js` | Available — service worker |
| `GET /.well-known/security.txt` | Available |
| `GET /server/docs/swagger` | Available — OpenAPI UI |
| `GET /api/{api_version}/server/swagger` | Available — OpenAPI document |

There is no GraphQL or GraphiQL route in the current server.

## Other integrations

- **OIDC, LDAP, and SAML:** not implemented.
- **Webhooks:** the model has event constants, but the router does not expose
  webhook management or delivery endpoints.
- **Android App Links and Apple Universal Links:** not implemented.
- **Federation and ActivityPub:** not implemented.
- **Tor:** the server checks for Tor at startup and attempts to start its
  managed hidden service when available.

## Short links

Short links are a Cassocial feature. An authenticated Regular User creates one
at `POST /api/{api_version}/users/shortlinks`; public redirects use
`GET /s/{code}`. The API requires a profile ID when creating a short link.

Cassocial works with local accounts and without SMTP. When no working relay is
available, email-dependent account flows cannot deliver messages.
