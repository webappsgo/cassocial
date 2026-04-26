# Configuration

Cassocial configuration is stored in `server.yml` and can be edited via:
1. Admin web UI at `/admin/settings` (recommended)
2. Configuration file at `{configdir}/server.yml`
3. Environment variables (`CASSOCIAL_*`)
4. CLI flags

## Configuration File

Location: `/config/server.yml` (Docker) or `/etc/casapps/cassocial/server.yml` (system install)

```yaml
server:
  address: "0.0.0.0"
  port: 8080
  mode: production  # production or development
  debug: false

database:
  driver: sqlite  # sqlite, postgres, mysql
  name: /data/db/cassocial.db
  max_connections: 10
  max_idle_connections: 5

logging:
  level: info  # debug, info, warn, error
  format: text  # text or json

ssl:
  enabled: false
  letsencrypt: false
  domain: ""
  cert_file: ""
  key_file: ""

email:
  enabled: false
  host: smtp.example.com
  port: 587
  username: ""
  password: ""
  from: "noreply@example.com"
  tls: true

cassocial:
  site_name: "Cassocial"
  site_description: "Self-hosted link aggregator and social profile"
  allow_registration: true
  max_profiles_per_user: 5
  max_links_per_profile: 100
```

## Environment Variables

All settings can be overridden with environment variables:

```bash
# Server
CASSOCIAL_ADDRESS=0.0.0.0
CASSOCIAL_PORT=8080
CASSOCIAL_MODE=production
CASSOCIAL_DEBUG=false

# Database
CASSOCIAL_DB_DRIVER=sqlite
CASSOCIAL_DB_NAME=/data/db/cassocial.db
CASSOCIAL_DB_HOST=localhost
CASSOCIAL_DB_PORT=5432
CASSOCIAL_DB_USER=cassocial
CASSOCIAL_DB_PASSWORD=password

# Email
CASSOCIAL_EMAIL_HOST=smtp.example.com
CASSOCIAL_EMAIL_PORT=587
CASSOCIAL_EMAIL_USERNAME=user
CASSOCIAL_EMAIL_PASSWORD=pass
```

## CLI Flags

All settings can be overridden with CLI flags:

```bash
cassocial --port 8080
cassocial --mode production
cassocial --config /etc/cassocial
cassocial --data /var/lib/cassocial
cassocial --debug
```

## Database Configuration

### SQLite (Default)

```yaml
database:
  driver: sqlite
  name: /data/db/cassocial.db
```

### PostgreSQL

```yaml
database:
  driver: pgx
  host: localhost
  port: 5432
  name: cassocial
  user: cassocial
  password: secret
  ssl_mode: require
```

### MySQL/MariaDB

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  name: cassocial
  user: cassocial
  password: secret
```

## Email Configuration

Configure SMTP for email notifications:

```yaml
email:
  enabled: true
  host: smtp.gmail.com
  port: 587
  username: your-email@gmail.com
  password: your-app-password
  from: "Cassocial <noreply@example.com>"
  tls: true
```

Test SMTP configuration from admin panel at `/admin/smtp`.

## SSL/TLS Configuration

### Manual Certificates

```yaml
ssl:
  enabled: true
  cert_file: /config/ssl/local/cert.pem
  key_file: /config/ssl/local/key.pem
```

### Let's Encrypt

```yaml
ssl:
  enabled: true
  letsencrypt: true
  domain: cassocial.example.com
  email: admin@example.com
```

## Live Reload

Configuration changes apply immediately without restart for:
- Branding and SEO settings
- Email settings
- Rate limiting rules
- Theme changes
- Most feature toggles

Settings requiring restart (admin UI shows warning):
- `server.port`
- `server.address`
- `database.*` settings
- `ssl.*` settings (except domain)
