# Regular User guide

Cassocial keeps Regular User accounts separate from Server Admin accounts.
Regular Users manage their own profiles and API tokens; they do not receive
Server Admin access. Multi-user mode is disabled by default. When enabled,
public self-registration defaults to `open`; the operator can set the mode to
`private` to disable the public registration form. Email verification applies
only when `server.users.registration.require_email_verification` is `true` and
SMTP is working. If verification is disabled or SMTP is unavailable, email
addresses are auto-verified. There is no admin-approval step in the
registration flow. The account settings live in
`{config_dir}/yaml.d/accounts.yml`.

## Sign in and dashboard

Use `/server/auth/register` to create an account when `server.users.enabled` is
true and `server.users.registration.mode` is `open`, or use
`/server/auth/login` to sign in. In `private` mode, existing users can still
sign in and a Server Admin creates accounts through an invite or direct-create
flow; both let the user set their own password through a one-time link. The
dashboard at `/users/` lists profiles owned by the signed-in account. Browser
pages are server-rendered and their core forms work without JavaScript.

## Profiles and links

Create a profile at `/users/profiles/new`, edit it at
`/users/profiles/{id}`, and manage its links at
`/users/profiles/{id}/links`. A profile can be public or private according to
its visibility setting. Private profiles return 404 to other visitors. Public
profiles are available at `/{slug}`; the account profile page is
`/users/{username}`. Public account pages also provide `/users/{username}.txt`.

The profile form controls its title, slug, description, visibility, and theme
where available. Link forms accept a destination URL and display information,
and support reordering and removal. The server validates ownership and all
submitted values before saving changes.

## Account settings and security

| Path | Purpose |
|------|---------|
| `/users/settings` | Account settings |
| `/users/settings/privacy` | Profile visibility and privacy choices |
| `/users/settings/notifications` | Notification preferences |
| `/users/settings/appearance` | Appearance preferences |
| `/users/tokens` | Create, rotate, and revoke API tokens |
| `/users/security/sessions` | Review and revoke active browser sessions |

Notification preferences are set per event and channel. Optional events have
separate WebUI and email choices; required security events stay enabled and
are not stored as user overrides. Email preferences apply only to events that
have an email template, and delivery still depends on working SMTP. Session
expiry, API token creation/revocation, and recovery-key use are WebUI events.
Email verification messages belong to the account-verification flow.

The server stores passwords using Argon2id. API tokens are displayed once at
creation; keep the raw token private and revoke it if it is exposed. Browser
sessions and API bearer tokens are separate authentication mechanisms.
Date and time formats are part of `/users/settings/appearance` alongside the
theme, font size, and reduced-motion preference.

## Guest preferences

Use `/server/preferences` to set the browser's theme and language, manage
cookie-consent choices, and opt out of sale or sharing where applicable. These
preferences are stored in cookies on this device and are independent of a
Regular User account. Preference cookies are written only after the browser
accepts the Preferences cookie category.

The page at `/server/preferences/sync` shows a shareable URL and short code
for transferring theme, language, and documented application preferences to
another device. Open the shared link or paste its URL, query string, or code
into the import form. Import applies valid values additively and respects the
destination device's consent choice. Reset removes this site's noncredential
browser storage while leaving authentication cookies alone.
