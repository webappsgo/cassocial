# Integrations

Cassocial supports SMTP delivery, OIDC browser and API authentication, the
standard browser discovery endpoints, OpenAPI, and a public read-only GraphQL
API. OIDC providers must currently be configured in `server.yml`; the
Server Admin interface for managing providers is not implemented. LDAP and
SAML authentication, webhooks, and platform app-link verification are not
implemented.

## SMTP email

SMTP supports email verification, password recovery, and notification mail.
Settings are stored under `server.notifications.email.smtp` in `server.yml`;
the password is stored separately in `{config_dir}/secrets.yml`.

Server Admins can review and update SMTP settings at
`/server/{admin_path}/email`. The JSON API is:

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/{api_version}/server/{admin_path}/email/config` | Read SMTP settings without returning the password |
| `PUT` | `/api/{api_version}/server/{admin_path}/email/config` | Update SMTP settings |
| `POST` | `/api/{api_version}/server/{admin_path}/email/test` | Test the configured relay |

The SMTP password is never echoed in the configuration response. For
autodetection, TLS modes, and environment variables, see the
[SMTP example](examples/smtp.md) and [Configuration](configuration.md).

## OpenID Connect (OIDC)

OIDC providers are configured under `server.auth.oidc` in `server.yml` (or
the corresponding account configuration drop-in). The family and each
provider must be enabled. Provider settings are not yet editable in the admin
panel. Configure the client secret through the protected secrets configuration;
the generated configuration does not expose it in public settings output.

```yaml
server:
  auth:
    oidc:
      enabled: true
      providers:
        - name: example
          enabled: true
          display_name: Example Login
          issuer: https://identity.example
          client_id: cassocial
```

Store the provider secret in `{config_dir}/secrets.yml`, with mode `0600`:

```yaml
server:
  auth:
    oidc:
      providers:
        - name: example
          client_secret: REPLACE_WITH_PROVIDER_SECRET
```

The provider name must match the entry in `server.yml`. The server keeps
provider secrets in the protected secrets file rather than returning them in
settings output.

The server discovers provider endpoints from issuer metadata. Configure scopes,
PKCE, state and nonce checks, account registration, claims mapping, and role
mapping per provider. Account registration defaults to disabled; enable it only
when Regular User accounts should be created on first login. New external
accounts use the first-login username confirmation flow. Server Admin and
Regular User accounts remain separate, and a role mapping alone never grants
Server Admin access.

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/server/auth/oidc/{provider}` | Begin browser login |
| `GET` | `/server/auth/oidc/{provider}/callback` | Validate the provider response and complete browser login |
| `GET`, `POST` | `/server/auth/external/username` | Choose a username when first-login confirmation is required |
| `POST` | `/server/auth/oidc/{provider}/backchannel-logout` | Accept enabled OIDC back-channel logout notifications |
| `GET` | `/api/{api_version}/server/auth/oidc/{provider}` | Begin native-client login; returns a browser redirect URL and request ID |
| `POST` | `/api/{api_version}/server/auth/oidc/{provider}/callback` | Exchange the completed browser login for a session; returns `202` while pending |

Browser and API logout revoke the Cassocial session. When provider
RP-initiated logout is enabled and the provider exposes a valid end-session
endpoint, browser logout also redirects to the identity provider. API-mode
login uses a short-lived server-side transaction; the native client opens the
returned URL in a system browser and then polls the callback exchange endpoint.
Provider administration and LDAP/SAML flows remain incomplete.

## Discovery endpoints

| Path | Status |
|------|--------|
| `GET /api/autodiscover` | Available — API version, public origins, exposed overlay addresses, and checksum-verified CLI release metadata |
| `GET /cli/binaries/{filename}` | Available — serves a provisioned CLI binary after SHA-256 verification; public by default, optionally bearer-protected |
| `GET /robots.txt` | Available |
| `GET /sitemap.xml` | Available |
| `GET /manifest.json` | Available — PWA manifest |
| `GET /service-worker.js` | Available — service worker |
| `GET /.well-known/security.txt` | Available |
| `GET /server/docs/swagger` | Available — OpenAPI UI |
| `GET /api/{api_version}/server/swagger` | Available — OpenAPI document |
| `GET /server/docs/graphql` | Available — GraphiQL UI |
| `GET`, `POST /api/{api_version}/server/graphql` | Available — public read-only GraphQL endpoint |

The direct unversioned aliases `/api/swagger` and `/api/graphql` use the same
handlers as their configured-version endpoints. GraphQL exposes anonymous
public reads only and does not forward cookies or bearer credentials. Binary
downloads remain available through REST and are not GraphQL fields. Public
profile responses have selectable fields; responses without an authoritative
Go output type currently use the JSON scalar.

## Other integrations

- **OIDC provider management:** not implemented in the admin interface; configure providers in `server.yml` as described above.
- **LDAP and SAML:** authentication flows are not implemented.
- **Webhooks:** the model has event constants, but the router does not expose
  webhook management or delivery endpoints.
- **Android App Links and Apple Universal Links:** not implemented.
- **Federation and ActivityPub:** not implemented.
- **Tor:** the server checks for Tor at startup and attempts to start its
  managed hidden service when available. The onion address is published only
  when `server.tor.expose` is enabled.
- **I2P:** when enabled, the server prefers a dedicated `i2pd` process and
  falls back to an existing SAM bridge. The server persists its own destination
  key and serves the eepsite through a separate loopback listener. The optional
  address can be exposed through autodiscovery when `server.i2p.expose` is on.

## Short links

Short links are a Cassocial feature. An authenticated Regular User creates one
at `POST /api/{api_version}/users/shortlinks`; public redirects use
`GET /s/{code}`. The API requires a profile ID when creating a short link.

Cassocial works with local accounts and without SMTP. When no working relay is
available, email-dependent account flows cannot deliver messages.
