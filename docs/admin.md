# Server Administration

Server Admin accounts manage the Cassocial installation. They are separate
from Regular User accounts and use a separate browser session. The default
administration path is `/server/administration`; set `server.admin_path` to
change it.

## First Server Admin

On first start, Cassocial prints a one-time setup token and the setup URL.
Open `/server/{admin_path}/setup` and use the token to create the primary
Server Admin. The primary account cannot be deleted.

## Available pages

The current browser administration interface includes these pages:

| Path | Purpose |
|------|---------|
| `/server/{admin_path}/` | Server dashboard and overview |
| `/server/{admin_path}/settings` | Update the settings exposed by the current form |
| `/server/{admin_path}/smtp` | Configure and test SMTP |
| `/server/{admin_path}/users` | Manage Regular User status and delete Regular User accounts |
| `/server/{admin_path}/security/reports` | Review encrypted vulnerability reports and update triage status |

The shared pages `/server/about`, `/server/privacy`, `/server/contact`,
`/server/help`, `/server/terms`, and `/server/security` show installation-level
information. Terms and contact details are controlled by the server operator.

## API access

The administration API is under
`/api/{api_version}/server/{admin_path}/...` and requires a Server Admin bearer
token beginning with `adm_`. The currently mounted endpoints cover Regular
User management, settings, statistics, backup trigger, service import, cache
clearing, SMTP, and notification preferences. See the
[API reference](api.md#server-admin-api) for the route list.

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

## Configuration safety

SMTP passwords and other credentials are not returned by settings APIs or
forms. Do not include setup tokens, passwords, or API tokens in support
requests, screenshots, or logs. See [Security](security.md) for account and
token handling.
