# Installation

Cassocial can be deployed using Docker, Docker Compose, or as a standalone binary.

## Docker (Recommended)

### Quick Start

```bash
docker run -d \
  --name cassocial \
  -p 64580:80 \
  -v ./rootfs/config:/config:z \
  -v ./rootfs/data:/data:z \
  ghcr.io/casapps/cassocial:latest
```

Access at `http://localhost:64580`

### With Docker Compose

1. Download docker-compose.yml:
```bash
curl -O https://raw.githubusercontent.com/casapps/cassocial/main/docker/docker-compose.yml
```

2. Start the service:
```bash
docker compose up -d
```

### First Run

On first run, Cassocial will:
1. Generate default configuration at `/config/server.yml`
2. Create SQLite databases at `/data/db/`
3. Display a setup token in the logs
4. Redirect you to the setup wizard at `/setup`

View logs to get the setup token:
```bash
docker logs cassocial
```

## Binary Installation

### Download

```bash
# Linux AMD64
curl -LO https://github.com/casapps/cassocial/releases/latest/download/cassocial-linux-amd64
chmod +x cassocial-linux-amd64
sudo mv cassocial-linux-amd64 /usr/local/bin/cassocial

# macOS ARM64 (Apple Silicon)
curl -LO https://github.com/casapps/cassocial/releases/latest/download/cassocial-darwin-arm64
chmod +x cassocial-darwin-arm64
sudo mv cassocial-darwin-arm64 /usr/local/bin/cassocial
```

### Run

```bash
# Start server
cassocial

# Start on custom port
cassocial --port 8080

# Run as system service
sudo cassocial --service --install
sudo cassocial --service start
```

## System Requirements

- **OS**: Linux, macOS, BSD, or Windows
- **Architecture**: AMD64 or ARM64
- **Memory**: 512MB minimum, 1GB recommended
- **Disk**: 1GB minimum for application and data
- **Database**: SQLite (included), PostgreSQL, or MySQL

## Post-Installation

After installation:

1. **Access setup wizard**: Navigate to `http://localhost:64580/setup`
2. **Configure basic settings**: Site name, description
3. **Create admin account**: Username, email, password
4. **Configure email** (optional): SMTP settings for notifications
5. **Complete setup**: Review and finish

The admin panel will be available at `/admin` after setup.

## Upgrading

### Docker

```bash
# Pull latest image
docker pull ghcr.io/casapps/cassocial:latest

# Restart container
docker compose down
docker compose up -d
```

### Binary

```bash
# Download new version
curl -LO https://github.com/casapps/cassocial/releases/latest/download/cassocial-linux-amd64

# Stop service
sudo systemctl stop cassocial

# Replace binary
sudo mv cassocial-linux-amd64 /usr/local/bin/cassocial
sudo chmod +x /usr/local/bin/cassocial

# Start service
sudo systemctl start cassocial
```

Backups are created automatically daily at 2:30am and stored in `/data/backup/`.
