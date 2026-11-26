# Configuration

Cassocial reads `{config_dir}/server.yml`. Command-line flags override
environment variables, which override file values and defaults. The default
API version is `v1`; the default administration path is `administration`.

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

## Main settings

The generated file includes additional scheduler, rate-limit, and product
settings. The following is the canonical SMTP and web-security structure:

```yaml
server:
  address: "0.0.0.0"
  port: 64500
  fqdn: "YOUR_SERVER_HOSTNAME"
  mode: production
  admin_path: administration
  api_version: v1
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
| `LOG_DIR` | Log directory |
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
`/server/{admin_path}/smtp`. Settings that affect listeners, database
connections, or TLS setup may require a restart.
