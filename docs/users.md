# Regular User guide

Cassocial keeps Regular User accounts separate from Server Admin accounts.
Regular Users manage their own profiles and API tokens; they do not receive
Server Admin access. Registration is available only when the operator enables
it and any configured email verification and approval requirements are met.

## Sign in and dashboard

Use `/server/auth/register` to create an account when registration is open, or
`/server/auth/login` to sign in. The dashboard at `/users/` lists profiles
owned by the signed-in account. Browser pages are server-rendered and their
core forms work without JavaScript.

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

The server stores passwords using Argon2id. API tokens are displayed once at
creation; keep the raw token private and revoke it if it is exposed. Browser
sessions and API bearer tokens are separate authentication mechanisms.
