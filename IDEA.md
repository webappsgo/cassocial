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
| **Server Admin** | Administrative account, separate from Regular Users | Full access: manage users and profiles, system settings, backups, and GeoIP |
| **user** | Registered account holder | Create and manage own profiles and links; view own analytics; change own password/2FA |
| **viewer** | Read-only account (future) | View own profile stats; cannot create links or profiles |
| **public** (unauthenticated) | Anyone on the internet | View public profiles at `/{slug}`; follow redirect links; cannot access `/dashboard` or `/admin` |

Server Admin accounts are separate from Regular User accounts. The first Server Admin is created by the setup wizard; additional Server Admins are managed through the admin panel.

An active user (`status = "active"`, `email_verified = true`) can log in. Suspended users (`status = "suspended"`) receive a clear error on login. Pending users (`status = "pending"`) must verify their email before the first login.

### Data model & sensitivity

**users** — regular-user credentials and profile metadata. Password credentials, 2FA secrets, and password-reset credentials are sensitive and never exposed in responses or logs.

**profiles** — public-facing landing pages. Each profile belongs to one user via `user_id`. `protection_password` (hashed) is sensitive. `custom_domain` is user-supplied and must be validated against DNS before `domain_verified` is set to true.

**links** — ordered list of clickable cards on a profile. Only safe web destinations are accepted; active-content schemes and internal destinations are rejected.

**profile_theme** — per-profile visual customization. User styling cannot introduce executable content or affect other profiles.

System settings may contain SMTP credentials and other secrets. These values are private to Server Admins and are never exposed to Regular Users or the public.

**analytics** — view and click activity includes device type, referrer, and optional GeoIP country code. Raw visitor IP addresses are not retained, and retention is configurable.

**shortlinks** — short codes that redirect to a target URL. Target URL is validated the same as link URLs. Optional expiry.

**Classification:**

| Sensitivity | Fields |
|-------------|--------|
| Secret | `password_hash`, `two_factor_secret`, `password_reset_token`, `protection_password`, `smtp_password`, `bypass_token` |
| Private | `email`, `last_login`, `two_factor_enabled` |
| Internal | `role`, `status`, `created_at`, settings keys |
| Public | `username`, `slug`, `display_name`, `bio`, `avatar_url`, `links` (for public profiles) |

### Trust boundaries & external services

**SMTP (outbound only):** Used for email verification, password reset, and Server Admin notifications. The operator supplies and controls the SMTP service and credentials. If delivery fails, the user receives a clear message and the failure does not expose credentials.

**GeoIP database:** An operator-supplied local database may enrich analytics with a country code. GeoIP is informational only; if the database is unavailable, profile and analytics features remain usable without country enrichment.

**QR code generation:** Profile QR codes are generated by Cassocial without sending profile data to third parties.

**Let's Encrypt / ACME (optional TLS):** Operators may enable automatic certificates for domains they control. Certificate setup failures must not expose secrets or prevent access to the configured service.

**Custom domains (user-supplied):** A profile owner must prove control of a domain before it serves their profile. Domains resolving to private or local destinations cannot be claimed.

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
| Privilege escalation | Only Server Admins can access administrative controls; Regular Users cannot grant themselves higher privileges |
| SSRF via custom domains | Domain ownership is verified before use; private and local destinations cannot be claimed |
| SSRF via link URLs | Only safe web destinations are accepted; active-content schemes and internal destinations are rejected |
| Malicious link URLs (phishing) | Links are presented as-is; the platform does not warrant link safety. Moderation tools allow admins to review and remove flagged links |
| Stored XSS via custom CSS | User styling cannot execute scripts, escape its profile, or expose visitors to stored cross-site scripting |
| Session hijacking | Sessions are protected from script access and cross-site request abuse, and their lifetime is configurable |
| Path traversal in static file serving | User-controlled paths cannot expose private files or application data |
| Denial of service on analytics writes | Analytics collection is best-effort and may drop events under overload without interrupting profile access |
