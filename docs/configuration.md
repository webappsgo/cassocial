# Configuration

Cassocial reads `{config_dir}/server.yml` and the ordered YAML files in
`{config_dir}/yaml.d/`. `server.yml` holds node-specific settings; shared
application settings live in their owning area file. Command-line flags
override environment variables, which override file values and defaults. The
default API version is `v1`; the default administration path is
`administration`.

## Paths

On Linux and BSD, system installs use `/etc/casapps/cassocial` for
configuration, `/var/lib/casapps/cassocial` for data, and
`/var/log/casapps/cassocial` for logs. User installs use
`~/.config/casapps/cassocial` for configuration and
`~/.local/share/casapps/cassocial` for data and logs. macOS and Windows use
their platform application-support and program-data directories. Docker uses
`/config` and `/data` mounts. See [Installation](installation.md) for the
platform path table.

Set `CONFIG_DIR`, `DATA_DIR`, or `LOG_DIR` to override the corresponding
directory. The server accepts `--config`, `--data`, and `--log` as command-line
overrides.

## Drop-in configuration

Keep node-specific values in `server.yml` and put other configuration areas in
`{config_dir}/yaml.d/`. The server reads `.yml` and `.yaml` files in
byte-wise filename order after `server.yml`; later files override earlier
values. Nested maps merge by key, while a scalar or list replaces the earlier
value. Environment variables and command-line flags retain higher precedence.

The account area is stored in `{config_dir}/yaml.d/accounts.yml`. It owns
`server.auth`, `server.users`, `server.session`, `server.admin`, and
`server.admin_path`. For example, an operator can enable Regular User accounts
and choose registration mode in this file:

```yaml
server:
  users:
    enabled: true
    registration:
      mode: open
      require_email_verification: true
  admin_path: administration
  admin:
    email: ""
```

The admin settings page can enable Regular User accounts and choose `open` or
`private` registration. It writes settings to their owning file. When another
configuration file overrides an account setting, the admin page shows that
setting as read-only and identifies the overriding file; edit the owning
configuration file instead.

For example, an operator can enable I2P in
`{config_dir}/yaml.d/50-i2p.yml`:

```yaml
server:
  i2p:
    enabled: true
```

CLI release distribution and minimum-version settings belong in
`{config_dir}/yaml.d/maintenance.yml`. The server reads CLI release files from
`{data_dir}/cli` by default; operators provision `version.txt`, `sha256.txt`,
and the platform-named binaries from a release. Downloads are public by
default. Set `require_auth: true` to require a bearer token.

```yaml
cli:
  binary_download:
    dir: "{data_dir}/cli"
    require_auth: false
server:
  update:
    cli_min_version: ""
```

The onion address exposure policy belongs in `{config_dir}/yaml.d/tor.yml`:

```yaml
server:
  tor:
    expose: false
    contact_email: ""
```

The `yaml.d` directory must be mode `0700`; `server.yml`, `secrets.yml`, and
each drop-in must be mode `0600`. The server corrects these modes at startup.
Hidden files, backup files, non-YAML files, and subdirectories are ignored.

## Main settings

The following shows a merged view of canonical SMTP, network-security, and web
settings; it is not a single-file template. Store each section in its owning
area file as described above. See the complete setting reference below for
every supported key.

```yaml
server:
  address: "0.0.0.0"
  port: 64500
  fqdn: "YOUR_SERVER_HOSTNAME"
  mode: production
  api_version: v1
  baseurl: /
  healthz:
    root:
      enabled: false
  trusted_proxies:
    additional: []
    client_ip_header: auto
  branding:
    title: Cassocial
    tagline: ""
    description: ""
  i18n:
    default_language: en
    available_languages: [en, es, fr, de, it, ar, ja, zh]
  security:
    reports:
      retention_days: 365
    allowlist: []
    blocklists:
      enabled: true
      auto_update: true
      extra_sources: []
      action: reject
      reject_message: ""
      log_blocked: true
  geoip:
    enabled: true
    deny_countries: []
    allow_countries: []
    presets: {}
  i2p:
    enabled: false
    expose: false
    binary: ""
    sam_address: "127.0.0.1:7656"
    virtual_port: 80
    inbound_length: 3
    outbound_length: 3
    inbound_quantity: 5
    outbound_quantity: 5
    signature_type: 7
    bootstrap_timeout: 300
  notifications:
    email:
      smtp:
        host: ""
        port: 587
        username: ""
        tls: auto
      from:
        name: Cassocial
        email: "no-reply@YOUR_SERVER_HOSTNAME"

database:
  driver: sqlite
  name: "{data_dir}/db/server.db"

web:
  ui:
    theme: dark
  csrf:
    enabled: true
    token_length: 32
    cookie_name: csrf_token
    header_name: X-CSRF-Token
    secure: auto
    exempt_paths: []
```

The port shown is an example value; the generated configuration selects an
available port in the 64000–64999 range. Use the configured port when
connecting directly, or use the container and reverse-proxy settings in
[Installation](installation.md).

## Regular User registration

Regular User accounts are disabled by default. Enable multi-user mode to
activate account and profile features. In that mode, `open` is the default
registration mode; `private` disables public self-registration while allowing
Server Admin invitations and direct account creation.

These settings belong in `{config_dir}/yaml.d/accounts.yml` (see the
[account area-file example](#drop-in-configuration)). The admin settings page
can enable accounts and select either mode unless a different file owns the
effective setting; externally overridden values are read-only in the page.

```yaml
server:
  users:
    enabled: true
    registration:
      mode: open
      require_email_verification: true
      allowed_domains: []
      blocked_domains: []
      invite_expiration_days: 7
```

When open registration is active, `allowed_domains` restricts email domains;
an empty list permits all domains. `blocked_domains` excludes specific
domains. User invite and activation links default to seven days; configure
`invite_expiration_days` to change that validity period.

Email verification is required only when
`server.users.registration.require_email_verification` is `true` and SMTP is
working. If the setting is `false` or SMTP is unavailable, email addresses are
auto-verified. Configure SMTP in `server.notifications.email`; see the
[SMTP guide](examples/smtp.md).

## Setting reference

The generated configuration contains the settings below. Nested names are
YAML paths. See the owner file for each area; omitted values use the generated
defaults and unknown keys are rejected.

| YAML path | Purpose and behavior |
|-----------|----------------------|
| `server.address`, `server.port` | Listen address and port; the generated port is selected from 64000–64999. |
| `server.fqdn`, `server.mode`, `server.debug` | Public hostname, `production` or `development` mode, and debug setting. `DOMAIN` and `MODE` can override the corresponding values. |
| `server.api_version` | API version segment; stored in node-local `server.yml`. |
| `server.admin_path` | Server Admin URL path segment; stored in `yaml.d/accounts.yml`. |
| `server.admin.email` | Private fallback recipient for Server Admin notices; stored in `yaml.d/accounts.yml`. |
| `server.users.enabled` | Enable Regular User accounts and multi-user features; defaults to `false`; stored in `yaml.d/accounts.yml`. |
| `server.users.registration.mode` | Public self-registration mode: `open` (default when multi-user mode is enabled) or `private`; stored in `yaml.d/accounts.yml`. |
| `server.users.registration.require_email_verification` | Require email verification only when enabled and SMTP is working; otherwise addresses are auto-verified. Stored in `yaml.d/accounts.yml`. |
| `server.users.registration.allowed_domains`, `.blocked_domains` | Domain allow-list and block-list for open registration; an empty allow-list permits every domain. Stored in `yaml.d/accounts.yml`. |
| `server.users.registration.invite_expiration_days` | Validity period for user invite and activation links; defaults to seven days. Stored in `yaml.d/accounts.yml`. |
| `server.baseurl` | URL path prefix for the installation, such as `/myproject`; defaults to `/`. Trusted reverse-proxy prefix headers take precedence per request. |
| `server.trusted_proxies.additional`, `server.trusted_proxies.client_ip_header` | Explicit trusted proxy addresses/CIDRs and the forwarding header policy (`auto` or a supported header name). |
| `server.healthz.root.enabled` | Mount `/healthz` as a direct alias of `/server/healthz`; disabled by default. |
| `server.branding.title`, `.tagline`, `.description`, `.content.about`, `.content.privacy`, `.content.terms` | Site branding and operator-authored information-page content. |
| `server.seo.keywords` | Keywords used by page metadata. |
| `server.contact.admin.email`, `.security.email`, `.abuse.email`, `.general.email`, `.dpo.name`, `.dpo.email` | Public operator contact roles. |
| `server.pages.contact.enabled`, `.captcha`, `.captcha_site_key`, `.success_message` | Public contact form availability and response behavior. Secret captcha credentials are not serialized in this file. |
| `server.pages.contact.antispam.honeypot`, `.min_fill_seconds`, `.max_age_seconds`, `.per_ip`, `.per_email`, `.global`, `.window`, `.max_links`, `.reject_markup`, `.blocked_words`, `.blocked_email_domains`, `.strikes`, `.ban_hours`, `.review_min_seconds`, `.review_max_seconds`, `.duplicate_window_hours` | Honeypot, submission timing, IP/email/global limits, window, link and markup limits, blocked words/domains, strike and ban thresholds, review timing, and duplicate-report window. |
| `server.pages.help.content` | Operator-authored Help page content. |
| `server.pages.help.mcp` | Project-specific MCP usage guidance shown on `/server/help/mcp`. Stored in `yaml.d/pages.yml`; editable by Server Admins under Pages → Help and applied to the running server after saving. |
| `server.privacy.data.sold`, `.stored_on_server`, `.sharing` | Public privacy disclosures about storage and conditional sharing. Each sharing entry has `condition`, `when`, and `data`. |
| `server.privacy.retention.period`, `.export_available`, `.deletion_available` | Public retention, export, and deletion disclosures. |
| `server.privacy.cookies.*` | Named cookie categories with `enabled` and `description` values. |
| `server.privacy.third_party.services` | Third-party service disclosures; each item has `name`, `purpose`, and `policy_url`. |
| `server.compliance.gdpr`, `.ccpa`, `.hipaa`, `.soc2`, `.pci_dss`, `.iso27001`, `.fedramp`, `.lgpd`, `.pipeda`, `.appi`, `.pdpa` | Compliance standards stored in `yaml.d/privacy.yml`; all default to `false`. Enabling any standard requires encrypted backups. |
| `server.tracking.type`, `.id`, `.url` | Consent-gated analytics provider settings stored in `yaml.d/privacy.yml`. Supported providers are `google`, `matomo`, `piwik`, `owa`, `fathom`, `plausible`, `umami`, `simple`, and `cloudflare`; empty or `none` disables tracking. Matomo, Piwik, OWA, and Umami require a URL; Fathom and Plausible allow an optional URL. |
| `server.security.reports.retention_days` | Retention period for security reports. |
| `server.security.encryption_key`, `.encryption_key_version` | Managed encryption key metadata. The application provisions and rotates this key; operators should not set or copy key material manually. |
| `server.security.allowlist` | Trusted IP/CIDR entries, each with `cidr` and a required `description`. |
| `server.security.blocklists.enabled`, `.auto_update`, `.action`, `.reject_message`, `.log_blocked` | Blocklist enforcement policy. `action` is `reject` or `drop`; the built-in FireHOL Level 1 source remains available. |
| `server.security.blocklists.extra_sources` | Additional HTTPS sources with `id`, `type` (`ip` or `domain`), `url`, and optional `enabled`. |
| `server.geoip.enabled`, `.deny_countries`, `.allow_countries`, `.presets` | Local GeoIP lookups and ISO 3166-1 alpha-2 country filtering. An allow-list takes precedence over a deny-list. |
| `server.i2p.enabled`, `.expose`, `.binary`, `.sam_address`, `.virtual_port`, `.inbound_length`, `.outbound_length`, `.inbound_quantity`, `.outbound_quantity`, `.signature_type`, `.bootstrap_timeout` | I2P provider, SAM bridge, tunnel, and bootstrap settings. The generated container default enables I2P; bare-metal defaults disable it. |
| `server.tor.expose`, `.contact_email` | Opt in to publishing the running onion address through autodiscovery; Tor contact email is shown only on Tor responses. Both default to empty/false. |
| `server.update.cli_min_version`, `.agent_min_version` | Minimum client versions included in autodiscovery responses when configured. |
| `cli.binary_download.dir`, `.require_auth` | Operator-provisioned CLI release directory and whether downloads require a bearer token; default directory is `{data_dir}/cli`, downloads are public by default. |
| `server.backup.retention.max_backups`, `.keep_weekly`, `.keep_monthly`, `.keep_yearly`, `.max_total_size` | Full-backup retention and storage cap. The size cap accepts bytes or a percentage of the backup volume. |
| `server.backup.encryption.enabled` | Backup encryption state. The password is held in the protected secret store and is not a YAML value. |
| `server.restore.max_archive_size`, `.max_expanded_size`, `.max_files`, `.max_file_size`, `.max_ratio` | Limits checked before and during backup archive extraction. |
| `scheduler.tasks.backup_daily.enabled`, `.schedule` | Daily full backup and differential task state and schedule. Retention remains under `server.backup.retention`. |
| `scheduler.tasks.backup_hourly.enabled`, `.schedule` | Optional cumulative hourly differential task; disabled by default. |
| `server.notifications.email.smtp.host`, `.port`, `.username`, `.password`, `.tls`, `.from.name`, `.from.email` | Canonical SMTP relay and sender configuration. A legacy YAML password is migrated to `{config_dir}/secrets.yml`; the admin UI never returns it. |
| `server.notifications.email.events.{event}` | Server default email preference for each named notification event. User preferences can override these defaults for events with an email template; required security events remain enabled. |
| `server.admin.email` | Default Server Admin contact address. Account credentials are stored in the database, not here. |
| `server.i18n.default_language`, `.available_languages` | Default locale and enabled translation locales. |
| `server.user`, `server.group` | Operating-system user and group used by service operations. |
| `server.pidfile`, `server.daemonize` | Enable PID-file creation and configure daemon startup behavior. |
| `database.driver`, `.url`, `.host`, `.port`, `.name`, `.user`, `.password`, `.ssl_mode`, `.max_connections`, `.max_idle_connections` | Database driver and connection pool. A configured `url` takes precedence over individual connection fields. |
| `logging.level`, `logging.format` | Log threshold (`debug`, `info`, `warn`, `error`) and format (`text`, `json`). |
| `ssl.enabled`, `.cert_file`, `.key_file`, `.letsencrypt`, `.domain` | HTTPS certificate settings. Certificate files and private keys require restrictive permissions. |
| `email.enabled`, `.host`, `.port`, `.username`, `.password`, `.from_name`, `.from`, `.tls`, `.tls_mode` | Legacy SMTP settings accepted during migration to `server.notifications.email`; use the canonical `server.notifications` keys for new configuration. Passwords are moved into the protected secret file. |
| `scheduler.enabled`, `.tasks` | Built-in scheduled jobs. Each named task accepts `enabled`, `schedule`, and applicable `retry_on_fail`, `retry_delay`, `retention`, or `renew_before` values. |
| `rate_limit.global_burst`, `.read`, `.write`, `.health`, `.strict`, `.auth_failures`, `.auth.login`, `.auth.password_reset`, `.auth.registration` | Always-on request limits. Each rule contains `requests` and `window` in seconds; anonymous traffic is counted per IP, valid bearer tokens per token, strict bearer actions per owner, and failed authentication per IP. |
| `web.ui.theme`, `web.cors` | Default UI theme (`dark`, `light`, or `auto`) and CORS policy. |
| `web.csrf.enabled`, `.token_length`, `.cookie_name`, `.header_name`, `.secure`, `.exempt_paths` | CSRF protection for cookie-authenticated frontend forms. API paths cannot be exempted through this setting. |

## Analytics tracking and consent

Analytics settings belong in `{config_dir}/yaml.d/privacy.yml`, alongside the
privacy disclosures. The provider is disabled by default. Example for a
self-hosted Matomo instance:

```yaml
server:
  tracking:
    type: matomo
    id: "1"
    url: "https://analytics.example.com"
```

Optional analytics scripts are included only after a visitor consents to
analytics cookies and while `server.privacy.cookies.analytics.enabled` is
true. Disabling that category hides its consent toggle and suppresses provider
scripts, the tracking endpoint, and provider hosts in the Content Security
Policy. The same-origin
`/server/tracking.js` endpoint returns `404` without consent and on Tor or I2P
requests, so analytics requests do not leave an overlay network.

Provider identifiers and URLs are validated when configuration loads. The
server supports Google Analytics, Matomo/Piwik, Open Web Analytics, Fathom,
Plausible, Umami, Simple Analytics, and Cloudflare Web Analytics. Self-hosted
providers use `url`; cloud providers use their standard endpoint when `url` is
empty.

### CORS and trusted proxy origins

An empty `web.cors` disables cross-origin access. The generated `*` value is
treated as unset: Cassocial combines explicit configured origins, HTTPS origins
from `DOMAIN`, and HTTPS hosts observed in `X-Forwarded-Host` from trusted
proxies. If those sources provide no origins, the policy falls back to `*` and
does not allow credentials. A resolved origin list allows credentials and is
also used for the CSP `connect-src` directive.

CORS preflights return `204` and allow `GET`, `POST`, `PUT`, `PATCH`, `DELETE`,
and `OPTIONS`. Allowed request headers are `Content-Type`, `Accept`,
`X-Requested-With`, `Authorization`, `X-CSRF-Token`, `X-XSRF-Token`,
`MCP-Protocol-Version`, `Mcp-Method`, and `Mcp-Name`; preflight results are
cached for 24 hours.

Learned hosts are lowercased RFC 1123 domain names; IP literals and invalid
hosts are ignored. A host is refreshed at most once per minute, expires after
30 days without observation, and the table retains at most 1,000 hosts.

The built-in scheduler currently provides `geoip_update`, `blocklist_update`,
`cve_update`, `log_rotation`, `session_cleanup`, `backup`, `ssl_renewal`,
`health_check`, and `tor_health`. Each task can be enabled, disabled, or
rescheduled in `scheduler.tasks`.

SMTP credentials are not written into `server.yml`. The SMTP password is
stored in `{config_dir}/secrets.yml` with restricted permissions. The web
administration page redacts the password and indicates whether one is
configured. See [SMTP](examples/smtp.md) for relay behavior.

GeoIP downloads checksum-verified ASN, country, and city MMDB files from
ip-location-db on first run and refreshes them weekly. Country codes are
ISO 3166-1 alpha-2 values. `allow_countries` takes precedence over
`deny_countries`; leaving both empty allows all countries. Lookup failures,
missing datasets, and stale datasets fail open, while authentication and
authorization remain required. Private and internal addresses are not looked
up. Operators can manage country lists and presets at
`/server/{admin_path}/network/geoip`.

`server.security.allowlist` accepts IP addresses or CIDRs with a required
description. Administrators can add and remove database-backed entries at
`/server/{admin_path}/security/allowlist`; configuration entries remain
managed in `server.yml`. Allowlisted clients bypass blocklists, rate limits,
and GeoIP country filtering, while authentication remains required.

Blocklists always include the built-in `firehol_level1` IP source. Add
operator-managed HTTPS sources under `server.security.blocklists.extra_sources`
using an `id`, a `type` (`ip` or `domain`), a URL, and optional `enabled`.
Sources refresh daily; the last valid copy remains active when an update fails.
Manage source and enforcement settings at
`/server/{admin_path}/network/blocklists`. The default rejection response is
403; allowlisted requests bypass blocklist enforcement but still require
authentication.

The canonical frontend health route is `/server/healthz`. Set
`server.healthz.root.enabled: true` to mount `/healthz` as a direct alias;
the default is false. API health JSON is available at
`/api/{api_version}/server/healthz`, with dotted text at its `.txt` alias.

The I2P eepsite is disabled on bare metal and in VMs. Application containers
enable it by default; `I2P_ENABLED` or `--i2p=false` can opt out. When enabled,
the server prefers its own managed `i2pd` process and otherwise tries the
configured SAM bridge. The destination key is stored under `{data_dir}/i2p/site`
with mode `0600`; the generated `.b32.i2p` address remains stable across
restarts. Manage the provider at `/server/{admin_path}/network/i2p`.

## Environment variables

These server environment variables override configuration values:

| Variable | Setting |
|----------|---------|
| `CONFIG_DIR` | Configuration directory |
| `DATA_DIR` | Data directory |
| `CACHE_DIR` | Cache directory |
| `LOG_DIR` | Log directory |
| `BACKUP_DIR` | Backup directory |
| `PORT` | `server.port` |
| `LISTEN` | `server.address` |
| `DOMAIN` | `server.fqdn` |
| `MODE` | `server.mode`; `debug` selects development mode and enables debug output unless overridden |
| `DEBUG` | `server.debug` |
| `I2P_ENABLED` | `server.i2p.enabled` |
| `DATABASE_DRIVER` | `database.driver` |
| `DATABASE_URL` | `database.url` |
| `SMTP_HOST`, `SMTP_PORT` | `server.notifications.email.smtp.host` and `.port` |
| `SMTP_USERNAME`, `SMTP_PASSWORD` | SMTP credentials |
| `SMTP_FROM_NAME`, `SMTP_FROM_EMAIL` | Sender identity |
| `SMTP_TLS` | SMTP transport mode: `auto`, `starttls`, `tls`, or `none` |
| `CASSOCIAL_PID` | PID file path |

Use `SMTP_PASSWORD` through the service's secret-management facility. Do not
put a real credential in shell history or a committed environment file.

## Database

SQLite is the default and stores the database under `{data_dir}/db/`. The
configuration also accepts the `pgx`/`postgres` and `mysql` drivers. Use
`database.url` or the corresponding connection fields for a remote database.
Back up the data directory before changing database settings.

## Web security

`web.csrf` controls protection for cookie-authenticated frontend forms. The
default is enabled with a 32-byte token, a `csrf_token` cookie, the
`X-CSRF-Token` header, and automatic `Secure` cookie selection based on the
request scheme. `exempt_paths` applies to frontend paths. API routes are
outside the frontend CSRF scope.

## TLS and administration

TLS may use operator-provided certificate and key files or the automatic
certificate service for a domain the operator controls. Certificate and
private-key files belong in the configuration directory with restrictive
permissions.

Server Admins can change the settings exposed by the current interface at
`/server/{admin_path}/settings` and SMTP at
`/server/{admin_path}/email`. Settings that affect listeners, database
connections, or TLS setup may require a restart.
