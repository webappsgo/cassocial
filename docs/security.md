# Security

Cassocial treats public requests, profile content, destination URLs, and uploaded configuration as untrusted input. Validation and authorization happen on the server. The product uses separate Server Admin and Regular User account boundaries.

## Accounts and passwords

Passwords are stored as Argon2id hashes. Passwords, reset credentials, recovery codes, setup tokens, and API tokens must never appear in API responses or logs. Server Admins are stored separately from Regular Users. A Regular User role does not grant access to Server Admin routes.

## Tokens and sessions

API tokens are presented as bearer credentials. Server Admin tokens use the `adm_` prefix; Regular User tokens use `usr_`. Tokens are scoped to their account type and owner. Store tokens in a permission-restricted file, do not share them, and revoke them when no longer needed.

Browser sign-in uses secure session handling. Administrative and user sessions remain separate. The administration interface uses its session cookie; API clients use bearer tokens. A browser cookie is not an API authorization credential.

## Private profiles and link destinations

Private Regular User profiles are not disclosed through public routes. Anonymous direct requests return 404, and public searches, lists, links, QR endpoints, and click tracking must not reveal the profile. The owner may access their own profile; Server Admin access is available only through the authenticated administration panel.

Profile links and shortlink destinations accept safe web URLs. Active-content schemes and internal or local destinations are rejected. Custom domains must be verified before routing to a profile and cannot resolve to private or local addresses.

## Data handling

Analytics do not retain raw visitor IP addresses. GeoIP is informational and does not control access. Settings secrets are visible only to Server Admins. QR generation stays within Cassocial and does not send profile details to external services.

## Health and operational endpoints

The public health page is `GET /server/healthz`; the JSON endpoint is `GET /api/v1/server/healthz`. Health responses report service state without returning credentials or internal error details. Administrative, authentication, and account pages must not be included in public sitemaps or indexing.

## Report a vulnerability

Use [GitHub private vulnerability reporting](https://github.com/webappsgo/cassocial/security/advisories/new). Do not post exploitable details in a public issue.
