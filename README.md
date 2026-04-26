# Cassocial

[![Build Status](https://jenkins.casjay.cc/buildStatus/icon?job=casapps/cassocial)](https://jenkins.casjay.cc/job/casapps/job/cassocial/)
[![GitHub release](https://img.shields.io/github/v/release/casapps/cassocial)](https://github.com/casapps/cassocial/releases)
[![License](https://img.shields.io/github/license/casapps/cassocial)](LICENSE.md)

## About

Cassocial is a self-hosted link aggregator and social profile landing page platform. Create beautiful, customizable profile pages with links to all your social media, websites, and content in one place - like Linktree, but fully under your control.

## Official Site

https://cassocial.casapps.us

## Features

- **Profile Management** - Create unlimited profiles with custom slugs
- **Link Aggregation** - Add links to 5000+ supported services with auto-detected icons
- **Custom Domains** - Use your own domain for branded profile pages
- **Analytics** - Track page views, link clicks, and visitor analytics
- **QR Codes** - Generate QR codes for profiles with customizable styles
- **Themes** - Beautiful dark/light themes with custom CSS support
- **Multi-User** - Support for organizations and teams
- **Import/Export** - Import from Linktree, Linkstack, Carrd, and others
- **API Access** - Full REST, GraphQL, and OpenAPI support
- **Privacy-Focused** - Self-hosted, GDPR compliant, no tracking scripts
- **Shortlinks** - Create custom shortlinks with analytics

## Production

### Docker (Recommended)

```bash
docker run -d \
  --name cassocial \
  -p 64580:80 \
  -v ./rootfs/config:/config:z \
  -v ./rootfs/data:/data:z \
  ghcr.io/casapps/cassocial:latest
```

Access the admin panel at `http://localhost:64580/admin`

### Docker Compose

```bash
curl -O https://raw.githubusercontent.com/casapps/cassocial/main/docker/docker-compose.yml
docker compose up -d
```

### Binary

```bash
# Download latest release
curl -LO https://github.com/casapps/cassocial/releases/latest/download/cassocial-linux-amd64

# Make executable and run
chmod +x cassocial-linux-amd64
./cassocial-linux-amd64
```

The server will start and display a one-time setup token for admin panel access.

## Configuration

Configuration is auto-generated on first run. Edit via admin panel at `http://localhost:64580/admin`.

Key settings:

- **Server**
  - `server.port` - Listen port (default: 8080)
  - `server.address` - Listen address (default: 0.0.0.0)
  - `server.mode` - production or development

- **Database**
  - `database.driver` - sqlite (default), postgres, or mysql
  - `database.name` - Database name/path

- **Cassocial**
  - `cassocial.site_name` - Site name (default: Cassocial)
  - `cassocial.allow_registration` - Allow new user signups
  - `cassocial.max_profiles_per_user` - Max profiles per user (default: 5)
  - `cassocial.max_links_per_profile` - Max links per profile (default: 100)

All settings can be configured via:
1. Admin web UI at `/admin` (recommended)
2. Configuration file at `{configdir}/server.yml`
3. Environment variables (`CASSOCIAL_*`)
4. CLI flags (`--port`, `--address`, etc.)

## API

API documentation available at `/openapi` when running.

| Endpoint | Description |
|----------|-------------|
| `GET /healthz` | Health check |
| `GET /api/v1/profiles` | List profiles |
| `GET /api/v1/profiles/{slug}` | Get profile by slug |
| `GET /api/v1/links` | List links |
| `POST /api/v1/links` | Create link (requires auth) |
| `GET /api/v1/analytics` | Get analytics (requires auth) |
| `GET /api/v1/services` | List supported services |
| `GET /openapi` | Swagger UI |
| `GET /openapi.json` | OpenAPI specification |
| `GET /graphql` | GraphQL playground |

Full API documentation: `http://localhost:64580/openapi`

## Other

### Default Admin Access

On first run, a setup token is displayed in the console. Use this token to access `/admin` and create your admin account.

### Custom Domains

To use your own domain:

1. Point your domain DNS to your server
2. Configure domain in admin panel
3. Cassocial will automatically handle SSL via Let's Encrypt

### Troubleshooting

**Check logs:**
```bash
# Docker
docker logs cassocial

# Binary
tail -f /var/log/casapps/cassocial/cassocial.log
```

**Health check:**
```bash
curl http://localhost:64580/healthz
```

**Status:**
```bash
./cassocial-linux-amd64 --status
```

## Development

**Development instructions are for contributors only.**

### Prerequisites

- Docker (for containerized builds)
- Go 1.23+ (optional - Docker build works without Go installed)

### Build

```bash
# Clone
git clone https://github.com/casapps/cassocial
cd cassocial

# Quick dev build (outputs to OS temp dir)
make dev

# Full build (all platforms, outputs to binaries/)
make build

# Test
make test
```

### Project Structure

```
src/           # Source code
  config/      # Configuration package
  server/      # HTTP server, handlers, models, store
  service/     # Business services (email, etc.)
  swagger/     # OpenAPI/Swagger
  graphql/     # GraphQL API
  mode/        # Application mode detection
  paths/       # Path resolution
  ssl/         # SSL/TLS handling
  scheduler/   # Background tasks
  admin/       # Admin panel

docker/        # Docker files
tests/         # Test files
binaries/      # Built binaries (gitignored)
```

### Testing with Docker

```bash
# Build
make build

# Test in container
docker run --rm -v $(pwd)/binaries:/app alpine:latest /app/cassocial --help
```

## License

MIT - See [LICENSE.md](LICENSE.md)
