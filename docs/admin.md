# Server Administration

Server Admin accounts manage Cassocial itself. They are separate from Regular User accounts and use a separate session. The default admin path is `/server/administration`; operators can change it with `server.admin_path`.

The public interface does not link to or advertise the administration path. Enter the configured path directly. Unauthenticated requests go to the shared sign-in page. A signed-in Regular User is returned to their own account area and cannot use Server Admin routes.

## First Server Admin

On a new installation, the server prints a one-time setup token. Visit `/server/administration/setup` and use that token to create the primary Server Admin. The primary Server Admin cannot be deleted. Additional Server Admins are managed through the administration panel.

## Pages

Administration routes use `/server/{admin_path}/...`. Common sections include:

| Path | Purpose |
|------|---------|
| `/server/{admin_path}/` | Server dashboard |
| `/server/{admin_path}/settings` | Server configuration |
| `/server/{admin_path}/users` | Regular User management when PART 34 is enabled |
| `/server/{admin_path}/security` | Security and authentication settings |
| `/server/{admin_path}/email` | Email configuration |
| `/server/{admin_path}/backup` | Backup and restore |
| `/server/{admin_path}/logs` | Application and audit logs |
| `/server/{admin_path}/scheduler` | Built-in scheduled tasks |
| `/server/{admin_path}/updates` | Application updates |
| `/server/{admin_path}/network` | Network and overlay settings |

The admin's personal account is under `/server/{admin_path}/{admin_username}/profile`. Server-wide settings stay in their section; they do not belong under the personal account path.

## Web and API authentication

Browser administration uses the Server Admin session and standard HTML forms. Core operations must work with JavaScript disabled. The JSON API uses the versioned route family `/api/{api_version}/server/{admin_path}/...` and Server Admin bearer tokens beginning with `adm_`; a browser session cookie does not replace API token authentication.

Never put passwords, one-time setup tokens, or API tokens in documentation, URLs, screenshots, or logs. Review [Security](security.md) before changing authentication settings.

## Regular User separation

When PART 34 is active, Regular Users manage their own profiles and account settings in the user area. They do not become Server Admins by receiving a role in the `users` table. Server Admin exceptions to private profile visibility are available only through the authenticated administration panel, never through a public profile route.
