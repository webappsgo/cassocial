# Development Guide

This guide is for contributors developing Cassocial.

## Prerequisites

- Docker (for containerized builds)
- Go 1.23+ (optional - Docker build works without Go installed)
- Git

## Setup

```bash
# Clone repository
git clone https://github.com/casapps/cassocial
cd cassocial

# Quick dev build (outputs to OS temp dir)
make dev

# Full build (all 8 platforms)
make build

# Run tests
make test
```

## Project Structure

```
src/
├── main.go                 # Entry point
├── config/                 # Configuration
├── mode/                   # Mode detection
├── paths/                  # Path resolution
├── signal/                 # Signal handling
├── ssl/                    # SSL/TLS handling
├── swagger/                # OpenAPI/Swagger
├── graphql/                # GraphQL API
├── admin/                  # Admin panel
├── scheduler/              # Background tasks
├── server/
│   ├── handler/            # HTTP handlers
│   ├── service/            # Business logic
│   ├── model/              # Data models
│   ├── store/              # Database layer
│   ├── static/             # Static assets
│   └── template/           # HTML templates
└── service/                # Email services
```

## Building

### Quick Dev Build

```bash
make dev
```

Outputs to random temp directory for isolation.

### Full Build (All Platforms)

```bash
make build
```

Builds for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)
- FreeBSD (amd64, arm64)

Output: `binaries/cassocial-{os}-{arch}`

### Docker Build

```bash
make docker
```

Builds multi-arch image and pushes to ghcr.io.

## Testing

### Unit Tests

```bash
make test
```

Runs all tests via Docker.

### Manual Testing

```bash
# Build
make dev

# The output shows the binary location, e.g.:
# Built: /tmp/casapps.XXXXXX/cassocial

# Test in container
docker run --rm -v /tmp/casapps.XXXXXX:/app alpine:latest /app/cassocial --help
```

### Docker Testing

```bash
# Use temp directory workflow
TEMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/casapps.XXXXXX")
mkdir -p "$TEMP_DIR/rootfs/config" "$TEMP_DIR/rootfs/data"
cp docker/docker-compose.test.yml "$TEMP_DIR/docker-compose.yml"
cd "$TEMP_DIR" && docker compose up --abort-on-container-exit
rm -rf "$TEMP_DIR"
```

## Adding Features

### 1. Create Handler

```go
// src/server/handler/myfeature.go
package handler

func (h *MyHandler) HandleMyFeature(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### 2. Register Route

```go
// src/server/handler/router.go — add to SetupRoutes()
rt.mux.Handle("GET /api/myfeature", rt.middleware.RequireAuth(http.HandlerFunc(rt.myHandlers.HandleMyFeature)))
```

### 3. Add to OpenAPI Spec

```go
// src/swagger/swagger.go
// Add path to generatePaths()
```

### 4. Add to GraphQL Schema

```go
// src/graphql/graphql.go
// Add field to schema
```

### 5. Update Documentation

- Update `docs/api.md`
- Update `README.md` if user-facing
- Update AI.md PART 36 if significant

## Code Style

- Follow TEMPLATE.md NON-NEGOTIABLE rules
- Comments ABOVE code, never inline
- Descriptive variable names
- Handle all errors
- Validate all input
- Use Argon2id for passwords
- Use SHA-256 for tokens

## Database Migrations

Add migrations to `src/server/store/migrations/`:

```sql
-- 003_add_my_feature.sql
CREATE TABLE IF NOT EXISTS my_table (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Migrations run automatically on startup.

## Debugging

### Enable Debug Mode

```bash
cassocial --debug
```

or

```bash
export CASSOCIAL_DEBUG=true
docker compose up
```

Debug mode enables:
- Verbose logging
- Request/response logging
- Stack traces in errors

### View Logs

```bash
# Docker
docker logs cassocial

# Binary
tail -f /var/log/casapps/cassocial/server.log
```

## CI/CD

Workflows are in `.github/workflows/` and `.gitea/workflows/`:

- `release.yml` - Stable releases (on version tag)
- `beta.yml` - Beta releases (on beta branch push)
- `daily.yml` - Daily builds (3am UTC + main branch push)
- `docker.yml` - Docker images (on every push)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes following code style
4. Test thoroughly
5. Update documentation
6. Submit pull request

See AI.md for complete specification.
