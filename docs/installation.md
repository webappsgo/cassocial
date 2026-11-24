# Installation

Cassocial is available as a Docker image or a standalone binary. Releases support Linux, macOS, Windows, and FreeBSD on amd64 and arm64.

## Docker

The repository Compose file stores persistent configuration and data in `rootfs/` beside the Compose file. Download it and start the service:

```bash
curl -q -LSsf -O https://raw.githubusercontent.com/webappsgo/cassocial/main/docker/docker-compose.yml
docker compose up -d
```

The Compose file binds the application to the Docker bridge address on port `64580`; place a reverse proxy on the same network to provide external access. The container listens on port 80 internally.

For a direct Docker run, preserve the same private bind and mount paths:

```bash
docker run -d \
  --name cassocial \
  -p 172.17.0.1:64580:80 \
  -v ./volumes/config:/config:z \
  -v ./volumes/data:/data:z \
  ghcr.io/webappsgo/cassocial:latest
```

## Binary

Download the release matching the operating system and architecture from [GitHub Releases](https://github.com/webappsgo/cassocial/releases). For Linux amd64:

```bash
curl -q -LSsf -O https://github.com/webappsgo/cassocial/releases/latest/download/cassocial-linux-amd64
chmod +x cassocial-linux-amd64
./cassocial-linux-amd64
```

Install a system service with `sudo cassocial --service --install`. Service commands include `start`, `stop`, `restart`, `reload`, `--uninstall`, `--disable`, and `--help`. Installation and service management run on the target host; development and test workflows must use the isolated environment required by AI.md.

## First run

On first start, Cassocial creates its default configuration and database, checks for SMTP, chooses an available port in the 64xxx range, creates a one-time setup token, prints startup details, and begins serving. The application remains usable before setup.

The console displays the setup token once and directs the operator to `/server/administration/setup`. Enter the token there to create the primary Server Admin. Keep the token private; it is not a substitute for the admin password.

For a Docker deployment, view the startup output with `docker logs cassocial`. Server Admin credentials are stored separately from Regular User accounts.

## Platform paths

Stable server paths use `internal_org: casapps` and `internal_name: cassocial`:

| Purpose | Linux / BSD | macOS | Windows |
|---------|-------------|-------|---------|
| Configuration | `/etc/casapps/cassocial` | `/Library/Application Support/casapps/cassocial` | `%PROGRAMDATA%\casapps\cassocial` |
| Data | `/var/lib/casapps/cassocial` | `/Library/Application Support/casapps/cassocial` | `%PROGRAMDATA%\casapps\cassocial` |
| Logs | `/var/log/casapps/cassocial` | `/Library/Logs/casapps/cassocial` | `%PROGRAMDATA%\casapps\cassocial\logs` |

Docker stores configuration under `/config` and application data under `/data`. See [Configuration](configuration.md) for directory overrides and supported settings.

## Upgrading

Back up the data directory before replacing a binary or updating the container image. Use the same persistent mounts after an image update. Review release notes for behavior or configuration changes, then restart the service.
