## Project description

Cassocial is a self-hosted, open-source link aggregator and social profile platform. Users create a public profile page with links to their social media accounts, websites, and content — think Linktree, but self-hosted and fully open. Each user can own one or more named profiles, each with a unique URL slug, a customizable theme, an avatar, a bio, and an ordered list of links. A built-in admin panel lets server administrators manage users, profiles, services, and system settings. The application is distributed as a single self-contained binary.

## Project variables

```
project_name:        cassocial
project_org:         webappsgo
internal_name:       cassocial
internal_org:        casapps
app_name:            Cassocial
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
| **Server Admin** | Administrative account, separate from Regular Users | Full access: manage users and profiles, system settings, backups, and GeoIP |
| **user** | Registered account holder | Create and manage own profiles and links; view own analytics; change own password/2FA |
| **viewer** | Read-only account (future) | View own profile stats; cannot create links or profiles |
| **public** (unauthenticated) | Anyone on the internet | View public profiles at `/{slug}`; follow redirect links; cannot access `/dashboard` or `/admin` |

Server Admin accounts are separate from Regular User accounts. The first Server Admin is created by the setup wizard; additional Server Admins are managed through the admin panel.

An active user (`status = "active"`, `email_verified = true`) can log in. Suspended users (`status = "suspended"`) receive a clear error on login. Pending users (`status = "pending"`) must verify their email before the first login.

### Data model & sensitivity

**users** — regular-user credentials and profile metadata. Password credentials, 2FA secrets, and password-reset credentials are sensitive and never exposed in responses or logs.

**profiles** — public-facing landing pages. Each profile belongs to one user via `user_id`. Visibility is public or private; private profiles return 404 to anyone other than their owner. Profiles do not have password protection. `custom_domain` is user-supplied and must be validated against DNS before `domain_verified` is set to true.

**links** — ordered list of clickable cards on a profile. Only safe web destinations are accepted; active-content schemes and internal destinations are rejected.

**profile_theme** — per-profile visual customization. User styling cannot introduce executable content or affect other profiles.

System settings may contain SMTP credentials and other secrets. These values are private to Server Admins and are never exposed to Regular Users or the public.

**analytics** — view and click activity includes device type, referrer, and optional GeoIP country code. Raw visitor IP addresses are not retained, and retention is configurable.

**shortlinks** — short codes that redirect to a target URL. Target URL is validated the same as link URLs. Optional expiry.

**Classification:**

| Sensitivity | Fields |
|-------------|--------|
| Secret | `password_hash`, `two_factor_secret`, `password_reset_token`, `smtp_password`, `bypass_token` |
| Private | `email`, `last_login`, `two_factor_enabled` |
| Internal | `role`, `status`, `created_at`, settings keys |
| Public | `username`, `slug`, `display_name`, `bio`, `avatar_url`, `links` (for public profiles) |

### API routes

These are Cassocial-specific profile, link, service-catalog, analytics, and
short-link operations. The route families supplied by AI.md (authentication,
current-user accounts, custom-domain management, and Server Admin) are defined
there and are not duplicated here. `{api_version}` is the configured API
version. Public profile routes return only fields marked public; private
profiles and their links return `404` to callers other than the owner.

| Method | Route | Authentication | Request and response |
|--------|-------|----------------|----------------------|
| `GET` | `/api/{api_version}/users/profiles` | Regular User bearer token (`usr_`) | No body; returns the caller's profiles. |
| `POST` | `/api/{api_version}/users/profiles` | Regular User bearer token (`usr_`) | JSON profile fields: `slug`, `display_name`, optional `bio`, `avatar_url`, `header_image_url`, `show_usernames`, `visibility`, `meta_title`, and `meta_description`; returns the created profile. |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}` | Regular User bearer token (`usr_`) | No body; returns the caller-owned profile. |
| `PATCH` | `/api/{api_version}/users/profiles/{profile_id}` | Regular User bearer token (`usr_`) | JSON object containing fields to update; returns the updated profile. |
| `DELETE` | `/api/{api_version}/users/profiles/{profile_id}` | Regular User bearer token (`usr_`) | No body; deletes the caller-owned profile and returns `204`. |
| `GET` | `/api/{api_version}/profiles/{slug}` | None | No body; returns the public profile representation or `404`. |
| `GET` | `/api/{api_version}/profiles/{slug}/links` | None | No body; returns active public links in display order or `404`. |
| `GET` | `/api/{api_version}/profiles/{slug}/qr` | None for a public profile; owner token for a private profile | Optional `format` (`png` or `svg`) and `size` query values; returns the generated QR image. |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/links` | Regular User bearer token (`usr_`) | No body; returns the caller-owned profile's ordered links. |
| `POST` | `/api/{api_version}/users/profiles/{profile_id}/links` | Regular User bearer token (`usr_`) | JSON link fields: `title`, `url`, optional `service_id`, `username`, `icon_url`, `background_color`, and `text_color`; returns the created link. |
| `PUT` | `/api/{api_version}/users/links/{link_id}` | Regular User bearer token (`usr_`) | JSON object containing link fields to update; returns the updated link. |
| `DELETE` | `/api/{api_version}/users/links/{link_id}` | Regular User bearer token (`usr_`) | No body; deletes the caller-owned link and returns `204`. |
| `POST` | `/api/{api_version}/users/profiles/{profile_id}/links/order` | Regular User bearer token (`usr_`) | JSON `{ "link_ids": ["..."] }`; returns the updated link order. |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics` | Regular User bearer token (`usr_`) | Optional `period` query value (`day`, `week`, `month`, `year`, or `all`); returns view, click, geography, device, and referrer aggregates. |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics/links` | Regular User bearer token (`usr_`) | No body; returns click counts by link. |
| `GET` | `/api/{api_version}/users/profiles/{profile_id}/analytics/export` | Regular User bearer token (`usr_`) | Returns the profile analytics export in the requested supported format. |
| `GET` | `/api/{api_version}/services` | None | Optional `category`, `limit`, and `offset`; returns active services. |
| `GET` | `/api/{api_version}/services/search` | None | Search query and optional category/pagination values; returns matching active services. |
| `GET` | `/api/{api_version}/services/categories` | None | No body; returns service categories. |
| `GET` | `/api/{api_version}/services/popular` | None | No body; returns popular active services. |
| `GET` | `/api/{api_version}/services/{service_id}` | None | No body; returns one active service or `404`. |
| `GET` | `/api/{api_version}/users/shortlinks` | Regular User bearer token (`usr_`) | No body; returns the caller's short links. |
| `POST` | `/api/{api_version}/users/shortlinks` | Regular User bearer token (`usr_`) | JSON `{ "profile_id": "...", "url": "https://...", "custom_code": "...", "expires_in": 0 }`; returns the created short link and its public short URL. |
| `DELETE` | `/api/{api_version}/users/shortlinks/{code}` | Regular User bearer token (`usr_`) | No body; deletes the caller-owned short link and returns `204`. |
| `GET` | `/api/{api_version}/shortlinks/{code}` | None | No body; returns public short-link metadata or `404`. |
| `GET` | `/api/{api_version}/links/{link_id}/click` | None | No body; records an allowed click and redirects to the validated link destination. |

Import and export use JSON payloads and are owned by the authenticated user:

| Method | Route | Authentication | Request and response |
|--------|-------|----------------|----------------------|
| `POST` | `/api/{api_version}/users/profiles/import` | Regular User bearer token (`usr_`) | JSON `{ "source": "json", "data": { ... } }`; returns an import job identifier and result. |
| `GET` | `/api/{api_version}/users/profiles/import/{job_id}` | Regular User bearer token (`usr_`) | No body; returns the caller-owned import result. |
| `GET` | `/api/{api_version}/users/profiles/export/{profile_id}` | Regular User bearer token (`usr_`) | Returns the caller-owned profile and links as JSON. |

### Trust boundaries & external services

**SMTP (outbound only):** Used for email verification, password reset, and Server Admin notifications. The operator supplies and controls the SMTP service and credentials. If delivery fails, the user receives a clear message and the failure does not expose credentials.

**GeoIP:** The server downloads checksum-verified ASN, country, IPv4 city, and IPv6 city MMDB datasets from sapics/ip-location-db on first run and refreshes them weekly. GeoIP is enabled by default, with country allow/deny lists empty. Country filtering is a fail-open risk signal that never replaces authentication or authorization; private/internal and Tor client addresses are not looked up.

**QR code generation:** Profile QR codes are generated by Cassocial without sending profile data to third parties.

**Let's Encrypt / ACME (optional TLS):** Operators may enable automatic certificates for domains they control. Certificate setup failures must not expose secrets or prevent access to the configured service.

**Custom domains (user-supplied):** A profile owner must prove control of a domain before it serves their profile. Domains resolving to private or local destinations cannot be claimed.

### Threat model & abuse cases

**Primary assets being protected:**
- User credentials (passwords, 2FA secrets, reset tokens)
- Private profile data (private profiles and analytics)
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
- GeoIP country codes derived from public client IP addresses (no raw IP retention)

**Abuse cases and required defenses:**

| Threat | Defense |
|--------|---------|
| Credential stuffing on `/auth/login` | Rate limit: max 5 attempts per IP per 15-minute window; exponential backoff; account lockout after 10 failures; identical error message for wrong password vs unknown user |
| Spam account registration | Optional email verification (`email_verification_required = true`); optional admin approval (`registration_requires_approval = true`); registration rate limit per IP |
| Link scraping | Public profiles are intentionally public; click-tracking URLs go through a server redirect so the target URL is not in the raw HTML; robots.txt disallows `/api` and `/admin` |
| Privilege escalation | Only Server Admins can access administrative controls; Regular Users cannot grant themselves higher privileges |
| SSRF via custom domains | Domain ownership is verified before use; private and local destinations cannot be claimed |
| SSRF via link URLs | Only safe web destinations are accepted; active-content schemes and internal destinations are rejected |
| Malicious link URLs (phishing) | Links are presented as-is; the platform does not warrant link safety. Moderation tools allow admins to review and remove flagged links |
| Stored XSS via custom CSS | User styling cannot execute scripts, escape its profile, or expose visitors to stored cross-site scripting |
| Session hijacking | Sessions are protected from script access and cross-site request abuse, and their lifetime is configurable |
| Path traversal in static file serving | User-controlled paths cannot expose private files or application data |
| Denial of service on analytics writes | Analytics collection is best-effort and may drop events under overload without interrupting profile access |
