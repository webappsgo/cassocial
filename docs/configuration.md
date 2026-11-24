# Configuration

Cassocial reads `{config_dir}/server.yml`. Command-line overrides take precedence over environment variables, which take precedence over the file and defaults. Secret values belong in the restricted `secrets.yml` file where supported; do not put credentials in examples, logs, or issue reports.

## Paths

On Linux, system installs use `/etc/casapps/cassocial` for configuration, `/var/lib/casapps/cassocial` for data, and `/var/log/casapps/cassocial` for logs. User installs use `~/.config/casapps/cassocial`, `~/.local/share/casapps/cassocial`, and the `logs` directory under the data path. Docker uses `/config` and `/data` mounts. See [Installation](installation.md) for platform-specific locations.

## Main settings

```yaml
server:
  address: "0.0.0.0"
  mode: production
  admin_path: administration
  api_version: v1
  i18n:
    default_language: en
    available_languages: [en, es, fr, de, it, ar]
  branding:
    title: Cassocial
    tagline: ""
    description: ""

database:
  driver: sqlite
  name: "{data_dir}/db/cassocial.db"

logging:
  level: info
  format: text

ssl:
  enabled: false
  letsencrypt: false

email:
  enabled: false
  host: ""
  port: 587
  username: ""
  password: ""
  from: ""
  tls_mode: auto

web:
  ui:
    theme: dark
```

The generated file contains additional scheduler, rate-limit, and product settings. Keep existing unknown settings when editing; Cassocial reads and writes the complete configuration structure.

## Environment variables

The server recognizes these overrides:

| Variable | Setting |
|----------|---------|
| `CONFIG_DIR` | Configuration directory |
| `DATA_DIR` | Data directory |
| `LOG_DIR` | Log directory |
| `PORT` | `server.port` |
| `LISTEN` | `server.address` |
| `DOMAIN` | Detected server name / `server.fqdn` |
| `MODE` | `server.mode` (`debug` also enables debug output) |
| `DEBUG` | `server.debug` |
| `DATABASE_DRIVER` | `database.driver` |
| `DATABASE_URL` | `database.url` |
| `SMTP_HOST`, `SMTP_PORT` | SMTP host and port |
| `SMTP_USERNAME`, `SMTP_PASSWORD` | SMTP credentials |
| `SMTP_FROM_NAME`, `SMTP_FROM_EMAIL` | Sender name and address |
| `SMTP_TLS` | SMTP transport mode |
| `CASSOCIAL_PID` | PID file path |

For example, set `SMTP_HOST=YOUR_SMTP_HOST` and `SMTP_FROM_EMAIL=YOUR_EMAIL` in the service environment. Avoid placing secrets directly in shell history or committed files.

## Command-line overrides

The server accepts `--config`, `--data`, `--log`, `--address`, `--port`, `--mode`, `--pid`, and `--debug`. Run `cassocial --help` for the complete current option list. The client has its own `--server`, `--token`, and `--token-file` options; see the [CLI reference](cli.md).

## Database

SQLite is the default and stores its database under `{data_dir}/db/`. The current configuration also accepts PostgreSQL (`pgx`/`postgres`) and MySQL (`mysql`) drivers. Use the `database.url` connection string or the supported connection fields. Back up the database before changing database settings.

## SMTP and TLS

SMTP is optional. When configured, Cassocial can send account and administrative mail. TLS settings use `email.tls_mode`; valid values are `auto`, `starttls`, `tls`, and `none`. Keep SMTP credentials private.

TLS may use operator-provided certificate and key files or the automatic certificate service for a domain the operator controls. Certificate and private-key files belong in the configuration directory with restrictive permissions.

## Administration

Server Admins manage supported settings under `/server/administration`. The administration interface does not make every setting dynamically reloadable; restart the service after changes that affect listeners, database connections, or TLS setup.
