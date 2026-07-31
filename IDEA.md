## Project description

Cassocial is a self-hosted, open-source link aggregator and social profile platform. Users create a public profile page with links to their social media accounts, websites, and content — think Linktree, but self-hosted and fully open. Each user can own one or more named profiles, each with a unique URL slug, a customizable theme, an avatar, a bio, and an ordered list of links. A built-in admin panel lets server administrators manage users, profiles, services, and system settings. The application ships as a single self-contained binary with an embedded SQLite database and zero required external dependencies.

## Project variables

```
project_name:        cassocial
project_org:         casapps
internal_name:       cassocial
internal_org:        casapps
app_name:            Cassocial
official_site:       cassocial.example.com
maintainer_email:    casjay@yahoo.com
```

## Business logic

### Product scope & non-goals

**In scope:**
- Public profile pages reachable at `/{slug}` (or a verified custom domain)
- Per-profile ordered link cards with title, icon, click tracking, and optional username display
- Per-profile theming: background color/gradient/image, button style, font override, custom CSS
- QR code generation for any profile URL (PNG/SVG, configurable size and error-correction)
- Short-link creation tied to a profile
- Per-profile analytics: view count, click count, geographic breakdown (GeoIP), device type
- Import/export of profile link data (JSON)
- First-run setup wizard that creates the initial admin account and persists settings to the database
- Admin panel: user management (create, list, suspend, delete), system settings, backup, aggregate stats
- SMTP email for registration verification, password reset, and admin notifications
- Maintenance mode (system-wide or per-profile) with a bypass token
- 2FA (TOTP) for user accounts
- Password protection on individual profiles
- Accessibility pass (WCAG 2.1 AA): skip links, ARIA labels, keyboard navigation, reduced-motion support

**Non-goals:**
- Paid tiers, feature gating, or subscription billing
- Federation or ActivityPub
- Native mobile apps (responsive web only)
- Multi-tenancy with isolated database schemas per tenant
- Real-time collaboration or multi-user editing of a single profile
- Serving user-uploaded binary files (only URLs to external images are stored; avatars are external URLs)

### Roles & permissions

| Role | Description | Key Permissions |
|------|-------------|-----------------|
| **admin** | Server administrator | Full access: manage all users, all profiles, system settings, backup, GeoIP refresh |
| **user** | Registered account holder | Create and manage own profiles and links; view own analytics; change own password/2FA |
| **viewer** | Read-only account (future) | View own profile stats; cannot create links or profiles |
| **public** (unauthenticated) | Anyone on the internet | View public profiles at `/{slug}`; follow redirect links; cannot access `/dashboard` or `/admin` |

Admins are identified by `role = "admin"` in the `users` table. The first admin is created by the setup wizard. Subsequent admins are promoted via the admin panel. A `ServerAdmin` record in the store is the primary admin contact used for notifications; it mirrors the `users` row for the first admin.

An active user (`status = "active"`, `email_verified = true`) can log in. Suspended users (`status = "suspended"`) receive a clear error on login. Pending users (`status = "pending"`) must verify their email before the first login.

### Data model & sensitivity

**users** — account credentials and metadata. `password_hash` (Argon2id), `two_factor_secret`, `password_reset_token` are sensitive and never returned in API responses or logged.

**profiles** — public-facing landing pages. Each profile belongs to one user via `user_id`. `protection_password` (hashed) is sensitive. `custom_domain` is user-supplied and must be validated against DNS before `domain_verified` is set to true.

**links** — ordered list of clickable cards on a profile. `url` is user-supplied and must pass scheme + host validation; SSRF-risky schemes (`file://`, `javascript:`, internal hostnames) are rejected at write time.

**profile_theme** — per-profile visual customization. `custom_css` is user-supplied; it is rendered inside a `<style>` tag scoped to the profile page and must be sanitized server-side before storage.

**settings** — key-value configuration store. Contains SMTP credentials (`smtp_password`) and other secrets; these values are masked in logs and never returned to non-admin callers.

**analytics** — view and click events stored with a SHA-256 hashed IP (never raw IP), user-agent string, device type, referrer, and GeoIP country code. Retention is configurable (`analytics_retention_days`).

**shortlinks** — short codes that redirect to a target URL. Target URL is validated the same as link URLs. Optional expiry.

**Classification:**

| Sensitivity | Fields |
|-------------|--------|
| Secret | `password_hash`, `two_factor_secret`, `password_reset_token`, `protection_password`, `smtp_password`, `bypass_token` |
| Private | `email`, `last_login`, `two_factor_enabled`, raw analytics event IP |
| Internal | `role`, `status`, `created_at`, settings keys |
| Public | `username`, `slug`, `display_name`, `bio`, `avatar_url`, `links` (for public profiles) |

### Trust boundaries & external services

**SMTP (outbound only):** Used for email verification, password reset, and admin notifications. Credentials are stored in the `settings` table. Connection is made from the server to the configured SMTP host. Failure mode: email is not sent; registration and password-reset flows must degrade gracefully with a clear user message. SMTP host/port are operator-supplied and are trusted as server configuration.

**GeoIP database (MaxMind GeoLite2 or compatible):** Loaded from a local file path at startup. Used exclusively to derive a country code from a hashed visitor IP for analytics. The database file is operator-supplied; no runtime network fetch is performed. If the file is absent or stale, GeoIP enrichment silently returns an empty country code — analytics still record the event.

**QR code generation:** Performed entirely in-process using a Go library. No external network call. Output is a PNG or SVG byte slice served directly. Untrusted input is only the profile URL (already a validated URL owned by an authenticated user).

**Let's Encrypt / ACME (optional TLS):** The server can request and auto-renew TLS certificates via the ACME protocol. Operator must point DNS to the server before enabling. Failure mode: serve on HTTP only; log the error and notify the admin via email if SMTP is configured.

**Custom domains (user-supplied):** A profile owner can set a `custom_domain`. The domain is not trusted until `domain_verified = true`, which requires a DNS TXT record check performed server-side. Custom domain routing must not allow a user to claim a domain that resolves to an internal address (SSRF mitigation: resolve the domain and reject RFC 1918 / loopback addresses).

### Threat model & abuse cases

**Primary assets being protected:**
- User credentials (passwords, 2FA secrets, reset tokens)
- Private profile data (password-protected profiles, analytics)
- Server integrity (preventing privilege escalation, SSRF, malicious link injection)
- Availability (rate limits protect against brute force and scraping)

**Trusted inputs:**
- Authenticated admin session (role = "admin", active status, valid session cookie)
- Operator-supplied configuration file and environment variables
- Signed ACME challenges from Let's Encrypt

**Untrusted inputs:**
- All HTTP request bodies, query parameters, and headers
- User-supplied URLs (links, custom domains, shortlink targets)
- User-supplied profile content (display name, bio, custom CSS)
- GeoIP-derived country codes (informational only, never used as an access gate)

**Abuse cases and required defenses:**

| Threat | Defense |
|--------|---------|
| Credential stuffing on `/auth/login` | Rate limit: max 5 attempts per IP per 15-minute window; exponential backoff; account lockout after 10 failures; identical error message for wrong password vs unknown user |
| Spam account registration | Optional email verification (`email_verification_required = true`); optional admin approval (`registration_requires_approval = true`); registration rate limit per IP |
| Link scraping | Public profiles are intentionally public; click-tracking URLs go through a server redirect so the target URL is not in the raw HTML; robots.txt disallows `/api` and `/admin` |
| Privilege escalation | Role is stored server-side in the database; never derived from a request parameter; role check middleware applied to every `/admin` and `/api/admin` route |
| SSRF via custom domains | Domain verified by DNS TXT record; IP of resolved domain is checked against RFC 1918 / loopback ranges before `domain_verified` is set; custom domain routing does not perform outbound HTTP to the claimed domain |
| SSRF via link URLs | Link URLs validated for `http://` or `https://` scheme and non-empty host at write time; `file://`, `javascript:`, and bare internal hostnames are rejected |
| Malicious link URLs (phishing) | Links are presented as-is; the platform does not warrant link safety. Moderation tools allow admins to review and remove flagged links |
| Stored XSS via custom CSS | Custom CSS is stored verbatim but rendered inside a scoped `<style>` tag; the template engine HTML-escapes all user text fields; CSP header set to `default-src 'self'` on public profile pages |
| Session hijacking | Session tokens are random 32-byte values stored server-side (not in JWT); `HttpOnly`, `Secure`, `SameSite=Lax` flags; session TTL is configurable |
| Path traversal in static file serving | Static files served from an embedded filesystem; no user-controlled path segments reach the filesystem |
| Denial of service on analytics writes | Click and view events are written asynchronously with a bounded queue; if the queue is full, the event is silently dropped to protect the database |

### Security decisions & exceptions

- **Admin runtime privilege:** The server process does not drop privileges after binding to the configured port. The operator is responsible for running the binary under a non-root user account via a systemd unit or container entrypoint.
- **Custom CSS injection:** User-supplied `custom_css` is stored and rendered without a sandbox. This is an intentional product decision to enable rich profile customization. Only authenticated profile owners can set custom CSS for their own profiles; anonymous users cannot trigger its rendering path.
- **GeoIP country codes in analytics:** Country code is derived from a hashed IP using a local GeoIP database. It is used for informational analytics only, not as an access control gate. This is explicitly documented to satisfy the "GeoIP is a signal, not a gate" rule.
- **Viewer role:** The `viewer` role is defined in code constants but not yet surfaced in the UI. Assignment requires direct database manipulation until a future release exposes it. This is a known limitation, not a security exception.
