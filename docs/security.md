# Security

Cassocial validates requests on the server and treats profile content,
destination URLs, headers, and configuration input as untrusted. Server Admin
and Regular User accounts have separate records and authorization boundaries.

## Accounts and passwords

Passwords are stored as Argon2id hashes. Passwords, reset credentials, setup
tokens, and stored API token values never appear in ordinary API responses or
logs. The authorized first-run setup flow displays a newly generated API token
and optional TOTP recovery codes exactly once so the operator can save them.
A Regular User role does not grant access to Server Admin routes.

## Tokens and sessions

API tokens are sent in the `Authorization: Bearer` header. Server Admin tokens
use the `adm_` prefix; Regular User tokens use `usr_`. Keep tokens in
permission-restricted files and revoke them when they are no longer needed.

Browser sessions use separate admin and user cookies. Those cookies authorize
browser pages only; API routes require their documented bearer token. Session
cookies are `HttpOnly`, use `SameSite=Strict`, and are marked `Secure` for
HTTPS requests.

## CSRF protection

Frontend forms receive a double-submit CSRF token in the `csrf_token` cookie
and hidden form field. State-changing frontend forms, including public forms
such as login, registration, and contact, require a valid token when the
request is cross-site or its origin is unknown. Same-origin mutations with a
Strict session cookie can rely on the browser's cookie posture. API routes and
Bearer-authenticated requests are outside frontend CSRF validation.

Do not disable `web.csrf.enabled` on deployments that serve browser forms.
Use `web.csrf.exempt_paths` only for a required frontend exception; API routes
are already outside the CSRF scope.

## Private profiles and links

Private Regular User profiles return 404 to other visitors. The owner can view
their own profile after signing in; Server Admin access is through the
authenticated administration interface. Profiles do not have password
protection. Private profiles and links must not be exposed through public API
responses, search, QR codes, or click redirects.

Profile link and short-link destinations are validated by the server. Avoid
active-content schemes and internal or local destinations. Custom domains
must be verified before they route to a profile.

## Data and operations

Analytics avoid retaining raw visitor IP addresses. GeoIP uses locally stored,
MMDB-validated ASN, country, and DB-IP city datasets from ip-location-db.
Downloads are checked for the expected database type and metadata and must
answer a known public-address lookup before installation. Country filtering is
an optional fail-open risk signal; authentication and authorization remain
required. The IP allowlist bypasses country filtering,
and private/internal or loopback addresses are never looked up. For Tor
requests, only the observed exit address may be evaluated; the visitor's real
IP is never inferred or logged. Configure countries and operator-owned presets
at `/server/{admin_path}/network/geoip`. The public `/server/about` page and
the repository's LICENSE.md include the required DB-IP attribution.
Trusted IP addresses and CIDRs can also be managed at
`/server/{admin_path}/security/allowlist`. Allowlisted requests bypass IP
blocklists, rate limits, and GeoIP country filtering, but still require normal
authentication and authorization. YAML entries are operator-managed; entries
added in the administration page are persisted in the server database. The
admin API exposes list, add, delete, and address-check operations under
`/api/{api_version}/server/{admin_path}/security/allowlist`.
The built-in `firehol_level1` IP blocklist and any enabled HTTPS IP or domain
sources are checked before rate limiting and request handlers. Source health
checks appear in the server health endpoint; failed refreshes keep the last
valid source data active. Administrators can manage sources and view aggregate
rule statistics at `/server/{admin_path}/network/blocklists`.
I2P eepsite traffic reaches a dedicated loopback-only listener separate from
the clearnet port. Untrusted `X-I2P-*` headers are stripped, and I2P visitors
share a global rate-limit bucket because the SAM forwarder connects from the
same loopback address. The I2P destination key is kept in a `0600` file under
`{data_dir}/i2p/site`.
SMTP passwords are stored in `{config_dir}/secrets.yml`; do not commit that
file or copy credentials into issue reports. Protect configuration, data, and
backup directories with operating-system permissions appropriate to the
deployment.

The public health endpoints include `GET /server/healthz`,
`GET /api/{api_version}/server/healthz`, and the versionless `GET /api/healthz`.
The root `GET /healthz` alias is disabled unless `server.healthz.root.enabled`
is true. Health reports public service state and aggregate counters; they never
return credentials, private addresses, filesystem paths, or provider error
details. Do not include administration, authentication, or account pages in
public sitemaps.

## Report a vulnerability

Use `/server/security` to review the coordinated disclosure policy and submit
a report through `/server/contact`. Researchers can use the one-time status
link in the acknowledgment email to check triage updates. Administrators review
reports at `/server/{admin_path}/security/reports`; closed reports are retained
for the configured `server.security.reports.retention_days` period (365 days
by default).

Use [GitHub private vulnerability reporting](https://github.com/webappsgo/cassocial/security/advisories/new).
Do not post exploitable details in a public issue.
