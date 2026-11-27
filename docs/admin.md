# Server Administration

Server Admin accounts manage the Cassocial installation. They are separate
from Regular User accounts and use a separate browser session. The default
administration path is `/server/administration`; set `server.admin_path` to
change it.

## First Server Admin

On first start, Cassocial prints a one-time setup token and the setup URL.
Open `/server/{admin_path}` and submit the token. The server then guides you
through creating the primary Server Admin, generating an admin-bound API
token, setting the app name/domain/mode/timezone, optionally configuring
encrypted scheduled backups and admin TOTP, choosing HTTPS and Regular User
registration, and completing setup. Setup forms are server-rendered and include
CSRF protection; JavaScript only enhances token copying. The API token and TOTP
recovery codes are displayed once. Completion invalidates the setup token and
signs the new primary admin into the browser. The primary account cannot be
deleted.

## Available pages

The current browser administration interface includes these pages:

| Path | Purpose |
|------|---------|
| `/server/{admin_path}/` | Server dashboard and overview |
| `/server/{admin_path}/settings` | Update the settings exposed by the current form |
| `/server/{admin_path}/email` | Configure and test SMTP |
| `/server/{admin_path}/backup` | Review and trigger backup operations |
| `/server/{admin_path}/network/i2p` | Configure the I2P eepsite, inspect provider status, restart, or regenerate its address |
| `/server/{admin_path}/network/geoip` | Configure country filtering and GeoIP dataset presets |
| `/server/{admin_path}/network/blocklists` | Manage built-in and operator-defined IP/domain blocklists |
| `/server/{admin_path}/security/allowlist` | Manage trusted IP/CIDR entries and check an address |
| `/server/{admin_path}/users` | Manage Regular User status and delete Regular User accounts |
| `/server/{admin_path}/security/reports` | Review encrypted vulnerability reports and update triage status |
| `POST /server/{admin_path}/{admin_username}/logout` | Revoke the authenticated Server Admin's browser session |

The shared pages `/server/about`, `/server/privacy`, `/server/contact`,
`/server/help`, `/server/terms`, and `/server/security` show installation-level
information. Terms and contact details are controlled by the server operator.

## API access

The administration API is under
`/api/{api_version}/server/{admin_path}/...` and requires a Server Admin bearer
token beginning with `adm_`. The currently mounted endpoints cover Regular
User management, settings, statistics, backup trigger, service import, cache
clearing, SMTP, notification preferences, I2P, blocklists, and allowlists. GeoIP
is currently managed through its browser page; no GeoIP admin API is mounted. See the
[API reference](api.md#server-admin-api) for the route list.

The I2P API supports `GET` and `PATCH` on
`/api/{api_version}/server/{admin_path}/network/i2p`, plus `POST` operations at
`/validate`, `/restart`, and `/regenerate`. Regeneration creates a new I2P
destination and changes the public address.

Browser session cookies do not authorize API requests. Regular User accounts
and Server Admin accounts are stored and authorized separately; a Regular User
role never grants Server Admin access.

## Regular User management

When multi-user support is active, Server Admins can review user accounts,
change account status, and delete accounts from the Users page. The user
management API is available under
`/api/{api_version}/server/{admin_path}/users`. Users manage their own profiles
under `/users/`; private profiles return 404 to other visitors, including
through public API routes.

## Regular User browser pages

Regular Users sign in through `/server/auth/login`. The server-rendered account
area provides:

| Path | Purpose |
|------|---------|
| `/users/` | Dashboard and owned profiles |
| `/users/profiles/new` | Create a profile |
| `/users/profiles/{id}` | Edit profile details |
| `/users/profiles/{id}/links` | Add, edit, remove, and order profile links |
| `/users/settings` | Account settings |
| `/users/settings/privacy` | Public profile and privacy preferences |
| `/users/settings/notifications` | Notification preferences |
| `/users/settings/appearance` | Theme and appearance preferences |
| `/users/tokens` | Create, rotate, and revoke API tokens |
| `/users/security/sessions` | Review and revoke browser sessions |
| `/users/{username}` | Public account profile, when visible |

Forms work without JavaScript and use session authentication and CSRF
protection. API tokens are shown only once when created; store them securely.

## Configuration safety

SMTP passwords and other credentials are not returned by settings APIs or
forms. Do not include setup tokens, passwords, or API tokens in support
requests, screenshots, or logs. See [Security](security.md) for account and
token handling.
